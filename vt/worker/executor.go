package worker

import (
	"fmt"
	"time"

	"github.com/youtube/vitess/go/vt/discovery"
	"github.com/youtube/vitess/go/vt/throttler"
	"github.com/youtube/vitess/go/vt/topo/topoproto"
	"github.com/youtube/vitess/go/vt/wrangler"
	"golang.org/x/net/context"

	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
)

// executor takes care of the write-side of the copy.
// There is one executor for each destination shard and writer thread.
// To-be-written data will be passed in through a channel.
// The main purpose of this struct is to aggregate the objects which won't
// change during the execution and remove them from method signatures.
type executor struct {
	wr          *wrangler.Wrangler
	healthCheck discovery.HealthCheck
	throttler   *throttler.Throttler
	keyspace    string
	shard       string
	threadID    int
	// statsKey is the cached metric key which we need when we increment the stats
	// variable when we get throttled.
	statsKey []string
}

func newExecutor(wr *wrangler.Wrangler, healthCheck discovery.HealthCheck, throttler *throttler.Throttler, keyspace, shard string, threadID int) *executor {
	return &executor{
		wr:          wr,
		healthCheck: healthCheck,
		throttler:   throttler,
		keyspace:    keyspace,
		shard:       shard,
		threadID:    threadID,
		statsKey:    []string{keyspace, shard, fmt.Sprint(threadID)},
	}
}

// fetchLoop loops over the provided insertChannel and sends the commands to the
// current master.
func (e *executor) fetchLoop(ctx context.Context, dbName string, insertChannel chan string) error {
	for {
		select {
		case cmd, ok := <-insertChannel:
			if !ok {
				// no more to read, we're done
				return nil
			}
			cmd = "INSERT INTO `" + dbName + "`." + cmd
			if err := e.fetchWithRetries(ctx, cmd); err != nil {
				return fmt.Errorf("ExecuteFetch failed: %v", err)
			}
		case <-ctx.Done():
			// Doesn't really matter if this select gets starved, because the other case
			// will also return an error due to executeFetch's context being closed. This case
			// does prevent us from blocking indefinitely on insertChannel when the worker is canceled.
			return nil
		}
	}
}

// fetchWithRetries will attempt to run ExecuteFetch for a single command, with
// a reasonably small timeout.
// If will keep retrying the ExecuteFetch (for a finite but longer duration) if
// it fails due to a timeout or a retriable application error.
//
// executeFetchWithRetries will always get the current MASTER tablet from the
// healthcheck instance. If no MASTER is available, it will keep retrying.
func (e *executor) fetchWithRetries(ctx context.Context, command string) error {
	retryDuration := 2 * time.Hour
	// We should keep retrying up until the retryCtx runs out.
	retryCtx, retryCancel := context.WithTimeout(ctx, retryDuration)
	defer retryCancel()
	// Is this current attempt a retry of a previous attempt?
	isRetry := false
	for {
		var master *discovery.EndPointStats
		var err error

		// Get the current master from the HealthCheck.
		masters := discovery.GetCurrentMaster(
			e.healthCheck.GetEndPointStatsFromTarget(e.keyspace, e.shard, topodatapb.TabletType_MASTER))
		if len(masters) == 0 {
			e.wr.Logger().Warningf("ExecuteFetch failed for keyspace/shard %v/%v because no MASTER is available; will retry until there is MASTER again", e.keyspace, e.shard)
			statsRetryCount.Add(1)
			statsRetryCounters.Add(retryCategoryNoMasterAvailable, 1)
			goto retry
		}
		master = masters[0]

		// Block if we are throttled.
		if e.throttler != nil {
			for {
				backoff := e.throttler.Throttle(e.threadID)
				if backoff == throttler.NotThrottled {
					break
				}
				statsThrottledCounters.Add(e.statsKey, 1)
				time.Sleep(backoff)
			}
		}

		// Run the command (in a block since goto above does not allow to introduce
		// new variables until the label is reached.)
		{
			tryCtx, cancel := context.WithTimeout(retryCtx, 2*time.Minute)
			_, err = e.wr.TabletManagerClient().ExecuteFetchAsApp(tryCtx, endPointToTabletInfo(master), command, 0)
			cancel()

			if err == nil {
				// success!
				return nil
			}

			succeeded, finalErr := e.checkError(err, isRetry, master)
			if succeeded {
				// We can ignore the error and don't have to retry.
				return nil
			}
			if finalErr != nil {
				// Non-retryable error.
				return finalErr
			}
		}

	retry:
		masterAlias := "no-master-was-available"
		if master != nil {
			masterAlias = topoproto.TabletAliasString(master.Alias())
		}
		tabletString := fmt.Sprintf("%v (%v/%v)", masterAlias, e.keyspace, e.shard)

		select {
		case <-retryCtx.Done():
			if retryCtx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("failed to connect to destination tablet %v after retrying for %v", tabletString, retryDuration)
			}
			return fmt.Errorf("interrupted while trying to run %v on tablet %v", command, tabletString)
		case <-time.After(*executeFetchRetryTime):
			// Retry 30s after the failure using the current master seen by the HealthCheck.
		}
		isRetry = true
	}
}

// checkError returns true if the error can be ignored and the command
// succeeded, false if the error is retryable and a non-nil error if the
// command must not be retried.
func (e *executor) checkError(err error, isRetry bool, master *discovery.EndPointStats) (bool, error) {
	tabletString := fmt.Sprintf("%v (%v/%v)", topoproto.TabletAliasString(master.Alias()), e.keyspace, e.shard)
	// If the ExecuteFetch call failed because of an application error, we will try to figure out why.
	// We need to extract the MySQL error number, and will attempt to retry if we think the error is recoverable.
	match := errExtract.FindStringSubmatch(err.Error())
	var errNo string
	if len(match) == 2 {
		errNo = match[1]
	}
	switch {
	case e.wr.TabletManagerClient().IsTimeoutError(err):
		e.wr.Logger().Warningf("ExecuteFetch failed on %v; will retry because it was a timeout error: %v", tabletString, err)
		statsRetryCount.Add(1)
		statsRetryCounters.Add(retryCategoryTimeoutError, 1)
	case errNo == "1290":
		e.wr.Logger().Warningf("ExecuteFetch failed on %v; will reresolve and retry because it's due to a MySQL read-only error: %v", tabletString, err)
		statsRetryCount.Add(1)
		statsRetryCounters.Add(retryCategoryReadOnly, 1)
	case errNo == "2002" || errNo == "2006":
		e.wr.Logger().Warningf("ExecuteFetch failed on %v; will reresolve and retry because it's due to a MySQL connection error: %v", tabletString, err)
		statsRetryCount.Add(1)
		statsRetryCounters.Add(retryCategoryConnectionError, 1)
	case errNo == "1062":
		if !isRetry {
			return false, fmt.Errorf("ExecuteFetch failed on %v on the first attempt; not retrying as this is not a recoverable error: %v", tabletString, err)
		}
		e.wr.Logger().Infof("ExecuteFetch failed on %v with a duplicate entry error; marking this as a success, because of the likelihood that this query has already succeeded before being retried: %v", tabletString, err)
		return true, nil
	default:
		// Unknown error.
		return false, err
	}
	return false, nil
}

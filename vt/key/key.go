// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package key

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// MinKey is smaller than all KeyspaceId (the value really is).
var MinKey = KeyspaceId("")

// MaxKey is bigger than all KeyspaceId (by convention).
var MaxKey = KeyspaceId("")

// KeyspaceId is the type we base sharding on.
type KeyspaceId string

// Hex prints a KeyspaceId in capital hex.
func (kid KeyspaceId) Hex() HexKeyspaceId {
	return HexKeyspaceId(strings.ToUpper(hex.EncodeToString([]byte(kid))))
}

// MarshalJSON turns a KeyspaceId into json (using hex encoding).
func (kid KeyspaceId) MarshalJSON() ([]byte, error) {
	return []byte("\"" + string(kid.Hex()) + "\""), nil
}

// UnmarshalJSON reads a KeyspaceId from json (hex decoding).
func (kid *KeyspaceId) UnmarshalJSON(data []byte) (err error) {
	*kid, err = HexKeyspaceId(data[1 : len(data)-1]).Unhex()
	return err
}

// Uint64Key is a uint64 that can be converted into a KeyspaceId.
type Uint64Key uint64

func (i Uint64Key) String() string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint64(i))
	return buf.String()
}

// KeyspaceId returns the KeyspaceId associated with a Uint64Key.
func (i Uint64Key) KeyspaceId() KeyspaceId {
	return KeyspaceId(i.String())
}

// HexKeyspaceId is the hex represention of a KeyspaceId.
type HexKeyspaceId string

// Unhex converts a HexKeyspaceId into a KeyspaceId (hex decoding).
func (hkid HexKeyspaceId) Unhex() (KeyspaceId, error) {
	b, err := hex.DecodeString(string(hkid))
	if err != nil {
		return KeyspaceId(""), err
	}
	return KeyspaceId(string(b)), nil
}

// KeyRange is an interval of KeyspaceId values. It contains Start,
// but excludes End. In other words, it is: [Start, End[
type KeyRange struct {
	Start KeyspaceId
	End   KeyspaceId
}

func (kr KeyRange) MapKey() string {
	return string(kr.Start) + "-" + string(kr.End)
}

func (kr KeyRange) Contains(i KeyspaceId) bool {
	return kr.Start <= i && (kr.End == MaxKey || i < kr.End)
}

func (kr KeyRange) String() string {
	return fmt.Sprintf("{Start: %v, End: %v}", string(kr.Start.Hex()), string(kr.End.Hex()))
}

// Returns true if the KeyRange does not cover the entire space.
func (kr KeyRange) IsPartial() bool {
	return !(kr.Start == MinKey && kr.End == MaxKey)
}

// KeyRangesIntersect returns true if some Keyspace values exist in both ranges.
//
// See: http://stackoverflow.com/questions/4879315/what-is-a-tidy-algorithm-to-find-overlapping-intervals
// two segments defined as (a,b) and (c,d) (with a<b and c<d):
// intersects = (b > c) && (a < d)
// overlap = min(b, d) - max(c, a)
func KeyRangesIntersect(first, second KeyRange) bool {
	return (first.End == MaxKey || second.Start < first.End) &&
		(second.End == MaxKey || first.Start < second.End)
}

// KeyRangesOverlap returns the overlap between two KeyRanges.
// They need to overlap, otherwise an error is returned.
func KeyRangesOverlap(first, second KeyRange) (KeyRange, error) {
	if !KeyRangesIntersect(first, second) {
		return KeyRange{}, fmt.Errorf("Keyranges %v and %v don't overlap", first, second)
	}
	result := first
	if second.Start > first.Start {
		result.Start = second.Start
	}
	if second.End != MaxKey && second.End < first.End {
		result.End = second.End
	}
	return result, nil
}

// KeyspaceIdArray is an array of KeyspaceId that can be sorted
type KeyspaceIdArray []KeyspaceId

func (p KeyspaceIdArray) Len() int { return len(p) }

func (p KeyspaceIdArray) Less(i, j int) bool {
	return p[i] < p[j]
}

func (p KeyspaceIdArray) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p KeyspaceIdArray) Sort() { sort.Sort(p) }

// KeyRangeArray is an array of KeyRange that can be sorted
type KeyRangeArray []KeyRange

func (p KeyRangeArray) Len() int { return len(p) }

func (p KeyRangeArray) Less(i, j int) bool {
	return p[i].Start < p[j].Start
}

func (p KeyRangeArray) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p KeyRangeArray) Sort() { sort.Sort(p) }

// ParseShardingSpec parses a string that describes a sharding
// specification. a-b-c-d will be parsed as a-b, b-c, c-d. The empty
// string may serve both as the start and end of the keyspace: -a-b-
// will be parsed as start-a, a-b, b-end.
func ParseShardingSpec(spec string) (KeyRangeArray, error) {
	parts := strings.Split(spec, "-")
	if len(parts) == 1 {
		return nil, fmt.Errorf("malformed spec: doesn't define a range: %q", spec)
	}
	old := parts[0]
	ranges := make([]KeyRange, len(parts)-1)

	for i, p := range parts[1:] {
		if p == "" && i != (len(parts)-2) {
			return nil, fmt.Errorf("malformed spec: MinKey/MaxKey cannot be in the middle of the spec: %q", spec)
		}
		if p != "" && p <= old {
			return nil, fmt.Errorf("malformed spec: shard limits should be in order: %q", spec)
		}
		s, err := HexKeyspaceId(old).Unhex()
		if err != nil {
			return nil, err
		}
		e, err := HexKeyspaceId(p).Unhex()
		if err != nil {
			return nil, err
		}
		ranges[i] = KeyRange{Start: s, End: e}
		old = p
	}
	return ranges, nil
}

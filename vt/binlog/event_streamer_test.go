// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binlog

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/youtube/vitess/go/vt/binlog/proto"

	pb "github.com/youtube/vitess/go/vt/proto/binlogdata"
)

var dmlErrorCases = []string{
	"query",
	"query /* _stream 10 (eid id name ) (null 1 'bmFtZQ==' ); */",
	"query /* _stream _table_ eid id name ) (null 1 'bmFtZQ==' ); */",
	"query /* _stream _table_ (10 id name ) (null 1 'bmFtZQ==' ); */",
	"query /* _stream _table_ (eid id name  (null 1 'bmFtZQ==' ); */",
	"query /* _stream _table_ (eid id name)  (null 'aaa' 'bmFtZQ==' ); */",
	"query /* _stream _table_ (eid id name)  (null 'bmFtZQ==' ); */",
	"query /* _stream _table_ (eid id name)  (null 1.1 'bmFtZQ==' ); */",
	"query /* _stream _table_ (eid id name)  (null a 'bmFtZQ==' ); */",
}

func TestEventErrors(t *testing.T) {
	var got *proto.StreamEvent
	evs := &EventStreamer{
		sendEvent: func(event *proto.StreamEvent) error {
			if event.Category != "POS" {
				got = event
			}
			return nil
		},
	}
	for _, sql := range dmlErrorCases {
		trans := &pb.BinlogTransaction{
			Statements: []*pb.BinlogTransaction_Statement{
				{
					Category: pb.BinlogTransaction_Statement_BL_DML,
					Sql:      sql,
				},
			},
		}
		err := evs.transactionToEvent(trans)
		if err != nil {
			t.Errorf("%s: %v", sql, err)
			continue
		}
		want := &proto.StreamEvent{
			Category: "ERR",
			Sql:      sql,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("error for SQL: '%v' got: %+v, want: %+v", sql, got, want)
		}
	}
}

func TestSetErrors(t *testing.T) {
	evs := &EventStreamer{
		sendEvent: func(event *proto.StreamEvent) error {
			return nil
		},
	}
	trans := &pb.BinlogTransaction{
		Statements: []*pb.BinlogTransaction_Statement{
			{
				Category: pb.BinlogTransaction_Statement_BL_SET,
				Sql:      "SET INSERT_ID=abcd",
			},
		},
	}
	before := binlogStreamerErrors.Counts()["EventStreamer"]
	err := evs.transactionToEvent(trans)
	if err != nil {
		t.Error(err)
	}
	got := binlogStreamerErrors.Counts()["EventStreamer"]
	if got != before+1 {
		t.Errorf("got: %v, want: %+v", got, before+1)
	}
}

func TestDMLEvent(t *testing.T) {
	trans := &pb.BinlogTransaction{
		Statements: []*pb.BinlogTransaction_Statement{
			{
				Category: pb.BinlogTransaction_Statement_BL_SET,
				Sql:      "SET TIMESTAMP=2",
			}, {
				Category: pb.BinlogTransaction_Statement_BL_SET,
				Sql:      "SET INSERT_ID=10",
			}, {
				Category: pb.BinlogTransaction_Statement_BL_DML,
				Sql:      "query /* _stream _table_ (eid id name)  (null -1 'bmFtZQ==' ) (null 18446744073709551615 'bmFtZQ==' ); */",
			}, {
				Category: pb.BinlogTransaction_Statement_BL_DML,
				Sql:      "query",
			},
		},
		Timestamp:     1,
		TransactionId: "MariaDB/0-41983-20",
	}
	evs := &EventStreamer{
		sendEvent: func(event *proto.StreamEvent) error {
			switch event.Category {
			case "DML":
				want := `&{DML _table_ [{eid 8 0} {id 8 0} {name 253 0}] [[10 -1 name] [11 18446744073709551615 name]]  1 }`
				got := fmt.Sprintf("%v", event)
				if got != want {
					t.Errorf("got \n%s, want \n%s", got, want)
				}
			case "ERR":
				want := `&{ERR  [] [] query 1 }`
				got := fmt.Sprintf("%v", event)
				if got != want {
					t.Errorf("got %s, want %s", got, want)
				}
			case "POS":
				want := `&{POS  [] []  1 MariaDB/0-41983-20}`
				got := fmt.Sprintf("%v", event)
				if got != want {
					t.Errorf("got %s, want %s", got, want)
				}
			default:
				t.Errorf("unexppected: %#v", event)
			}
			return nil
		},
	}
	err := evs.transactionToEvent(trans)
	if err != nil {
		t.Error(err)
	}
}

func TestDDLEvent(t *testing.T) {
	trans := &pb.BinlogTransaction{
		Statements: []*pb.BinlogTransaction_Statement{
			{
				Category: pb.BinlogTransaction_Statement_BL_SET,
				Sql:      "SET TIMESTAMP=2",
			}, {
				Category: pb.BinlogTransaction_Statement_BL_DDL,
				Sql:      "DDL",
			},
		},
		Timestamp:     1,
		TransactionId: "MariaDB/0-41983-20",
	}
	evs := &EventStreamer{
		sendEvent: func(event *proto.StreamEvent) error {
			switch event.Category {
			case "DDL":
				want := `&{DDL  [] [] DDL 1 }`
				got := fmt.Sprintf("%v", event)
				if got != want {
					t.Errorf("got %s, want %s", got, want)
				}
			case "POS":
				want := `&{POS  [] []  1 MariaDB/0-41983-20}`
				got := fmt.Sprintf("%v", event)
				if got != want {
					t.Errorf("got %s, want %s", got, want)
				}
			default:
				t.Errorf("unexppected: %#v", event)
			}
			return nil
		},
	}
	err := evs.transactionToEvent(trans)
	if err != nil {
		t.Error(err)
	}
}

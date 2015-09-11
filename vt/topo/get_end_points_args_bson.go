// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package topo

import (
	"bytes"

	"github.com/youtube/vitess/go/bson"
	"github.com/youtube/vitess/go/bytes2"
)

// DO NOT EDIT.
// FILE GENERATED BY BSONGEN.

// MarshalBson bson-encodes GetEndPointsArgs.
func (getEndPointsArgs *GetEndPointsArgs) MarshalBson(buf *bytes2.ChunkedWriter, key string) {
	bson.EncodeOptionalPrefix(buf, bson.Object, key)
	lenWriter := bson.NewLenWriter(buf)

	bson.EncodeString(buf, "Cell", getEndPointsArgs.Cell)
	bson.EncodeString(buf, "Keyspace", getEndPointsArgs.Keyspace)
	bson.EncodeString(buf, "Shard", getEndPointsArgs.Shard)
	getEndPointsArgs.TabletType.MarshalBson(buf, "TabletType")

	lenWriter.Close()
}

// UnmarshalBson bson-decodes into GetEndPointsArgs.
func (getEndPointsArgs *GetEndPointsArgs) UnmarshalBson(buf *bytes.Buffer, kind byte) {
	switch kind {
	case bson.EOO, bson.Object:
		// valid
	case bson.Null:
		return
	default:
		panic(bson.NewBsonError("unexpected kind %v for GetEndPointsArgs", kind))
	}
	bson.Next(buf, 4)

	for kind := bson.NextByte(buf); kind != bson.EOO; kind = bson.NextByte(buf) {
		switch bson.ReadCString(buf) {
		case "Cell":
			getEndPointsArgs.Cell = bson.DecodeString(buf, kind)
		case "Keyspace":
			getEndPointsArgs.Keyspace = bson.DecodeString(buf, kind)
		case "Shard":
			getEndPointsArgs.Shard = bson.DecodeString(buf, kind)
		case "TabletType":
			getEndPointsArgs.TabletType.UnmarshalBson(buf, kind)
		default:
			bson.Skip(buf, kind)
		}
	}
}

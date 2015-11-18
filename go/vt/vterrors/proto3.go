// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vterrors

import (
	"errors"

	vtrpcpb "github.com/youtube/vitess/go/vt/proto/vtrpc"
)

// This files contains the necessary methods to send and receive errors
// as payloads of proto3 structures. It converts VitessError to and from
// vtrpcpb.Error

// FromVtRPCError recovers a VitessError from a *vtrpcpb.RPCError (which is how VitessErrors
// are transmitted across proto3 RPC boundaries).
func FromVtRPCError(rpcErr *vtrpcpb.RPCError) error {
	if rpcErr == nil {
		return nil
	}
	return &VitessError{
		Code: rpcErr.Code,
		err:  errors.New(rpcErr.Message),
	}
}

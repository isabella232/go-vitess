// Copyright 2016, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grpcthrottlerserver contains the gRPC implementation of the server
// side of the throttler service.
package grpcthrottlerserver

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/youtube/vitess/go/vt/servenv"
	"github.com/youtube/vitess/go/vt/throttler"

	"github.com/youtube/vitess/go/vt/proto/throttlerdata"
	"github.com/youtube/vitess/go/vt/proto/throttlerservice"
)

// Server is the gRPC server implementation of the Throttler service.
type Server struct {
	manager *throttler.Manager
}

// NewServer creates a new RPC server for a given throttler manager.
func NewServer(m *throttler.Manager) *Server {
	return &Server{m}
}

// SetMaxRate implements the gRPC server interface. It sets the rate on all
// throttlers controlled by the manager.
func (s *Server) SetMaxRate(ctx context.Context, request *throttlerdata.SetMaxRateRequest) (*throttlerdata.SetMaxRateResponse, error) {
	names := s.manager.SetMaxRate(request.Rate)
	return &throttlerdata.SetMaxRateResponse{
		Names: names,
	}, nil
}

// StartServer registers the Server instance with the gRPC server.
func StartServer(s *grpc.Server, m *throttler.Manager) {
	throttlerservice.RegisterThrottlerServer(s, NewServer(m))
}

func init() {
	servenv.OnRun(func() {
		if servenv.GRPCCheckServiceMap("throttler") {
			StartServer(servenv.GRPCServer, throttler.GlobalManager)
		}
	})
}

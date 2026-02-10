// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package grpcservice

import (
	"context"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/net"
)

func (s *GrpcService) RttTest(ctx context.Context, req *RttTestRequest) (*RttTestResponse, error) {
	dest := net.AddressPort{
		Address: net.ParseAddress(req.Addr),
		Port:    net.Port(req.Port),
	}

	dl, err := s.Client.DialerFactory.GetDialer(nil)
	if err != nil {
		return nil, err
	}
	l, err := s.Client.DialerFactory.GetPacketListener(nil)
	if err != nil {
		return nil, err
	}
	rtt, err := util.RttTest(ctx, dest, dl, l, s.Client.IPResolver)
	if err != nil {
		return nil, err
	}
	return &RttTestResponse{Ping: uint32(rtt)}, nil
}

// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package clientgrpc

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/app/create"
	"github.com/rs/zerolog/log"
)

func (s *ClientGrpc) AddInbound(ctx context.Context, req *AddInboundRequest) (*AddInboundResponse, error) {
	log.Info().Str("tag", req.HandlerConfig.Tag).Msg("AddInbound")
	in, err := create.NewInbound(req.HandlerConfig, s.Client.Dispatcher, s.Client.Policy)
	if err != nil {
		return nil, err
	}
	err = s.Client.InboundManager.AddInbound(in)
	if err != nil {
		return nil, err
	}
	return &AddInboundResponse{}, nil
}

func (s *ClientGrpc) RemoveInbound(ctx context.Context, in *RemoveInboundRequest) (*RemoveInboundResponse, error) {
	log.Info().Str("tag", in.Tag).Msg("RemoveInbound")
	if err := s.Client.InboundManager.RemoveInbound(in.Tag); err != nil {
		return nil, fmt.Errorf("failed to remove inbound handler: %w", err)
	}
	return &RemoveInboundResponse{}, nil
}

// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package grpcservice

import (
	"context"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/common/signal/done"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
)

type GrpcService struct {
	streamLock        sync.RWMutex
	communicateStream GrpcService_CommunicateServer

	Client *client.Client

	Done *done.Instance

	// if runningInService, when flutter side disconnect(meaning the app exits), call OnExit after 2 seconds
	RunningInService bool
	timeoutExit      *time.Timer
	OnExit           func()

	UpdateLantency bool

	UnimplementedGrpcServiceServer
}

type GrpcServiceConfig struct {
	Client         *client.Client
	GrpcServer     *grpc.Server
	UpdateLantency bool
}

func NewGrpcService(grpcConfig *GrpcServiceConfig) (*GrpcService, error) {
	client := &GrpcService{
		Client:         grpcConfig.Client,
		Done:           done.New(),
		UpdateLantency: grpcConfig.UpdateLantency,
	}
	RegisterGrpcServiceServer(grpcConfig.GrpcServer, client)
	return client, nil
}

func (s *GrpcService) Start() error {
	s.Client.Selectors.RegisterSelectedHandlersChangeObserver(s)
	selector := s.Client.Selectors.GetSelector("代理")
	if selector != nil {
		handlers := selector.GetHandlersBeingUsed()
		if len(handlers) == 1 {
			handler4BeingUsed.Store(handlers[0])
		}
	}
	return nil
}

func (s *GrpcService) Close() error {
	s.Done.Close()
	s.Client.Selectors.UnregisterSelectedHandlersChangeObserver(s)
	// TODO:
	// s.Subscription.Close()
	return nil
}

func (s *GrpcService) SetSubscriptionInterval(ctx context.Context, req *SetSubscriptionIntervalRequest) (*SetSubscriptionIntervalResponse, error) {
	log.Debug().Msgf("set subscription interval: %d", req.Interval)
	s.Client.Subscription.SetInterval(time.Duration(req.Interval) * time.Minute)
	return &SetSubscriptionIntervalResponse{}, nil
}

func (s *GrpcService) SetAutoSubscriptionUpdate(ctx context.Context, req *SetAutoSubscriptionUpdateRequest) (*Receipt, error) {
	log.Debug().Msgf("set auto subscription update: %t", req.Enable)
	s.Client.Subscription.SetAutoUpdate(req.Enable)
	return &Receipt{}, nil
}

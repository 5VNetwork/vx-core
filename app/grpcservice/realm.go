// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package grpcservice

import (
	"time"

	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/rs/zerolog/log"
)

// realmStatusProvider is implemented by any inbound that exposes realm server state.
// It is satisfied by *hysteria2.Inbound without an import cycle.
type realmStatusProvider interface {
	RealmStatus() (active bool, realmID string, publicAddrs []string, peers int)
}

const defaultRealmStatusInterval = 5 * time.Second

func (s *GrpcService) GetRealmStatusStream(req *GetRealmStatusStreamRequest, stream GrpcService_GetRealmStatusStreamServer) error {
	log.Debug().Msg("GetRealmStatusStream")
	interval := time.Duration(req.Interval) * time.Second
	if interval < time.Second {
		interval = defaultRealmStatusInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var hysteriaInbound realmStatusProvider
	if ib, err := s.Client.InboundManager.GetInbound("realmServer"); err == nil {
		for _, worker := range ib.(*proxy.ProxyInbound).GetWorkers() {
			if provider, ok := worker.(realmStatusProvider); ok {
				hysteriaInbound = provider
				break
			}
		}
	} else {
		log.Error().Err(err).Msg("failed to get realm server inbound")
		return err
	}

	sendStatus := func() error {
		log.Debug().Msg("sending realm status")
		status := &RealmServerStatus{}
		active, realmID, addrs, peers := hysteriaInbound.RealmStatus()
		log.Debug().Msgf("realm status: active=%v, realmID=%v, addrs=%v, peers=%v", active, realmID, addrs, peers)
		status.Active = active
		status.RealmId = realmID
		status.PublicAddresses = addrs
		status.Peers = int32(peers)
		return stream.Send(status)
	}

	if err := sendStatus(); err != nil {
		return err
	}
	for {
		select {
		case <-s.Done.Wait():
			return nil
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			if err := sendStatus(); err != nil {
				return err
			}
		}
	}
}

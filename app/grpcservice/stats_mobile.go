// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

//go:build ios || android || darwin
// +build ios android darwin

package grpcservice

import (
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
)

func (s *GrpcService) GetStatsStream(in *GetStatsRequest,
	stream GrpcService_GetStatsStreamServer) error {
	log.Debug().Msg("get outbound stats stream request received")
	s.Client.Policy.StatsPolicy.SetOutboundStats(true)
	var m runtime.MemStats
	timer := time.NewTicker(time.Duration(in.Interval) * time.Second)
	defer timer.Stop()

	sendStats := func() error {
		st := s.Client.OutStats

		statsList := make([]*OutboundStats, 0, len(st.Map))
		st.Lock()
		for tag, stats := range st.Map {
			statsList = append(statsList, &OutboundStats{
				Up:       stats.UpCounter.Swap(0),
				Down:     stats.DownCounter.Swap(0),
				Rate:     stats.Throughput.Load(),
				Ping:     stats.Ping.Load(),
				Id:       tag,
				Interval: float32(time.Since(stats.Interval.Swap(time.Now()).(time.Time)).Seconds()),
			})
		}
		st.Unlock()

		runtime.ReadMemStats(&m)
		memory := m.Sys

		return stream.Send(&StatsResponse{
			Connections: s.Client.Dispatcher.Flows.Load() +
				s.Client.Dispatcher.PacketConns.Load(),
			Memory: memory,
			Stats:  statsList,
		})
	}

	err := sendStats()
	if err != nil {
		log.Error().Err(err).Msg("failed to send outbound stats response")
		return err
	}

	for {
		if s.Done.Done() {
			return nil
		}
		select {
		case <-s.Done.Wait():
			return nil
		case <-stream.Context().Done():
			return nil
		case <-timer.C:
			err := sendStats()
			if err != nil {
				log.Error().Err(err).Msg("failed to send outbound stats response")
				return err
			}
		}
	}
}

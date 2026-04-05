// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dispatcher

import (
	"context"
	"sync"

	"github.com/5vnetwork/vx-core/app/inbound/monitor"
	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/common/buf"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/units"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type StatsHook struct {
	StatsPolicy  i.StatsSetting
	Um           *user.Manager
	LinkStats    sync.Map              //key is prefix string, value is *LinkStats
	InboundStats *monitor.InboundStats //key is inbound tag, value is *InboundStats
}

func (p *StatsHook) AfterHandlerSelection(ctx context.Context, info *session.Info, rw any,
	handler i.Outbound) (context.Context, any, error) {
	var ups session.UpCounters
	var downs session.DownCounters

	if p.StatsPolicy.CalculateOutboundLinkStats() || p.StatsPolicy.CalculateInboundLinkStats() {
		var throughputAdder linkStatsAdder
		if p.StatsPolicy.CalculateInboundLinkStats() &&
			info.Source.Address != nil && info.Source.Address.Family().IsIP() {
			network := mynet.PrefixStringFromIP(info.Source.Address.IP())
			stats, ok := p.LinkStats.Load(network)
			if !ok {
				stats = &LinkStats{}
				p.LinkStats.Store(network, stats)
			}
			throughputAdder = stats.(*LinkStats)
		}
		if throughputAdder != nil {
			ls := &linkStats{
				ctx:     ctx,
				ohStats: throughputAdder,
			}
			ups = append(ups, ls)
			downs = append(downs, ls)
		}
	}

	// server
	if p.StatsPolicy.CalculateInboundStats() && p.InboundStats != nil {
		inboundStats := p.InboundStats.Get(info.InboundTag)
		ups = append(ups, session.AtomicCounter{
			Counter: &inboundStats.Traffic,
		})
		downs = append(downs, session.AtomicCounter{
			Counter: &inboundStats.Traffic,
		})
	}
	if p.StatsPolicy.CalculateUserStats() && p.Um != nil {
		us := p.Um.GetUser(info.User.Uid())
		if us != nil {
			ups = append(ups, session.AtomicCounter{
				Counter: us.Counter(),
			})
			downs = append(downs, session.AtomicCounter{
				Counter: us.Counter(),
			})
		} else {
			log.Warn().Str("uid", info.User.Uid()).Msg("no user stats found")
		}
	}
	if p.StatsPolicy.CalculateSessionStats() {
		ups = append(ups, session.AtomicCounter{
			Counter: &info.SessionUpCounter,
		})
		downs = append(downs, session.AtomicCounter{
			Counter: &info.SessionDownCounter,
		})
	}

	if len(ups) > 0 || len(downs) > 0 {
		if r, ok := rw.(i.DeadlineRW); ok {
			rw = &StatsDeadlineRW{
				DeadlineRW:  r,
				upCounter:   ups,
				downCounter: downs,
			}
		} else if r, ok := rw.(buf.ReaderWriter); ok {
			rw = &StatsReaderWriter{
				ReaderWriter: r,
				upCounter:    ups,
				downCounter:  downs,
			}
		} else {
			rw = &StatsPacketConn{
				PacketReaderWriter: rw.(udp.PacketReaderWriter),
				upCounter:          ups,
				downCounter:        downs,
			}
		}
	}
	return ctx, rw, nil
}

type LinkStats struct {
	sync.Mutex
	Num       uint32
	BWTotal   uint32 //MBps
	PingTotal uint32 //ms
}

func (l *LinkStats) AddPing(pingMs uint64) {
	l.Lock()
	defer l.Unlock()
	l.PingTotal += uint32(pingMs)
}

func (l *LinkStats) AddThroughput(bytesPerSec uint64) {
	l.Lock()
	defer l.Unlock()
	l.Num++
	l.BWTotal += uint32(units.BytesToMB(bytesPerSec))
}

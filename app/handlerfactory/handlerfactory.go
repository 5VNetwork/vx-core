package handlerfactory

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/dispatcher/linkstats"
	"github.com/5vnetwork/vx-core/app/dispatcher/variants"
	outboundstats "github.com/5vnetwork/vx-core/app/outbound/stats"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
)

type HandlerFactory struct {
	DialerFactory               transport.DialerFactory
	Policy                      *policy.Policy
	IPResolver                  i.IPResolver
	IPResolverForRequestAddress i.IPResolver
	EchResolver                 i.ECHResolver
	Hysteria2RejectQuic         bool
	HandlerLinkStats            bool
	HandlerMeter                bool
	OutStats                    *outboundstats.OutStats
	TotalUpCounter              *atomic.Uint64
	TotalDownCounter            *atomic.Uint64
}

func (c *HandlerFactory) CreateHandler(hs ...*configs.HandlerConfig) (i.Outbound, error) {
	df := c.DialerFactory
	var outbound i.Outbound
	if len(hs) > 1 {
		handlers := make([]*configs.OutboundHandlerConfig, 0)
		for _, handlerConfig := range hs {
			if handlerConfig.GetOutbound() != nil {
				handlers = append(handlers, handlerConfig.GetOutbound())
			} else if handlerConfig.GetChain() != nil {
				handlers = append(handlers, handlerConfig.GetChain().GetHandlers()...)
			}
		}

		var tags []string
		for _, handler := range hs {
			if handler.GetOutbound() != nil {
				tags = append(tags, handler.GetOutbound().GetTag())
			} else if handler.GetChain() != nil {
				tags = append(tags, handler.GetChain().GetTag())
			}
		}
		tag := strings.Join(tags, "-")

		ch, err := create.NewChainHandler(&create.ChainHandlerConfig{
			ChainHandlerConfig: &configs.ChainHandlerConfig{
				Tag:      tag,
				Handlers: handlers,
			},
			Policy:                      c.Policy,
			IPResolver:                  c.IPResolver,
			DF:                          c.DialerFactory,
			IPResolverForRequestAddress: c.IPResolverForRequestAddress,
			RejectQuic:                  c.Hysteria2RejectQuic,
			EchResolver:                 c.EchResolver,
		})
		if err != nil {
			return nil, err
		}
		outbound = ch
	} else {
		handler, err := create.NewHandler(&create.HandlerConfig{
			HandlerConfig:               hs[0],
			DialerFactory:               df,
			Policy:                      c.Policy,
			IPResolver:                  c.IPResolver,
			EchResolver:                 c.EchResolver,
			IPResolverForRequestAddress: c.IPResolverForRequestAddress,
			RejectQuic:                  c.Hysteria2RejectQuic,
		})
		if err != nil {
			return nil, err
		}
		outbound = handler
	}

	if c.OutStats != nil && outbound.Tag() != "direct" && outbound.Tag() != "dns" {
		stats := c.OutStats.Get(outbound.Tag())

		var ups session.UpCounters
		var downs session.DownCounters
		if c.HandlerMeter {
			ups = append(ups, session.AtomicCounter{
				Counter: &stats.UpCounter,
			})
			downs = append(downs, session.AtomicCounter{
				Counter: &stats.DownCounter,
			})
		}
		if c.TotalUpCounter != nil {
			ups = append(ups, session.AtomicCounter{
				Counter: c.TotalUpCounter,
			})
		}
		if c.TotalDownCounter != nil {
			downs = append(downs, session.AtomicCounter{
				Counter: c.TotalDownCounter,
			})
		}

		so := &StatsOutbound{
			Outbound:      outbound,
			ups:           ups,
			downs:         downs,
			activeChecker: &stats.ActiveTime,
		}
		if c.HandlerLinkStats {
			so.outStats = c.OutStats.Get(outbound.Tag())
		}
		return so, nil
	}

	return outbound, nil
}

type StatsOutbound struct {
	i.Outbound
	outStats      *outboundstats.OutboundHandlerStats
	ups           session.UpCounters
	downs         session.DownCounters
	activeChecker *atomic.Value
}

func (s *StatsOutbound) Support6() bool {
	if h, ok := s.Outbound.(i.HandlerWith6Info); ok {
		return h.Support6()
	}
	return false
}

func (s *StatsOutbound) HandleFlow(ctx context.Context,
	dst net.Destination, rw buf.ReaderWriter) error {
	var ups session.UpCounters
	var downs session.DownCounters
	if s.outStats != nil {
		ups = append(ups, s.ups...)
		downs = append(downs, s.downs...)
		ups = append(ups, linkstats.NewLinkStats(ctx, s.outStats))
		downs = append(downs, linkstats.NewLinkStats(ctx, s.outStats))
	} else {
		ups = s.ups
		downs = s.downs
	}
	if r, ok := rw.(i.DeadlineRW); ok {
		rw = variants.NewStatsDeadlineRW(r, ups, downs, s.activeChecker)
	} else {
		rw = variants.NewStatsReaderWriter(r, ups, downs, s.activeChecker)
	}
	return s.Outbound.HandleFlow(ctx, dst, rw)
}

func (s *StatsOutbound) HandlePacketConn(ctx context.Context,
	dst net.Destination, pc udp.PacketReaderWriter) error {
	var ups session.UpCounters
	var downs session.DownCounters
	if s.outStats != nil {
		ups = append(ups, s.ups...)
		downs = append(downs, s.downs...)
		ups = append(ups, linkstats.NewLinkStats(ctx, s.outStats))
		downs = append(downs, linkstats.NewLinkStats(ctx, s.outStats))
	} else {
		ups = s.ups
		downs = s.downs
	}
	rw := variants.NewStatsPacketConn(pc, ups, downs, s.activeChecker)
	return s.Outbound.HandlePacketConn(ctx, dst, rw)
}

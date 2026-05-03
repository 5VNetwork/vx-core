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
		so := &StatsOutbound{
			Outbound: outbound,
			statsFunc: func(ctx context.Context, a any) any {
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
				if c.HandlerLinkStats {
					ls := linkstats.NewLinkStats(ctx, stats)
					ups = append(ups, ls)
					downs = append(downs, ls)
				}
				if r, ok := a.(buf.ReaderWriter); ok {
					return variants.NewStatsReaderWriter(r, ups, downs, &stats.ActiveTime)
				} else if pc, ok := a.(i.DeadlineRW); ok {
					return variants.NewStatsDeadlineRW(pc, ups, downs, &stats.ActiveTime)
				} else if pc, ok := a.(udp.PacketReaderWriter); ok {
					return variants.NewStatsPacketConn(pc, ups, downs, &stats.ActiveTime)
				} else {
					return a
				}
			},
		}
		return so, nil
	}

	return outbound, nil
}

type StatsOutbound struct {
	i.Outbound
	statsFunc func(ctx context.Context, a any) any
}

func (s *StatsOutbound) Support6() bool {
	if h, ok := s.Outbound.(i.HandlerWith6Info); ok {
		return h.Support6()
	}
	return false
}

func (s *StatsOutbound) HandleFlow(ctx context.Context,
	dst net.Destination, rw buf.ReaderWriter) error {
	return s.Outbound.HandleFlow(ctx, dst, s.statsFunc(ctx, rw).(buf.ReaderWriter))
}

func (s *StatsOutbound) HandlePacketConn(ctx context.Context,
	dst net.Destination, pc udp.PacketReaderWriter) error {
	return s.Outbound.HandlePacketConn(ctx, dst, s.statsFunc(ctx, pc).(udp.PacketReaderWriter))
}

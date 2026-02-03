// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dispatcher

import (
	"context"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type IdleHook struct {
	TimeoutPolicy i.TimeoutSetting
}

func (p *IdleHook) getTimeout(info *session.Info) time.Duration {
	if info.Target.Port == 22 {
		return p.TimeoutPolicy.SshIdleTimeout()
	}
	if info.Target.Port == 53 {
		return p.TimeoutPolicy.DnsIdleTimeout()
	}
	if info.Target.Network == mynet.Network_TCP {
		return p.TimeoutPolicy.TcpIdleTimeout()
	}
	return p.TimeoutPolicy.UdpIdleTimeout()
}

func (p *IdleHook) BeforeHandlerSelection(ctx context.Context, info *session.Info,
	rw any) (context.Context, any, error) {
	// idle
	idleTimeout := p.getTimeout(info)
	if idleTimeout != 0 {
		var cancelCause context.CancelCauseFunc
		ctx, cancelCause = context.WithCancelCause(ctx)
		idleChecker := signal.NewActivityChecker(func() {
			cancelCause(errors.ErrIdle)
			log.Ctx(ctx).Debug().Msg("idle timeout")
		}, idleTimeout)
		if r, ok := rw.(i.DeadlineRW); ok {
			rw = &TimeoutDeadlineRW{
				timeout:    p.TimeoutPolicy,
				idle:       idleChecker,
				DeadlineRW: r,
				upOnly:     info.Target.Network == mynet.Network_UDP,
			}
		} else if r, ok := rw.(buf.ReaderWriter); ok {
			rw = &TimeoutReaderWriter{
				timeout:      p.TimeoutPolicy,
				idle:         idleChecker,
				ReaderWriter: r,
				upOnly:       info.Target.Network == mynet.Network_UDP,
			}
		} else if pc, ok := rw.(udp.PacketReaderWriter); ok {
			rw = &TimeoutPacketConn{
				idle:               idleChecker,
				PacketReaderWriter: pc,
			}
		}
		info.ActivityChecker = idleChecker
		return ctx, rw, nil
	} else {
		log.Ctx(ctx).Debug().Msg("no idle timeout")
		return ctx, rw, nil
	}
}

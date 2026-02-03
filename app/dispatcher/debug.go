// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dispatcher

import (
	"context"
	"strings"

	"github.com/5vnetwork/vx-core/common/appid"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DebugHook struct {
}

func (p *DebugHook) BeforeHandlerSelection(ctx context.Context, info *session.Info,
	rw any) (context.Context, any, error) {
	if info.AppId == "" {
		if (zerolog.GlobalLevel() == zerolog.DebugLevel) &&
			(!strings.Contains(info.InboundTag, "dns")) &&
			!strings.Contains(info.InboundTag, "DNS") {
			appId, err := appid.GetAppId(ctx, info.Source, &info.Target)
			if err != nil {
				log.Ctx(ctx).Debug().Err(err).Msg("failed to get appId")
			}
			info.AppId = appId
		}
	}
	return ctx, rw, nil
}

func (p *DebugHook) AfterHandlerSelection(ctx context.Context, info *session.Info, rw any,
	handler i.Outbound) (context.Context, any, error) {
	tag := ""
	if handler != nil {
		tag = handler.Tag()
	}
	if _, ok := rw.(buf.ReaderWriter); ok {
		log.Ctx(ctx).Debug().Str("dst", info.Target.String()).Str("out_tag", tag).
			Any("net", info.Target.Network.String()).Str("in_tag", info.InboundTag).
			Str("src", info.Source.String()).Str("sniffed_domain", info.SniffedDomain).
			Str("app", info.AppId).Str("protocol", info.Protocol).Msg("flow info")
	} else {
		log.Ctx(ctx).Debug().Str("udp src", info.Source.String()).Str("inbound", info.InboundTag).
			Str("sniff", info.SniffedDomain).Str("dst", info.Target.String()).Str("outbound", tag).
			Str("app", info.AppId).Str("protocol", info.Protocol).Msg("packetconn info")
	}
	return ctx, rw, nil
}

func (p *DebugHook) FlowSessionEnd(ctx context.Context, info *session.Info, err error) {
	log.Ctx(ctx).Debug().Uint64("up", info.SessionUpCounter.Load()).
		Uint64("down", info.SessionDownCounter.Load()).Msg("flow session end")
}

func (p *DebugHook) PacketConnSessionEnd(ctx context.Context, info *session.Info, err error) {
	log.Ctx(ctx).Debug().Uint64("up", info.SessionUpCounter.Load()).
		Uint64("down", info.SessionDownCounter.Load()).Msg("packet conn session end")
}

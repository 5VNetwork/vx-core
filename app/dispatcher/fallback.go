// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dispatcher

import (
	"context"
	"strings"

	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/common/buf"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type Fallbacker struct {
	FallbackToProxy  bool
	FallbackToDomain bool
	Sm               *selector.Selectors
	Om               i.OutboundManager
	Logger           fallbackLogger
}

type fallbackLogger interface {
	LogFallback(info *session.Info, tag string)
}

func (p *Fallbacker) Fallback(ctx context.Context, info *session.Info,
	rw buf.ReaderWriter, handler i.Outbound, err error) error {
	if handler.Tag() == "direct" && p.FallbackToProxy {
		log.Ctx(ctx).Warn().Str("dst", info.Target.String()).Str("domain", info.SniffedDomain).Msg("fallback to proxy")
		// since ip might be polluted, replace it with the domain
		if info.Target.Address.Family().IsIP() && info.SniffedDomain != "" {
			info.Target.Address = mynet.DomainAddress(info.SniffedDomain)
		}
		proxySelector := p.Sm.GetSelector("代理")
		var handler i.Outbound
		if proxySelector != nil {
			handler = proxySelector.GetHandler(info)
		} else {
			for _, selector := range p.Sm.GetAllSelectors() {
				handler = selector.GetHandler(info)
				if handler != nil {
					break
				}
			}
			for _, h := range p.Om.GetAllHandlers() {
				if h != nil && h.Tag() != "direct" && h.Tag() != "dns" {
					handler = h
					break
				}
			}
		}
		if handler != nil {
			if p.Logger != nil {
				p.Logger.LogFallback(info, handler.Tag())
			}
			err = handler.HandleFlow(ctx, info.Target, rw)
		}
	} else if p.FallbackToDomain && handler.Tag() != "direct" && info.Target.Address.Family().IsIP() &&
		(info.GetTargetDomain() != "") && strings.Contains(err.Error(), "i/o timeout") {
		// This might due to polluted ip
		log.Ctx(ctx).Warn().Str("dst", info.Target.String()).Str("domain", info.GetTargetDomain()).Msg("retry domain")
		info.Target.Address = mynet.DomainAddress(info.GetTargetDomain())
		err = handler.HandleFlow(ctx, info.Target, rw)
	}
	return err
}

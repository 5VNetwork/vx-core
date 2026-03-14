// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dispatcher

import (
	"context"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

// type Fallbacker struct {
// 	FallbackToProxy bool
// 	Sm              *selector.Selectors
// 	Om              i.OutboundManager
// 	Logger          fallbackLogger
// }

// type fallbackLogger interface {
// 	LogFallback(info *session.Info, tag string)
// }

// func (p *Fallbacker) Fallback(ctx context.Context, info *session.Info,
// 	rw buf.ReaderWriter, handler i.Outbound, err error) error {
// 	if handler.Tag() == "direct" && p.FallbackToProxy {
// 		log.Ctx(ctx).Warn().Str("dst", info.Target.String()).
// 			Str("domain", info.SniffedDomain).Msg("fallback to proxy")
// 		// since ip might be polluted, replace it with the domain
// 		if info.Target.Address.Family().IsIP() && info.SniffedDomain != "" {
// 			info.Target.Address = mynet.DomainAddress(info.SniffedDomain)
// 		}
// 		proxySelector := p.Sm.GetSelector("代理")
// 		var handler i.Outbound
// 		if proxySelector != nil {
// 			handler = proxySelector.GetHandler(info)
// 		} else {
// 			for _, selector := range p.Sm.GetAllSelectors() {
// 				handler = selector.GetHandler(info)
// 				if handler != nil {
// 					break
// 				}
// 			}
// 			for _, h := range p.Om.GetAllHandlers() {
// 				if h != nil && h.Tag() != "direct" && h.Tag() != "dns" {
// 					handler = h
// 					break
// 				}
// 			}
// 		}
// 		if handler != nil {
// 			if p.Logger != nil {
// 				p.Logger.LogFallback(info, handler.Tag())
// 			}
// 			err = handler.HandleFlow(ctx, info.Target, rw)
// 		}
// 	} /* else if p.FallbackToDomain && info.Target.Address.Family().IsIP() &&
// 		(info.GetTargetDomain() != "") && strings.Contains(err.Error(), "i/o timeout") {
// 		// This might due to polluted ip
// 		log.Ctx(ctx).Warn().Str("dst", info.Target.String()).Str("domain", info.GetTargetDomain()).Msg("retry domain")
// 		info.Target.Address = mynet.DomainAddress(info.GetTargetDomain())
// 		err = handler.HandleFlow(ctx, info.Target, rw)
// 	} */
// 	return err
// }

type cacheReaderWriter struct {
	buf.DdlReaderWriter
	mb               buf.MultiBuffer
	stopCaching      bool
	maximumCacheSize int
	reading          bool
	done             bool
	ctx              context.Context
}

func (f *cacheReaderWriter) Done(ctx context.Context) {
	f.done = true
	err := f.DdlReaderWriter.SetReadDeadline(time.Now().Add(-100 * time.Millisecond))
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("set read deadline")
	}
	log.Ctx(ctx).Debug().Msg("set read deadline to -100ms")
}

func (f *cacheReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if f.done {
		return nil, errors.New("done")
	}
	if f.stopCaching {
		return f.DdlReaderWriter.ReadMultiBuffer()
	}

	f.reading = true
	mb, err := f.DdlReaderWriter.ReadMultiBuffer()
	f.reading = false
	if err != nil {
		log.Ctx(f.ctx).Error().Err(err).Msg("read multi buffer")
		return nil, err
	}

	if len(f.mb)+len(mb) > f.maximumCacheSize {
		f.stopCaching = true
		buf.ReleaseMulti(f.mb)
		f.mb = nil
	} else {
		clone := mb.Clone()
		f.mb = append(f.mb, clone...)
	}

	return mb, nil
}

func (f *cacheReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if f.done {
		return errors.New("done")
	}
	if !f.stopCaching {
		f.stopCaching = true
		buf.ReleaseMulti(f.mb)
		f.mb = nil
	}
	return f.DdlReaderWriter.WriteMultiBuffer(mb)
}

type FallbackDeadlineRW struct {
	buf.ReaderWriter
	i.DeadlineRW
	mb               buf.MultiBuffer
	stopCaching      bool
	maximumCacheSize int
}

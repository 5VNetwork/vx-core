// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package userlogger

import (
	"context"
	"io"
	"strings"
	"sync/atomic"
	"time"

	vx "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/userlogger"
	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/common/appid"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/hysteria2"
	"github.com/5vnetwork/vx-core/proxy/vless/outbound"
	"github.com/rs/zerolog/log"
)

type UserLogger struct {
	LogAppId         atomic.Bool
	enabled          atomic.Bool
	logSessionEnd    atomic.Bool
	logRealtimeUsage atomic.Bool

	ch   chan struct{}
	done *done.Instance
	buf  *buf.RingBuffer[*vx.UserLogMessage]

	dns ipToDomain
}

type ipToDomain interface {
	GetDomain(ip net.IP) []string
	GetResolvers(domain string, ip net.IP) []net.Address
}

func NewUserLogger(enabled bool, logAppId bool, size int) *UserLogger {
	ul := &UserLogger{
		enabled: atomic.Bool{},
		buf:     buf.NewRingBuffer[*vx.UserLogMessage](size),
		done:    done.New(),
		ch:      make(chan struct{}),
	}
	ul.SetEnabled(enabled)
	ul.LogAppId.Store(logAppId)
	return ul
}

func (s *UserLogger) SetLogSessionEnd(enabled bool) {
	s.logSessionEnd.Store(enabled)
}

func (s *UserLogger) SetLogRealtimeUsage(enabled bool) {
	s.logRealtimeUsage.Store(enabled)
}

func (s *UserLogger) SetDns(dnsConn ipToDomain) {
	s.dns = dnsConn
}

func (s *UserLogger) SetEnabled(enabled bool) {
	s.enabled.Store(enabled)
	if !enabled {
		s.buf.Clear()
	}
}

func (s *UserLogger) OnFallback(info *session.Info, previous, tag string) {
	if !s.enabled.Load() {
		return
	}
	if s.done.Done() {
		return
	}

	msg := &vx.UserLogMessage{
		Message: &vx.UserLogMessage_Fallback{
			Fallback: &vx.Fallback{
				Tag: tag,
				Sid: uint32(info.ID),
			}},
	}
	s.buf.Add(msg)
	select {
	case s.ch <- struct{}{}:
	default:
	}
}

func (s *UserLogger) LogReject(info *session.Info, reason string) {
	if !s.enabled.Load() {
		return
	}
	if s.done.Done() {
		return
	}

	if info.AppId == "" && s.LogAppId.Load() {
		target := &info.Target
		if info.FakeIP != nil {
			target = &net.Destination{
				Address: net.IPAddress(info.FakeIP),
				Port:    info.Target.Port,
				Network: info.Target.Network,
			}
		}
		appId, err := appid.GetAppId(context.Background(), info.Source, target)
		if err != nil {
			log.Debug().Err(err).Msg("failed to get appId")
		}
		info.AppId = appId
	}

	msg := &vx.UserLogMessage{
		Message: &vx.UserLogMessage_RejectMessage{
			RejectMessage: &vx.RejectMessage{
				Dst:       info.Target.Address.String(),
				Domain:    info.SniffedDomain,
				Timestamp: time.Now().Unix(),
				Reason:    reason,
				AppId:     info.AppId,
			},
		},
	}
	s.buf.Add(msg)
	select {
	case s.ch <- struct{}{}:
	default:
	}
}

func (s *UserLogger) AfterHandlerSelection(ctx context.Context, info *session.Info, rw any,
	handler i.Outbound) (context.Context, any, error) {
	tag := ""
	if handler != nil {
		tag = handler.Tag()
	}
	s.LogRoute(info, tag)

	if s.logRealtimeUsage.Load() {
		go s.logSessionUsageLoop(ctx, info)
	}

	return ctx, rw, nil
}

func (s *UserLogger) logSessionUsageLoop(ctx context.Context, info *session.Info) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		if !s.enabled.Load() || s.done.Done() || !s.logRealtimeUsage.Load() {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			up := info.SessionUpCounter.Load()
			down := info.SessionDownCounter.Load()

			msg := &vx.UserLogMessage{
				Message: &vx.UserLogMessage_SessionUsage{
					SessionUsage: &vx.SessionUsage{
						Sid:  uint32(info.ID),
						Up:   up,
						Down: down,
						Ts:   time.Now().Unix(),
					},
				},
			}

			s.buf.Add(msg)
			select {
			case s.ch <- struct{}{}:
			default:
			}
		}
	}
}

func (s *UserLogger) LogRoute(info *session.Info, tag string) {
	if !s.enabled.Load() {
		return
	}
	if s.done.Done() {
		return
	}
	if tag == "dns" || strings.Contains(info.InboundTag, "dns") {
		return
	}

	if info.AppId == "" && s.LogAppId.Load() {
		target := &info.Target
		if info.FakeIP != nil {
			target = &net.Destination{
				Address: net.IPAddress(info.FakeIP),
				Port:    info.Target.Port,
				Network: info.Target.Network,
			}
		}
		appId, err := appid.GetAppId(context.Background(), info.Source, target)
		if err != nil {
			log.Debug().Err(err).Msg("failed to get appId")
		}
		info.AppId = appId
	}

	ipToDomain := ""
	if info.SniffedDomain == "" && info.Target.Address.Family().IsIP() && s.dns != nil {
		ipToDomain = strings.Join(s.dns.GetDomain(info.Target.Address.IP()), ",")
	}

	msg := &vx.UserLogMessage{
		Message: &vx.UserLogMessage_RouteMessage{
			RouteMessage: &vx.RouteMessage{
				Sid:           uint32(info.ID),
				Dst:           info.Target.Address.String(),
				Tag:           tag,
				SniffDomain:   info.SniffedDomain,
				AppId:         info.AppId,
				IpToDomain:    ipToDomain,
				Timestamp:     time.Now().Unix(),
				SelectorTag:   info.UsedSelector,
				MatchedRule:   info.MatchedRule,
				InboundTag:    info.InboundTag,
				Network:       info.Target.Network.String(),
				SniffProtofol: info.Protocol,
				Source:        info.Source.String(),
			},
		},
	}
	s.buf.Add(msg)
	select {
	case s.ch <- struct{}{}:
	default:
	}
}

func (s *UserLogger) FlowSessionEnd(ctx context.Context, info *session.Info, err error) {
	if err != nil {
		s.logSessionError(info, err)
	}
	if s.logSessionEnd.Load() {
		s.logSessionEndMessage(info)
	}
}

func (s *UserLogger) PacketConnSessionEnd(ctx context.Context, info *session.Info, err error) {
	if err != nil {
		s.logSessionError(info, err)
	}
	if s.logSessionEnd.Load() {
		s.logSessionEndMessage(info)
	}
}

func (s *UserLogger) logSessionEndMessage(info *session.Info) {
	if !s.enabled.Load() {
		return
	}
	if s.done.Done() {
		return
	}

	se := &vx.SessionEnd{
		Sid:   uint32(info.ID),
		Up:    info.SessionUpCounter.Load(),
		Down:  info.SessionDownCounter.Load(),
		Start: info.StartTime,
		End:   time.Now().Unix(),
	}

	msg := &vx.UserLogMessage{
		Message: &vx.UserLogMessage_SessionEnd{
			SessionEnd: se,
		},
	}
	s.buf.Add(msg)
	select {
	case s.ch <- struct{}{}:
	default:
	}
}

// when down link is 0, this is called. err might be nil
func (s *UserLogger) logSessionError(info *session.Info, err error) {
	if !s.enabled.Load() {
		return
	}
	if s.done.Done() {
		return
	}

	if errors.Is(err, hysteria2.ErrRejectQuic) {
		return
	} else if errors.Is(err, outbound.ErrRejectQuic) {
		return
	}
	if err == router.ErrBlocked {
		s.LogReject(info, err.Error())
	} else if err == router.ErrNoHandler {
		s.LogRoute(info, "")
	} else if info.SessionDownCounter.Load() == 0 || info.SessionUpCounter.Load() == 0 {
		// udp idle is not considered an error
		if info.Target.Network == net.Network_UDP && errors.Is(err, errors.ErrIdle) {
			return
		}
		if errors.Is(err, io.EOF) {
			return
		}
		se := &vx.SessionError{
			Up:   uint32(info.SessionUpCounter.Load()),
			Down: uint32(info.SessionDownCounter.Load()),
			Sid:  uint32(info.ID),
		}
		if err != nil {
			se.Message = err.Error()
		}
		domain := info.SniffedDomain
		if info.Target.Address.Family().IsIP() && domain != "" {
			if s.dns != nil {
				resolvers := s.dns.GetResolvers(domain, info.Target.Address.IP())
				if len(resolvers) > 0 {
					resolversStr := make([]string, len(resolvers))
					for i, resolver := range resolvers {
						resolversStr[i] = resolver.String()
					}
					se.Dns = strings.Join(resolversStr, ",")
				}
			}
		}

		msg := &vx.UserLogMessage{
			Message: &vx.UserLogMessage_SessionError{
				SessionError: se,
			},
		}
		s.buf.Add(msg)
		select {
		case s.ch <- struct{}{}:
		default:
		}
	}

}

func (s *UserLogger) LogError(err error) {
	if !s.enabled.Load() {
		return
	}
	if s.done.Done() {
		return
	}
	msg := &vx.UserLogMessage{
		Message: &vx.UserLogMessage_ErrorMessage{
			ErrorMessage: &vx.ErrorMessage{
				Message:   err.Error(),
				Timestamp: time.Now().Unix(),
			},
		},
	}
	s.buf.Add(msg)
	select {
	case s.ch <- struct{}{}:
	default:
	}
}

func (s *UserLogger) Close() error {
	if s.done.Done() {
		return nil
	}
	s.done.Close()
	s.buf.Clear()
	return nil
}

func (s *UserLogger) ReadLog(ctx context.Context, slice []*vx.UserLogMessage) (int, error) {
	for {
		if !s.enabled.Load() {
			return 0, errors.New("user logger disabled")
		}
		select {
		case <-s.ch:
			n, err := s.buf.Read(slice)
			if err != nil {
				return 0, err
			}
			return n, nil
		case <-s.done.Wait():
			return 0, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
}

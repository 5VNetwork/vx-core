package userlogger

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/appid"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/rs/zerolog/log"
)

type UserLogger struct {
	LogAppId atomic.Bool
	enabled  atomic.Bool

	ch   chan struct{}
	done *done.Instance
	buf  *buf.RingBuffer[*UserLogMessage]

	dns ipToDomain
}

type ipToDomain interface {
	GetDomain(ip net.IP) []string
	GetResolvers(domain string, ip net.IP) []string
}

func NewUserLogger(enabled bool, logAppId bool, size int) *UserLogger {
	ul := &UserLogger{
		enabled: atomic.Bool{},
		buf:     buf.NewRingBuffer[*UserLogMessage](size),
		done:    done.New(),
		ch:      make(chan struct{}),
	}
	ul.SetEnabled(enabled)
	ul.LogAppId.Store(logAppId)
	return ul
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

func (s *UserLogger) LogFallback(info *session.Info, tag string) {
	if !s.enabled.Load() {
		return
	}
	if s.done.Done() {
		return
	}

	msg := &UserLogMessage{
		Message: &UserLogMessage_Fallback{
			Fallback: &Fallback{
				Tag: tag,
				Sid: uint32(info.ID),
			},
		},
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

	// ipToDomain := info.IpToDomain
	// if s.dns != nil && ipToDomain == "" &&
	// 	info.Target.Address != nil && info.Target.Address.Family().IsIP() {
	// 	ipToDomain = s.dns.GetDomain(info.Target.Address.IP())
	// }

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

	msg := &UserLogMessage{
		Message: &UserLogMessage_RejectMessage{
			RejectMessage: &RejectMessage{
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

	msg := &UserLogMessage{
		Message: &UserLogMessage_RouteMessage{
			RouteMessage: &RouteMessage{
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

// when down link is 0, this is called. err might be nil
func (s *UserLogger) LogSessionError(info *session.Info, err error) {
	if !s.enabled.Load() {
		return
	}
	if s.done.Done() {
		return
	}
	se := &SessionError{
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
				se.Dns = strings.Join(resolvers, ",")
			}
		}
	}

	msg := &UserLogMessage{
		Message: &UserLogMessage_SessionError{
			SessionError: se,
		},
	}
	s.buf.Add(msg)
	select {
	case s.ch <- struct{}{}:
	default:
	}
}

func (s *UserLogger) LogError(err error) {
	if !s.enabled.Load() {
		return
	}
	if s.done.Done() {
		return
	}
	msg := &UserLogMessage{
		Message: &UserLogMessage_ErrorMessage{
			ErrorMessage: &ErrorMessage{
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

func (s *UserLogger) ReadLog(ctx context.Context, slice []*UserLogMessage) (int, error) {
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

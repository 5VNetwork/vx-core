// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dispatcher

import (
	"context"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/dispatcher/variants"
	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/mux"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/retry"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/uot"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/hysteria2"
	vless_out "github.com/5vnetwork/vx-core/proxy/vless/outbound"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Dispatcher struct {
	BeforeHandlerSelectionHooks []BeforeHandlerSelectionHook
	Router                      i.Router
	OnHandlerSelectedHooks      []AfterHandlerSelectedHook
	SessionStartHooks           []SessionStartHook
	SessionEndHooks             []SessionEndHook
	TimeoutSetting              i.TimeoutSetting
	OnFallbacks                 []OnFallback
	FallbackTimeout             time.Duration

	SessionStats        bool
	RewriteIpv6ToDomain bool

	Flows       atomic.Int32
	PacketConns atomic.Int32

	observerLock          sync.Mutex
	HandlerErrorObservers []i.HandlerErrorObserver
}

type OnFallback interface {
	OnFallback(info *session.Info, previousTag, tag string)
}

func (p *Dispatcher) AddHandlerErrorObserver(observer i.HandlerErrorObserver) {
	p.observerLock.Lock()
	defer p.observerLock.Unlock()
	p.HandlerErrorObservers = append(p.HandlerErrorObservers, observer)
}

func (p *Dispatcher) RemoveHandlerErrorObserver(observer i.HandlerErrorObserver) {
	p.observerLock.Lock()
	defer p.observerLock.Unlock()
	for i, o := range p.HandlerErrorObservers {
		if o == observer {
			p.HandlerErrorObservers = append(p.HandlerErrorObservers[:i], p.HandlerErrorObservers[i+1:]...)
			break
		}
	}
}

type GetCounters interface {
	GetCounters(ctx context.Context, info *session.Info, handler i.Outbound) (session.UpCounters, session.DownCounters)
}

type Fallback interface {
	Fallback(ctx context.Context, info *session.Info,
		rw buf.ReaderWriter, handler i.Outbound, err error) error
}

type BeforeHandlerSelectionHook interface {
	BeforeHandlerSelection(ctx context.Context, info *session.Info,
		rw any) (context.Context, any, error)
}

type AfterHandlerSelectedHook interface {
	AfterHandlerSelection(ctx context.Context, info *session.Info, rw any,
		handler i.Outbound) (context.Context, any, error)
}

type SessionStartHook interface {
	SessionStart(ctx context.Context, info *session.Info)
}

type SessionEndHook interface {
	// should return quickly
	FlowSessionEnd(ctx context.Context, info *session.Info, err error)
	PacketConnSessionEnd(ctx context.Context, info *session.Info, err error)
}

type SessionErrorLogger interface {
	LogSessionError(info *session.Info, err error)
}

func (d *Dispatcher) AddBeforeHandlerSelectionHook(hook BeforeHandlerSelectionHook) {
	d.BeforeHandlerSelectionHooks = append(d.BeforeHandlerSelectionHooks, hook)
}

func (d *Dispatcher) AddAfterHandlerSelectionHook(hook AfterHandlerSelectedHook) {
	d.OnHandlerSelectedHooks = append(d.OnHandlerSelectedHooks, hook)
}

func (d *Dispatcher) AddSessionEndHook(hook SessionEndHook) {
	d.SessionEndHooks = append(d.SessionEndHooks, hook)
}

func (d *Dispatcher) AddSessionStartHook(hook SessionStartHook) {
	d.SessionStartHooks = append(d.SessionStartHooks, hook)
}

func (d *Dispatcher) AddOnFallback(fallback OnFallback) {
	d.OnFallbacks = append(d.OnFallbacks, fallback)
}

func (d *Dispatcher) HandleFlow(ctx context.Context, dst net.Destination,
	rw buf.ReaderWriter) error {
	if dst.Address == mux.MuxCoolAddressDst {
		return mux.Serve(ctx, rw, d)
	}
	if dst.Address == uot.Addr {
		return uot.Serve(ctx, rw, d)
	}

	info := infoFromContext(ctx, dst)
	ctx = session.ContextWithInfo(ctx, info)
	defer info.CleanUp()

	for _, hook := range d.SessionStartHooks {
		hook.SessionStart(ctx, info)
	}

	d.Flows.Add(1)
	defer d.Flows.Add(-1)

	var err error
	defer func() {
		for _, hook := range d.SessionEndHooks {
			hook.FlowSessionEnd(ctx, info, err)
		}
	}()

	var rw0 any
	for _, pre := range d.BeforeHandlerSelectionHooks {
		ctx, rw0, err = pre.BeforeHandlerSelection(ctx, info, rw)
		rw = rw0.(buf.ReaderWriter)
		if err != nil {
			return err
		}
	}

	var handler i.Outbound
	var retries []i.Fallback
	rw0, handler, retries, err = d.Router.PickHandlerWithData(ctx, info, rw)
	rw = rw0.(buf.ReaderWriter)
	if err != nil {
		return err
	}

	for _, hook := range d.OnHandlerSelectedHooks {
		ctx, rw0, err = hook.AfterHandlerSelection(ctx, info, rw, handler)
		rw = rw0.(buf.ReaderWriter)
		if err != nil {
			return err
		}
	}

	rw = d.sessionStats(info, rw).(buf.ReaderWriter)

	// if dialer, ok := handler.(i.ProxyDialer); ok {
	// 	var conn i.FlowConn
	// 	if ddlRw, ok := rw.(i.DeadlineRW); ok &&
	// 		ddlRw.SetReadDeadline(time.Now().Add(time.Millisecond*10)) == nil {
	// 		var mb buf.MultiBuffer
	// 		mb, err = ddlRw.ReadMultiBuffer()
	// 		ddlRw.SetReadDeadline(time.Time{})
	// 		if err == nil {
	// 			conn, err = dialer.ProxyDial(ctx, info.Target, mb)
	// 		}
	// 	} else {
	// 		conn, err = dialer.ProxyDial(ctx, info.Target, nil)
	// 	}
	// 	if conn != nil {
	// 		defer conn.Close()
	// 		err = d.Relay(ctx, info, rw, conn, handler)
	// 		if err != nil {
	// 			d.onHandlerError(ctx, info, handler.Tag(), err)
	// 		}
	// 	}
	// } else {
	// }

loop:
	for {
		retryable := len(retries) > 0
		if retryable {
			var mb buf.MultiBuffer
			var fallbackable bool
			fallbackable, mb, err = d.cacheHandle(ctx, info, rw, handler)
			if err != nil {
				defer buf.ReleaseMulti(mb)
				if fallbackable {
					log.Ctx(ctx).Debug().Err(err).Str("tag", handler.Tag()).
						Int32("cacheSize", mb.Len()).Uint64("upCounter", info.SessionUpCounter.Load()).
						Uint64("downCounter", info.SessionDownCounter.Load()).
						Msg("cacheHandle err and fallbackable")
					// try to find next handler
					for len(retries) > 0 {
						nextHandler, final := retries[0].GetHandler(ctx, info, err)
						retries = retries[1:]
						if nextHandler != nil {
							if final {
								retries = retries[:0]
							}
							for _, fallback := range d.OnFallbacks {
								fallback.OnFallback(info, handler.Tag(), nextHandler.Tag())
							}
							handler = nextHandler
							rw = buf.NewSecondDdl(rw.(buf.DdlReaderWriter), mb)
							// continue the outer loop. i.e. try next handler
							continue loop
						}
					}
					// no next handler is available. Return error.
					return err
				} else {
					// unfallbackable. Return error.
					return err
				}
			} else {
				// nothing wrong
				return nil
			}
		} else {
			err = handler.HandleFlow(ctx, d.getDst(ctx, info, handler), rw)
			if err != nil {
				d.onHandlerError(ctx, info, handler.Tag(), err)
			}
			return err
		}
	}
}

var errFallbackTimeout = errors.New("fallback timeout")

func (d *Dispatcher) cacheHandle(ctx context.Context, info *session.Info,
	rw buf.ReaderWriter, handler i.Outbound) (fallbackable bool, allRequestData buf.MultiBuffer, err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	cacheRw := &cacheReaderWriter{
		DdlReaderWriter:  rw.(buf.DdlReaderWriter),
		maximumCacheSize: 1024 * 16,
		ctx:              ctx,
	}
	if d.FallbackTimeout > 0 {
		t := time.AfterFunc(d.FallbackTimeout, func() {
			cacheRw.lock.Lock()
			defer cacheRw.lock.Unlock()
			if cacheRw.stopCaching || cacheRw.done {
				return
			}
			log.Ctx(ctx).Debug().Msg("fallback timeout")
			cancel(errFallbackTimeout)
		})
		defer t.Stop()
	}

	err = handler.HandleFlow(ctx, d.getDst(ctx, info, handler), cacheRw)
	if err != nil {
		d.onHandlerError(ctx, info, handler.Tag(), err)

		mb, err1 := cacheRw.Done(ctx)
		if err1 != nil {
			log.Ctx(ctx).Debug().Err(err1).Msg("unfallbackable due to")
			return false, nil, err
		}
		rw.(buf.DdlReaderWriter).SetReadDeadline(time.Time{})
		// this means the problem occur on the left. Unfallbackable.
		if (errors.Is(err, errors.LeftToRightError{}) && errors.Is(err, buf.ReadError{})) ||
			(errors.Is(err, errors.RightToLeftError{}) && buf.IsWriteError(err)) ||
			(ctx.Err() != nil && context.Cause(ctx) != errFallbackTimeout) {
			buf.ReleaseMulti(mb)
			return false, nil, err
		}
		return true, mb, err
	}

	return false, nil, nil
}

func (d *Dispatcher) HandlePacketConn(ctx context.Context, dst net.Destination, pc udp.PacketReaderWriter) error {
	info := infoFromContext(ctx, dst)
	ctx = session.ContextWithInfo(ctx, info)
	defer info.CleanUp()

	for _, hook := range d.SessionStartHooks {
		hook.SessionStart(ctx, info)
	}

	d.PacketConns.Add(1)
	defer d.PacketConns.Add(-1)

	var err error
	defer func() {
		for _, hook := range d.SessionEndHooks {
			hook.PacketConnSessionEnd(ctx, info, err)
		}
	}()

	var pc0 any
	for _, pre := range d.BeforeHandlerSelectionHooks {
		ctx, pc0, err = pre.BeforeHandlerSelection(ctx, info, pc)
		pc = pc0.(udp.PacketReaderWriter)
		if err != nil {
			return err
		}
	}

	// var fallbackers []Fallbacker
	var handler i.Outbound
	pc0, handler, _, err = d.Router.PickHandlerWithData(ctx, info, pc)
	pc = pc0.(udp.PacketReaderWriter)
	if err != nil {
		return err
	}

	for _, hook := range d.OnHandlerSelectedHooks {
		ctx, pc0, err = hook.AfterHandlerSelection(ctx, info, pc, handler)
		pc = pc0.(udp.PacketReaderWriter)
		if err != nil {
			return err
		}
	}
	pc = d.sessionStats(info, pc).(udp.PacketReaderWriter)

	// if listener, ok := handler.(i.ProxyPacketListener); ok {
	// 	var udpConn udp.UdpConn
	// 	udpConn, err = listener.ListenPacket(ctx, info.Target)
	// 	if err != nil {
	// 		p.logError(ctx, info, err)
	// 		return err
	// 	}
	// 	defer udpConn.Close()
	// 	err = helper.RelayUDPPacketConn(ctx, pc, udpConn)
	// } else {
	err = handler.HandlePacketConn(ctx, d.getDst(ctx, info, handler), pc)
	// }
	if err != nil {
		d.onHandlerError(ctx, info, handler.Tag(), err)
	}

	return err
}

func (p *Dispatcher) onHandlerError(ctx context.Context, info *session.Info, tag string, err error) {
	if tag == "dns" || tag == "direct" {
		return
	}
	if info.SessionDownCounter.Load() != 0 {
		return
	}
	if errors.Is(err, context.Canceled) {
		return
	}
	if errors.Is(err, io.EOF) {
		return
	}
	if errors.Is(err, hysteria2.ErrRejectQuic) {
		return
	}
	if errors.Is(err, vless_out.ErrRejectQuic) {
		return
	}
	var closeError *websocket.CloseError
	if errors.As(err, &closeError) && closeError.Code == websocket.CloseNormalClosure {
		return
	}

	p.observerLock.Lock()
	defer p.observerLock.Unlock()
	if len(p.HandlerErrorObservers) == 0 {
		return
	}

	log.Ctx(ctx).Debug().Str("tag", tag).Err(err).Msg("handler error")
	for _, observer := range p.HandlerErrorObservers {
		go observer.OnHandlerError(tag, err)
	}
}

type NullSessionErrorLogger struct{}

func (n *NullSessionErrorLogger) LogSessionError(info *session.Info, err error) {
}

// TODO: improve error handling
// TODO: a udp session should only be closed when idle

type OnHandlerErrorFunc func(tag string, err error)

func infoFromContext(ctx context.Context, dst net.Destination) *session.Info {
	info := session.Info{
		Target:    dst,
		StartTime: time.Now().Unix(),
	}
	info.InboundTag, _ = inbound.InboundTagFromContext(ctx)
	info.Source, _ = inbound.SrcFromContext(ctx)
	info.Gateway, _ = inbound.GatewayFromContext(ctx)
	info.User, _ = proxy.UserFromContext(ctx)
	id, _ := session.IDFromContext(ctx)
	info.ID = session.ID(id)
	return &info
}

func udpShouldReconnect(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, errors.ErrIdle) {
		return false
	}
	if errors.Is(err, vless_out.ErrRejectQuic) {
		return false
	}
	if errors.Is(err, retry.ErrRetryFailed) {
		return false
	}
	if errors.Is(err, outbound.ErrIpv6NotSupported) {
		return false
	}
	if errors.Is(err, errors.ErrClosed) {
		return false
	}
	if strings.Contains(err.Error(), "connection was refused") {
		return false
	}
	return true
}

func (d *Dispatcher) sessionStats(info *session.Info, rw any) any {
	if d.SessionStats {
		if r, ok := rw.(i.DeadlineRW); ok {
			rw = variants.NewStatsDeadlineRW(r,
				session.AtomicCounter{
					Counter: &info.SessionUpCounter,
				},
				session.AtomicCounter{
					Counter: &info.SessionDownCounter,
				}, nil)
		} else if r, ok := rw.(buf.ReaderWriter); ok {
			rw = variants.NewStatsReaderWriter(r,
				session.AtomicCounter{
					Counter: &info.SessionUpCounter,
				},
				session.AtomicCounter{
					Counter: &info.SessionDownCounter,
				}, nil)
		} else {
			rw = variants.NewStatsPacketConn(rw.(udp.PacketReaderWriter),
				session.AtomicCounter{
					Counter: &info.SessionUpCounter,
				},
				session.AtomicCounter{
					Counter: &info.SessionDownCounter,
				}, nil)
		}
	}
	return rw
}

func (d *Dispatcher) getDst(ctx context.Context, info *session.Info, handler i.Outbound) net.Destination {
	if d.RewriteIpv6ToDomain && info.Target.Address.Family().IsIP() && info.Target.Address.Family().IsIPv6() {
		if handlerSupport6, ok := handler.(i.HandlerWith6Info); ok && !handlerSupport6.Support6() && info.SniffedDomain != "" {
			log.Ctx(ctx).Debug().Str("handler", handler.Tag()).Str("dst", info.Target.String()).Msg("ipv6 not supported, replace it with the domain")
			return net.Destination{
				Address: net.DomainAddress(info.SniffedDomain),
				Port:    info.Target.Port,
				Network: info.Target.Network,
			}
		}
	}

	return info.Target
}

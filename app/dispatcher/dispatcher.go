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

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/mux"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/retry"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/units"
	"github.com/5vnetwork/vx-core/common/uot"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	vless_out "github.com/5vnetwork/vx-core/proxy/vless/outbound"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Dispatcher struct {
	BeforeHandlerSelectionHooks []BeforeHandlerSelectionHook
	Router                      i.Router
	OnHandlerSelectedHooks      []AfterHandlerSelectedHook
	SessionEndHooks             []SessionEndHook
	Fallback                    Fallback

	SessinoErrorLogger SessionErrorLogger
	Flows              atomic.Int32
	PacketConns        atomic.Int32

	observerLock          sync.Mutex
	HandlerErrorObservers []i.HandlerErrorObserver
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

	d.Flows.Add(1)
	defer d.Flows.Add(-1)

	var err error
	var rw0 any
	for _, pre := range d.BeforeHandlerSelectionHooks {
		ctx, rw0, err = pre.BeforeHandlerSelection(ctx, info, rw)
		rw = rw0.(buf.ReaderWriter)
		if err != nil {
			d.SessinoErrorLogger.LogSessionError(info, err)
			return err
		}
	}

	rw0, handler, err := d.Router.PickHandlerWithData(ctx, info, rw)
	rw = rw0.(buf.ReaderWriter)
	if err != nil {
		d.SessinoErrorLogger.LogSessionError(info, err)
		return err
	}

	for _, hook := range d.OnHandlerSelectedHooks {
		ctx, rw0, err = hook.AfterHandlerSelection(ctx, info, rw, handler)
		rw = rw0.(buf.ReaderWriter)
		if err != nil {
			d.SessinoErrorLogger.LogSessionError(info, err)
			return err
		}
	}

	err = handler.HandleFlow(ctx, info.Target, rw)
	if err != nil {
		d.onHandlerError(ctx, info, handler.Tag(), err)
		if d.Fallback != nil {
			err = d.Fallback.Fallback(ctx, info, rw, handler, err)
		}
	}
	if err != nil && d.SessinoErrorLogger != nil {
		d.SessinoErrorLogger.LogSessionError(info, err)
	}

	for _, hook := range d.SessionEndHooks {
		hook.FlowSessionEnd(ctx, info, err)
	}

	return err
}

func (d *Dispatcher) HandlePacketConn(ctx context.Context, dst net.Destination, pc udp.PacketReaderWriter) error {
	info := infoFromContext(ctx, dst)
	ctx = session.ContextWithInfo(ctx, info)
	defer info.CleanUp()

	d.PacketConns.Add(1)
	defer d.PacketConns.Add(-1)

	var err error
	var pc0 any
	for _, pre := range d.BeforeHandlerSelectionHooks {
		ctx, pc0, err = pre.BeforeHandlerSelection(ctx, info, pc)
		pc = pc0.(udp.PacketReaderWriter)
		if err != nil {
			d.SessinoErrorLogger.LogSessionError(info, err)
			return err
		}
	}

	pc0, handler, err := d.Router.PickHandlerWithData(ctx, info, pc)
	pc = pc0.(udp.PacketReaderWriter)
	if err != nil {
		d.SessinoErrorLogger.LogSessionError(info, err)
		return err
	}

	for _, hook := range d.OnHandlerSelectedHooks {
		ctx, pc0, err = hook.AfterHandlerSelection(ctx, info, pc, handler)
		pc = pc0.(udp.PacketReaderWriter)
		if err != nil {
			d.SessinoErrorLogger.LogSessionError(info, err)
			return err
		}
	}

	err = handler.HandlePacketConn(ctx, info.Target, pc)
	if err != nil {
		d.onHandlerError(ctx, info, handler.Tag(), err)
		d.SessinoErrorLogger.LogSessionError(info, err)
	}

	for _, hook := range d.SessionEndHooks {
		hook.PacketConnSessionEnd(ctx, info, err)
	}

	return err
}

func (p *Dispatcher) onHandlerError(ctx context.Context, info *session.Info, tag string, err error) {
	if tag == "dns" || tag == "direct" {
		return
	}
	if errors.Is(err, context.Canceled) {
		return
	}
	var closeError *websocket.CloseError
	if errors.As(err, &closeError) && closeError.Code == websocket.CloseNormalClosure {
		return
	}
	if errors.Is(err, io.EOF) {
		return
	}
	if info.SessionDownCounter.Load() != 0 {
		return
	}

	p.observerLock.Lock()
	defer p.observerLock.Unlock()
	if len(p.HandlerErrorObservers) == 0 {
		return
	}

	errStr := err.Error()
	// this error occurs if the src closes the connection, x continues to write response data, and src send rst.
	if strings.Contains(errStr, "endpoint is closed for send") {
		return
	}
	if strings.Contains(errStr, "An established connection was aborted by the software in your host machine.") {
		return
	}
	if strings.Contains(errStr, "write: broken pipe") {
		return
	}
	if strings.Contains(errStr, "connection reset by peer") {
		return
	}
	if strings.Contains(errStr, "reject quic over hysteria2") {
		return
	}
	if strings.Contains(errStr, "XTLS rejected QUIC traffic") {
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

func shouldOverride(info *session.Info, domainOverride []string, fakeIPNotFound bool) bool {
	if info.SniffedDomain == "" {
		return false
	}
	if fakeIPNotFound {
		return true
	}
	protocolString := info.Protocol
	if protocolString == "" {
		return false
	}
	for _, p := range domainOverride {
		if strings.HasPrefix(protocolString, p) || strings.HasSuffix(protocolString, p) {
			return true
		}
	}
	return false
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

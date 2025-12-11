// Package vless contains the implementation of VLess protocol and transportation.
//
// VLess contains both inbound and outbound connections. VLess inbound is usually used on servers
// together with 'freedom' to talk to final destination, while VLess outbound is usually used on
// clients with 'socks' for proxying.
package vless

import (
	"context"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal"
)

const (
	XRV = "xtls-rprx-vision"
)

type InboundInfo struct {
	CanSpliceCopy int
	Conn          net.Conn
	Timer         *signal.ActivityChecker
	UpCounter     session.UpCounters
	DownCounter   session.DownCounters
}

var ContextKeyInbound = 0

func InboundFromContext(ctx context.Context) *InboundInfo {
	ib, _ := ctx.Value(ContextKeyInbound).(*InboundInfo)
	return ib
}

func WithInbound(ctx context.Context, ib *InboundInfo) context.Context {
	return context.WithValue(ctx, ContextKeyInbound, ib)
}

type OutboundInfo struct {
	Target net.Destination
	// 1 yes, 2 maybe, we'll see, 3 no
	CanSpliceCopy int
}

var ContextKeyOutbound = 1

func OutboundsFromContext(ctx context.Context) []*OutboundInfo {
	ob, _ := ctx.Value(ContextKeyOutbound).([]*OutboundInfo)
	return ob
}

func WithOutbounds(ctx context.Context, ob []*OutboundInfo) context.Context {
	return context.WithValue(ctx, ContextKeyOutbound, ob)
}

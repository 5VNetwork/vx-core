package inbound

import (
	"context"
	"math/rand/v2"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/rs/zerolog/log"
)

type InboundCtxKey int

const (
	InboundTagKey InboundCtxKey = iota
	SrcKey
	GatewayKey
	IDKey
	RawConnKey
)

func ContextWithID(ctx context.Context, id uint32) context.Context {
	return context.WithValue(ctx, IDKey, id)
}
func IDFromContext(ctx context.Context) (uint32, bool) {
	id, ok := ctx.Value(IDKey).(uint32)
	return id, ok
}

func ContextWithInboundTag(ctx context.Context, tag string) context.Context {
	return context.WithValue(ctx, InboundTagKey, tag)
}
func InboundTagFromContext(ctx context.Context) (string, bool) {
	tag, ok := ctx.Value(InboundTagKey).(string)
	return tag, ok
}
func ContextWithSrc(ctx context.Context, src net.Destination) context.Context {
	return context.WithValue(ctx, SrcKey, src)
}
func SrcFromContext(ctx context.Context) (net.Destination, bool) {
	src, ok := ctx.Value(SrcKey).(net.Destination)
	return src, ok
}
func ContextWithGateway(ctx context.Context, gateway net.Destination) context.Context {
	return context.WithValue(ctx, GatewayKey, gateway)
}
func GatewayFromContext(ctx context.Context) (net.Destination, bool) {
	gateway, ok := ctx.Value(GatewayKey).(net.Destination)
	return gateway, ok
}

func GetCtx(ctx context.Context, src, gateway net.Destination, tag string) (context.Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(ctx)
	id := rand.Uint32()
	ctx = log.With().Uint32("sid", id).Logger().WithContext(ctx)
	ctx = ContextWithInboundTag(ctx, tag)
	ctx = ContextWithSrc(ctx, src)
	ctx = ContextWithGateway(ctx, gateway)
	ctx = ContextWithID(ctx, id)
	log.Ctx(ctx).Debug().Str("tag", tag).Str("network", src.Network.SystemString()).
		Str("src", src.String()).Str("gateway", gateway.String()).Msg("new session")
	return ctx, cancel
}

// only add conn to ctx when it is a tcp conn
func ContextWithRawConn(ctx context.Context, conn net.Conn) context.Context {
	if _, ok := conn.(*net.TCPConn); ok {
		return context.WithValue(ctx, RawConnKey, conn)
	} else {
		return ctx
	}
}
func RawConnFromContext(ctx context.Context) (net.Conn, bool) {
	conn, ok := ctx.Value(RawConnKey).(net.Conn)
	return conn, ok
}

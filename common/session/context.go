package session

import "context"

type sessionKey int

const (
	idKey sessionKey = iota
	srcKey
	inboundTagKey
	gatewayKey
	targetKey
	udpUuidKey
	infoKey
	sockoptSessionKey
	userKey
	bufferSizeKey
	networkKey
	viaKey
	dnsResolverKey
)

func ContextWithInfo(ctx context.Context, info *Info) context.Context {
	return context.WithValue(ctx, infoKey, info)
}
func InfoFromContext(ctx context.Context) *Info {
	if ctx.Value(infoKey) == nil {
		return nil
	}
	return ctx.Value(infoKey).(*Info)
}

// // id
// func ContextWithId(ctx context.Context, id ID) context.Context {
// 	return context.WithValue(ctx, idKey, id)
// }
// func IDFromContext(ctx context.Context) ID {
// 	return ctx.Value(infoKey).(ID)
// }

// // src
// func ContextWithSource(ctx context.Context, src net.Tuple3) context.Context {
// 	return context.WithValue(ctx, srcKey, src)
// }
// func SourceFromContext(ctx context.Context) net.Tuple3 {
// 	return ctx.Value(srcKey).(net.Tuple3)
// }

// // inbound tag
// func ContextWithInboundTag(ctx context.Context, tag string) context.Context {
// 	return context.WithValue(ctx, inboundTagKey, tag)
// }
// func InboundTagFromContext(ctx context.Context) string {
// 	return ctx.Value(inboundTagKey).(string)
// }

// // gateway
// func ContextWithGateway(ctx context.Context, gateway net.Tuple3) context.Context {
// 	return context.WithValue(ctx, gatewayKey, gateway)
// }
// func GatewayFromContext(ctx context.Context) net.Tuple3 {
// 	return ctx.Value(gatewayKey).(net.Tuple3)
// }

// // target
// func ContextWithTarget(ctx context.Context, target net.Tuple3) context.Context {
// 	return context.WithValue(ctx, targetKey, &target)
// }
// func TargetFromContext(ctx context.Context) *net.Tuple3 {
// 	if ctx.Value(targetKey) == nil {
// 		return nil
// 	}
// 	return ctx.Value(targetKey).(*net.Tuple3)
// }

// // udpUuid
// func ContextWithUdpUuid(ctx context.Context, udpUuid *uuid.UUID) context.Context {
// 	return context.WithValue(ctx, udpUuidKey, udpUuid)
// }
// func UdpUuidFromContext(ctx context.Context) *uuid.UUID {
// 	if ctx.Value(udpUuidKey) == nil {
// 		return nil
// 	} else {
// 		return ctx.Value(udpUuidKey).(*uuid.UUID)
// 	}
// }

// // ContextWithSockopt returns a new context with Socket configs included
// func ContextWithSockopt(ctx context.Context, s *Sockopt) context.Context {
// 	return context.WithValue(ctx, sockoptSessionKey, s)
// }

// // SockoptFromContext returns Socket configs in this context, or nil if not contained.
func SockoptFromContext(ctx context.Context) *Sockopt {
	info := InfoFromContext(ctx)
	if info == nil {
		return nil
	}
	return info.Sockopt
}

// // user
// func ContextWithUser(ctx context.Context, user user.User) context.Context {
// 	return context.WithValue(ctx, userKey, user)
// }
// func UserFromContext(ctx context.Context) user.User {
// 	return ctx.Value(userKey).(user.User)
// }

// // buffer size
// func ContextWithBufferSize(ctx context.Context, size int32) context.Context {
// 	return context.WithValue(ctx, bufferSizeKey, size)
// }
// func BufferSizeFromContext(ctx context.Context) int32 {
// 	return ctx.Value(bufferSizeKey).(int32)
// }

// // network
// func NetworkFromContext(ctx context.Context) net.Network {
// 	if t := TargetFromContext(ctx); t != nil {
// 		return t.Network
// 	}
// 	return SourceFromContext(ctx).Network
// }

// // via
// func ContextWithVia(ctx context.Context, via net.Address) context.Context {
// 	return context.WithValue(ctx, viaKey, via)
// }
// func ViaFromContext(ctx context.Context) net.Address {
// 	if ctx.Value(viaKey) == nil {
// 		return nil
// 	}
// 	return ctx.Value(viaKey).(net.Address)
// }

// dns resolver
// func ContextWithResolver(ctx context.Context, resolver func(ctx context.Context, domain string) net.Address) context.Context {
// 	return context.WithValue(ctx, dnsResolverKey, resolver)
// }
// func ResolverFromContext(ctx context.Context) func(ctx context.Context, domain string) (net.Address, error) {
// 	return ctx.Value(dnsResolverKey).(func(ctx context.Context, domain string) (net.Address, error))
// }

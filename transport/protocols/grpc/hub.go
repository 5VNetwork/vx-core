package grpc

import (
	"context"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/protocols/grpc/encoding"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Listener struct {
	encoding.UnimplementedGRPCServiceServer
	ctx     context.Context
	handler func(net.Conn)
	local   net.Addr
	config  *GrpcConfig

	s *grpc.Server
}

func (l Listener) Tun(server encoding.GRPCService_TunServer) error {
	tunCtx, cancel := context.WithCancel(l.ctx)
	l.handler(encoding.NewHunkConn(server, cancel))
	<-tunCtx.Done()
	return nil
}

func (l Listener) TunMulti(server encoding.GRPCService_TunMultiServer) error {
	tunCtx, cancel := context.WithCancel(l.ctx)
	l.handler(encoding.NewMultiHunkConn(server, cancel))
	<-tunCtx.Done()
	return nil
}

func (l Listener) Close() error {
	l.s.Stop()
	return nil
}

func (l Listener) Addr() net.Addr {
	return l.local
}

func Listen(ctx context.Context, addr net.Destination,
	grpcSettings *GrpcConfig, li i.Listener,
	handler func(net.Conn)) (*Listener, error) {

	listener := &Listener{
		handler: handler,
		config:  grpcSettings,
		local:   addr.Addr(),
	}

	listener.ctx = ctx

	var options []grpc.ServerOption
	var s *grpc.Server
	if grpcSettings.IdleTimeout > 0 || grpcSettings.HealthCheckTimeout > 0 {
		options = append(options, grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    time.Second * time.Duration(grpcSettings.IdleTimeout),
			Timeout: time.Second * time.Duration(grpcSettings.HealthCheckTimeout),
		}))
	}

	s = grpc.NewServer(options...)
	listener.s = s

	go func() {
		streamListener, err := li.Listen(ctx, addr.Addr())
		if err != nil {
			log.Error().Err(err).
				Str("address", addr.String()).Msg("failed to listen")
			return
		}
		log.Debug().Msg("gRPC listen for service name `" +
			grpcSettings.getServiceName() + "` tun `" + grpcSettings.getTunStreamName() +
			"` multi tun `" + grpcSettings.getTunMultiStreamName() + "`")
		encoding.RegisterGRPCServiceServerX(s, listener, grpcSettings.getServiceName(),
			grpcSettings.getTunStreamName(), grpcSettings.getTunMultiStreamName())

		if err = s.Serve(streamListener); err != nil {
			log.Error().Err(err).Msg("failed to serve")
		}
	}()

	return listener, nil
}

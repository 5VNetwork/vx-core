package grpc

import (
	"context"
	gonet "net"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/protocols/grpc/encoding"
	"github.com/5vnetwork/vx-core/transport/security"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type connectionWrapper struct {
	net.Conn
	dialer     *Dialer
	clientConn *clientConnWrapper
}

func (cw *connectionWrapper) Close() error {
	err := cw.Conn.Close()
	cw.dialer.decrementConnCount(cw.clientConn)
	return err
}

type Dialer struct {
	config       *GrpcConfig
	engine       security.Engine
	socketConfig i.Dialer

	lock sync.Mutex
	// when default interface changed, this destToConn will be replaced.
	destToConn map[net.Destination]*clientConnWrapper
}

type clientConnWrapper struct {
	*grpc.ClientConn
	dest       net.Destination
	connCounts int
	timer      *time.Timer
}

func (d *Dialer) decrementConnCount(c *clientConnWrapper) {
	d.lock.Lock()
	defer d.lock.Unlock()

	c.connCounts--
	if c.connCounts <= 0 {
		if c.timer == nil {
			c.timer = time.AfterFunc(time.Second*5, func() {
				d.lock.Lock()
				defer d.lock.Unlock()
				if c.connCounts <= 0 {
					// log.Debug().Int("remaining", len(d.destToConn)).Msg("grpc close client conn")
					c.ClientConn.Close()
					delete(d.destToConn, c.dest)
				}
			})
		}
	}
}

func (d *Dialer) closeClientConn(c *clientConnWrapper) {
	d.lock.Lock()
	defer d.lock.Unlock()
	c.ClientConn.Close()
	delete(d.destToConn, c.dest)
}

func NewGrpcDialer(config *GrpcConfig, engine security.Engine, socketConfig i.Dialer) *Dialer {
	return &Dialer{
		config:       config,
		engine:       engine,
		socketConfig: socketConfig,
	}
}

func (d *Dialer) Dial(ctx context.Context, dest net.Destination) (net.Conn, error) {
	conn, err := d.dialgRPC(ctx, dest)
	if err != nil {
		return nil, errors.New("failed to dial gRPC").Base(err)
	}
	return conn, nil
}

func (d *Dialer) dialgRPC(ctx context.Context, dest net.Destination) (net.Conn, error) {
	conn, err := d.getGrpcClient(ctx, dest)
	if err != nil {
		return nil, errors.New("Cannot dial gRPC").Base(err)
	}
	client := encoding.NewGRPCServiceClient(conn)

	if d.config.MultiMode {
		grpcService, err := client.(encoding.GRPCServiceClientX).TunMultiCustomName(ctx, getServiceName(d.config), getTunMultiStreamName(d.config))
		if err != nil {
			d.decrementConnCount(conn)
			return nil, errors.New("Cannot dial gRPC").Base(err)
		}
		return &connectionWrapper{
			Conn:       encoding.NewMultiHunkConn(grpcService, nil),
			dialer:     d,
			clientConn: conn,
		}, nil
	}

	grpcService, err := client.(encoding.GRPCServiceClientX).TunCustomName(ctx,
		getServiceName(d.config), getTunStreamName(d.config))
	if err != nil {
		d.decrementConnCount(conn)
		return nil, errors.New("Cannot dial gRPC").Base(err)
	}

	return &connectionWrapper{
		Conn:       encoding.NewHunkConn(grpcService, nil),
		dialer:     d,
		clientConn: conn,
	}, nil
}

func (d *Dialer) incrementConnCount(c *clientConnWrapper) {
	c.connCounts++
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
}

func (d *Dialer) getGrpcClient(ctx context.Context, dest net.Destination) (*clientConnWrapper, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.destToConn == nil {
		d.destToConn = make(map[net.Destination]*clientConnWrapper)
	}

	if client, found := d.destToConn[dest]; found {
		if client.GetState() != connectivity.Shutdown {
			d.incrementConnCount(client)
			return client, nil
		} else {
			client.ClientConn.Close()
			delete(d.destToConn, dest)
		}
	}

	connWrapper := &clientConnWrapper{
		dest: dest,
	}

	dialOptions := []grpc.DialOption{
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  500 * time.Millisecond,
				Multiplier: 1.5,
				Jitter:     0.2,
				MaxDelay:   19 * time.Second,
			},
			MinConnectTimeout: 5 * time.Second,
		}),
		grpc.WithContextDialer(func(gctx context.Context, s string) (gonet.Conn, error) {
			select {
			case <-gctx.Done():
				return nil, gctx.Err()
			default:
			}

			rawHost, rawPort, err := net.SplitHostPort(s)
			if err != nil {
				return nil, err
			}
			if len(rawPort) == 0 {
				rawPort = "443"
			}
			port, err := net.PortFromString(rawPort)
			if err != nil {
				return nil, err
			}
			address := net.ParseAddress(rawHost)

			c, err := d.socketConfig.Dial(gctx, net.TCPDestination(address, port))
			if err != nil {
				return nil, err
			}
			log.Debug().Str("laddr", c.LocalAddr().String()).Msg("grpc dial success")
			if d.engine == nil {
				return c, nil
			}
			return d.engine.GetClientConn(c,
				security.OptionWithDestination{Dest: dest})
		}),
	}

	dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))

	authority := ""
	if d.config.Authority != "" {
		authority = d.config.Authority
	}
	dialOptions = append(dialOptions, grpc.WithAuthority(authority))

	if d.config.IdleTimeout > 0 || d.config.HealthCheckTimeout > 0 || d.config.PermitWithoutStream {
		dialOptions = append(dialOptions, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * time.Duration(d.config.IdleTimeout),
			Timeout:             time.Second * time.Duration(d.config.HealthCheckTimeout),
			PermitWithoutStream: d.config.PermitWithoutStream,
		}))
	}

	if d.config.InitialWindowsSize > 0 {
		dialOptions = append(dialOptions, grpc.WithInitialWindowSize(d.config.InitialWindowsSize))
	}

	if d.config.UserAgent != "" {
		dialOptions = append(dialOptions, grpc.WithUserAgent(d.config.UserAgent))
	}

	var grpcDestHost string
	if dest.Address.Family().IsDomain() {
		grpcDestHost = dest.Address.Domain()
	} else {
		grpcDestHost = dest.Address.IP().String()
	}

	conn, err := grpc.Dial(
		gonet.JoinHostPort(grpcDestHost, dest.Port.String()),
		dialOptions...,
	)
	if err != nil {
		return nil, err
	}

	connWrapper.ClientConn = conn

	d.incrementConnCount(connWrapper)
	d.destToConn[dest] = connWrapper

	log.Debug().Msg("grpc new client conn")

	return connWrapper, nil
}

package transport

import (
	"context"
	"errors"
	"net"

	gotls "crypto/tls"

	net1 "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/5vnetwork/vx-core/transport/protocols/grpc"
	"github.com/5vnetwork/vx-core/transport/protocols/http"
	"github.com/5vnetwork/vx-core/transport/protocols/httpupgrade"
	"github.com/5vnetwork/vx-core/transport/protocols/kcp"
	"github.com/5vnetwork/vx-core/transport/protocols/splithttp"
	"github.com/5vnetwork/vx-core/transport/protocols/tcp"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	goreality "github.com/xtls/reality"
)

type Listener interface {
	Close() error
	Addr() net.Addr
}

func (config *Config) Listen(ctx context.Context, netAddr net.Addr) (net.Listener, error) {
	return Listen(net1.DestinationFromAddr(netAddr), config)
}

type ListenerAdapter struct {
	Listener
	channel chan net.Conn
	done    *done.Instance
}

func (l *ListenerAdapter) handleConn(conn net.Conn) {
	if l.done.Done() {
		conn.Close()
		return
	}
	select {
	case l.channel <- conn:
	default:
		conn.Close()
	}
}

func (l *ListenerAdapter) Accept() (net.Conn, error) {
	select {
	case conn := <-l.channel:
		return conn, nil
	case <-l.done.Wait():
		return nil, errors.New("listener closed")
	}
}

func (l *ListenerAdapter) Close() error {
	l.done.Close()
	err := l.Listener.Close()
	return err
}

// Keeps accepting and use h to handle the conn.
func Listen(addr net1.Destination, config *Config) (net.Listener, error) {
	if config == nil {
		config = &Config{}
	}
	if config.Protocol == nil {
		config.Protocol = &tcp.TcpConfig{}
	}

	listenerAdapter := &ListenerAdapter{
		channel: make(chan net.Conn, 128),
		done:    done.New(),
	}

	var err error
	switch Protocol := config.Protocol.(type) {
	case *tcp.TcpConfig:
		listenerImpl := &ListenerImpl{
			Socket:         config.Socket,
			SecurityConfig: config.Security,
		}
		listener, err := listenerImpl.Listen(context.Background(), addr.Addr())
		if err != nil {
			return nil, err
		}
		return listener, nil
	case *kcp.KcpConfig:
		if tls, ok := config.Security.(*tls.TlsConfig); ok {
			tlsConfig, err1 := tls.GetTLSConfig()
			if err1 != nil {
				return nil, err1
			}
			listenerAdapter.Listener, err = kcp.Listen(context.Background(), addr, Protocol,
				tlsConfig, config.Socket, listenerAdapter.handleConn)
		} else {
			listenerAdapter.Listener, err = kcp.Listen(context.Background(), addr, Protocol, nil,
				config.Socket, listenerAdapter.handleConn)
		}
		if err != nil {
			return nil, err
		}
		return listenerAdapter, nil
	case *websocket.WebsocketConfig:
		listenerAdapter.Listener, err = websocket.Listen(context.Background(), addr, Protocol, &ListenerImpl{
			Socket:         config.Socket,
			SecurityConfig: config.Security,
		}, listenerAdapter.handleConn)
		if err != nil {
			return nil, err
		}
		return listenerAdapter, nil
	case *httpupgrade.HttpUpgradeConfig:
		listenerAdapter.Listener, err = httpupgrade.Listen(context.Background(), addr, Protocol, &ListenerImpl{
			Socket:         config.Socket,
			SecurityConfig: config.Security,
		}, listenerAdapter.handleConn)
		if err != nil {
			return nil, err
		}
		return listenerAdapter, nil
	case *http.HttpConfig:
		listenerAdapter.Listener, err = http.Listen(context.Background(), addr, Protocol, &ListenerImpl{
			Socket:         config.Socket,
			SecurityConfig: config.Security,
			useH2:          true,
		}, listenerAdapter.handleConn)
		if err != nil {
			return nil, err
		}
		return listenerAdapter, nil
	case *grpc.GrpcConfig:
		listenerAdapter.Listener, err = grpc.Listen(context.Background(), addr, Protocol, &ListenerImpl{
			Socket:         config.Socket,
			SecurityConfig: config.Security,
			useH2:          true,
		}, listenerAdapter.handleConn)
		if err != nil {
			return nil, err
		}
		return listenerAdapter, nil
	case *splithttp.SplitHttpConfig:
		listenerAdapter.Listener, err = splithttp.ListenXH(context.Background(), addr, Protocol, &ListenerImpl{
			Socket:         config.Socket,
			SecurityConfig: config.Security,
		}, listenerAdapter.handleConn)
		if err != nil {
			return nil, err
		}
		return listenerAdapter, nil
	}
	return nil, errors.New("invalid transport config")
}

type ListenerImpl struct {
	Socket         *dlhelper.SocketSetting
	SecurityConfig interface{}
	useH2          bool
}

func (l *ListenerImpl) Listen(ctx context.Context, addr net.Addr) (net.Listener, error) {
	listener, err := l.Socket.Listen(ctx, addr)
	if err != nil {
		return nil, err
	}
	if t, ok := l.SecurityConfig.(*tls.TlsConfig); ok {
		var opts []tls.Option
		if l.useH2 {
			opts = append(opts, tls.WithNextProtocol([]string{"h2"}))
		} else {
			opts = append(opts, tls.WithNextProtocol([]string{"h2", "http/1.1"}))
		}
		tlsConfig, err := t.GetTLSConfig(opts...)
		if err != nil {
			return nil, err
		}
		return gotls.NewListener(listener, tlsConfig), nil
	} else if reality, ok := l.SecurityConfig.(*reality.RealityConfig); ok {
		return goreality.NewListener(listener, reality.GetREALITYConfig()), nil
	}
	return listener, nil
}

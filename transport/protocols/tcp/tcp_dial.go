package tcp

import (
	"context"
	"fmt"
	"net"

	net1 "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/headers"
	"github.com/5vnetwork/vx-core/transport/security"
)

type tcpDialer struct {
	config       *TcpConfig
	engine       security.Engine
	socketConfig i.Dialer
}

func NewTcpDialer(config *TcpConfig, engine security.Engine, socketConfig i.Dialer) *tcpDialer {
	return &tcpDialer{
		config:       config,
		engine:       engine,
		socketConfig: socketConfig,
	}
}

func (d *tcpDialer) Dial(ctx context.Context, dest net1.Destination) (net.Conn, error) {
	return Dial(ctx, dest, d.config, d.engine, d.socketConfig)
}

// Dial dials a new TCP connection to the given destination.
func Dial(ctx context.Context, dest net1.Destination, c *TcpConfig,
	securityConfig security.Engine, socketConfig i.Dialer) (net.Conn, error) {
	conn, err := socketConfig.Dial(ctx, dest)
	if err != nil {
		return nil, err
	}

	if securityConfig != nil {
		conn, err = securityConfig.GetClientConn(conn, security.OptionWithDestination{Dest: dest})
		if err != nil {
			return nil, err
		}
	}

	if c != nil && c.HeaderSettings != nil {
		c, err := serial.GetInstanceOf(c.HeaderSettings)
		if err != nil {
			return nil, err
		}
		auth, err := headers.CreateConnectionAuthenticator(c)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection authenticator: %w", err)
		}
		conn = auth.Client(conn)
	}

	return conn, nil
}

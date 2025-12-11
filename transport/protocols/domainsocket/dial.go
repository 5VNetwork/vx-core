//go:build !windows && !wasm && server
// +build !windows,!wasm,server

package domainsocket

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/transport/security"
)

type domainsocketDialer struct {
	config *DomainSocketConfig
	engine security.Engine
}

func NewDomainSocketDialer(config *DomainSocketConfig, engine security.Engine) *domainsocketDialer {
	return &domainsocketDialer{
		config: config,
		engine: engine,
	}
}

func Dial(ctx context.Context, dest net.Destination, config *DomainSocketConfig, securityConfig security.Engine) (net.Conn, error) {
	addr, err := config.GetUnixAddr()
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial domain socket: %w", err)
	}

	if securityConfig != nil {
		return securityConfig.GetClientConn(conn)
	}

	return conn, nil
}

package security

import (
	"crypto/tls"
	"net"

	net1 "github.com/5vnetwork/vx-core/common/net"
)

type Engine interface {
	GetTLSConfig(...Option) *tls.Config
	GetClientConn(net.Conn, ...Option) (net.Conn, error)
}

type Option interface {
	isSecurityOption()
}

type OptionWithALPN struct {
	ALPNs []string
}

func (a OptionWithALPN) isSecurityOption() {
}

type OptionWithDestination struct {
	Dest net1.Destination
}

func (a OptionWithDestination) isSecurityOption() {
}

type ConnectionApplicationProtocol interface {
	GetConnectionApplicationProtocol() (string, error)
}

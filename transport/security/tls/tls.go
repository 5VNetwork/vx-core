package tls

import (
	"crypto/tls"

	"github.com/5vnetwork/vx-core/common/net"
)

// a wrapper of tls.Conn
type Conn struct {
	*tls.Conn
}

func (c *Conn) GetConnectionApplicationProtocol() (string, error) {
	if err := c.Handshake(); err != nil { //in xray, this is omitted
		return "", err
	}
	return c.ConnectionState().NegotiatedProtocol, nil
}

func (c *Conn) HandshakeAddress() net.Address {
	if err := c.Handshake(); err != nil {
		return nil
	}
	state := c.ConnectionState()
	if state.ServerName == "" {
		return nil
	}
	return net.ParseAddress(state.ServerName)
}

type Option func(*tls.Config)

// if c.ServerName has been specified, dont apply dest as ServerName
func WithDestination(dest net.Destination) Option {
	return func(c *tls.Config) {
		if c.ServerName == "" {
			switch dest.Address.Family() {
			case net.AddressFamilyDomain:
				c.ServerName = dest.Address.Domain()
			case net.AddressFamilyIPv4, net.AddressFamilyIPv6:
				c.ServerName = dest.Address.IP().String()
			}
		}
	}
}

func WithNextProtocol(protos []string) Option {
	return func(c *tls.Config) {
		if len(c.NextProtos) == 0 {
			c.NextProtos = protos
		}
	}
}

//go:build server

package domainsocket

import (
	"errors"

	"github.com/5vnetwork/vx-core/common/net"
)

const (
	protocolName  = "domainsocket"
	sizeofSunPath = 108
)

func (c *DomainSocketConfig) GetUnixAddr() (*net.UnixAddr, error) {
	path := c.Path
	if path == "" {
		return nil, errors.New("empty domain socket path")
	}
	if c.Abstract && path[0] != '\x00' {
		path = "\x00" + path
	}
	if c.Abstract && c.Padding {
		raw := []byte(path)
		addr := make([]byte, sizeofSunPath)
		copy(addr, raw)
		path = string(addr)
	}
	return &net.UnixAddr{
		Name: path,
		Net:  "unix",
	}, nil
}

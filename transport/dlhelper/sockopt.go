package dlhelper

import "github.com/5vnetwork/vx-core/common/errors"

func isTCPSocket(network string) bool {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return true
	default:
		return false
	}
}

func isUDPSocket(network string) bool {
	switch network {
	case "udp", "udp4", "udp6":
		return true
	default:
		return false
	}
}

type Sockopt struct {
	// Mark of the socket connection.
	Mark           uint32
	InterfaceName4 string
	InterfaceName6 string
}

var ErrBindToDeviceNotFound = errors.New("bind to device not found")

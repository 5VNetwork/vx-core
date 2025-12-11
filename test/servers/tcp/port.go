package tcp

import "github.com/5vnetwork/vx-core/common/net"

// PickPort returns an unused TCP port of the system.
func PickPort() net.Port {
	listener := pickPort()
	addr := listener.Addr().(*net.TCPAddr)
	listener.Close()
	return net.Port(addr.Port)
}

func pickPort() net.Listener {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		listener = pickPort()
	}
	return listener
}

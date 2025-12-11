package dlhelper

import (
	"context"

	"github.com/5vnetwork/vx-core/common/net"
)

var effectiveListener = DefaultListener{}

// ListenSystem listens on a local address for incoming TCP connections.
func ListenSystem(ctx context.Context, addr net.Addr, sockopt *SocketSetting) (net.Listener, error) {
	return effectiveListener.Listen(ctx, addr, sockopt)
}

// ListenSystemPacket listens on a local address for incoming UDP connections.
func ListenSystemPacket(ctx context.Context, network, address string, sockopt *SocketSetting) (net.PacketConn, error) {
	return effectiveListener.ListenPacket(ctx, network, address, sockopt)
}

var effectiveSystemDialer systemDialer = &DefaultSystemDialer{}

// DialSystemConn calls system dialer to create a network connection.
func DialSystemConn(ctx context.Context, dest net.Destination, sockopt *SocketSetting) (net.Conn, error) {
	return effectiveSystemDialer.DialConn(ctx, dest, sockopt)
}

// systemDialer is to create network connections. It returns raw network conn,
// which can be wrapped by security.
type systemDialer interface {
	DialConn(ctx context.Context, destination net.Destination, sockopt *SocketSetting) (net.Conn, error)
}

// TODO DialPacketConn and ListenPacket are same, need to merge them
type systemListener interface {
	Listen(ctx context.Context, addr net.Addr, sockopt *SocketSetting) (net.Listener, error)
	ListenPacket(ctx context.Context, network, address string, sockopt *SocketSetting) (net.PacketConn, error)
}

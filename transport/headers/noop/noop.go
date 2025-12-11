package noop

import (
	"context"
	"net"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/creator"
)

type Header struct{}

func (Header) Size() int32 {
	return 0
}

// Serialize implements PacketHeader.
func (Header) Serialize([]byte) {}

func NewHeader(context.Context, interface{}) (interface{}, error) {
	return Header{}, nil
}

type ConnectionHeader struct{}

func (ConnectionHeader) Client(conn net.Conn) net.Conn {
	return conn
}

func (ConnectionHeader) Server(conn net.Conn) net.Conn {
	return conn
}

func NewConnectionHeader(context.Context, interface{}) (interface{}, error) {
	return ConnectionHeader{}, nil
}

func init() {
	common.Must(creator.RegisterConfig((*Config)(nil), NewHeader))
	common.Must(creator.RegisterConfig((*ConnectionConfig)(nil), NewConnectionHeader))
}

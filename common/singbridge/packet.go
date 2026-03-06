package singbridge

import (
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	M "github.com/sagernet/sing/common/metadata"

	B "github.com/sagernet/sing/common/buf"
)

type PacketConnWrapper struct {
	udp.PacketReaderWriter
}

func (w *PacketConnWrapper) LocalAddr() net.Addr {
	return &net.UDPAddr{
		IP:   net.IP{},
		Port: 0,
	}
}

func (w *PacketConnWrapper) SetDeadline(t time.Time) error {
	if deadline, ok := w.PacketReaderWriter.(udp.DdlPacketReaderWriter); ok {
		return deadline.SetDeadline(t)
	}
	return nil
}

func (w *PacketConnWrapper) SetReadDeadline(t time.Time) error {
	if deadline, ok := w.PacketReaderWriter.(udp.DdlPacketReaderWriter); ok {
		return deadline.SetReadDeadline(t)
	}
	return nil
}

func (w *PacketConnWrapper) SetWriteDeadline(t time.Time) error {
	if deadline, ok := w.PacketReaderWriter.(udp.DdlPacketReaderWriter); ok {
		return deadline.SetWriteDeadline(t)
	}
	return nil
}

// This ReadPacket implemented a timeout to avoid goroutine leak like PipeConnWrapper.Read()
// as a temporarily solution
func (w *PacketConnWrapper) ReadPacket(buffer *B.Buffer) (M.Socksaddr, error) {
	packet, err := w.PacketReaderWriter.ReadPacket()
	if err != nil {
		return M.Socksaddr{}, err
	}
	defer packet.Release()
	buffer.Write(packet.Payload.Bytes())
	return ToSocksaddr(packet.Target), nil
}

func (w *PacketConnWrapper) WritePacket(buffer *B.Buffer, destination M.Socksaddr) error {
	vBuf := buf.New()
	vBuf.Write(buffer.Bytes())
	endpoint := ToDestination(destination, net.Network_UDP)
	packet := &udp.Packet{
		Source:  endpoint,
		Payload: vBuf,
	}
	return w.PacketReaderWriter.WritePacket(packet)
}

func (w *PacketConnWrapper) Close() error {
	return nil
}

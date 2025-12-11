package mocks

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/common/buf"
	net "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
)

type LoopbackHandler struct {
}

func NewLoopbackHandler() *LoopbackHandler {
	return &LoopbackHandler{}
}

func (p *LoopbackHandler) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	err := buf.Copy(rw, rw, buf.OnEOFCopyOption(func() {
		rw.CloseWrite()
	}))
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}
	return nil
}

func (p *LoopbackHandler) HandlePacketConn(ctx context.Context, dst net.Destination, conn udp.PacketReaderWriter) error {
	for {
		pa, err := conn.ReadPacket()
		if err != nil {
			return fmt.Errorf("failed to read packet: %w", err)
		}
		pa.Source = pa.Target
		if err := conn.WritePacket(pa); err != nil {
			return fmt.Errorf("failed to write packet: %w", err)
		}
	}
}

package gvisor

import (
	"fmt"
	"io"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func NewLinkWriterToWriter(writer stack.LinkWriter) io.Writer {
	return &linkWriterToWriter{writer: writer}
}

type linkWriterToWriter struct {
	writer stack.LinkWriter
}

func (l linkWriterToWriter) Write(p []byte) (n int, err error) {
	buffer := buffer.MakeWithData(p)
	packetBufferPtr := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Payload: buffer,
		OnRelease: func() {
			buffer.Release()
		},
	})
	packetList := stack.PacketBufferList{}
	packetList.PushBack(packetBufferPtr)
	_, terr := l.writer.WritePackets(packetList)
	if terr != nil {
		return 0, fmt.Errorf("failed to write packet: %s", terr.String())
	}
	return len(p), nil
}

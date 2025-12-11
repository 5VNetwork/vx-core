package gvisor

import (
	"io"
	"sync"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net/gtcpip"

	"github.com/rs/zerolog/log"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type ReadWriteCloserSplitUdp struct {
	Rw     io.ReadWriteCloser
	Offset int32

	channel chan *buf.Buffer
	once    sync.Once
}

func NewReadWriteCloserSplitUdp(rw io.ReadWriteCloser, offset int32) *ReadWriteCloserSplitUdp {
	return &ReadWriteCloserSplitUdp{
		Rw:      rw,
		Offset:  offset,
		channel: make(chan *buf.Buffer, 100),
	}
}

func (u *ReadWriteCloserSplitUdp) Start() error {
	return nil
}

func (u *ReadWriteCloserSplitUdp) Close() error {
	var err error
	u.once.Do(func() {
		close(u.channel)
		for b := range u.channel {
			b.Release()
		}
		err = u.Rw.Close()
	})
	return err
}

func (u *ReadWriteCloserSplitUdp) Read(p []byte) (int, error) {
	for {
		n, err := u.Rw.Read(p)
		if err != nil {
			return n, err
		}

		ipPacket := gtcpip.NewIPPacket(p[u.Offset:n])
		if ipPacket == nil {
			log.Error().Msg("invalid ip packet")
			continue
		}

		if ipPacket.TransportProtocol() == header.UDPProtocolNumber {
			b := buf.New()
			b.Write(p[u.Offset:n])
			u.channel <- b
		} else {
			return n, nil
		}
	}
}

func (u *ReadWriteCloserSplitUdp) Write(p []byte) (int, error) {
	return u.Rw.Write(p)
}

func (u *ReadWriteCloserSplitUdp) Name() string {
	return ""
}

func (u *ReadWriteCloserSplitUdp) ReadPacket() (*buf.Buffer, error) {
	b, ok := <-u.channel
	if !ok {
		return nil, errors.ErrClosed
	}
	return b, nil
}

func (u *ReadWriteCloserSplitUdp) WritePacket(p *buf.Buffer) error {
	defer p.Release()
	if u.Offset > 0 {
		ipv4 := header.IPVersion(p.Bytes()) == header.IPv4Version
		p.RetreatStart(u.Offset)
		if ipv4 {
			p.Write(ipv4FourBytes)
		} else {
			p.Write(ipv6FourBytes)
		}
	}
	_, err := u.Rw.Write(p.Bytes())
	return err
}

package udp

import (
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/signal/done"
)

type PacketLink struct {
	WritePipe *PacketPipe
	ReadPipe  *PacketPipe
}

func NewLink(bufSize int) (*PacketLink, *PacketLink) {
	pipeA := NewPacketPipe(bufSize)
	pipeB := NewPacketPipe(bufSize)
	linkA := &PacketLink{
		WritePipe: pipeA,
		ReadPipe:  pipeB,
	}
	linkB := &PacketLink{
		WritePipe: pipeB,
		ReadPipe:  pipeA,
	}
	return linkA, linkB
}

func (l *PacketLink) WritePacket(packet *Packet) error {
	return l.WritePipe.WritePacket(packet)
}

func (l *PacketLink) ReadPacket() (*Packet, error) {
	return l.ReadPipe.ReadPacket()
}

func (l *PacketLink) Close() error {
	l.WritePipe.Close()
	l.ReadPipe.Close()
	return nil
}

func (l *PacketLink) SetReadDeadline(t time.Time) error {
	l.ReadPipe.SetReadDeadline(t)
	return nil
}

func (l *PacketLink) SetWriteDeadline(t time.Time) error {
	l.WritePipe.SetWriteDeadline(t)
	return nil
}

func (l *PacketLink) SetDeadline(t time.Time) error {
	l.SetReadDeadline(t)
	l.SetWriteDeadline(t)
	return nil
}

type LinkToNetPacketConn struct {
	*PacketLink
	LocalDestination net.Destination
}

func (pc *LinkToNetPacketConn) LocalAddr() net.Addr {
	return pc.LocalDestination.Addr()
}

func (pc *LinkToNetPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	packet, err := pc.PacketLink.ReadPacket()
	if err != nil {
		return 0, nil, err
	}
	defer packet.Release()
	n = copy(p, packet.Payload.Bytes())
	if n != int(packet.Payload.Len()) {
		return 0, nil, errors.New("not enough space")
	}
	return n, packet.Source.Addr(), nil
}

func (pc *LinkToNetPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	buffer := buf.NewWithSize(int32(len(p)))
	buffer.Write(p)

	packet := &Packet{
		Payload: buffer,
		Source:  pc.LocalDestination,
		Target:  net.DestinationFromAddr(addr),
	}
	err = pc.PacketLink.WritePacket(packet)
	return len(p), err
}

// implements net.PacketConn and PacketConn
type PacketPipe struct {
	c             chan *Packet
	done          *done.Instance
	readDeadline  pipeDeadline
	writeDeadline pipeDeadline
}

func NewPacketPipe(bufSize int) *PacketPipe {
	pc := &PacketPipe{
		c:             make(chan *Packet, bufSize),
		done:          done.New(),
		readDeadline:  makePipeDeadline(),
		writeDeadline: makePipeDeadline(),
	}
	return pc
}

func (p *PacketPipe) ReadPacket() (*Packet, error) {
	switch {
	case common.IsClosedChan(p.readDeadline.wait()):
		return nil, os.ErrDeadlineExceeded
	}

	select {
	case packet := <-p.c:
		return packet, nil
	case <-p.done.Wait():
		return nil, errors.New("closed")
	case <-p.readDeadline.wait():
		return nil, os.ErrDeadlineExceeded
	}
}

func (p *PacketPipe) WritePacket(packet *Packet) error {
	switch {
	case common.IsClosedChan(p.writeDeadline.wait()):
		packet.Release()
		return os.ErrDeadlineExceeded
	}

	select {
	case p.c <- packet:
		return nil
	case <-p.done.Wait():
		packet.Release()
		return io.ErrClosedPipe
	case <-p.writeDeadline.wait():
		packet.Release()
		return os.ErrDeadlineExceeded
	}
}

func (p *PacketPipe) Close() error {
	p.done.Close()
	return nil
}

// SetDeadline sets both read and write deadlines.
func (p *PacketPipe) SetDeadline(t time.Time) error {
	p.readDeadline.set(t)
	p.writeDeadline.set(t)
	return nil
}

// SetReadDeadline sets the read deadline.
func (p *PacketPipe) SetReadDeadline(t time.Time) error {
	p.readDeadline.set(t)
	return nil
}

// SetWriteDeadline sets the write deadline.
func (p *PacketPipe) SetWriteDeadline(t time.Time) error {
	p.writeDeadline.set(t)
	return nil
}

// pipeDeadline is an abstraction for handling timeouts.
type pipeDeadline struct {
	mu     sync.Mutex // Guards timer and cancel
	timer  *time.Timer
	cancel chan struct{} // Must be non-nil
}

func makePipeDeadline() pipeDeadline {
	return pipeDeadline{cancel: make(chan struct{})}
}

// set sets the point in time when the deadline will time out.
// A timeout event is signaled by closing the channel returned by waiter.
// Once a timeout has occurred, the deadline can be refreshed by specifying a
// t value in the future.
//
// A zero value for t prevents timeout.
func (d *pipeDeadline) set(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil && !d.timer.Stop() {
		<-d.cancel // Wait for the timer callback to finish and close cancel
	}
	d.timer = nil

	// Time is zero, then there is no deadline.
	closed := common.IsClosedChan(d.cancel)
	if t.IsZero() {
		if closed {
			d.cancel = make(chan struct{})
		}
		return
	}

	// Time in the future, setup a timer to cancel in the future.
	if dur := time.Until(t); dur > 0 {
		if closed {
			d.cancel = make(chan struct{})
		}
		d.timer = time.AfterFunc(dur, func() {
			close(d.cancel)
		})
		return
	}

	// Time in the past, so close immediately.
	if !closed {
		close(d.cancel)
	}
}

// wait returns a channel that is closed when the deadline is exceeded.
func (d *pipeDeadline) wait() chan struct{} {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.cancel
}

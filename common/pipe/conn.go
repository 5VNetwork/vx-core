package pipe

import (
	"io"
	"net"
	"os"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
)

type LinkConn struct {
	link          *Link
	mb            buf.MultiBuffer
	readDeadline  time.Time
	writeDeadline time.Time
	localAddr     net.Addr
	remoteAddr    net.Addr
	udp           bool
}

func NewLinkConn(link *Link, localAddr net.Addr,
	remoteAddr net.Addr, udp bool) *LinkConn {
	return &LinkConn{
		link:       link,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
		udp:        udp,
	}
}

func (l *LinkConn) WriteMultiBuffer(mb buf.MultiBuffer) error {
	return l.link.Writer.WriteMultiBuffer(mb)
}

func (l *LinkConn) CloseWrite() error {
	return l.link.CloseWrite()
}

func (l *LinkConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if !l.mb.IsEmpty() {
		mb := l.mb
		l.mb = nil
		return mb, nil
	}
	return l.link.ReadMultiBuffer()
}

func (l *LinkConn) Read(b []byte) (n int, err error) {
	if !l.readDeadline.IsZero() && time.Now().After(l.readDeadline) {
		return 0, os.ErrDeadlineExceeded
	}

	if l.mb.IsEmpty() {
		l.mb, err = l.link.ReadMultiBuffer()
	}
	if !l.mb.IsEmpty() {
		if l.udp {
			buffer := l.mb[0]
			l.mb = l.mb[1:]
			defer buffer.Release()
			n, _ = buffer.Read(b)
			if !buffer.IsEmpty() {
				return 0, io.ErrShortBuffer
			}
		} else {
			l.mb, n = buf.SplitBytes(l.mb, b)
		}
	}
	return n, err
}

func (l *LinkConn) Write(b []byte) (n int, err error) {
	if !l.writeDeadline.IsZero() && time.Now().After(l.writeDeadline) {
		return 0, os.ErrDeadlineExceeded
	}

	buffer := buf.NewWithSize(int32(len(b)))
	n, err = buffer.Write(b)
	if err != nil {
		return 0, err
	}

	err = l.link.Writer.WriteMultiBuffer(buf.MultiBuffer{buffer})
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (l *LinkConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = l.Read(p)
	return n, l.remoteAddr, err
}

func (l *LinkConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return l.Write(p)
}

func (l *LinkConn) Close() error {
	l.link.Close()
	buf.ReleaseMulti(l.mb)
	return nil
}

func (l *LinkConn) LocalAddr() net.Addr {
	return l.localAddr
}

func (l *LinkConn) RemoteAddr() net.Addr {
	return l.remoteAddr
}

func (l *LinkConn) SetDeadline(t time.Time) error {
	l.readDeadline = t
	l.writeDeadline = t
	l.link.SetDeadline(t)
	return nil
}

func (l *LinkConn) SetReadDeadline(t time.Time) error {
	l.readDeadline = t
	l.link.SetReadDeadline(t)
	return nil
}

func (l *LinkConn) SetWriteDeadline(t time.Time) error {
	l.writeDeadline = t
	l.link.SetWriteDeadline(t)
	return nil
}

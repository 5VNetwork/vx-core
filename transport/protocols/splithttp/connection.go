package splithttp

import (
	"io"
	"net"
	"time"
)

type splitConn struct {
	writer     io.WriteCloser
	reader     io.ReadCloser
	remoteAddr net.Addr
	localAddr  net.Addr
	onClose    func()
}

func (c *splitConn) Write(b []byte) (int, error) {
	return c.writer.Write(b)
}

func (c *splitConn) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}

func (c *splitConn) Close() error {
	if c.onClose != nil {
		c.onClose()
	}

	err := c.writer.Close()
	err2 := c.reader.Close()
	if err != nil {
		return err
	}

	if err2 != nil {
		return err
	}

	return nil
}

func (c *splitConn) LocalAddr() net.Addr {
	if c.localAddr == nil {
		return &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 0,
		}
	}
	return c.localAddr
}

func (c *splitConn) RemoteAddr() net.Addr {
	if c.remoteAddr == nil {
		return &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 0,
		}
	}
	return c.remoteAddr
}

func (c *splitConn) SetDeadline(t time.Time) error {
	// TODO cannot do anything useful
	return nil
}

func (c *splitConn) SetReadDeadline(t time.Time) error {
	// TODO cannot do anything useful
	return nil
}

func (c *splitConn) SetWriteDeadline(t time.Time) error {
	// TODO cannot do anything useful
	return nil
}

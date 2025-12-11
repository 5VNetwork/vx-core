package proxy

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"
)

type ProxyCtxKey int

const (
	userKey ProxyCtxKey = iota
)

func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, userKey, user)
}
func UserFromContext(ctx context.Context) (string, bool) {
	if ctx.Value(userKey) == nil {
		return "", false
	}
	return ctx.Value(userKey).(string), true
}

const FirstPayloadTimeout = 10 * time.Millisecond

type CustomConn struct {
	net.Conn
	io.Writer
	io.Reader
	closeWrite func()
	close      func()
}

type CustomConnOption struct {
	Conn       net.Conn
	Writer     io.Writer
	Reader     io.Reader
	CloseWrite func()
	Close      func()
}

func NewProxyConn(customConn CustomConnOption) *CustomConn {
	return &CustomConn{
		Conn:       customConn.Conn,
		Writer:     customConn.Writer,
		Reader:     customConn.Reader,
		closeWrite: customConn.CloseWrite,
		close:      customConn.Close,
	}
}

func (c *CustomConn) NetConn() net.Conn {
	return c.Conn
}

func (c *CustomConn) Write(b []byte) (int, error) {
	return c.Writer.Write(b)
}

func (c *CustomConn) Read(b []byte) (int, error) {
	return c.Reader.Read(b)
}

func (c *CustomConn) Close() error {
	if c.close != nil {
		c.close()
	}
	return c.Conn.Close()
}

func (c *CustomConn) CloseWrite() error {
	if c.closeWrite != nil {
		c.closeWrite()
	}
	return nil
}

type NetConnPacketConn struct {
	net.Conn
	udp.ReadFromer
	udp.WriteToer
}

func NewNetConnPacketConn(conn net.Conn, readFromer udp.ReadFromer, writeToer udp.WriteToer) *NetConnPacketConn {
	return &NetConnPacketConn{
		Conn:       conn,
		ReadFromer: readFromer,
		WriteToer:  writeToer,
	}
}

type FlowConn struct {
	buf.Reader
	buf.Writer
	i.ReadDeadline
	close      func() error
	closeWrite func() error
}

type FlowConnOption struct {
	Reader      buf.Reader
	Writer      buf.Writer
	Close       func() error
	CloseWrite  func() error
	SetDeadline i.ReadDeadline
}

func NewFlowConn(flowConnOption FlowConnOption) *FlowConn {
	return &FlowConn{
		Reader:       flowConnOption.Reader,
		Writer:       flowConnOption.Writer,
		close:        flowConnOption.Close,
		closeWrite:   flowConnOption.CloseWrite,
		ReadDeadline: flowConnOption.SetDeadline,
	}
}

func (c *FlowConn) Close() error {
	if c.close != nil {
		return c.close()
	}
	return nil
}

func (c *FlowConn) CloseWrite() error {
	if c.closeWrite != nil {
		err := c.closeWrite()
		if err != nil {
			return err
		}
	}
	return c.Writer.CloseWrite()
}

func (c *FlowConn) OkayToUnwrapReader() int {
	return 1
}

func (c *FlowConn) OkayToUnwrapWriter() int {
	return 1
}

func (c *FlowConn) UnwrapWriter() any {
	return c.Writer
}

func (c *FlowConn) UnwrapReader() any {
	return c.Reader
}

var ErrUDPNotSupport = errors.New("udp is not supported")

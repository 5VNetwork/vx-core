package net

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/common/buf"
)

type HasNetConn interface {
	NetConn() net.Conn
}

func GetInnerMostConn(conn net.Conn) net.Conn {
	for {
		if c, ok := conn.(HasNetConn); ok {
			conn = c.NetConn()
		} else {
			break
		}
	}
	return conn
}

// read will read from bytes first
type BytesConn struct {
	net.Conn
	bytes []byte
}

func NewBytesConn(conn net.Conn, b []byte) *BytesConn {
	return &BytesConn{
		Conn:  conn,
		bytes: b,
	}
}

func (c *BytesConn) Read(b []byte) (int, error) {
	if len(c.bytes) > 0 {
		n := copy(b, c.bytes)
		c.bytes = c.bytes[n:]
		return n, nil
	}
	return c.Conn.Read(b)
}

// read will read from Mb first
type MbConn struct {
	net.Conn
	Mb buf.MultiBuffer
}

func NewMbConn(conn net.Conn, cache buf.MultiBuffer) *MbConn {
	return &MbConn{
		Conn: conn,
		Mb:   cache,
	}
}

func (c *MbConn) NetConn() net.Conn {
	return c.Conn
}

func (c *MbConn) Read(b []byte) (n int, err error) {
	if c.Mb.Len() > 0 {
		c.Mb, n = buf.SplitBytes(c.Mb, b)
		return n, nil
	}
	return c.Conn.Read(b)
}

func (c *MbConn) CloseWrite() error {
	if cw, ok := c.Conn.(buf.CloseWriter); ok {
		return cw.CloseWrite()
	}
	return nil
}

func (c *MbConn) OkayToUnwrapWriter() int {
	return 1
}

func (c *MbConn) UnwrapWriter() any {
	return c.Conn
}

func (c *MbConn) OkayToUnwrapReader() int {
	if c.Mb.Len() > 0 {
		return 0
	}
	return 1
}

func (c *MbConn) UnwrapReader() any {
	return c.Conn
}

type ConnWithClose struct {
	net.Conn
	closeCallback func()
}

func NewConnWithClose(conn net.Conn, cb func()) *ConnWithClose {
	return &ConnWithClose{
		Conn:          conn,
		closeCallback: cb,
	}
}

func (c *ConnWithClose) Close() error {
	err := c.Conn.Close()
	c.closeCallback()
	return err
}

type StatsConn struct {
	net.Conn
	ReadCounter  *atomic.Uint64
	WriteCounter *atomic.Uint64
}

func NewStatsConn(conn net.Conn, readCounter, writeCounter *atomic.Uint64) *StatsConn {
	return &StatsConn{
		Conn:         conn,
		ReadCounter:  readCounter,
		WriteCounter: writeCounter,
	}
}

func (c *StatsConn) HasNetConn() net.Conn {
	return c.Conn
}

func (c *StatsConn) Read(b []byte) (int, error) {
	nBytes, err := c.Conn.Read(b)
	if c.ReadCounter != nil {
		c.ReadCounter.Add(uint64(nBytes))
	}

	return nBytes, err
}

func (c *StatsConn) Write(b []byte) (int, error) {
	nBytes, err := c.Conn.Write(b)
	if c.WriteCounter != nil {
		c.WriteCounter.Add(uint64(nBytes))
	}
	return nBytes, err
}

type StatsPacketConn struct {
	net.PacketConn
	ReadCounter  *atomic.Uint64
	WriteCounter *atomic.Uint64
}

func NewStatsPacketConn(pc net.PacketConn, readCounter, writeCounter *atomic.Uint64) *StatsPacketConn {
	return &StatsPacketConn{
		PacketConn:   pc,
		ReadCounter:  readCounter,
		WriteCounter: writeCounter,
	}
}

func (c *StatsPacketConn) ReadFrom(p []byte) (n int, addr Addr, err error) {
	n, addr, err = c.PacketConn.ReadFrom(p)
	if c.ReadCounter != nil {
		c.ReadCounter.Add(uint64(n))
	}
	return
}

func (c *StatsPacketConn) WriteTo(p []byte, addr Addr) (n int, err error) {
	n, err = c.PacketConn.WriteTo(p, addr)
	if c.WriteCounter != nil {
		c.WriteCounter.Add(uint64(n))
	}
	return
}

type NetDialer struct {
	net.Dialer
}

func (d *NetDialer) Dial(ctx context.Context, dst Destination) (net.Conn, error) {
	addr := dst.NetAddr()
	return d.Dialer.DialContext(ctx, dst.Network.SystemString(), addr)
}

type NetPacketListener struct {
	net.ListenConfig
}

func (l *NetPacketListener) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	return l.ListenConfig.ListenPacket(ctx, network, address)
}

// all data read from Conn will be saved to history until stop
type MemoryConn struct {
	net.Conn

	history buf.MultiBuffer
	stop    bool
}

func NewMemoryRConn(conn net.Conn) *MemoryConn {
	return &MemoryConn{
		Conn: conn,
	}
}

func (c *MemoryConn) History() buf.MultiBuffer {
	return c.history
}

func (c *MemoryConn) StopMemorize() {
	c.stop = true
	buf.ReleaseMulti(c.history)
}

func (c *MemoryConn) Read(p []byte) (int, error) {
	if c.stop {
		return c.Conn.Read(p)
	}

	n, err := c.Conn.Read(p)
	if err != nil {
		return n, err
	}

	c.history = buf.MergeBytes(c.history, p[:n])
	return n, nil
}

func (c *MemoryConn) OkayToUnwrapReader() int {
	if c.stop {
		return 1
	}
	return 0
}

func (c *MemoryConn) UnwrapReader() any {
	return c.Conn
}

func (c *MemoryConn) OkayToUnwrapWriter() int {
	return 1
}

func (c *MemoryConn) UnwrapWriter() any {
	return c.Conn
}

type NetConnToPacketConn struct {
	net.Conn
}

func (c *NetConnToPacketConn) ReadFrom(p []byte) (n int, addr Addr, err error) {
	n, err = c.Read(p)
	return n, c.Conn.RemoteAddr(), err
}

func (c *NetConnToPacketConn) WriteTo(p []byte, addr Addr) (n int, err error) {
	return c.Write(p)
}

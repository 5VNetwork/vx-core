package websocket

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
)

var _ buf.Writer = (*connection)(nil)

// connection is a wrapper for net.Conn over WebSocket connection.
type connection struct {
	conn       *websocket.Conn
	reader     io.Reader
	remoteAddr net.Addr

	shouldWait        bool
	delayedDialFinish context.Context
	finishedDial      context.CancelFunc
	dialer            DelayedDialer
}

type DelayedDialer interface {
	Dial(earlyData []byte) (*websocket.Conn, error)
}

func newConnection(conn *websocket.Conn, remoteAddr net.Addr) *connection {
	return &connection{
		conn:       conn,
		remoteAddr: remoteAddr,
	}
}

func newConnectionWithEarlyData(conn *websocket.Conn, remoteAddr net.Addr, earlyData io.Reader) *connection {
	return &connection{
		conn:       conn,
		remoteAddr: remoteAddr,
		reader:     earlyData,
	}
}

func newConnectionWithDelayedDial(dialer DelayedDialer) *connection {
	delayedDialContext, cancelFunc := context.WithCancel(context.Background())
	return &connection{
		shouldWait:        true,
		delayedDialFinish: delayedDialContext,
		finishedDial:      cancelFunc,
		dialer:            dialer,
	}
}

func newRelayedConnectionWithDelayedDial(dialer DelayedDialerForwarded) *connectionForwarder {
	delayedDialContext, cancelFunc := context.WithCancel(context.Background())
	return &connectionForwarder{
		shouldWait:        true,
		delayedDialFinish: delayedDialContext,
		finishedDial:      cancelFunc,
		dialer:            dialer,
	}
}

func newRelayedConnection(conn io.ReadWriteCloser) *connectionForwarder {
	return &connectionForwarder{
		ReadWriteCloser: conn,
		shouldWait:      false,
	}
}

// Read implements net.Conn.Read()
func (c *connection) Read(b []byte) (int, error) {
	for {
		reader, err := c.getReader()
		if err != nil {
			return 0, err
		}

		nBytes, err := reader.Read(b)
		if errors.Is(err, io.EOF) {
			c.reader = nil
			continue
		}
		return nBytes, err
	}
}

func (c *connection) getReader() (io.Reader, error) {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			return nil, errors.New("unable to read delayed dial websocket connection as it do not exist")
		}
	}
	if c.reader != nil {
		return c.reader, nil
	}

	_, reader, err := c.conn.NextReader()
	if err != nil {
		return nil, err
	}
	c.reader = reader
	return reader, nil
}

// Write implements io.Writer.
func (c *connection) Write(b []byte) (int, error) {
	if c.shouldWait {
		var err error
		c.conn, err = c.dialer.Dial(b)
		c.finishedDial()
		if err != nil {
			return 0, errors.New("Unable to proceed with delayed write").Base(err)
		}
		c.remoteAddr = c.conn.RemoteAddr()
		c.shouldWait = false
		return len(b), nil
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *connection) WriteMultiBuffer(mb buf.MultiBuffer) error {
	mb = buf.Compact(mb)
	mb, err := buf.WriteMultiBuffer(c, mb)
	buf.ReleaseMulti(mb)
	return err
}

func (c *connection) CloseWrite() error {
	return nil
}

func (c *connection) Close() error {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			return errors.New("unable to close delayed dial websocket connection as it do not exist")
		}
	}
	var errs []error
	if err := c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second*5)); err != nil {
		errs = append(errs, err)
	}
	if err := c.conn.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close websocket connection: %v", errors.Join(errs...))
	}
	return nil
}

func (c *connection) LocalAddr() net.Addr {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			log.Warn().Msg("websocket transport is not materialized when LocalAddr() is called")
			return &net.UnixAddr{
				Name: "@placeholder",
				Net:  "unix",
			}
		}
	}
	return c.conn.LocalAddr()
}

func (c *connection) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *connection) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

func (c *connection) SetReadDeadline(t time.Time) error {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			log.Warn().Msg("websocket transport is not materialized when SetReadDeadline() is called")
			return nil
		}
	}
	return c.conn.SetReadDeadline(t)
}

func (c *connection) SetWriteDeadline(t time.Time) error {
	if c.shouldWait {
		<-c.delayedDialFinish.Done()
		if c.conn == nil {
			log.Warn().Msg("websocket transport is not materialized when SetWriteDeadline() is called")
			return nil
		}
	}
	return c.conn.SetWriteDeadline(t)
}

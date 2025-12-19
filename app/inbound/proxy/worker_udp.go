package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	nethelper "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type udpWorker struct {
	sync.RWMutex
	connHandler        connHandler
	udpConn            *net.UDPConn
	addr               *net.UDPAddr
	address            net.IP
	port               uint16
	tag                string
	listener           i.PacketListener
	periodicCleaner    *task.Periodic
	activeMyUdpConnMap map[nethelper.Destination]*myUdpConn
	closed             bool
}

func (w *udpWorker) Start() error {
	w.activeMyUdpConnMap = make(map[nethelper.Destination]*myUdpConn, 16)

	udpConn, err := w.listener.ListenPacket(context.Background(), w.addr.Network(), w.addr.String())
	if err != nil {
		return fmt.Errorf("failed to listen udp: %w", err)
	}

	log.Debug().Str("address", w.addr.String()).Msg("udp listening")

	w.periodicCleaner = &task.Periodic{
		Interval: time.Second * 16,
		Execute:  w.clean,
	}
	w.periodicCleaner.Start()

	w.udpConn = udpConn.(*net.UDPConn)

	// w.conn receive packets and
	go func() {
		for {
			buffer := buf.New()
			rawBytes := buffer.Extend(buf.Size)

			nBytes, addr, err := w.udpConn.ReadFromUDP(rawBytes)
			if err != nil {
				buffer.Release()
				if w.closed && errors.Is(err, net.ErrClosed) {
					return
				}
				log.Err(err).Msg("udpWorker's udpConn failed in readFromUDP")
				return
			}
			buffer.Resize(0, int32(nBytes))

			if buffer.IsEmpty() {
				buffer.Release()
				continue
			}

			w.handlePacket(buffer, addr)
		}
	}()
	return nil
}

func (w *udpWorker) Close() error {
	w.Lock()
	defer w.Unlock()

	w.closed = true
	var errs []error
	if w.udpConn != nil {
		errs = append(errs, w.udpConn.Close())
		w.udpConn = nil
	}

	if w.periodicCleaner != nil {
		errs = append(errs, w.periodicCleaner.Close())
		w.periodicCleaner = nil
	}

	errs = append(errs, common.Close(w.connHandler))
	return errors.Join(errs...)
}

func (w *udpWorker) closeAndRemoveMyUdpConn(addr nethelper.Destination) {
	w.Lock()
	defer w.Unlock()
	if myUConn, found := w.activeMyUdpConnMap[addr]; found {
		myUConn.Close()
		delete(w.activeMyUdpConnMap, addr)
	}
}

func (w *udpWorker) handlePacket(b *buf.Buffer, source *net.UDPAddr) {
	w.Lock()
	defer w.Unlock()
	src := nethelper.DestinationFromAddr(source)
	myUConn, found := w.activeMyUdpConnMap[src]
	// create a new myUdpConn
	if !found || myUConn.done.Done() {
		ctx, cancelCause := inbound.GetCtx(
			src, nethelper.UDPDestination(nethelper.IPAddress(w.address),
				nethelper.Port(w.port)), w.tag)
		myUConn = newMyUdpConn(w.udpConn, source, w.addr)
		w.activeMyUdpConnMap[src] = myUConn
		go func() {
			err := w.connHandler.Process(ctx, myUConn)
			if err != nil {
				log.Ctx(ctx).Err(err).Send()
			}
			cancelCause(err)
			w.closeAndRemoveMyUdpConn(src)
		}()
	}
	// payload will be discarded in pipe is full.
	myUConn.pipe.WriteMultiBuffer(buf.MultiBuffer{b})
}

func (w *udpWorker) Port() uint16 {
	return w.port
}

// close and delete the inactive udpConn
func (w *udpWorker) clean() error {
	nowSec := time.Now().Unix()
	w.Lock()
	defer w.Unlock()
	for addr, conn := range w.activeMyUdpConnMap {
		if nowSec-atomic.LoadInt64(&conn.lastActivityTime) > int64(policy.DefaultTimeout.UdpIdleTimeout()) {
			conn.Close()
			delete(w.activeMyUdpConnMap, addr)
		}
	}
	if len(w.activeMyUdpConnMap) == 0 {
		w.activeMyUdpConnMap = make(map[nethelper.Destination]*myUdpConn, 16)
	}
	return nil
}

type myUdpConn struct {
	pipe    *pipe.Pipe
	udpConn *net.UDPConn
	remote  *net.UDPAddr
	local   *net.UDPAddr
	done    *done.Instance
	// only request is used when checking idle
	lastActivityTime int64
}

func newMyUdpConn(udpConn *net.UDPConn, remote *net.UDPAddr, local *net.UDPAddr) *myUdpConn {
	c := &myUdpConn{
		pipe:    pipe.NewPipe(16*1024, true),
		remote:  remote,
		local:   local,
		done:    done.New(),
		udpConn: udpConn,
	}
	return c
}

func (c *myUdpConn) Close() error {
	c.done.Close()
	c.pipe.Close()
	return nil
}

// ReadMultiBuffer implements buf.Reader
func (c *myUdpConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := c.pipe.ReadMultiBuffer()
	if err != nil {
		return nil, err
	}

	atomic.StoreInt64(&c.lastActivityTime, time.Now().Unix())

	return mb, nil
}

func (c *myUdpConn) Read(buf []byte) (int, error) {
	panic("not implemented")
}

// Write implements io.Writer.
func (c *myUdpConn) Write(buf []byte) (int, error) {
	return c.udpConn.WriteToUDP(buf, c.remote)
}

func (c *myUdpConn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *myUdpConn) LocalAddr() net.Addr {
	return c.local
}

func (*myUdpConn) SetDeadline(time.Time) error {
	return nil
}

func (*myUdpConn) SetReadDeadline(time.Time) error {
	return nil
}

func (*myUdpConn) SetWriteDeadline(time.Time) error {
	return nil
}

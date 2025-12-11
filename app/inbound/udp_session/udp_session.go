package udp_session

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/pipe"
)

type UdpSession interface {
	Ctx() context.Context
	Src() net.Destination
	// return quickly
	// Send p to its destination
	Dispatch(p *udp.Packet) error
	// close all related resources
	Close()
}

type SessionManager struct {
	sync.RWMutex
	udpSrcToUdpSession map[net.Destination]UdpSession
}

func NewManager() *SessionManager {
	m := make(map[net.Destination]UdpSession)
	manager := &SessionManager{udpSrcToUdpSession: m}
	return manager
}

func (u *SessionManager) Start() {
}

func (u *SessionManager) Close() {
	u.Lock()
	for _, s := range u.udpSrcToUdpSession {
		s.Close()
	}
	u.udpSrcToUdpSession = make(map[net.Destination]UdpSession)
	u.Unlock()
}

func (u *SessionManager) AddUdpSession(src net.Destination, s UdpSession) {
	u.Lock()
	u.udpSrcToUdpSession[src] = s
	u.Unlock()
}

func (u *SessionManager) GetUdpSession(src net.Destination) (s UdpSession, found bool) {
	u.RLock()
	defer u.RUnlock()
	s, found = u.udpSrcToUdpSession[src]
	return
}

func (u *SessionManager) CloseAndRemoveUdpSession(src net.Destination) {
	u.Lock()
	if s, found := u.udpSrcToUdpSession[src]; found {
		delete(u.udpSrcToUdpSession, src)
		s.Close()
	}
	u.Unlock()
}

type UDPSessionCommon struct {
	Ctx         context.Context
	Src         net.Destination
	CancelCause context.CancelCauseFunc
	Ipv4        bool
	Time        time.Time
	CloseOnce   sync.Once
}

func (u *UDPSessionCommon) HasTimeout() bool {
	return time.Since(u.Time) > time.Minute
}

type udpSessionFullCone struct {
	ctx            context.Context
	src            net.Destination
	cancelCause    context.CancelCauseFunc
	responseWriter responseWriter
	requestChan    chan *udp.Packet
	sync.RWMutex
	closed bool
}

func NewUdpSessionFullCone(ctx context.Context, src net.Destination,
	cancelCause context.CancelCauseFunc, responseWriter responseWriter,
	requestChan chan *udp.Packet) *udpSessionFullCone {
	return &udpSessionFullCone{
		ctx:            ctx,
		src:            src,
		cancelCause:    cancelCause,
		responseWriter: responseWriter,
		requestChan:    requestChan,
	}
}

func (u *udpSessionFullCone) Ctx() context.Context {
	return u.ctx
}

func (u *udpSessionFullCone) Src() net.Destination {
	return u.src
}

func (u *udpSessionFullCone) Dispatch(p *udp.Packet) error {
	u.RLock()
	defer u.RUnlock()
	if u.closed {
		return errors.New("session is closed")
	}

	select {
	case u.requestChan <- p:
		return nil
	default:
		return errors.New("request channel is full")
	}
}

func (s *udpSessionFullCone) Close() {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.cancelCause(nil)
	close(s.requestChan)
	for p := range s.requestChan {
		p.Release()
	}
}

type responseWriter interface {
	WritePacket(p *udp.Packet) error
}

func (u *udpSessionFullCone) WritePacket(p *udp.Packet) error {
	p.Target = u.src
	return u.responseWriter.WritePacket(p)
}

func (u *udpSessionFullCone) ReadPacket() (*udp.Packet, error) {
	p, open := <-u.requestChan
	if !open {
		return nil, common.ErrClosed
	}
	return p, nil
}

// every src address has a udpSessionSymetric
type udpSessionSymetric struct {
	ctx         context.Context
	src         net.Destination
	dst         net.Destination
	CancelCause context.CancelCauseFunc
	*pipe.Pipe
	responseWriter responseWriter
}

func NewUdpSessionSymetric(ctx context.Context, src net.Destination,
	dst net.Destination, cancelCause context.CancelCauseFunc,
	responseWriter responseWriter, pipe *pipe.Pipe) *udpSessionSymetric {
	return &udpSessionSymetric{
		ctx:            ctx,
		src:            src,
		dst:            dst,
		CancelCause:    cancelCause,
		responseWriter: responseWriter,
		Pipe:           pipe,
	}
}

func (u *udpSessionSymetric) Ctx() context.Context {
	return u.ctx
}

func (u *udpSessionSymetric) Dispatch(p *udp.Packet) error {
	return u.Pipe.WriteMultiBuffer(buf.MultiBuffer{p.Payload})
}

func (u *udpSessionSymetric) Src() net.Destination {
	return u.src
}

func (u *udpSessionSymetric) Close() {
	u.CancelCause(nil)
	u.Pipe.Close()
}

func (u *udpSessionSymetric) WriteMultiBuffer(mb buf.MultiBuffer) error {
	defer buf.ReleaseMulti(mb)
	for i, b := range mb {
		err := u.responseWriter.WritePacket(&udp.Packet{
			Source:  u.dst,
			Target:  u.src,
			Payload: b,
		})
		mb[i] = nil
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *udpSessionSymetric) CloseWrite() error {
	return nil
}

//go:build server

package shadowsocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
)

type Server struct {
	lock                      sync.RWMutex
	memoryAccount             *MemoryAccount
	cipher                    CipherType
	reducedIVEntropy, ivCheck bool
	PolicyManager             i.TimeoutSetting
	Handler                   i.Handler
}

func (s *Server) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UDP}
}

type ServerSettings struct {
	Cipher                    CipherType
	ReducedIVEntropy, IvCheck bool
	PolicyManager             i.TimeoutSetting
	Handler                   i.Handler
}

// NewServer create a new Shadowsocks server.
func NewServer(settings ServerSettings) *Server {
	s := &Server{
		cipher:           settings.Cipher,
		reducedIVEntropy: settings.ReducedIVEntropy,
		ivCheck:          settings.IvCheck,
		PolicyManager:    settings.PolicyManager,
		Handler:          settings.Handler,
	}
	return s
}

func (s *Server) AddUser(user i.User) {
	s.lock.Lock()
	defer s.lock.Unlock()
	memoryAccount, _ := NewMemoryAccount(user.Uid(), s.cipher, user.Secret(),
		s.reducedIVEntropy, s.ivCheck)
	s.memoryAccount = memoryAccount
}

func (s *Server) RemoveUser(uid, secret string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.memoryAccount.Uid == uid {
		s.memoryAccount = nil
	}
}

func (h *Server) WithOnUnauthorizedRequest(f i.UnauthorizedReport) {
}

func (s *Server) GetMemoryAccount() *MemoryAccount {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.memoryAccount
}

func (s *Server) Process(ctx context.Context, conn net.Conn) error {
	account := s.GetMemoryAccount()
	if account == nil {
		return errors.New("user not set")
	}

	switch net.NetworkFromAddr(conn.LocalAddr()) {
	case net.Network_TCP:
		return s.handleConnection(ctx, conn, account)
	case net.Network_UDP:
		return s.handlerUDPPayload(ctx, conn, account)
	default:
		return errors.New("unknown network")
	}
}

func (s *Server) FallbackProcess(ctx context.Context, conn net.Conn) (bool, buf.MultiBuffer, error) {
	account := s.GetMemoryAccount()
	if account == nil {
		return true, nil, errors.New("user not set")
	}

	conn.SetReadDeadline(time.Now().Add(s.PolicyManager.HandshakeTimeout()))

	cacheReader := buf.NewMemoryReader(conn)

	bufferedReader := buf.BufferedReader{Reader: buf.NewReader(cacheReader)}
	request, bodyReader, err := ReadTCPSession(account, &bufferedReader, false)
	if err != nil {
		return true, cacheReader.History(), err
	}
	conn.SetReadDeadline(time.Time{})

	cacheReader.StopMemorize()

	return false, nil, s.handleConnectionCommon(ctx, conn, request, bodyReader, account)
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn, account *MemoryAccount) error {
	conn.SetReadDeadline(time.Now().Add(s.PolicyManager.HandshakeTimeout()))

	bufferedReader := buf.BufferedReader{Reader: buf.NewReader(conn)}
	request, bodyReader, err := ReadTCPSession(account, &bufferedReader, true)
	if err != nil {
		return errors.New("failed to create request from: ", conn.RemoteAddr()).Base(err)
	}
	conn.SetReadDeadline(time.Time{})

	return s.handleConnectionCommon(ctx, conn, request, bodyReader, account)
}

// func (s *Server) handleConnectionCommon(ctx context.Context, conn net.Conn,
// 	request *protocol.RequestHeader, bodyReader buf.Reader) error {
// 	var err error
// 	ctx = proxy.ContextWithUser(ctx, s.user.Uid)

// 	iLink, oLink := helper.GetLinks(request.Destination().Network, 0, s.policyManager.(i.BufferPolicy))

// 	go func() {
// 		err = s.d.HandleFlow(ctx, request.Destination(), oLink)
// 		if err != nil {
// 			log.Ctx(ctx).Error().Err(err).Msg("handler stops to handle")
// 		}
// 		oLink.Interrupt(err)
// 	}()

// 	responseDone := func() error {
// 		bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
// 		responseWriter, err := WriteTCPResponse(request, bufferedWriter)
// 		if err != nil {
// 			return fmt.Errorf("failed to write response: %w", err)
// 		}
// 		{
// 			payload, err := iLink.ReadMultiBuffer()
// 			if !payload.IsEmpty() {
// 				if err := responseWriter.WriteMultiBuffer(payload); err != nil {
// 					return err
// 				}
// 			}
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		if err := bufferedWriter.SetBuffered(false); err != nil {
// 			return err
// 		}
// 		if err := buf.Copy(iLink, responseWriter); err != nil {
// 			return fmt.Errorf("failed to transport all TCP response: %w", err)
// 		}
// 		responseWriter.CloseWrite()
// 		return nil
// 	}

// 	requestDone := func() error {
// 		if err := buf.Copy(bodyReader, iLink); err != nil {
// 			return fmt.Errorf("failed to transport all TCP request: %w", err)
// 		}
// 		iLink.CloseWrite()
// 		return nil
// 	}

// 	err = task.Run(ctx, requestDone, responseDone)
// 	if err != nil {
// 		return fmt.Errorf("connnection ends with error: %w", err)
// 	}
// 	return nil
// }

func (s *Server) handleConnectionCommon(ctx context.Context, conn net.Conn,
	request *protocol.RequestHeader, bodyReader buf.Reader, account *MemoryAccount) error {
	var err error
	ctx = proxy.ContextWithUser(ctx, account.Uid)

	bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
	responseWriter, err := WriteTCPResponse(request, bufferedWriter)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return s.Handler.HandleFlow(ctx, request.Destination(), &rw{
		Reader:         bodyReader,
		bufferedWriter: bufferedWriter,
		Writer:         responseWriter,
		ReadDeadline:   conn,
	})
}

type rw struct {
	buf.Reader
	writeLock      sync.Mutex
	bufferedWriter *buf.BufferedWriter
	buf.Writer
	i.ReadDeadline
}

func (r *rw) WriteMultiBuffer(mb buf.MultiBuffer) error {
	r.writeLock.Lock()
	defer r.writeLock.Unlock()
	if r.bufferedWriter != nil {
		if err := r.Writer.WriteMultiBuffer(mb); err != nil {
			return err
		}
		r.bufferedWriter.SetBuffered(false)
		r.bufferedWriter = nil
		return nil
	}
	return r.Writer.WriteMultiBuffer(mb)
}

type UDPConn struct {
	ctx     context.Context
	cache   buf.MultiBuffer
	reader  buf.Reader
	conn    net.Conn
	account *MemoryAccount
}

func (c *UDPConn) ReadPacket() (*udp.Packet, error) {
	var p *buf.Buffer
	if len(c.cache) > 0 {
		p = c.cache[0]
		c.cache = c.cache[1:]
	} else {
		var err error
		mb, err := c.reader.ReadMultiBuffer()
		if err != nil {
			return nil, fmt.Errorf("failed to read UDP packet, %w", err)
		}
		if len(mb) == 0 {
			return nil, errors.New("empty UDP packet")
		}
		p = mb[0]
		c.cache = mb[1:]
	}
	r, data, err := DecodeUDPPacket(c.account, p)
	if err != nil {
		p.Release()
		return nil, fmt.Errorf("failed to decode UDP packet, %w", err)
	}
	return &udp.Packet{
		Payload: data,
		Target:  r.Destination(),
	}, nil
}

func (c *UDPConn) WritePacket(p *udp.Packet) error {
	defer p.Release()
	request := &protocol.RequestHeader{
		Port:    p.Source.Port,
		Address: p.Source.Address,
		User:    c.account.Uid,
		Account: c.account,
	}
	data, err := EncodeUDPPacket(request, p.Payload.Bytes())
	if err != nil {
		return fmt.Errorf("failed to encode UDP packet, %w", err)
	}
	defer data.Release()
	if _, err = c.conn.Write(data.Bytes()); err != nil {
		return fmt.Errorf("failed to write UDP packet, %w", err)
	} //send the returned udp packet(from server) back to the client
	return nil
}

func (c *UDPConn) CloseWrite() error {
	return nil
}

func (c *UDPConn) Close() error {
	return nil
}

func (s *Server) handlerUDPPayload(ctx context.Context, conn net.Conn, account *MemoryAccount) error {
	udpConn := &UDPConn{
		ctx:     ctx,
		conn:    conn,
		account: account,
		reader:  buf.NewReader(conn),
	}
	ctx = proxy.ContextWithUser(ctx, account.Uid)
	err := s.Handler.HandlePacketConn(ctx, net.AnyUdpDest, udpConn)
	if err != nil {
		return fmt.Errorf("failed to dispatch UDP, %w", err)
	}
	return nil
}

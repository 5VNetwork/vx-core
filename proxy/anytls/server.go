package anytls

import (
	"anytls/proxy/padding"
	"anytls/proxy/session"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"sync"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/serial/address_parser"
	S "github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/rs/zerolog/log"
)

type Server struct {
	ServerSettings
	secrets sync.Map // key: [32]byte sha256(password), value: string uid
}

type ServerSettings struct {
	Handler               i.Handler
	OnUnauthorizedRequest i.UnauthorizedReport
}

func NewServer(settings ServerSettings) *Server {
	return &Server{
		ServerSettings: settings,
	}
}

func (h *Server) AddUser(user i.User) {
	var sum = sha256.Sum256([]byte(user.Secret()))
	h.secrets.Store(sum, user.Uid())
}

func (h *Server) RemoveUser(uid, secret string) {
	if secret == "" {
		h.secrets.Range(func(key, value any) bool {
			if value.(string) == uid {
				h.secrets.Delete(key)
				return false
			}
			return true
		})
	} else {
		var sum = sha256.Sum256([]byte(secret))
		h.secrets.Delete(sum)
	}
}

func (h *Server) WithOnUnauthorizedRequest(f i.UnauthorizedReport) {
	h.OnUnauthorizedRequest = f
}

func (d *Server) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

func (h *Server) GetUser(sha256 [32]byte) (string, error) {
	uid, ok := h.secrets.Load(sha256)
	if !ok {
		return "", errors.New("user not found")
	}
	return uid.(string), nil
}

func (s *Server) FallbackProcess(ctx context.Context, conn net.Conn) (bool, buf.MultiBuffer, error) {
	cacheReader := buf.NewMemoryReader(conn)
	bufferedReader := &buf.BufferedReader{
		Reader: buf.NewReader(conn),
	}
	user, err := s.auth(ctx, bufferedReader, conn)
	if err != nil {
		return true, cacheReader.History(), err
	}
	ctx = proxy.ContextWithUser(ctx, user)
	cacheReader.StopMemorize()

	return false, nil, s.processCommon(ctx, conn, bufferedReader)
}

func (d *Server) Process(ctx context.Context, conn net.Conn) error {
	bufferedReader := &buf.BufferedReader{
		Reader: buf.NewReader(conn),
	}
	user, err := d.auth(ctx, bufferedReader, conn)
	if err != nil {
		return err
	}
	ctx = proxy.ContextWithUser(ctx, user)
	return d.processCommon(ctx, conn, bufferedReader)
}

func (d *Server) auth(ctx context.Context, reader io.Reader, conn net.Conn) (string, error) {
	var sha256 [32]byte
	n, err := reader.Read(sha256[:])
	if err != nil {
		return "", err
	}
	if n < 32 {
		return "", errors.New("not anytls protocol")
	}
	user, err := d.GetUser(sha256)
	if err != nil {
		if d.OnUnauthorizedRequest != nil {
			d.OnUnauthorizedRequest.ReportUnauthorized(conn.RemoteAddr().String(), "")
		}
		return "", err
	}
	return user, nil
}

func (d *Server) processCommon(ctx context.Context, conn net.Conn,
	bufferedReader *buf.BufferedReader) error {
	paddingLen, err := serial.ReadUint16(bufferedReader)
	if err != nil {
		return err
	}
	if paddingLen > 0 {
		err = buf.Copy(buf.NewSizedReader(bufferedReader), buf.Discard)
		if err != nil {
			return err
		}
	}

	conn = net.NewMbConn(conn, bufferedReader.Buffer)
	session := session.NewServerSession(conn, func(stream *session.Stream) {
		defer stream.Close()

		ctx = S.GetCtx(ctx)
		address, port, err := address_parser.SocksAddressSerializer.ReadAddressPort(nil, stream)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("ReadAddressPort")
			return
		}
		destination := net.Destination{
			Address: address,
			Port:    port,
			Network: net.Network_TCP,
		}
		err = d.Handler.HandleFlow(ctx, destination,
			buf.NewRWD(buf.NewReader(stream), buf.NewWriter(stream), stream))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("HandleFlow")
			return
		}
	}, &padding.DefaultPaddingFactory)
	session.Run()
	return session.Close()
}

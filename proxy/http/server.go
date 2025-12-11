package http

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	http_proto "github.com/5vnetwork/vx-core/common/protocol/http"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"
)

// Server is an HTTP proxy server.
type Server struct {
	ServerSettings
}

type ServerSettings struct {
	PolicyManager    i.TimeoutSetting
	Handler          i.Handler
	AllowTransparent bool
}

// NewServer creates a new HTTP inbound handler.
func NewServer(settings ServerSettings) *Server {
	s := &Server{
		ServerSettings: settings,
	}
	return s
}

// Network implements proxy.Inbound.
func (*Server) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

type readerOnly struct {
	io.Reader
}

func (s *Server) Process(ctx context.Context, conn net.Conn) error {
	reader := bufio.NewReaderSize(readerOnly{conn}, buf.Size)

Start:
	if err := conn.SetReadDeadline(time.Now().Add(s.PolicyManager.HandshakeTimeout())); err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to set deadline")
	}

	request, err := http.ReadRequest(reader)
	if err != nil {
		return fmt.Errorf("failed to read http request: %w", err)
	}

	log.Ctx(ctx).Debug().Str("method", request.Method).Str("host", request.Host).Str("url", request.URL.String()).Msg("http request")
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to clear read deadline")
	}

	defaultPort := net.Port(80)
	if strings.EqualFold(request.URL.Scheme, "https") {
		defaultPort = net.Port(443)
	}
	host := request.Host
	if host == "" {
		host = request.URL.Host
	}
	dest, err := http_proto.ParseHost(host, defaultPort)
	if err != nil {
		return fmt.Errorf("malformed proxy host: %w", err)
	}

	if strings.EqualFold(request.Method, "CONNECT") {
		return s.handleConnect(ctx, request, reader, conn, dest)
	}

	keepAlive := (strings.TrimSpace(strings.ToLower(request.Header.Get("Proxy-Connection"))) == "keep-alive")

	err = s.handlePlainHTTP(ctx, request, conn, dest)
	if err == errWaitAnother {
		if keepAlive {
			goto Start
		}
		err = nil
	}

	return err
}

func (s *Server) handleConnect(ctx context.Context, _ *http.Request, reader *bufio.Reader, conn net.Conn, dest net.Destination) error {
	_, err := conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		return errors.New("failed to write back OK response").Base(err)
	}

	r := buf.NewReader(conn)
	if reader.Buffered() > 0 {
		payload, err := buf.ReadFrom(io.LimitReader(reader, int64(reader.Buffered())))
		if err != nil {
			buf.ReleaseMulti(payload)
			return err
		}
		r = &buf.BufferedReader{Reader: r, Buffer: payload}
		reader = nil
	}
	return s.Handler.HandleFlow(ctx, dest, buf.NewRWD(r, buf.NewWriter(conn), conn))
}

var errWaitAnother = errors.New("keep alive")

func (s *Server) handlePlainHTTP(ctx context.Context, request *http.Request, writer io.Writer, dest net.Destination) error {
	if !s.AllowTransparent && request.URL.Host == "" {
		// RFC 2068 (HTTP/1.1) requires URL to be absolute URL in HTTP proxy.
		response := &http.Response{
			Status:        "Bad Request",
			StatusCode:    400,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        http.Header(make(map[string][]string)),
			Body:          nil,
			ContentLength: 0,
			Close:         true,
		}
		response.Header.Set("Proxy-Connection", "close")
		response.Header.Set("Connection", "close")
		return response.Write(writer)
	}

	if len(request.URL.Host) > 0 {
		request.Host = request.URL.Host
	}
	http_proto.RemoveHopByHopHeaders(request.Header)

	// Prevent UA from being set to golang's default ones
	if request.Header.Get("User-Agent") == "" {
		request.Header.Set("User-Agent", "")
	}

	// Plain HTTP request is not a stream. The request always finishes before response. Hence, request has to be closed later.
	var result error = errWaitAnother

	linkA, linkB := pipe.NewLinks(buf.BufferSize, false)
	defer linkA.Interrupt(nil)

	go func() {
		if err := s.Handler.HandleFlow(ctx, dest, linkB); err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to handle")
		}
	}()

	requestDone := func() error {
		request.Header.Set("Connection", "close")

		requestWriter := buf.NewBufferedWriter(linkA)
		common.Must(requestWriter.SetBuffered(false))
		if err := request.Write(requestWriter); err != nil {
			return fmt.Errorf("failed to write whole request: %w", err)
		}
		linkA.CloseWrite()
		return nil
	}

	responseDone := func() error {
		responseReader := bufio.NewReaderSize(&buf.BufferedReader{Reader: linkA}, buf.Size)
		response, err := http.ReadResponse(responseReader, request)
		if err == nil {
			http_proto.RemoveHopByHopHeaders(response.Header)
			if response.ContentLength >= 0 {
				response.Header.Set("Proxy-Connection", "keep-alive")
				response.Header.Set("Connection", "keep-alive")
				response.Header.Set("Keep-Alive", "timeout=4")
				response.Close = false
			} else {
				response.Close = true
				result = nil
			}
			defer response.Body.Close()
		} else {
			log.Ctx(ctx).Err(err).Str("host", request.Host).Msg("failed to read response")
			response = &http.Response{
				Status:        "Service Unavailable",
				StatusCode:    503,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        http.Header(make(map[string][]string)),
				Body:          nil,
				ContentLength: 0,
				Close:         true,
			}
			response.Header.Set("Connection", "close")
			response.Header.Set("Proxy-Connection", "close")
		}
		if err := response.Write(writer); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
		return nil
	}

	if err := task.Run(ctx, requestDone, responseDone); err != nil {
		return fmt.Errorf("connection ends: %w", err)
	}
	return result
}

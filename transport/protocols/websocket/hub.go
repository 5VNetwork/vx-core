package websocket

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	gonet "net"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	http_proto "github.com/5vnetwork/vx-core/common/protocol/http"
	"github.com/5vnetwork/vx-core/i"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// TODO: SRP
type requestHandler struct {
	path                string
	ln                  *Listener
	earlyDataEnabled    bool
	earlyDataHeaderName string
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:   4 * 1024,
	WriteBufferSize:  4 * 1024,
	HandshakeTimeout: time.Second * 4,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *requestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	responseHeader := http.Header{}

	var earlyData io.Reader
	if !h.earlyDataEnabled { // nolint: gocritic
		if request.URL.Path != h.path {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
	} else if h.earlyDataHeaderName != "" {
		if request.URL.Path != h.path {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		earlyDataStr := request.Header.Get(h.earlyDataHeaderName)
		earlyData = base64.NewDecoder(base64.RawURLEncoding, bytes.NewReader([]byte(earlyDataStr)))
		if strings.EqualFold("Sec-WebSocket-Protocol", h.earlyDataHeaderName) {
			responseHeader.Set(h.earlyDataHeaderName, earlyDataStr)
		}
	} else {
		if strings.HasPrefix(request.URL.RequestURI(), h.path) {
			earlyDataStr := request.URL.RequestURI()[len(h.path):]
			earlyData = base64.NewDecoder(base64.RawURLEncoding, bytes.NewReader([]byte(earlyDataStr)))
		} else {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
	}

	conn, err := upgrader.Upgrade(writer, request, responseHeader)
	if err != nil {
		log.Err(err).Msg("failed to upgrade to WebSocket connection")
		return
	}

	forwardedAddrs := http_proto.ParseXForwardedFor(request.Header)
	remoteAddr := conn.RemoteAddr()
	if len(forwardedAddrs) > 0 && forwardedAddrs[0].Family().IsIP() {
		remoteAddr = &net.TCPAddr{
			IP:   forwardedAddrs[0].IP(),
			Port: int(0),
		}
	}
	if earlyData == nil {
		h.ln.addConn(newConnection(conn, remoteAddr))
	} else {
		h.ln.addConn(newConnectionWithEarlyData(conn, remoteAddr, earlyData))
	}
}

type Listener struct {
	sync.Mutex
	server   http.Server
	listener net.Listener
	config   *WebsocketConfig
	addConn  func(net.Conn)
	closed   bool
}

func Listen(ctx context.Context, addr net.Destination,
	config *WebsocketConfig, li i.Listener, ch func(net.Conn)) (*Listener, error) {
	l := &Listener{
		addConn: ch,
	}
	l.config = config

	listener, err := li.Listen(ctx, addr.Addr())
	if err != nil {
		return nil, err
	}

	l.listener = listener
	useEarlyData := false
	earlyDataHeaderName := ""
	if config.MaxEarlyData != 0 {
		useEarlyData = true
		earlyDataHeaderName = config.EarlyDataHeaderName
	}

	l.server = http.Server{
		Handler: &requestHandler{
			path:                config.GetNormalizedPath(),
			ln:                  l,
			earlyDataEnabled:    useEarlyData,
			earlyDataHeaderName: earlyDataHeaderName,
		},
		ReadHeaderTimeout: time.Second * 4,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	}

	go func() {
		if err := l.server.Serve(l.listener); err != nil {
			if !l.closed || !errors.Is(err, gonet.ErrClosed) {
				log.Err(err).Msg("failed to serve http for WebSocket")
			}
		}
	}()

	return l, err
}

// Addr implements net.Listener.Addr().
func (ln *Listener) Addr() net.Addr {
	return ln.listener.Addr()
}

// Close implements net.Listener.Close().
func (ln *Listener) Close() error {
	ln.closed = true
	return ln.listener.Close()
}

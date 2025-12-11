package websocket

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	gonet "net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/security"
)

type websocketDialer struct {
	config       *WebsocketConfig
	engine       security.Engine
	socketConfig i.Dialer
}

func NewWebsocketDialer(config *WebsocketConfig, engine security.Engine, socketConfig i.Dialer) *websocketDialer {
	return &websocketDialer{
		config:       config,
		engine:       engine,
		socketConfig: socketConfig,
	}
}

func (d *websocketDialer) Dial(ctx context.Context, dest net.Destination) (net.Conn, error) {
	return Dial(ctx, dest, d.config, d.engine, d.socketConfig)
}

// Dial dials a WebSocket connection to the given destination.
func Dial(ctx context.Context, dest net.Destination, config *WebsocketConfig, securityConfig security.Engine, so i.Dialer) (net.Conn, error) {
	log.Ctx(ctx).Debug().Str("dst", dest.String()).Msg("websocket creating connection to ")

	dialer := &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return so.Dial(ctx, dest)
		},
		ReadBufferSize:   4 * 1024,
		WriteBufferSize:  4 * 1024,
		HandshakeTimeout: time.Second * 8,
	}

	protocol := "ws"

	if securityConfig != nil {
		protocol = "wss"

		dialer.NetDialTLSContext = func(ctx context.Context, network, addr string) (gonet.Conn, error) {
			conn, err := dialer.NetDial(network, addr)
			if err != nil {
				return nil, errors.New("dial TLS connection failed").Base(err)
			}
			conn, err = securityConfig.GetClientConn(conn,
				security.OptionWithDestination{Dest: dest},
				security.OptionWithALPN{ALPNs: []string{"http/1.1"}})
			if err != nil {
				return nil, errors.New("unable to create security protocol client from security engine").Base(err)
			}
			return conn, nil
		}
	}

	host := dest.NetAddr()
	if (protocol == "ws" && dest.Port == 80) || (protocol == "wss" && dest.Port == 443) {
		host = dest.Address.String()
	}
	uri := protocol + "://" + host + config.GetNormalizedPath()

	if config.MaxEarlyData != 0 {
		return newConnectionWithDelayedDial(&dialerWithEarlyData{
			dialer:  dialer,
			uriBase: uri,
			config:  config,
		}), nil
	}

	conn, resp, err := dialer.Dial(uri, config.GetRequestHeader()) // nolint: bodyclose
	if err != nil {
		var reason string
		if resp != nil {
			reason = resp.Status
		}
		return nil, errors.New("failed to dial to (", uri, "): ", reason).Base(err)
	}

	return newConnection(conn, conn.RemoteAddr()), nil
}

type dialerWithEarlyData struct {
	dialer  *websocket.Dialer
	uriBase string
	config  *WebsocketConfig
}

func (d dialerWithEarlyData) Dial(earlyData []byte) (*websocket.Conn, error) {
	earlyDataBuf := bytes.NewBuffer(nil)
	base64EarlyDataEncoder := base64.NewEncoder(base64.RawURLEncoding, earlyDataBuf)

	earlydata := bytes.NewReader(earlyData)
	limitedEarlyDatareader := io.LimitReader(earlydata, int64(d.config.MaxEarlyData))
	n, encerr := io.Copy(base64EarlyDataEncoder, limitedEarlyDatareader)
	if encerr != nil {
		return nil, errors.New("websocket delayed dialer cannot encode early data").Base(encerr)
	}

	if errc := base64EarlyDataEncoder.Close(); errc != nil {
		return nil, errors.New("websocket delayed dialer cannot encode early data tail").Base(errc)
	}

	dialFunction := func() (*websocket.Conn, *http.Response, error) {
		return d.dialer.Dial(d.uriBase+earlyDataBuf.String(), d.config.GetRequestHeader())
	}

	if d.config.EarlyDataHeaderName != "" {
		dialFunction = func() (*websocket.Conn, *http.Response, error) {
			earlyDataStr := earlyDataBuf.String()
			currentHeader := d.config.GetRequestHeader()
			currentHeader.Set(d.config.EarlyDataHeaderName, earlyDataStr)
			return d.dialer.Dial(d.uriBase, currentHeader)
		}
	}

	conn, resp, err := dialFunction() // nolint: bodyclose
	if err != nil {
		var reason string
		if resp != nil {
			reason = resp.Status
		}
		return nil, errors.New("failed to dial to (", d.uriBase, ") with early data: ", reason).Base(err)
	}
	if n != int64(len(earlyData)) {
		if errWrite := conn.WriteMessage(websocket.BinaryMessage, earlyData[n:]); errWrite != nil {
			return nil, errors.New("failed to dial to (", d.uriBase, ") with early data as write of remainder early data failed: ").Base(err)
		}
	}
	return conn, nil
}

package httpupgrade

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type server struct {
	config *HttpUpgradeConfig

	addConn        func(net.Conn)
	innnerListener net.Listener
}

func Listen(ctx context.Context, addr net.Destination,
	config *HttpUpgradeConfig, li i.Listener,
	ch func(net.Conn)) (*server, error) {
	serverInstance := &server{config: config, addConn: ch}

	listener, err := li.Listen(ctx, addr.Addr())
	if err != nil {
		return nil, err
	}
	serverInstance.innnerListener = listener
	go serverInstance.keepAccepting()
	return serverInstance, nil
}

func (s *server) Close() error {
	return s.innnerListener.Close()
}

func (s *server) Addr() net.Addr {
	return nil
}

func (s *server) Handle(conn net.Conn) (net.Conn, error) {
	connReader := bufio.NewReader(conn)
	req, err := http.ReadRequest(connReader)
	if err != nil {
		return nil, err
	}
	connection := strings.ToLower(req.Header.Get("Connection"))
	upgrade := strings.ToLower(req.Header.Get("Upgrade"))
	if connection != "upgrade" || upgrade != "websocket" {
		_ = conn.Close()
		return nil, errors.New("unrecognized request")
	}
	resp := &http.Response{
		Status:     "101 Switching Protocols",
		StatusCode: 101,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	resp.Header.Set("Connection", "upgrade")
	resp.Header.Set("Upgrade", "websocket")
	err = resp.Write(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if s.config.Config.MaxEarlyData != 0 {
		if s.config.Config.EarlyDataHeaderName == "" {
			return nil, errors.New("EarlyDataHeaderName is not set")
		}
		earlyData := req.Header.Get(s.config.Config.EarlyDataHeaderName)
		if earlyData != "" {
			earlyDataBytes, err := base64.URLEncoding.DecodeString(earlyData)
			if err != nil {
				return nil, err
			}
			return newConnectionWithPendingRead(conn, conn.RemoteAddr(), bytes.NewReader(earlyDataBytes)), nil
		}
	}
	return conn, nil
}

func (s *server) keepAccepting() {
	for {
		conn, err := s.innnerListener.Accept()
		if err != nil {
			return
		}
		handledConn, err := s.Handle(conn)
		if err != nil {
			log.Err(err).Msg("failed to handle request")
			continue
		}
		s.addConn(handledConn)
	}
}

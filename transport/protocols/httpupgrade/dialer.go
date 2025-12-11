package httpupgrade

import (
	"bufio"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/security"
)

type httpupgradeDialer struct {
	config       *HttpUpgradeConfig
	engine       security.Engine
	socketConfig i.Dialer
}

func NewHttpUpgradeDialer(config *HttpUpgradeConfig, engine security.Engine, socketSetting i.Dialer) *httpupgradeDialer {
	return &httpupgradeDialer{
		config:       config,
		engine:       engine,
		socketConfig: socketSetting,
	}
}

func (d *httpupgradeDialer) Dial(ctx context.Context, dest net.Destination) (net.Conn, error) {
	return dial(ctx, dest, d.config, d.engine, d.socketConfig)
}

func dialhttpUpgrade(ctx context.Context, dest net.Destination, config *HttpUpgradeConfig, se security.Engine, so i.Dialer) (net.Conn, error) {
	dialer := func(earlyData []byte) (net.Conn, io.Reader, error) {
		conn, err := so.Dial(ctx, dest)
		if err != nil {
			return nil, nil, err
		}
		if se != nil {
			conn, err = se.GetClientConn(conn, security.OptionWithDestination{Dest: dest},
				security.OptionWithALPN{ALPNs: []string{"http/1.1"}})
			if err != nil {
				return nil, nil, err
			}
		}

		req, err := http.NewRequest("GET", config.GetNormalizedPath(), nil)
		if err != nil {
			return nil, nil, err
		}

		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Host = config.Config.Host

		if config.Config.Header != nil {
			for _, value := range config.Config.Header {
				req.Header.Set(value.Key, value.Value)
			}
		}

		earlyDataSize := len(earlyData)
		if earlyDataSize > int(config.Config.MaxEarlyData) {
			earlyDataSize = int(config.Config.MaxEarlyData)
		}

		if len(earlyData) > 0 {
			if config.Config.EarlyDataHeaderName == "" {
				return nil, nil, errors.New("EarlyDataHeaderName is not set")
			}
			req.Header.Set(config.Config.EarlyDataHeaderName, base64.URLEncoding.EncodeToString(earlyData))
		}

		err = req.Write(conn)
		if err != nil {
			return nil, nil, err
		}

		if earlyData != nil && len(earlyData[earlyDataSize:]) > 0 {
			_, err = conn.Write(earlyData[earlyDataSize:])
			if err != nil {
				return nil, nil, errors.New("failed to finish write early data").Base(err)
			}
		}

		bufferedConn := bufio.NewReader(conn)
		resp, err := http.ReadResponse(bufferedConn, req) // nolint:bodyclose
		if err != nil {
			return nil, nil, err
		}

		if resp.Status == "101 Switching Protocols" &&
			strings.ToLower(resp.Header.Get("Upgrade")) == "websocket" &&
			strings.ToLower(resp.Header.Get("Connection")) == "upgrade" {
			earlyReplyReader := io.LimitReader(bufferedConn, int64(bufferedConn.Buffered()))
			return conn, earlyReplyReader, nil
		}

		return nil, nil, errors.New("unrecognized reply")
	}

	if config.Config.MaxEarlyData == 0 {
		conn, earlyReplyReader, err := dialer(nil)
		if err != nil {
			return nil, err
		}
		remoteAddr := conn.RemoteAddr()

		return newConnectionWithPendingRead(conn, remoteAddr, earlyReplyReader), nil
	}

	return newConnectionWithDelayedDial(dialer), nil
}

func dial(ctx context.Context, dest net.Destination, c *HttpUpgradeConfig, engine security.Engine, socketConfig i.Dialer) (net.Conn, error) {
	conn, err := dialhttpUpgrade(ctx, dest, c, engine, socketConfig)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

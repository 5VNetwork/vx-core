package proxy

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type connHandler interface {
	Process(ctx context.Context, conn net.Conn) error
}

type tcpWorker struct {
	addr        net.Addr
	connHandler connHandler
	tag         string
	listener    i.Listener
	netListener net.Listener
}

func (w *tcpWorker) Start() error {
	listener, err := w.listener.Listen(context.Background(), w.addr)
	if err != nil {
		return fmt.Errorf("cannot create a listener on %s: %w", w.addr.String(), err)
	}
	log.Debug().Str("address", w.addr.String()).Msg("tcp listening")

	w.netListener = listener
	go w.keepAccepting()
	return nil
}

func (w *tcpWorker) keepAccepting() {
	for {
		conn, err := w.netListener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("failed to accept connection")
			return
		}
		go w.handleConn(conn)
	}
}

func (w *tcpWorker) handleConn(conn net.Conn) {
	ctx, cancel := inbound.GetCtx(
		net.DestinationFromAddr(conn.RemoteAddr()),
		net.DestinationFromAddr(w.addr), w.tag)
	ctx = inbound.ContextWithRawConn(ctx, conn)
	err := w.connHandler.Process(ctx, conn)
	if err != nil && !errors.Is(err, errors.ErrIdle) {
		log.Ctx(ctx).Debug().Err(err).Send()
	}

	cancel(err)
	conn.Close()
}

func (w *tcpWorker) Close() error {
	var errorList []error
	if w.netListener != nil {
		if err := w.netListener.Close(); err != nil {
			errorList = append(errorList, err)
		}
		if err := common.Close(w.connHandler); err != nil {
			errorList = append(errorList, err)
		}
	}
	if len(errorList) > 0 {
		return errors.Join(errorList...)
	}
	return nil
}

type proxyServers struct {
	fallbackProxyServers []FallbackProxyServer
	proxyServer          ProxyServer
}

func (w *proxyServers) Close() error {
	var errs []error
	for _, server := range w.fallbackProxyServers {
		errs = append(errs, common.Close(server))
	}
	if w.proxyServer != nil {
		errs = append(errs, common.Close(w.proxyServer))
	}
	return errors.Join(errs...)
}

func (w *proxyServers) Process(ctx context.Context, conn net.Conn) error {
	cachConn := net.NewMbConn(conn, nil)
	defer buf.ReleaseMulti(cachConn.Mb)
	for _, fallbackProxyServer := range w.fallbackProxyServers {
		okToFallback, cached, err := fallbackProxyServer.FallbackProcess(ctx, cachConn)
		if okToFallback {
			log.Ctx(ctx).Debug().Err(err).Msg("fallback")
			cachConn.Mb, _ = buf.MergeMulti(cached, cachConn.Mb)
			continue
		}
		return err
	}
	if w.proxyServer != nil {
		return w.proxyServer.Process(ctx, cachConn)
	}

	return errors.New("no proxy server to handle the conn")
}

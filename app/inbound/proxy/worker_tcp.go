// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type ConnHandler interface {
	Process(ctx context.Context, conn net.Conn) error
}

type tcpWorker struct {
	addr        net.Addr
	connHandler ConnHandler
	tag         string
	listener    i.Listener
	netListener net.Listener
}

type TcpWorkerConfig struct {
	Addr        net.Addr
	Listener    i.Listener
	Tag         string
	ConnHandler ConnHandler
}

func NewTcpWorker(config TcpWorkerConfig) *tcpWorker {
	return &tcpWorker{
		addr:        config.Addr,
		listener:    config.Listener,
		tag:         config.Tag,
		connHandler: config.ConnHandler,
	}

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

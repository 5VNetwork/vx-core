// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"errors"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type ProxyServer interface {
	// Network returns a list of networks that this inbound supports. Connections with not-supported networks will not be passed into Process().
	Network() []net.Network
	Process(context.Context, net.Conn) error
}

type FallbackProxyServer interface {
	ProxyServer
	// if okToFallback, it means all data (if any) that has been read from net.Conn is cached in cached and there
	// is no write into the conn, therefore, okay to fallback. cached might be nil; err is the reason for not able to processs.
	//
	// if not okToFallback, then cached is nil, and err is same as Process()
	FallbackProcess(context.Context, net.Conn) (okToFallback bool, cached buf.MultiBuffer, err error)
}

type ProxyServers struct {
	FallbackProxyServers []FallbackProxyServer
	ProxyServer          ProxyServer
}

func (w *ProxyServers) Close() error {
	var errs []error
	for _, server := range w.FallbackProxyServers {
		errs = append(errs, common.Close(server))
	}
	if w.ProxyServer != nil {
		errs = append(errs, common.Close(w.ProxyServer))
	}
	return errors.Join(errs...)
}

func (w *ProxyServers) Process(ctx context.Context, conn net.Conn) error {
	cachConn := net.NewMbConn(conn, nil)
	defer buf.ReleaseMulti(cachConn.Mb)
	for _, fallbackProxyServer := range w.FallbackProxyServers {
		okToFallback, cached, err := fallbackProxyServer.FallbackProcess(ctx, cachConn)
		if okToFallback {
			log.Ctx(ctx).Debug().Err(err).Msg("fallback")
			cachConn.Mb, _ = buf.MergeMulti(cached, cachConn.Mb)
			continue
		}
		return err
	}
	if w.ProxyServer != nil {
		return w.ProxyServer.Process(ctx, cachConn)
	}

	return errors.New("no proxy server to handle the conn")
}

// implements i.InboundHandler
type ProxyInbound struct {
	tag string
	// workers contain tcpWorker, udpWorker, hysteira inbound
	workers     []common.Runnable
	userManages []UserManage
}

func NewProxyInbound(tag string) *ProxyInbound {
	return &ProxyInbound{
		tag: tag,
	}
}

func (h *ProxyInbound) AddWorker(worker common.Runnable) {
	h.workers = append(h.workers, worker)
}

func (h *ProxyInbound) AddUserManage(userManage UserManage) {
	h.userManages = append(h.userManages, userManage)
}

// Start implements common.Runnable.
func (h *ProxyInbound) Start() error {
	for _, worker := range h.workers {
		if err := worker.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Close implements common.Closable.
func (h *ProxyInbound) Close() error {
	var errs []error
	for _, worker := range h.workers {
		errs = append(errs, worker.Close())
	}
	return errors.Join(errs...)
}

func (h *ProxyInbound) Tag() string {
	return h.tag
}

func (h *ProxyInbound) AddUser(user i.User) {
	for _, s := range h.userManages {
		s.AddUser(user)
	}

}

func (h *ProxyInbound) RemoveUser(user i.User) {
	for _, s := range h.userManages {
		s.RemoveUser(user)
	}
}

func (h *ProxyInbound) WithOnUnauthorizedRequest(f i.UnauthorizedReport) {
	for _, s := range h.userManages {
		s.WithOnUnauthorizedRequest(f)
	}
}

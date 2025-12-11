package proxy

import (
	"context"
	"errors"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
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

// implements i.InboundHandler
type ProxyInbound struct {
	tag string
	// workers contain tcpWorker, udpWorker, hysteira inbound
	workers     []worker
	userManages []UserManage
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

func (h *ProxyInbound) RemoveUser(uid, secret string) {
	for _, s := range h.userManages {
		s.RemoveUser(uid, secret)
	}
}

func (h *ProxyInbound) WithOnUnauthorizedRequest(f i.UnauthorizedReport) {
	for _, s := range h.userManages {
		s.WithOnUnauthorizedRequest(f)
	}
}

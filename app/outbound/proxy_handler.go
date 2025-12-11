package outbound

import (
	"context"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/mux"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/uot"
	"github.com/5vnetwork/vx-core/i"
)

// implement outbound.ProxyHandler and transport.Dialer
type ProxyHandler struct {
	i.Handler
	tag              string
	muxClientManager *mux.ClientManager
	uot              bool
}

type ProxyHandlerSettings struct {
	Tag       string
	Handler   i.Handler
	Uot       bool
	EnableMux bool
	MuxConfig mux.ClientStrategy
}

func NewProxyHandler(settings ProxyHandlerSettings) *ProxyHandler {
	h := &ProxyHandler{
		tag:     settings.Tag,
		Handler: settings.Handler,
		uot:     settings.Uot,
	}
	if settings.EnableMux {
		h.muxClientManager = mux.NewClientManager(settings.MuxConfig, h.Handler)
	}
	return h
}

func (h *ProxyHandler) Start() error {
	if h.muxClientManager != nil {
		h.muxClientManager.Start()
	}
	return common.Start(h.Handler)
}

func (h *ProxyHandler) Close() error {
	if h.muxClientManager != nil {
		h.muxClientManager.Close()
	}
	return common.Close(h.Handler)
}

func (h *ProxyHandler) Tag() string {
	return h.tag
}

func (h *ProxyHandler) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	if dst.Network == net.Network_UDP && h.uot {
		rw = uot.NewUotReaderWriter(rw, dst)
		dst = net.Destination{
			Address: uot.Addr,
			Network: net.Network_TCP,
		}
	}

	if h.muxClientManager != nil && dst.Address != mux.MuxCoolAddressDst {
		return h.muxClientManager.HandleReaderWriter(ctx, dst, rw)
	}
	return h.Handler.HandleFlow(ctx, dst, rw)
}

func (h *ProxyHandler) HandlePacketConn(ctx context.Context, dst net.Destination,
	rw udp.PacketReaderWriter) error {
	if h.uot {
		r := uot.NewUotPacketReaderWriter(rw, dst)
		dst = net.Destination{
			Address: uot.Addr,
			Network: net.Network_TCP,
		}
		return h.HandleFlow(ctx, dst, r)
	}
	return h.Handler.HandlePacketConn(ctx, dst, rw)
}

// func (h *ProxyHandler) DialFlowConn(ctx context.Context, dst net.Destination) (i.FlowConn, error) {
// 	return h.DialFlowConnWithInitialData(ctx, dst, nil)
// }

// func (h *ProxyHandler) DialFlowConnWithInitialData(ctx context.Context, dst net.Destination,
// 	initialData buf.MultiBuffer) (i.FlowConn, error) {
// 	if h.muxClientManager != nil && dst.Address != mux.MuxCoolAddressDst {
// 		l1, l2 := pipe.NewLinks(buf.Size, false)
// 		go func() {
// 			err := h.muxClientManager.HandleReaderWriter(ctx, dst, l2)
// 			if err != nil {
// 				l1.Interrupt(err)
// 				log.Ctx(ctx).Error().Err(err).Msg("failed to handle mux")
// 			}
// 		}()
// 		if initialData.Len() > 0 {
// 			err := l1.WriteMultiBuffer(initialData)
// 			if err != nil {
// 				l1.Interrupt(err)
// 				return nil, err
// 			}
// 		}
// 		return proxy.NewFlowConn(proxy.FlowConnOption{
// 			Reader:      l1,
// 			Writer:      l1,
// 			SetDeadline: l1,
// 			Close:       func() error { l1.Interrupt(nil); return nil },
// 		}), nil
// 	}

// 	if d, ok := h.Handler.(i.FlowConnDailerInitialData); ok {
// 		return d.DialFlowConnWithInitialData(ctx, dst, initialData)
// 	}

// 	var err error
// 	var c i.FlowConn
// 	if d, ok := h.Handler.(i.FlowConnDialer); ok {
// 		c, err = d.DialFlowConn(ctx, dst)
// 	} else {
// 		d := util.FlowHandlerToDialer{FlowHandler: h.Handler}
// 		var conn net.Conn
// 		conn, err = d.Dial(ctx, dst)
// 		if err == nil {
// 			c = conn.(i.FlowConn)
// 		}
// 	}
// 	if err != nil {
// 		buf.ReleaseMulti(initialData)
// 		return nil, err
// 	}
// 	if initialData.Len() > 0 {
// 		err = c.WriteMultiBuffer(initialData)
// 		if err != nil {
// 			c.Close()
// 			return nil, err
// 		}
// 	}
// 	return c, nil
// }

// func (h *ProxyHandler) ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error) {
// 	if l, ok := h.Handler.(i.ProxyPacketListener); ok {
// 		return l.ListenPacket(ctx, dst)
// 	}
// 	d := util.HandlerToProxyClient{Handler: h.Handler}
// 	return d.ListenPacket(ctx, dst)
// }

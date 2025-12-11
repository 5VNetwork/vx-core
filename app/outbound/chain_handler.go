package outbound

import (
	"context"
	"errors"
	"fmt"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
)

type ChainHandler struct {
	tag      string
	handlers []i.Handler
}

type ChainHandlerConfig struct {
	*configs.ChainHandlerConfig
	Policy                      *policy.Policy
	IPResolver                  i.IPResolver
	DF                          transport.DialerFactory
	IPResolverForRequestAddress i.IPResolver
	// used to lookup ech config
	DnsServer  i.ECHResolver
	RejectQuic bool
}

func NewChainHandler(config *ChainHandlerConfig) (*ChainHandler, error) {
	if len(config.GetHandlers()) == 0 {
		return nil, errors.New("no handlers in chain")
	}
	finalHandler, err := NewOutHandler(&Config{
		OutboundHandlerConfig:       config.Handlers[0],
		DialerFactory:               config.DF,
		Policy:                      config.Policy,
		IPResolver:                  config.IPResolver,
		IPResolverForRequestAddress: config.IPResolverForRequestAddress,
		RejectQuic:                  config.RejectQuic,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create final handler: %w", err)
	}

	handlers := make([]i.Handler, len(config.Handlers))
	handlers[len(config.Handlers)-1] = finalHandler

	for i := 1; i < len(config.Handlers); i++ {
		handlerConfig := config.Handlers[i]
		handler, err := NewOutHandler(&Config{
			OutboundHandlerConfig: handlerConfig,
			DialerFactory: &transport.HandlerDialerFactory{
				Handler:   handlers[len(config.Handlers)-i],
				DnsServer: config.DnsServer,
			},
			Policy:                      config.Policy,
			IPResolver:                  config.IPResolver,
			IPResolverForRequestAddress: config.IPResolverForRequestAddress,
			RejectQuic:                  config.RejectQuic,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create handler %d: %w", i, err)
		}
		handlers[len(config.Handlers)-1-i] = handler
	}

	return &ChainHandler{
		tag:      config.Tag,
		handlers: handlers,
	}, nil
}

func (c *ChainHandler) Tag() string {
	return c.tag
}

func (c *ChainHandler) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	return c.handlers[0].HandleFlow(ctx, dst, rw)
}

func (c *ChainHandler) HandlePacketConn(ctx context.Context, dst net.Destination, p udp.PacketReaderWriter) error {
	return c.handlers[0].HandlePacketConn(ctx, dst, p)
}

// func (c *ChainHandler) Dial(ctx context.Context, dst net.Destination) (net.Conn, error) {
// 	return c.handlers[0].Dial(ctx, dst)
// }

// func (c *ChainHandler) ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error) {
// 	return c.handlers[0].ListenPacket(ctx, dst)
// }

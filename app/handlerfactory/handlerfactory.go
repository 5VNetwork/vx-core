package handlerfactory

import (
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
)

type HandlerFactory struct {
	DialerFactory               transport.DialerFactory
	Policy                      *policy.Policy
	IPResolver                  i.IPResolver
	IPResolverForRequestAddress i.IPResolver
	EchResolver                 i.ECHResolver
	Hysteria2RejectQuic         bool
	HandlerLinkStats            bool
	HandlerMeter                bool
}

func (c *HandlerFactory) CreateHandler(hs ...*configs.HandlerConfig) (i.Outbound, error) {
	df := c.DialerFactory
	var outbound i.Outbound
	if len(hs) > 1 {
		handlers := make([]*configs.OutboundHandlerConfig, 0)
		for _, handlerConfig := range hs {
			if handlerConfig.GetOutbound() != nil {
				handlers = append(handlers, handlerConfig.GetOutbound())
			} else if handlerConfig.GetChain() != nil {
				handlers = append(handlers, handlerConfig.GetChain().GetHandlers()...)
			}
		}

		var tags []string
		for _, handler := range hs {
			if handler.GetOutbound() != nil {
				tags = append(tags, handler.GetOutbound().GetTag())
			} else if handler.GetChain() != nil {
				tags = append(tags, handler.GetChain().GetTag())
			}
		}
		tag := strings.Join(tags, "-")

		ch, err := create.NewChainHandler(&create.ChainHandlerConfig{
			ChainHandlerConfig: &configs.ChainHandlerConfig{
				Tag:      tag,
				Handlers: handlers,
			},
			Policy:                      c.Policy,
			IPResolver:                  c.IPResolver,
			DF:                          c.DialerFactory,
			IPResolverForRequestAddress: c.IPResolverForRequestAddress,
			RejectQuic:                  c.Hysteria2RejectQuic,
			EchResolver:                 c.EchResolver,
		})
		if err != nil {
			return nil, err
		}
		outbound = ch
	} else {
		handler, err := create.NewHandler(&create.HandlerConfig{
			HandlerConfig:               hs[0],
			DialerFactory:               df,
			Policy:                      c.Policy,
			IPResolver:                  c.IPResolver,
			EchResolver:                 c.EchResolver,
			IPResolverForRequestAddress: c.IPResolverForRequestAddress,
			RejectQuic:                  c.Hysteria2RejectQuic,
		})
		if err != nil {
			return nil, err
		}
		outbound = handler
	}

	return outbound, nil
}

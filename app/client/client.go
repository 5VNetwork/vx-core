package client

import (
	"strconv"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/logger"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/app/subscription"
	"github.com/5vnetwork/vx-core/app/tester"
	"github.com/5vnetwork/vx-core/app/userlogger"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
)

type Client struct {
	// All components and Inbounds will be started and closed
	Components *common.Components
	Inbounds   []interface{}

	Dispatcher      *dispatcher.Dispatcher
	Geo             *geo.GeoWrapper
	Subscription    *subscription.SubscriptionManager
	DialerFactory   transport.DialerFactory
	Policy          *policy.Policy
	InboundManager  *proxy.InboundManager
	UserLogger      *userlogger.UserLogger
	OutboundManager *outbound.Manager
	// might be nil
	DB        Db
	Tetser    *tester.Tester
	Router    *router.RouterWrapper
	Selectors *selector.Selectors
	Logger    *logger.Logger
	// used to handle dns requests
	Dns *dns.Dns
	// used to resolve domains when dial, typically node address and domains of direct connection
	IPResolver i.IPResolver
	// used to resolve domains of proxied connections, typically used for converting domain to real ip for udp connections
	IPResolverForRequestAddress i.IPResolver
	IPToDomain                  *dns.IPToDomain
	FakeDnsEnabled              atomic.Bool
	AllFakeDns                  *dns.AllFakeDns
	Hysteria2RejectQuic         bool
}

func (c *Client) SetFakeDnsEnabled(enabled bool) {
	c.FakeDnsEnabled.Store(enabled)
}

func (c *Client) GetFakeDnsEnabled() bool {
	return c.FakeDnsEnabled.Load()
}

type Db interface {
	selector.Db
	UpdateHandler(id int, m map[string]interface{}) error
}

func (c *Client) Start() error {
	components := []interface{}{c.Components}
	components = append(components, c.Inbounds...)
	err := common.StartAll(components...)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() error {
	var err error
	components := []interface{}{c.Components}
	components = append(c.Inbounds, components...)
	err = common.CloseAll(components...)
	c.Logger.Close()
	return err
}

func (c *Client) CreateHandler(h *configs.HandlerConfig, landHandlerIds []*xsqlite.OutboundHandler) (i.Outbound, error) {
	df := c.DialerFactory
	if len(landHandlerIds) > 0 {
		handlers := make([]*configs.OutboundHandlerConfig, 0)
		if h.GetOutbound() != nil {
			handlers = append(handlers, h.GetOutbound())
		} else if h.GetChain() != nil {
			handlers = append(handlers, h.GetChain().GetHandlers()...)
		}

		for _, handler := range landHandlerIds {
			handlerConfig := handler.ToConfig()
			if handlerConfig.GetOutbound() != nil {
				handlers = append(handlers, handlerConfig.GetOutbound())
			} else if handlerConfig.GetChain() != nil {
				handlers = append(handlers, handlerConfig.GetChain().GetHandlers()...)
			}
		}

		tag := h.GetTag()
		for _, id := range landHandlerIds {
			tag = tag + "-" + strconv.Itoa(id.ID)
		}

		ch, err := outbound.NewChainHandler(&outbound.ChainHandlerConfig{
			ChainHandlerConfig: &configs.ChainHandlerConfig{
				Tag:      tag,
				Handlers: handlers,
			},
			Policy:                      c.Policy,
			IPResolver:                  c.IPResolver,
			DF:                          c.DialerFactory,
			IPResolverForRequestAddress: c.IPResolverForRequestAddress,
			RejectQuic:                  c.Hysteria2RejectQuic,
			DnsServer:                   dns.NewDnsServerToResolver(&dns.DnsToDnsServer{Dns: c.Dns}),
		})
		if err != nil {
			return nil, err
		}
		return ch, nil
	}

	return outbound.NewHandler(&outbound.HandlerConfig{
		HandlerConfig:               h,
		DialerFactory:               df,
		Policy:                      c.Policy,
		IPResolver:                  c.IPResolver,
		IPResolverForRequestAddress: c.IPResolverForRequestAddress,
		RejectQuic:                  c.Hysteria2RejectQuic,
		DnsServer:                   dns.NewDnsServerToResolver(&dns.DnsToDnsServer{Dns: c.Dns}),
	})
}

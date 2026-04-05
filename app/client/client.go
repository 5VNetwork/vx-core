// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package client

import (
	"strconv"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/grpcserver"
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

	NetMon          i.DefaultInterfaceInfo
	Dispatcher      *dispatcher.Dispatcher
	Geo             *geo.GeoWrapper
	Subscription    *subscription.SubscriptionManager
	DialerFactory   transport.DialerFactory
	Policy          *policy.Policy
	InboundManager  *proxy.InboundManager
	UserLogger      *userlogger.UserLogger
	OutboundManager *outbound.Manager
	OutStats        *dispatcher.OutStats
	// might be nil
	DB         Db
	Tetser     *tester.Tester
	Router     *router.RouterWrapper
	Selectors  *selector.Selectors
	Logger     *logger.Logger
	GrpcServer *grpcserver.GrpcServer
	// used to handle dns requests
	Dns *dns.Dns
	// used to resolve domains when dial, typically node address and domains of direct connection
	IPResolver i.IPResolver
	// used to resolve ech config
	EchResolver i.ECHResolver
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

func (c *Client) CreateHandler(h *configs.HandlerConfig) (i.Outbound, error) {
	df := c.DialerFactory

	return create.NewHandler(&create.HandlerConfig{
		HandlerConfig:               h,
		DialerFactory:               df,
		Policy:                      c.Policy,
		IPResolver:                  c.IPResolver,
		EchResolver:                 c.EchResolver,
		IPResolverForRequestAddress: c.IPResolverForRequestAddress,
		RejectQuic:                  c.Hysteria2RejectQuic,
	})
}

func (c *Client) CreateHandlerWithLandHandlers(h *configs.HandlerConfig, landHandlerIds []*xsqlite.OutboundHandler) (i.Outbound, error) {
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

		tag := configs.HandlerTag(h)
		for _, id := range landHandlerIds {
			tag = tag + "-" + strconv.Itoa(id.ID)
		}

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
		return ch, nil
	}

	return create.NewHandler(&create.HandlerConfig{
		HandlerConfig:               h,
		DialerFactory:               df,
		Policy:                      c.Policy,
		IPResolver:                  c.IPResolver,
		EchResolver:                 c.EchResolver,
		IPResolverForRequestAddress: c.IPResolverForRequestAddress,
		RejectQuic:                  c.Hysteria2RejectQuic,
	})
}

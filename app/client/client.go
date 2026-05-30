// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package client

import (
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/grpcserver"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/logger"
	"github.com/5vnetwork/vx-core/app/outbound"
	outboundstats "github.com/5vnetwork/vx-core/app/outbound/stats"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/app/subscription"
	"github.com/5vnetwork/vx-core/app/userlogger"
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
	Geo             *geo.Geo
	Subscription    *subscription.SubscriptionManager
	DialerFactory   transport.DialerFactory
	Policy          *policy.Policy
	InboundManager  *proxy.InboundManager
	UserLogger      *userlogger.UserLogger
	OutboundManager *outbound.Manager
	OutStats        *outboundstats.OutStats
	// might be nil
	DB         Db
	Tester     selector.Tester
	Router     *router.RouterWrapper
	Selectors  *selector.Selectors
	Logger     *logger.Logger
	GrpcServer *grpcserver.GrpcServer
	// used to handle dns requests
	Dns           *dns.HijackDns
	AllDnsServers *dns.AllDnsServers
	// used to resolve domains when dial, typically node address and domains of direct connection
	IPResolver i.IPResolver
	// used to resolve ech config
	EchResolver i.ECHResolver
	// used to resolve domains of proxied connections, typically used for converting domain to real ip for udp connections
	IPResolverForRequestAddress i.IPResolver
	IPToDomain                  *dns.IPToDomain
	HandlerFactory              i.HandlerFactory
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

type Db interface {
	selector.Db
	UpdateHandler(id int, m map[string]interface{}) error
}

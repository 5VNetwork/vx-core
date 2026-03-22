// Package configs provides type aliases to buf.build generated protobuf modules,
// preserving the historical import path used across vx-core.
package configs

import (
	dispatcherpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/dispatcher"
	dnspb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/dns"
	geopb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/geo"
	grpcsvc "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/grpc"
	inboundpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/inbound"
	vxlog "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/log"
	outboundpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/outbound"
	routerpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/router"
	subscriptionpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/subscription"
	sysproxypb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/sysproxy"
	transportpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/transport"
	tunpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/tun"
	userpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/user"
	vx "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx"
)

// Root (vx)
type (
	PolicyConfig = vx.PolicyConfig
	ServerConfig = vx.ServerConfig
	TmConfig     = vx.TmConfig
	UserPolicy   = vx.UserPolicy
)

// DNS
type (
	DnsConfig                      = dnspb.DnsConfig
	DnsRuleConfig                  = dnspb.DnsRuleConfig
	DnsRules                       = dnspb.DnsRules
	DnsServerConfig                = dnspb.DnsServerConfig
	DnsServerConfig_DohDnsServer   = dnspb.DnsServerConfig_DohDnsServer
	DnsServerConfig_FakeDnsServer  = dnspb.DnsServerConfig_FakeDnsServer
	DnsServerConfig_PlainDnsServer = dnspb.DnsServerConfig_PlainDnsServer
	DnsServerConfig_QuicDnsServer  = dnspb.DnsServerConfig_QuicDnsServer
	DnsServerConfig_TlsDnsServer   = dnspb.DnsServerConfig_TlsDnsServer
	DnsType                        = dnspb.DnsType
	DohDnsServer                   = dnspb.DohDnsServer
	FakeDnsServer                  = dnspb.FakeDnsServer
	FakeDnsServer_PoolConfig       = dnspb.FakeDnsServer_PoolConfig
	PlainDnsServer                 = dnspb.PlainDnsServer
	QuicDnsServer                  = dnspb.QuicDnsServer
	Record                         = dnspb.Record
	TlsDnsServer                   = dnspb.TlsDnsServer
)

// Geo
type (
	AppSetConfig          = geopb.AppSetConfig
	AtomicDomainSetConfig = geopb.AtomicDomainSetConfig
	AtomicIPSetConfig     = geopb.AtomicIPSetConfig
	DomainSetConfig       = geopb.DomainSetConfig
	GeoConfig             = geopb.GeoConfig
	GeoIPConfig           = geopb.GeoIPConfig
	GeositeConfig         = geopb.GeositeConfig
	GreatDomainSetConfig  = geopb.GreatDomainSetConfig
	GreatIPSetConfig      = geopb.GreatIPSetConfig
)

// Inbound
type (
	InboundManagerConfig                         = inboundpb.InboundManagerConfig
	MultiProxyInboundConfig                      = inboundpb.MultiProxyInboundConfig
	MultiProxyInboundConfig_Protocol             = inboundpb.MultiProxyInboundConfig_Protocol
	MultiProxyInboundConfig_Protocol_Grpc        = inboundpb.MultiProxyInboundConfig_Protocol_Grpc
	MultiProxyInboundConfig_Protocol_Http        = inboundpb.MultiProxyInboundConfig_Protocol_Http
	MultiProxyInboundConfig_Protocol_Httpupgrade = inboundpb.MultiProxyInboundConfig_Protocol_Httpupgrade
	MultiProxyInboundConfig_Protocol_Splithttp   = inboundpb.MultiProxyInboundConfig_Protocol_Splithttp
	MultiProxyInboundConfig_Protocol_Tcp         = inboundpb.MultiProxyInboundConfig_Protocol_Tcp
	MultiProxyInboundConfig_Protocol_Websocket   = inboundpb.MultiProxyInboundConfig_Protocol_Websocket
	MultiProxyInboundConfig_Security             = inboundpb.MultiProxyInboundConfig_Security
	MultiProxyInboundConfig_Security_Reality     = inboundpb.MultiProxyInboundConfig_Security_Reality
	MultiProxyInboundConfig_Security_Tls         = inboundpb.MultiProxyInboundConfig_Security_Tls
	ProxyInboundConfig                           = inboundpb.ProxyInboundConfig
	WfpConfig                                    = inboundpb.WfpConfig
)

// Outbound
type (
	ChainHandlerConfig     = outboundpb.ChainHandlerConfig
	DomainStrategy         = outboundpb.DomainStrategy
	HandlerConfig          = outboundpb.HandlerConfig
	HandlerConfig_Chain    = outboundpb.HandlerConfig_Chain
	HandlerConfig_Outbound = outboundpb.HandlerConfig_Outbound
	HandlerConfigs         = outboundpb.HandlerConfigs
	MuxConfig              = outboundpb.MuxConfig
	OutboundConfig         = outboundpb.OutboundConfig
	OutboundHandlerConfig  = outboundpb.OutboundHandlerConfig
)

// Router
type (
	AppId                            = routerpb.AppId
	AppId_Type                       = routerpb.AppId_Type
	RouterConfig                     = routerpb.RouterConfig
	RuleConfig                       = routerpb.RuleConfig
	RuleConfig_Fallback              = routerpb.RuleConfig_Fallback
	RuleConfig_Fallback_Action       = routerpb.RuleConfig_Fallback_Action
	SelectorConfig                   = routerpb.SelectorConfig
	SelectorConfig_BalanceStrategy     = routerpb.SelectorConfig_BalanceStrategy
	SelectorConfig_Filter            = routerpb.SelectorConfig_Filter
	SelectorConfig_SelectingStrategy = routerpb.SelectorConfig_SelectingStrategy
	SelectorsConfig                  = routerpb.SelectorsConfig
)

// Transport
type (
	SocketConfig                   = transportpb.SocketConfig
	SocketConfig_TCPFastOpenState  = transportpb.SocketConfig_TCPFastOpenState
	SocketConfig_TProxyMode        = transportpb.SocketConfig_TProxyMode
	TransportConfig                = transportpb.TransportConfig
	TransportConfig_Grpc           = transportpb.TransportConfig_Grpc
	TransportConfig_Http           = transportpb.TransportConfig_Http
	TransportConfig_Httpupgrade    = transportpb.TransportConfig_Httpupgrade
	TransportConfig_Kcp            = transportpb.TransportConfig_Kcp
	TransportConfig_Reality        = transportpb.TransportConfig_Reality
	TransportConfig_Splithttp      = transportpb.TransportConfig_Splithttp
	TransportConfig_Tcp            = transportpb.TransportConfig_Tcp
	TransportConfig_Tls            = transportpb.TransportConfig_Tls
	TransportConfig_Websocket      = transportpb.TransportConfig_Websocket
)

// User
type (
	UserConfig        = userpb.UserConfig
	UserManagerConfig = userpb.UserManagerConfig
)

// Tun
type (
	TunConfig              = tunpb.TunConfig
	TunDeviceConfig        = tunpb.TunDeviceConfig
	TunConfig_TUN46Setting = tunpb.TunConfig_TUN46Setting
	Mode                   = tunpb.Mode
)

// Log
type (
	LoggerConfig     = vxlog.LoggerConfig
	UserLoggerConfig = vxlog.UserLoggerConfig
	Level            = vxlog.Level
)

// Dispatcher, subscription, sysproxy, gRPC service (client)
type (
	DispatcherConfig   = dispatcherpb.DispatcherConfig
	SubscriptionConfig = subscriptionpb.SubscriptionConfig
	SysProxyConfig       = sysproxypb.SysProxyConfig
	GrpcConfig           = grpcsvc.GrpcConfig
	GrpcServiceConfig    = grpcsvc.GrpcServiceConfig
)

const (
	DnsType_DnsType_A    = dnspb.DnsType_DnsType_A
	DnsType_DnsType_AAAA = dnspb.DnsType_DnsType_AAAA

	TunConfig_FOUR_ONLY = tunpb.TunConfig_FOUR_ONLY
	TunConfig_BOTH        = tunpb.TunConfig_BOTH
	TunConfig_DYNAMIC     = tunpb.TunConfig_DYNAMIC
	Mode_MODE_SYSTEM      = tunpb.Mode_MODE_SYSTEM
	Mode_MODE_GVISOR      = tunpb.Mode_MODE_GVISOR

	AppId_Keyword = routerpb.AppId_Keyword
	AppId_Prefix  = routerpb.AppId_Prefix
	AppId_Exact   = routerpb.AppId_Exact

	Level_DEBUG    = vxlog.Level_DEBUG
	Level_INFO     = vxlog.Level_INFO
	Level_WARN     = vxlog.Level_WARN
	Level_ERROR    = vxlog.Level_ERROR
	Level_FATAL    = vxlog.Level_FATAL
	Level_DISABLED = vxlog.Level_DISABLED
)

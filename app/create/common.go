// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package create

import (
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/5vnetwork/vx-core/transport/protocols/tcp"
	"github.com/sagernet/sing/common"
)

var OldTypeUrlToNewTypeUrl = map[string]string{
	"type.googleapis.com/x.proxy.Shadowsocks2022ClientConfig": "type.googleapis.com/vx.proxy.shadowsocks2022.Shadowsocks2022ClientConfig",
	"type.googleapis.com/x.proxy.Shadowsocks2022ServerConfig": "type.googleapis.com/vx.proxy.shadowsocks2022.Shadowsocks2022ServerConfig",
	"type.googleapis.com/x.proxy.ShadowsocksClientConfig":     "type.googleapis.com/vx.proxy.shadowsocks.ShadowsocksClientConfig",
	"type.googleapis.com/x.proxy.ShadowsocksServerConfig":     "type.googleapis.com/vx.proxy.shadowsocks.ShadowsocksServerConfig",
	"type.googleapis.com/x.proxy.VmessClientConfig":           "type.googleapis.com/vx.proxy.vmess.VmessClientConfig",
	"type.googleapis.com/x.proxy.VmessServerConfig":           "type.googleapis.com/vx.proxy.vmess.VmessServerConfig",
	"type.googleapis.com/x.proxy.TrojanClientConfig":          "type.googleapis.com/vx.proxy.trojan.TrojanClientConfig",
	"type.googleapis.com/x.proxy.TrojanServerConfig":          "type.googleapis.com/vx.proxy.trojan.TrojanServerConfig",
	"type.googleapis.com/x.proxy.SocksClientConfig":           "type.googleapis.com/vx.proxy.socks.SocksClientConfig",
	"type.googleapis.com/x.proxy.SocksServerConfig":           "type.googleapis.com/vx.proxy.socks.SocksServerConfig",
	"type.googleapis.com/x.proxy.VlessClientConfig":           "type.googleapis.com/vx.proxy.vless.VlessClientConfig",
	"type.googleapis.com/x.proxy.VlessServerConfig":           "type.googleapis.com/vx.proxy.vless.VlessServerConfig",
	"type.googleapis.com/x.proxy.Hysteria2ClientConfig":       "type.googleapis.com/vx.proxy.hysteria.Hysteria2ClientConfig",
	"type.googleapis.com/x.proxy.Hysteria2ServerConfig":       "type.googleapis.com/vx.proxy.hysteria.Hysteria2ServerConfig",
	"type.googleapis.com/x.proxy.AnytlsClientConfig":          "type.googleapis.com/vx.proxy.anytls.AnytlsClientConfig",
	"type.googleapis.com/x.proxy.AnytlsServerConfig":          "type.googleapis.com/vx.proxy.anytls.AnytlsServerConfig",
	"type.googleapis.com/x.proxy.DokodemoConfig":              "type.googleapis.com/vx.proxy.dokodemo.DokodemoConfig",
	"type.googleapis.com/x.proxy.HttpClientConfig":            "type.googleapis.com/vx.proxy.http.HttpClientConfig",
	"type.googleapis.com/x.proxy.HttpServerConfig":            "type.googleapis.com/vx.proxy.http.HttpServerConfig",
}

func TransportProtocolConfig(tc *configs.TransportConfig) interface{} {

	var protocolConfig interface{}
	switch c := tc.GetProtocol().(type) {
	case *configs.TransportConfig_Grpc:
		protocolConfig = c.Grpc
	case *configs.TransportConfig_Tcp:
		protocolConfig = c.Tcp
	case *configs.TransportConfig_Websocket:
		protocolConfig = c.Websocket
	case *configs.TransportConfig_Http:
		protocolConfig = c.Http
	case *configs.TransportConfig_Kcp:
		protocolConfig = c.Kcp
	case *configs.TransportConfig_Splithttp:
		protocolConfig = c.Splithttp
	case *configs.TransportConfig_Httpupgrade:
		protocolConfig = c.Httpupgrade
	}
	if protocolConfig == nil && tc.TransportProtocol != nil {
		msg, err := serial.GetInstanceOf(tc.TransportProtocol)
		common.Must(err)
		protocolConfig = msg
	}

	return protocolConfig
}

func TransportSecurityConfig(config interface{}) interface{} {
	var securityConfig interface{}
	switch c := config.(type) {
	case *configs.TransportConfig_Reality:
		securityConfig = c.Reality
	case *configs.TransportConfig_Tls:
		securityConfig = c.Tls
	}
	return securityConfig
}

func TransportConfigToMemoryConfig(config *configs.TransportConfig,
	readCounter, writeCounter *atomic.Uint64, dnsServer i.ECHResolver) *transport.Config {
	if config == nil {
		return &transport.Config{
			Protocol:  &tcp.TcpConfig{},
			Socket:    &dlhelper.SocketSetting{},
			DnsServer: dnsServer,
		}
	}
	return &transport.Config{
		Socket:    SocketConfigToMemoryConfig(config.GetSocket(), readCounter, writeCounter),
		Protocol:  TransportProtocolConfig(config),
		Security:  TransportSecurityConfig(config.GetSecurity()),
		DnsServer: dnsServer,
	}
}

func SocketConfigToMemoryConfig(config *configs.SocketConfig, readCounter, writeCounter *atomic.Uint64) *dlhelper.SocketSetting {
	if config == nil {
		return &dlhelper.SocketSetting{}
	}
	return &dlhelper.SocketSetting{
		Mark:                       config.Mark,
		Tfo:                        dlhelper.SocketConfig_TCPFastOpenState(config.Tfo),
		Tproxy:                     dlhelper.SocketConfig_TProxyMode(config.Tproxy),
		ReceiveOriginalDestAddress: config.ReceiveOriginalDestAddress,
		BindAddress:                config.BindAddress,
		BindPort:                   config.BindPort,
		AcceptProxyProtocol:        config.AcceptProxyProtocol,
		TcpKeepAliveInterval:       config.TcpKeepAliveInterval,
		TfoQueueLength:             config.TfoQueueLength,
		TcpKeepAliveIdle:           config.TcpKeepAliveIdle,
		BindToDevice4:              config.BindToDevice,
		BindToDevice6:              config.BindToDevice,
		RxBufSize:                  config.RxBufSize,
		TxBufSize:                  config.TxBufSize,
		ForceBufSize:               config.ForceBufSize,
		LocalAddr4:                 config.LocalAddr4,
		LocalAddr6:                 config.LocalAddr6,
		StatsReadCounter:           readCounter,
		StatsWriteCounter:          writeCounter,
		DialTimeout:                time.Duration(config.DialTimeout) * time.Second,
	}
}

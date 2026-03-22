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

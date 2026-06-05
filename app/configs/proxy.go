package configs

import (
	anytlspb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/anytls"
	dokodemopb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/dokodemo"
	freedompb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/freedom"
	httppb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/http"
	hysteriapb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/hysteria"
	sspb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/shadowsocks"
	shadowsocks2022pb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/shadowsocks2022"
	sockspb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/socks"
	trojanpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/trojan"
	vlesspb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/vless"
	vmesspb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/vmess"
	wireguardpb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/wireguard"
)

type (
	// Enums
	AuthType              = sockspb.AuthType
	SecurityType          = vmesspb.SecurityType
	ShadowsocksCipherType = sspb.ShadowsocksCipherType

	// Client configs
	AnytlsClientConfig          = anytlspb.AnytlsClientConfig
	FreedomConfig               = freedompb.FreedomConfig
	HttpClientConfig            = httppb.HttpClientConfig
	HttpServerConfig            = httppb.HttpServerConfig
	Hysteria2ClientConfig       = hysteriapb.Hysteria2ClientConfig
	Shadowsocks2022ClientConfig = shadowsocks2022pb.Shadowsocks2022ClientConfig
	WireguardClientConfig       = wireguardpb.DeviceConfig
	ShadowsocksClientConfig     = sspb.ShadowsocksClientConfig
	SocksClientConfig           = sockspb.SocksClientConfig
	TrojanClientConfig          = trojanpb.TrojanClientConfig
	VlessClientConfig           = vlesspb.VlessClientConfig
	VmessClientConfig           = vmesspb.VmessClientConfig

	// Server configs
	AnytlsServerConfig          = anytlspb.AnytlsServerConfig
	Hysteria2ServerConfig       = hysteriapb.Hysteria2ServerConfig
	Shadowsocks2022ServerConfig = shadowsocks2022pb.Shadowsocks2022ServerConfig
	ShadowsocksServerConfig     = sspb.ShadowsocksServerConfig
	SocksServerConfig           = sockspb.SocksServerConfig
	TrojanServerConfig          = trojanpb.TrojanServerConfig
	VlessServerConfig           = vlesspb.VlessServerConfig
	VmessServerConfig           = vmesspb.VmessServerConfig

	// Shared / nested
	DokodemoConfig        = dokodemopb.DokodemoConfig
	Account               = httppb.Account
	BandwidthConfig       = hysteriapb.BandwidthConfig
	ObfsConfig            = hysteriapb.ObfsConfig
	ObfsConfig_Salamander = hysteriapb.ObfsConfig_Salamander
	SalamanderConfig      = hysteriapb.SalamanderConfig
)

const (
	AuthType_NO_AUTH  = sockspb.AuthType_NO_AUTH
	AuthType_PASSWORD = sockspb.AuthType_PASSWORD

	SecurityType_SecurityType_UNKNOWN           = vmesspb.SecurityType_SecurityType_UNKNOWN
	SecurityType_SecurityType_LEGACY            = vmesspb.SecurityType_SecurityType_LEGACY
	SecurityType_SecurityType_AUTO              = vmesspb.SecurityType_SecurityType_AUTO
	SecurityType_SecurityType_AES128_GCM        = vmesspb.SecurityType_SecurityType_AES128_GCM
	SecurityType_SecurityType_CHACHA20_POLY1305 = vmesspb.SecurityType_SecurityType_CHACHA20_POLY1305
	SecurityType_SecurityType_NONE              = vmesspb.SecurityType_SecurityType_NONE
	SecurityType_SecurityType_ZERO              = vmesspb.SecurityType_SecurityType_ZERO

	ShadowsocksCipherType_AES_128_GCM       = sspb.ShadowsocksCipherType_AES_128_GCM
	ShadowsocksCipherType_AES_256_GCM       = sspb.ShadowsocksCipherType_AES_256_GCM
	ShadowsocksCipherType_CHACHA20_POLY1305 = sspb.ShadowsocksCipherType_CHACHA20_POLY1305
	ShadowsocksCipherType_NONE              = sspb.ShadowsocksCipherType_NONE
)

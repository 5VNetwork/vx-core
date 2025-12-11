package util

import (
	"fmt"
	"slices"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/transport/protocols/grpc"
	"github.com/5vnetwork/vx-core/transport/protocols/http"
	"github.com/5vnetwork/vx-core/transport/protocols/httpupgrade"
	"github.com/5vnetwork/vx-core/transport/protocols/kcp"
	"github.com/5vnetwork/vx-core/transport/protocols/splithttp"
	"github.com/5vnetwork/vx-core/transport/protocols/tcp"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"golang.org/x/crypto/curve25519"
	"google.golang.org/protobuf/types/known/anypb"
)

func InboundConfigToOutboundConfig(namePrefix string, inboundConfig *configs.ProxyInboundConfig,
	serverAddress string) ([]*configs.OutboundHandlerConfig, error) {
	if len(inboundConfig.Users) == 0 {
		return nil, fmt.Errorf("no users")
	}
	user := inboundConfig.Users[0]

	clientProtocols, err := proxyProtocolToOutboundConfig(inboundConfig.Protocols, user)
	if err != nil {
		return nil, fmt.Errorf("failed to convert proxy protocol: %w", err)
	}
	var securityProtocols []any
	if inboundConfig.GetTransport().GetSecurity() != nil {
		if tlsConfig, ok := inboundConfig.Transport.Security.(*configs.TransportConfig_Tls); ok {
			clientConfigs, err := serverTlsConfigToClientTlsConfig(tlsConfig.Tls)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tls config: %w", err)
			}
			securityProtocols = append(securityProtocols, clientConfigs...)
		} else if realityConfig, ok := inboundConfig.Transport.Security.(*configs.TransportConfig_Reality); ok {
			clientConfigs, err := serverRealityConfigToClientRealityConfig(realityConfig.Reality)
			if err != nil {
				return nil, fmt.Errorf("failed to convert reality config: %w", err)
			}
			securityProtocols = append(securityProtocols, clientConfigs...)
		} else {
			return nil, fmt.Errorf("invalid security config")
		}
	}
	var returns []*configs.OutboundHandlerConfig
	for _, proxyProtocol := range clientProtocols {
		transportConfig := &configs.TransportConfig{
			Protocol: inboundConfig.GetTransport().GetProtocol(),
		}
		if securityProtocols != nil {
			for _, securityProtocol := range securityProtocols {
				switch s := securityProtocol.(type) {
				case *tls.TlsConfig:
					transportConfig.Security = &configs.TransportConfig_Tls{
						Tls: s,
					}
				case *reality.RealityConfig:
					transportConfig.Security = &configs.TransportConfig_Reality{
						Reality: s,
					}
				}
				returns = append(returns, &configs.OutboundHandlerConfig{
					Tag:       fmt.Sprintf("%s-%s-%d", namePrefix, inboundConfig.Tag, len(returns)+1),
					Address:   serverAddress,
					Ports:     getPortRanges(inboundConfig.Ports),
					Protocol:  proxyProtocol,
					Transport: transportConfig,
				})
			}
		} else {
			returns = append(returns, &configs.OutboundHandlerConfig{
				Tag:       fmt.Sprintf("%s-%s-%d", namePrefix, inboundConfig.Tag, len(returns)+1),
				Address:   serverAddress,
				Ports:     getPortRanges(inboundConfig.Ports),
				Protocol:  proxyProtocol,
				Transport: transportConfig,
			})
		}
	}
	return returns, nil
}

// unencrypted options will not be included in the final results.
// For example, unencrypted proxy protocols without security will not be included
func MultiInboundConfigToOutboundConfig(namePrefix string, inboundConfig *configs.MultiProxyInboundConfig,
	serverAddress string) ([]*configs.OutboundHandlerConfig, error) {
	if len(inboundConfig.Users) == 0 {
		return nil, fmt.Errorf("no users")
	}
	user := inboundConfig.Users[0]

	clientProtocols, err := proxyProtocolToOutboundConfig(inboundConfig.Protocols, user)
	if err != nil {
		return nil, fmt.Errorf("failed to convert proxy protocol: %w", err)
	}
	// transport protocols
	var transportProtocols []any
	for _, transportProtocol := range inboundConfig.TransportProtocols {
		switch p := transportProtocol.Protocol.(type) {
		case *configs.MultiProxyInboundConfig_Protocol_Websocket:
			transportProtocols = append(transportProtocols, p.Websocket)
		case *configs.MultiProxyInboundConfig_Protocol_Http:
			transportProtocols = append(transportProtocols, p.Http)
		case *configs.MultiProxyInboundConfig_Protocol_Grpc:
			transportProtocols = append(transportProtocols, p.Grpc)
		case *configs.MultiProxyInboundConfig_Protocol_Httpupgrade:
			transportProtocols = append(transportProtocols, p.Httpupgrade)
		case *configs.MultiProxyInboundConfig_Protocol_Splithttp:
			transportProtocols = append(transportProtocols, p.Splithttp)
		}
	}
	// security
	var securitys []any
	{
		var securityConfigs []*configs.MultiProxyInboundConfig_Security
		// check whether there is forced security config. If so, only use it
		if index := slices.IndexFunc(inboundConfig.SecurityConfigs,
			func(securityConfig *configs.MultiProxyInboundConfig_Security) bool {
				return securityConfig.Always
			}); index != -1 {
			securityConfigs = append(securityConfigs, inboundConfig.SecurityConfigs[index])
		} else {
			// no forced security config. add all security configs
			securityConfigs = append(securityConfigs, inboundConfig.SecurityConfigs...)
		}
		for _, securityConfig := range securityConfigs {
			if tlsConfig, ok := securityConfig.Security.(*configs.MultiProxyInboundConfig_Security_Tls); ok {
				clientConfigs, err := serverTlsConfigToClientTlsConfig(tlsConfig.Tls)
				if err != nil {
					return nil, fmt.Errorf("failed to convert tls config: %w", err)
				}
				securitys = append(securitys, clientConfigs...)
			} else if realityConfig, ok := securityConfig.Security.(*configs.MultiProxyInboundConfig_Security_Reality); ok {
				clientConfigs, err := serverRealityConfigToClientRealityConfig(realityConfig.Reality)
				if err != nil {
					return nil, fmt.Errorf("failed to convert reality config: %w", err)
				}
				securitys = append(securitys, clientConfigs...)
			} else {
				return nil, fmt.Errorf("invalid security config")
			}
		}
	}

	return generate(fmt.Sprintf("%s-%s", namePrefix, inboundConfig.Tag), serverAddress,
		getPortRanges(inboundConfig.Ports),
		clientProtocols, transportProtocols, securitys), nil
}

func proxyProtocolToOutboundConfig(proxyProtocol []*anypb.Any, user *configs.UserConfig) ([]*anypb.Any, error) {
	var clientProtocols []*anypb.Any
	// convert to client config
	for _, protocol := range proxyProtocol {
		protocolConfig, err := protocol.UnmarshalNew()
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal protocol: %w", err)
		}
		switch pc := protocolConfig.(type) {
		case *proxy.VmessServerConfig:
			clientProtocols = append(clientProtocols, serial.ToTypedMessage(&proxy.VmessClientConfig{
				Id: user.Secret,
			}))
		case *proxy.ShadowsocksServerConfig:
			clientProtocols = append(clientProtocols, serial.ToTypedMessage(&proxy.ShadowsocksClientConfig{
				Password:   user.Secret,
				CipherType: pc.CipherType,
			}))
		case *proxy.TrojanServerConfig:
			clientProtocols = append(clientProtocols, serial.ToTypedMessage(&proxy.TrojanClientConfig{
				Password: user.Secret,
				Vision:   pc.Vision,
			}))
		case *proxy.AnytlsServerConfig:
			clientProtocols = append(clientProtocols, serial.ToTypedMessage(&proxy.AnytlsClientConfig{
				Password: user.Secret,
			}))
		case *proxy.SocksServerConfig:
			clientProtocols = append(clientProtocols, serial.ToTypedMessage(&proxy.SocksClientConfig{
				Name:     user.Id,
				Password: user.Secret,
			}))
		case *proxy.Hysteria2ServerConfig:
			tlsConfigs, err := serverTlsConfigToClientTlsConfig(pc.TlsConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tls config: %w", err)
			}
			for _, tlsConfig := range tlsConfigs {
				clientProtocols = append(clientProtocols, serial.ToTypedMessage(&proxy.Hysteria2ClientConfig{
					TlsConfig: tlsConfig.(*tls.TlsConfig),
					Quic:      pc.Quic,
					Obfs:      pc.Obfs,
					Auth:      user.Secret,
					Bandwidth: &proxy.BandwidthConfig{
						MaxTx: 10,
						MaxRx: 10,
					},
				}))
			}
		}
	}
	return clientProtocols, nil
}

func generate(tag, address string, ports []*net.PortRange, proxyProtocols []*anypb.Any,
	transportProtocols []any, securityProtocols []any) []*configs.OutboundHandlerConfig {
	var returns []*configs.OutboundHandlerConfig

	for _, proxyProtocol := range proxyProtocols {
		// for each encrypted protocol, generate one config
		if strings.Contains(proxyProtocol.TypeUrl, "Shadowsocks") || strings.Contains(proxyProtocol.TypeUrl, "Vmess") {
			returns = append(returns, &configs.OutboundHandlerConfig{
				Tag:      fmt.Sprintf("%s-%d", tag, len(returns)+1),
				Address:  address,
				Ports:    ports,
				Protocol: proxyProtocol,
			})
		} else if strings.Contains(proxyProtocol.TypeUrl, "Hysteria") {
			returns = append(returns, &configs.OutboundHandlerConfig{
				Tag:      fmt.Sprintf("%s-%d", tag, len(returns)+1),
				Address:  address,
				Ports:    ports,
				Protocol: proxyProtocol,
			})
			continue
		}
		for _, transportProtocol := range transportProtocols {
			for _, securityProtocol := range securityProtocols {
				transportConfig := &configs.TransportConfig{}
				switch p := transportProtocol.(type) {
				case *websocket.WebsocketConfig:
					transportConfig.Protocol = &configs.TransportConfig_Websocket{
						Websocket: p,
					}
				case *http.HttpConfig:
					transportConfig.Protocol = &configs.TransportConfig_Http{
						Http: p,
					}
				case *grpc.GrpcConfig:
					transportConfig.Protocol = &configs.TransportConfig_Grpc{
						Grpc: p,
					}
				case *httpupgrade.HttpUpgradeConfig:
					transportConfig.Protocol = &configs.TransportConfig_Httpupgrade{
						Httpupgrade: p,
					}
				case *kcp.KcpConfig:
					transportConfig.Protocol = &configs.TransportConfig_Kcp{
						Kcp: p,
					}
				case *tcp.TcpConfig:
					transportConfig.Protocol = &configs.TransportConfig_Tcp{
						Tcp: p,
					}
				case *splithttp.SplitHttpConfig:
					transportConfig.Protocol = &configs.TransportConfig_Splithttp{
						Splithttp: p,
					}
				}
				switch s := securityProtocol.(type) {
				case *tls.TlsConfig:
					transportConfig.Security = &configs.TransportConfig_Tls{
						Tls: s,
					}
				case *reality.RealityConfig:
					transportConfig.Security = &configs.TransportConfig_Reality{
						Reality: s,
					}
				}
				returns = append(returns, &configs.OutboundHandlerConfig{
					Tag:       fmt.Sprintf("%s-%d", tag, len(returns)+1),
					Address:   address,
					Ports:     ports,
					Protocol:  proxyProtocol,
					Transport: transportConfig,
				})
			}
		}
		for _, securityProtocol := range securityProtocols {
			transportConfig := &configs.TransportConfig{}
			switch s := securityProtocol.(type) {
			case *tls.TlsConfig:
				transportConfig.Security = &configs.TransportConfig_Tls{
					Tls: s,
				}
			case *reality.RealityConfig:
				transportConfig.Security = &configs.TransportConfig_Reality{
					Reality: s,
				}
			}
			returns = append(returns, &configs.OutboundHandlerConfig{
				Tag:       fmt.Sprintf("%s-%d", tag, len(returns)+1),
				Address:   address,
				Ports:     ports,
				Protocol:  proxyProtocol,
				Transport: transportConfig,
			})
		}
	}

	return returns
}

func getPortRanges(ports []uint32) []*net.PortRange {
	var returns []*net.PortRange
	for _, port := range ports {
		if port == 0 {
			continue
		}
		returns = append(returns, &net.PortRange{From: port, To: port})
	}
	return returns
}

func serverRealityConfigToClientRealityConfig(realityConfig *reality.RealityConfig) ([]any, error) {
	var returns []any
	pubKey, err := curve25519.X25519(realityConfig.PrivateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pubkey: %w", err)
	}
	var shortId []byte
	if len(realityConfig.ShortIds) > 0 {
		shortId = realityConfig.ShortIds[0]
	}
	for _, dest := range realityConfig.ServerNames {
		returns = append(returns, &reality.RealityConfig{
			ServerName: dest,
			PublicKey:  pubKey,
			ShortId:    shortId,
		})
	}
	return returns, nil
}

func serverTlsConfigToClientTlsConfig(serverTlsConfig *tls.TlsConfig) ([]any, error) {
	if len(serverTlsConfig.Certificates) == 0 {
		return nil, fmt.Errorf("no certificates")
	}
	var returns []any
	for _, certificate := range serverTlsConfig.Certificates {
		domain, err := cert.ExtractDomainFromCertificate(certificate.Certificate)
		if err != nil {
			return nil, err
		}
		trustedBySystemCa, err := cert.TrustedBySystemCA(certificate.Certificate)
		if err != nil {
			return nil, err
		}
		var sha256Hash []byte
		var allowInsecure bool
		if !trustedBySystemCa {
			allowInsecure = true
			sha256Hash, err = cert.GetCertHash(certificate.Certificate)
			if err != nil {
				return nil, err
			}
		}
		returns = append(returns, &tls.TlsConfig{
			ServerName:                       domain,
			AllowInsecure:                    allowInsecure,
			EchConfig:                        serverTlsConfig.EchConfig,
			PinnedPeerCertificateChainSha256: [][]byte{sha256Hash},
		})
	}
	return returns, nil
}

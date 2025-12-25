package clashparser

import (
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	grpcProto "github.com/5vnetwork/vx-core/transport/protocols/grpc"
	httpProto "github.com/5vnetwork/vx-core/transport/protocols/http"
	"github.com/5vnetwork/vx-core/transport/protocols/tcp"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"
)

// parseTransportConfig converts mihomo transport settings to vx-core TransportConfig
func parseTransportConfig(mapping map[string]any) (*configs.TransportConfig, error) {
	config := &configs.TransportConfig{}

	// Get network type (tcp, ws, grpc, h2, etc.)
	network, _ := mapping["network"].(string)
	if network == "" {
		network = "tcp"
	}

	// Parse protocol/network type
	switch network {
	case "tcp":
		tcpConfig := &tcp.TcpConfig{}
		config.Protocol = &configs.TransportConfig_Tcp{Tcp: tcpConfig}

	case "ws":
		wsConfig := &websocket.WebsocketConfig{}

		// Parse ws-opts
		if wsOptsRaw, ok := mapping["ws-opts"]; ok {
			if wsOptsMap, ok := wsOptsRaw.(map[string]any); ok {
				path, _ := wsOptsMap["path"].(string)
				wsConfig.Path = path

				earlyDataHeaderName, _ := wsOptsMap["early-data-header-name"].(string)
				wsConfig.EarlyDataHeaderName = earlyDataHeaderName

				// Get host from headers
				if headersRaw, ok := wsOptsMap["headers"]; ok {
					if headersMap, ok := headersRaw.(map[string]any); ok {
						if host, ok := headersMap["Host"].(string); ok {
							wsConfig.Host = host
						}
					}
				}
			}
		}

		config.Protocol = &configs.TransportConfig_Websocket{Websocket: wsConfig}

	case "grpc":
		grpcConfig := &grpcProto.GrpcConfig{}

		// Parse grpc-opts
		if grpcOptsRaw, ok := mapping["grpc-opts"]; ok {
			if grpcOptsMap, ok := grpcOptsRaw.(map[string]any); ok {
				serviceName, _ := grpcOptsMap["grpc-service-name"].(string)
				grpcConfig.ServiceName = serviceName
			}
		}

		config.Protocol = &configs.TransportConfig_Grpc{Grpc: grpcConfig}

	case "h2", "http":
		httpConfig := &httpProto.HttpConfig{}

		// Parse h2-opts
		if h2OptsRaw, ok := mapping["h2-opts"]; ok {
			if h2OptsMap, ok := h2OptsRaw.(map[string]any); ok {
				path, _ := h2OptsMap["path"].(string)
				httpConfig.Path = path

				// Parse host list
				if hostRaw, ok := h2OptsMap["host"]; ok {
					switch h := hostRaw.(type) {
					case string:
						httpConfig.Host = []string{h}
					case []any:
						hosts := make([]string, 0, len(h))
						for _, hv := range h {
							if hs, ok := hv.(string); ok {
								hosts = append(hosts, hs)
							}
						}
						httpConfig.Host = hosts
					case []string:
						httpConfig.Host = h
					}
				}
			}
		}

		config.Protocol = &configs.TransportConfig_Http{Http: httpConfig}
	}

	// Parse TLS settings
	if tlsEnabled, _ := mapping["tls"].(bool); tlsEnabled {
		tlsConfig := &tls.TlsConfig{}

		// Server name / SNI
		if sni, ok := mapping["servername"].(string); ok && sni != "" {
			tlsConfig.ServerName = sni
		} else if sni, ok := mapping["sni"].(string); ok && sni != "" {
			tlsConfig.ServerName = sni
		}

		// Skip cert verify
		if skipVerify, ok := mapping["skip-cert-verify"].(bool); ok {
			tlsConfig.AllowInsecure = skipVerify
		}

		// Fingerprint
		if fp, ok := mapping["fingerprint"].(string); ok {
			tlsConfig.Imitate = fp
		}

		// ALPN
		if alpnRaw, ok := mapping["alpn"]; ok {
			switch alpn := alpnRaw.(type) {
			case []string:
				tlsConfig.NextProtocol = alpn
			case []any:
				alpns := make([]string, 0, len(alpn))
				for _, a := range alpn {
					if as, ok := a.(string); ok {
						alpns = append(alpns, as)
					}
				}
				tlsConfig.NextProtocol = alpns
			case string:
				// Single ALPN or comma-separated
				if strings.Contains(alpn, ",") {
					tlsConfig.NextProtocol = strings.Split(alpn, ",")
				} else {
					tlsConfig.NextProtocol = []string{alpn}
				}
			}
		}

		config.Security = &configs.TransportConfig_Tls{Tls: tlsConfig}
	}

	// Parse Reality settings (mihomo reality-opts)
	if realityOptsRaw, ok := mapping["reality-opts"]; ok {
		if realityOptsMap, ok := realityOptsRaw.(map[string]any); ok {
			realityConfig := &reality.RealityConfig{}

			// Public key (required)
			if publicKey, ok := realityOptsMap["public-key"].(string); ok {
				realityConfig.Pbk = publicKey
			}

			// Short ID (required, must be 8 bytes)
			if shortID, ok := realityOptsMap["short-id"].(string); ok {
				realityConfig.Sid = shortID
			}

			// Server name (SNI)
			if sni, ok := mapping["sni"].(string); ok && sni != "" {
				realityConfig.ServerName = sni
			} else if servername, ok := mapping["servername"].(string); ok && servername != "" {
				realityConfig.ServerName = servername
			}

			// Fingerprint
			if fp, ok := mapping["fingerprint"].(string); ok {
				realityConfig.Fingerprint = fp
			}

			config.Security = &configs.TransportConfig_Reality{Reality: realityConfig}
			return config, nil
		}
	}

	return config, nil
}

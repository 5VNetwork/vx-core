// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/any"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/util/sub"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/transport/headers/http"
	"github.com/5vnetwork/vx-core/transport/headers/srtp"
	"github.com/5vnetwork/vx-core/transport/headers/utp"
	"github.com/5vnetwork/vx-core/transport/headers/wechat"
	"github.com/5vnetwork/vx-core/transport/headers/wireguard"
	"github.com/5vnetwork/vx-core/transport/protocols/grpc"
	httpProto "github.com/5vnetwork/vx-core/transport/protocols/http"
	"github.com/5vnetwork/vx-core/transport/protocols/httpupgrade"
	"github.com/5vnetwork/vx-core/transport/protocols/kcp"
	"github.com/5vnetwork/vx-core/transport/protocols/splithttp"
	"github.com/5vnetwork/vx-core/transport/protocols/tcp"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/rs/zerolog/log"
)

// Decode parses subscription content into outbound handler configurations
func DecodeCommon(content string) (*sub.DecodeResult, error) {
	// If content doesn't contain ':', assume it's base64 encoded
	if !strings.Contains(content, ":") {
		decoded, err := sub.DecodeBase64(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %v", err)
		}
		content = decoded
	}

	// Split content by line breaks (handle both Windows and Unix style)
	lines := strings.Split(content, "\r\n")
	if len(lines) == 1 {
		lines = strings.Split(content, "\n")
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no valid content found")
	}
	var description string
	configList := make([]*configs.OutboundHandlerConfig, 0)

	result := &sub.DecodeResult{}
	// Process each line
	for i, line := range lines {
		log.Debug().Str("link", line).Msg("decode")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var config *configs.OutboundHandlerConfig

		if strings.HasPrefix(line, "vmess://") {
			// Parse Vmess configuration
			vmessData := line[8:] // Remove "vmess://" prefix
			decodedVmess, err := sub.DecodeBase64(vmessData)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to decode vmess data")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}

			var vmessJson map[string]interface{}
			if err := json.Unmarshal([]byte(decodedVmess), &vmessJson); err != nil {
				log.Debug().Err(err).Msg("Failed to parse vmess JSON")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}

			vmessBytes, err := json.Marshal(vmessJson)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to re-marshal vmess JSON")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}

			vmessConfig, err := ParseVmessFromJSON(vmessBytes)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to parse vmess config")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}

			config, err = vmessConfig.ToProxyHandlerConfig()
			if err != nil {
				log.Debug().Err(err).Msg("Failed to convert vmess to handler config")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}

		} else if strings.HasPrefix(line, "ss://") {
			// Parse Shadowsocks configuration
			var err error
			config, err = ParseSsFromLink(line)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to parse ss link")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}

		} else if strings.HasPrefix(line, "trojan://") {
			// Parse Trojan configuration
			var err error
			config, err = ParseTrojan(line)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to convert trojan to handler config")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}

		} else if strings.HasPrefix(line, "vless://") {
			var err error
			config, err = ParseVlessFromLink(line)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to parse vless link")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}
		} else if strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "hy2://") {
			var err error
			config, err = ParseHysteriaFromLink(line)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to parse hysteria2 link")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}
		} else if strings.HasPrefix(line, "socks5://") {
			var err error
			config, err = ParseSocks5FromLink(line)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to parse socks link")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}
		} else if strings.HasPrefix(line, "anytls://") {
			var err error
			config, err = ParseAnytls(line)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to parse anytls link")
				result.FailedNodes = append(result.FailedNodes, line)
				continue
			}
		} else {
			// first line
			if i == 0 {
				description = line
			} else {
				log.Debug().Msgf("Unknown link format: %s\n", line)
				result.FailedNodes = append(result.FailedNodes, line)
			}
			continue
		}

		configList = append(configList, config)
	}
	result.Configs = configList
	result.Description = description
	return result, nil
}

func setOthers(config *configs.OutboundHandlerConfig, query url.Values) {
	if query.Get("mux") == "1" {
		config.EnableMux = true
		if maxConcurrency := query.Get("mux_max_concurrency"); maxConcurrency != "" {
			maxConcurrencyInt, _ := strconv.Atoi(maxConcurrency)
			config.MuxConfig.MaxConcurrency = uint32(maxConcurrencyInt)
		}
		if maxConnection := query.Get("mux_max_connection"); maxConnection != "" {
			maxConnectionInt, _ := strconv.Atoi(maxConnection)
			config.MuxConfig.MaxConnection = uint32(maxConnectionInt)
		}
	}
	if query.Get("uot") == "1" {
		config.Uot = true
	}
}

func setProtocol(config *configs.TransportConfig, query url.Values) error {
	var header *any.Any

	if query.Get("header") != "" {
		switch query.Get("header") {
		case "none":
			break
		case "http":
			httpHeader := &http.Config{}
			httpHeader.Request = &http.RequestConfig{
				Uri:    []string{query.Get("path")},
				Header: []*http.Header{},
			}
			if query.Get("host") != "" {
				httpHeader.Request.Header = append(httpHeader.Request.Header, &http.Header{
					Name:  "Host",
					Value: []string{query.Get("host")},
				})
			}
			header = serial.ToTypedMessage(httpHeader)
		case "srtp":
			srtpHeader := &srtp.Config{}
			header = serial.ToTypedMessage(srtpHeader)
		case "utp":
			utpHeader := &utp.Config{}
			header = serial.ToTypedMessage(utpHeader)
		case "wechat-video":
			wechatHeader := &wechat.VideoConfig{}
			header = serial.ToTypedMessage(wechatHeader)
		case "wireguard":
			wireguardHeader := &wireguard.WireguardConfig{}
			header = serial.ToTypedMessage(wireguardHeader)
		default:
			return fmt.Errorf("unknown transport header protocol: %s", query.Get("header"))
		}
	}

	protocol := query.Get("type")
	if protocol == "" {
		protocol = query.Get("network")
	}
	if protocol != "" {
		switch protocol {
		case "tcp":
			tcpConfig := &tcp.TcpConfig{
				HeaderSettings: header,
			}
			config.Protocol = &configs.TransportConfig_Tcp{Tcp: tcpConfig}
		case "ws":
			path := query.Get("path")
			maxEarlyData := 0
			if u, err := url.Parse(path); err == nil {
				q := u.Query()
				if ed := q.Get("ed"); ed != "" {
					maxEarlyData, _ = strconv.Atoi(ed)
					q.Del("ed")
					u.RawQuery = q.Encode()
					path = u.String()
				}
			}
			config.Protocol = &configs.TransportConfig_Websocket{
				Websocket: &websocket.WebsocketConfig{
					Host:                query.Get("host"),
					Path:                path,
					MaxEarlyData:        int32(maxEarlyData),
					EarlyDataHeaderName: "Sec-WebSocket-Protocol",
				},
			}
		case "grpc":
			config.Protocol = &configs.TransportConfig_Grpc{
				Grpc: &grpc.GrpcConfig{
					ServiceName: query.Get("serviceName"),
				},
			}
		case "h2", "http":
			var hosts []string
			if query.Get("host") != "" {
				hosts = strings.Split(query.Get("host"), ",")
			}
			httpConfig := &httpProto.HttpConfig{
				Host: hosts,
				Path: query.Get("path"),
			}
			config.Protocol = &configs.TransportConfig_Http{Http: httpConfig}
		case "httpupgrade":
			config.Protocol = &configs.TransportConfig_Httpupgrade{
				Httpupgrade: &httpupgrade.HttpUpgradeConfig{
					Config: &websocket.WebsocketConfig{
						Host: query.Get("host"),
						Path: query.Get("path"),
					},
				},
			}
		case "kcp":
			kcpConfig := &kcp.KcpConfig{
				Seed:         query.Get("path"),
				HeaderConfig: header,
			}
			config.Protocol = &configs.TransportConfig_Kcp{Kcp: kcpConfig}
		case "xhttp":
			config.Protocol = &configs.TransportConfig_Splithttp{
				Splithttp: &splithttp.SplitHttpConfig{
					Host: query.Get("host"),
					Mode: query.Get("mode"),
					Path: query.Get("path"),
				},
			}
		case "none", "---":
			break
		default:
			return fmt.Errorf("unknown transport protocol: %s", protocol)
		}
	}

	return nil
}

func getTransportConfig(query url.Values) (*configs.TransportConfig, error) {
	config := &configs.TransportConfig{}
	if err := setProtocol(config, query); err != nil {
		return nil, err
	}

	if security := query.Get("security"); security != "" {
		switch security {
		case "tls":
			allowInsecure := false
			if query.Get("allowInsecure") == "1" {
				allowInsecure = true
			} else if query.Get("insecure") == "1" {
				allowInsecure = true
			} else if query.Get("allow_insecure") == "1" {
				allowInsecure = true
			}
			var nextProtocol []string
			if alpn := query.Get("alpn"); alpn != "" {
				nextProtocol = strings.Split(alpn, ",")
			}
			var pinnedPeerCertificateChainSha256 [][]byte
			if pinned := query.Get("pinSHA256"); pinned != "" {
				for _, p := range strings.Split(pinned, ",") {
					hash, err := hex.DecodeString(p)
					if err != nil {
						return nil, fmt.Errorf("failed to decode pinned peer certificate chain sha256: %w", err)
					}
					pinnedPeerCertificateChainSha256 = append(pinnedPeerCertificateChainSha256, hash)
				}
			}
			var echConfig []byte
			if ech := query.Get("echConfig"); ech != "" {
				c, err := base64.StdEncoding.DecodeString(ech)
				if err != nil {
					return nil, fmt.Errorf("failed to decode echConfig: %w", err)
				}
				echConfig = c
			}
			config.Security = &configs.TransportConfig_Tls{
				Tls: &tls.TlsConfig{
					ServerName:                       query.Get("sni"),
					AllowInsecure:                    allowInsecure,
					Imitate:                          query.Get("fp"),
					NextProtocol:                     nextProtocol,
					PinnedPeerCertificateChainSha256: pinnedPeerCertificateChainSha256,
					EchConfig:                        echConfig,
					EnableEch:                        query.Get("enableEch") == "1",
				},
			}
		case "reality":
			realityConfig := &reality.RealityConfig{
				ServerName:  query.Get("sni"),
				SpiderX:     query.Get("spx"),
				Fingerprint: query.Get("fp"),
				Sid:         query.Get("sid"),
			}
			key, err := base64.RawURLEncoding.DecodeString(query.Get("pbk"))
			if err != nil {
				key, err = base64.URLEncoding.DecodeString(query.Get("pbk"))
				if err != nil {
					return nil, err
				}
			}
			realityConfig.PublicKey = key
			realityConfig.ShortId = make([]byte, 8)
			_, err = hex.Decode(realityConfig.ShortId, []byte(query.Get("sid")))
			if err != nil {
				return nil, err
			}
			config.Security = &configs.TransportConfig_Reality{
				Reality: realityConfig,
			}
		case "none":
		default:
			return nil, fmt.Errorf("unknown security protocol: %s", security)
		}
	}

	return config, nil
}

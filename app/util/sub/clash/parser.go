// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package clash

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/util/sub"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	mystrings "github.com/5vnetwork/vx-core/common/strings"
	"gopkg.in/yaml.v3"
)

// ClashConfig represents the structure of a Clash/Mihomo configuration file
type ClashConfig struct {
	Port               int        `yaml:"port"`
	SocksPort          int        `yaml:"socks-port"`
	AllowLan           bool       `yaml:"allow-lan"`
	Mode               string     `yaml:"mode"`
	LogLevel           string     `yaml:"log-level"`
	ExternalController string     `yaml:"external-controller"`
	DNS                *DNSConfig `yaml:"dns"`
	Proxies            []any      `yaml:"proxies"`
	ProxyGroups        []any      `yaml:"proxy-groups"`
	Rules              []string   `yaml:"rules"`
}

type DNSConfig struct {
	Enable            bool     `yaml:"enable"`
	IPv6              bool     `yaml:"ipv6"`
	DefaultNameserver []string `yaml:"default-nameserver"`
	EnhancedMode      string   `yaml:"enhanced-mode"`
	FakeIPRange       string   `yaml:"fake-ip-range"`
	Nameserver        []string `yaml:"nameserver"`
	Fallback          []string `yaml:"fallback"`
}

func ParseClashConfig(data []byte) (*sub.DecodeResult, error) {
	var clashConfig ClashConfig
	if err := yaml.Unmarshal(data, &clashConfig); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}
	results, failedNodes, err := ParseProxies(clashConfig.Proxies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxies: %w", err)
	}
	return &sub.DecodeResult{
		Configs:     results,
		FailedNodes: failedNodes,
	}, nil
}

// ParseProxies converts a "proxies" array from a clash/mihomo config to vx-core OutboundHandlerConfigs
func ParseProxies(proxies []any) ([]*configs.OutboundHandlerConfig, []string, error) {
	results := make([]*configs.OutboundHandlerConfig, 0, len(proxies))
	failedNodes := make([]string, 0, len(proxies))

	for _, proxyData := range proxies {
		mapping, ok := proxyData.(map[string]any)
		if !ok {
			failedNodes = append(failedNodes, mystrings.ToString(proxyData))
			continue
		}

		config, err := ParseProxy(mapping)
		if err != nil {
			yamlBytes, err := yaml.Marshal(mapping)
			if err != nil {
				failedNodes = append(failedNodes, mystrings.ToString(proxyData))
			} else {
				failedNodes = append(failedNodes, string(yamlBytes))
			}
			continue
		}

		results = append(results, config)
	}

	return results, failedNodes, nil
}

// ParseProxy converts a single mihomo proxy mapping to vx-core OutboundHandlerConfig
func ParseProxy(mapping map[string]any) (*configs.OutboundHandlerConfig, error) {
	proxyType, ok := mapping["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid type")
	}

	// Get the name (tag)
	name, _ := mapping["name"].(string)

	switch proxyType {
	case "vmess":
		return parseVmess(mapping, name)
	case "vless":
		return parseVless(mapping, name)
	case "trojan":
		return parseTrojan(mapping, name)
	case "ss":
		return parseShadowsocks(mapping, name)
	case "socks5":
		return parseSocks(mapping, name)
	case "http":
		return parseHTTP(mapping, name)
	case "anytls":
		return parseAnytls(mapping, name)
	case "hysteria":
		return parseHysteria(mapping, name)
	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", proxyType)
	}
}

// parseVmess converts mihomo vmess config to vx-core config
func parseVmess(mapping map[string]any, name string) (*configs.OutboundHandlerConfig, error) {
	// Extract basic fields
	server, _ := mapping["server"].(string)
	portNum := getPort(mapping["port"])
	if portNum == 0 {
		return nil, fmt.Errorf("invalid port")
	}

	uuid, _ := mapping["uuid"].(string)
	if uuid == "" {
		return nil, fmt.Errorf("missing uuid")
	}

	alterId := 0
	if alterIdVal, ok := mapping["alterId"]; ok {
		switch v := alterIdVal.(type) {
		case int:
			alterId = v
		case float64:
			alterId = int(v)
		case string:
			alterId, _ = strconv.Atoi(v)
		}
	}

	cipher, _ := mapping["cipher"].(string)

	// Create vx-core vmess config
	vmessConfig := &configs.VmessClientConfig{
		Id:      uuid,
		AlterId: uint32(alterId),
	}

	// Map cipher to security type
	switch cipher {
	case "auto":
		vmessConfig.Security = configs.SecurityType_SecurityType_AUTO
	case "aes-128-gcm":
		vmessConfig.Security = configs.SecurityType_SecurityType_AES128_GCM
	case "chacha20-poly1305":
		vmessConfig.Security = configs.SecurityType_SecurityType_CHACHA20_POLY1305
	case "none":
		vmessConfig.Security = configs.SecurityType_SecurityType_NONE
	case "zero":
		vmessConfig.Security = configs.SecurityType_SecurityType_ZERO
	default:
		vmessConfig.Security = configs.SecurityType_SecurityType_AUTO
	}

	outbound := &configs.OutboundHandlerConfig{
		Tag:      name,
		Address:  server,
		Port:     portNum,
		Protocol: serial.ToTypedMessage(vmessConfig),
	}

	// Parse transport config
	transport, err := parseTransportConfig(mapping)
	if err != nil {
		return nil, fmt.Errorf("parse transport: %w", err)
	}
	outbound.Transport = transport

	return outbound, nil
}

// parseVless converts mihomo vless config to vx-core config
func parseVless(mapping map[string]any, name string) (*configs.OutboundHandlerConfig, error) {
	server, _ := mapping["server"].(string)
	portNum := getPort(mapping["port"])
	if portNum == 0 {
		return nil, fmt.Errorf("invalid port")
	}

	uuid, _ := mapping["uuid"].(string)
	if uuid == "" {
		return nil, fmt.Errorf("missing uuid")
	}

	flow, _ := mapping["flow"].(string)
	encryption, _ := mapping["encryption"].(string)
	if encryption == "" {
		encryption = "none"
	}

	vlessConfig := &configs.VlessClientConfig{
		Id:         uuid,
		Flow:       flow,
		Encryption: encryption,
	}

	outbound := &configs.OutboundHandlerConfig{
		Tag:      name,
		Address:  server,
		Port:     portNum,
		Protocol: serial.ToTypedMessage(vlessConfig),
	}

	// Parse transport config
	transport, err := parseTransportConfig(mapping)
	if err != nil {
		return nil, fmt.Errorf("parse transport: %w", err)
	}
	outbound.Transport = transport

	return outbound, nil
}

// parseTrojan converts mihomo trojan config to vx-core config
func parseTrojan(mapping map[string]any, name string) (*configs.OutboundHandlerConfig, error) {
	server, _ := mapping["server"].(string)
	portNum := getPort(mapping["port"])
	if portNum == 0 {
		return nil, fmt.Errorf("invalid port")
	}

	password, _ := mapping["password"].(string)
	if password == "" {
		return nil, fmt.Errorf("missing password")
	}

	trojanConfig := &configs.TrojanClientConfig{
		Password: password,
	}

	outbound := &configs.OutboundHandlerConfig{
		Tag:      name,
		Address:  server,
		Port:     portNum,
		Protocol: serial.ToTypedMessage(trojanConfig),
	}

	// Parse transport config
	transport, err := parseTransportConfig(mapping)
	if err != nil {
		return nil, fmt.Errorf("parse transport: %w", err)
	}
	outbound.Transport = transport

	return outbound, nil
}

// parseShadowsocks converts mihomo shadowsocks config to vx-core config
func parseShadowsocks(mapping map[string]any, name string) (*configs.OutboundHandlerConfig, error) {
	server, _ := mapping["server"].(string)
	portNum := getPort(mapping["port"])
	if portNum == 0 {
		return nil, fmt.Errorf("invalid port")
	}

	password, _ := mapping["password"].(string)
	if password == "" {
		return nil, fmt.Errorf("missing password")
	}

	cipher, _ := mapping["cipher"].(string)
	if cipher == "" {
		return nil, fmt.Errorf("missing cipher")
	}

	// Map cipher string to CipherType enum
	var cipherType configs.ShadowsocksCipherType
	switch cipher {
	case "aes-128-gcm", "AEAD_AES_128_GCM":
		cipherType = configs.ShadowsocksCipherType_AES_128_GCM
	case "aes-256-gcm", "AEAD_AES_256_GCM":
		cipherType = configs.ShadowsocksCipherType_AES_256_GCM
	case "chacha20-poly1305", "chacha20-ietf-poly1305", "AEAD_CHACHA20_POLY1305":
		cipherType = configs.ShadowsocksCipherType_CHACHA20_POLY1305
	case "none", "plain":
		cipherType = configs.ShadowsocksCipherType_NONE
	default:
		return nil, fmt.Errorf("unsupported cipher: %s", cipher)
	}

	ssConfig := &configs.ShadowsocksClientConfig{
		Password:   password,
		CipherType: cipherType,
	}

	outbound := &configs.OutboundHandlerConfig{
		Tag:      name,
		Address:  server,
		Port:     portNum,
		Protocol: serial.ToTypedMessage(ssConfig),
	}

	if mapping["udp-over-tcp"] == true {
		outbound.Uot = true
	}

	// Parse transport config
	transport, err := parseTransportConfig(mapping)
	if err != nil {
		return nil, fmt.Errorf("parse transport: %w", err)
	}
	outbound.Transport = transport

	return outbound, nil
}

// parseSocks converts mihomo socks5 config to vx-core config
func parseSocks(mapping map[string]any, name string) (*configs.OutboundHandlerConfig, error) {
	server, _ := mapping["server"].(string)
	portNum := getPort(mapping["port"])
	if portNum == 0 {
		return nil, fmt.Errorf("invalid port")
	}

	username, _ := mapping["username"].(string)
	password, _ := mapping["password"].(string)

	socksConfig := &configs.SocksClientConfig{
		Name:     username,
		Password: password,
	}

	outbound := &configs.OutboundHandlerConfig{
		Tag:      name,
		Address:  server,
		Port:     portNum,
		Protocol: serial.ToTypedMessage(socksConfig),
	}

	return outbound, nil
}

// parseHTTP converts mihomo http config to vx-core config
func parseHTTP(mapping map[string]any, name string) (*configs.OutboundHandlerConfig, error) {
	server, _ := mapping["server"].(string)
	portNum := getPort(mapping["port"])
	if portNum == 0 {
		return nil, fmt.Errorf("invalid port")
	}

	username, _ := mapping["username"].(string)
	password, _ := mapping["password"].(string)

	var account *configs.Account
	if username != "" || password != "" {
		account = &configs.Account{
			Username: username,
			Password: password,
		}
	}
	httpConfig := &configs.HttpClientConfig{
		Account: account,
	}
	outbound := &configs.OutboundHandlerConfig{
		Tag:      name,
		Address:  server,
		Port:     portNum,
		Protocol: serial.ToTypedMessage(httpConfig),
	}

	return outbound, nil
}

// parseAnytls converts mihomo anytls config to vx-core config
func parseAnytls(mapping map[string]any, name string) (*configs.OutboundHandlerConfig, error) {
	server, _ := mapping["server"].(string)
	portNum := getPort(mapping["port"])
	if portNum == 0 {
		return nil, fmt.Errorf("invalid port")
	}

	password, _ := mapping["password"].(string)
	if password == "" {
		return nil, fmt.Errorf("missing password")
	}

	anytlsConfig := &configs.AnytlsClientConfig{
		Password: password,
	}
	if idleSessionCheckInterval, ok := mapping["idle-session-check-interval"].(int); ok {
		anytlsConfig.IdleSessionCheckInterval = uint32(idleSessionCheckInterval)
	}
	if idleSessionTimeout, ok := mapping["idle-session-timeout"].(int); ok {
		anytlsConfig.IdleSessionTimeout = uint32(idleSessionTimeout)
	}
	if minIdleSession, ok := mapping["min-idle-session"].(int); ok {
		anytlsConfig.MinIdleSession = uint32(minIdleSession)
	}

	outbound := &configs.OutboundHandlerConfig{
		Tag:      name,
		Address:  server,
		Port:     portNum,
		Protocol: serial.ToTypedMessage(anytlsConfig),
	}

	// AnyTLS always uses TLS, parse transport config
	transport, err := parseTransportConfig(mapping)
	if err != nil {
		return nil, fmt.Errorf("parse transport: %w", err)
	}
	outbound.Transport = transport

	return outbound, nil
}

func parseHysteria(mapping map[string]any, name string) (*configs.OutboundHandlerConfig, error) {
	server, _ := mapping["server"].(string)
	var ports []*net.PortRange
	if portVal, ok := mapping["port"].(string); ok {
		ports = append(ports, &net.PortRange{From: getPort(portVal), To: getPort(portVal)})
	} else if portVal, ok := mapping["ports"].(string); ok {
		ports = sub.TryParsePorts(portVal)
	}
	hysteriaConfig := &configs.Hysteria2ClientConfig{
		Bandwidth: &configs.BandwidthConfig{},
	}
	if auth, ok := mapping["password"].(string); ok {
		hysteriaConfig.Auth = auth
	}
	if mapping["obfs"] == "salamander" {
		hysteriaConfig.Obfs = &configs.ObfsConfig{
			Obfs: &configs.ObfsConfig_Salamander{
				Salamander: &configs.SalamanderConfig{
					Password: mapping["obfs-password"].(string),
				},
			},
		}
	}
	// bandwidth
	if up, ok := mapping["up"].(string); ok {
		if strings.Contains(up, "Mbps") {
			if u := strings.Split(up, " ")[0]; u != "" {
				upValue, _ := strconv.Atoi(u)
				hysteriaConfig.Bandwidth.MaxTx = uint32(upValue * 1024 * 1024 / 8)
			}
		} else {
			upValue, _ := strconv.Atoi(up)
			hysteriaConfig.Bandwidth.MaxTx = uint32(upValue * 1024 * 1024 / 8)
		}
	}
	if down, ok := mapping["down"].(string); ok {
		if strings.Contains(down, "Mbps") {
			if d := strings.Split(down, " ")[0]; d != "" {
				downValue, _ := strconv.Atoi(d)
				hysteriaConfig.Bandwidth.MaxRx = uint32(downValue * 1024 * 1024 / 8)
			}
		}
	} else {
		downValue, _ := strconv.Atoi(down)
		hysteriaConfig.Bandwidth.MaxRx = uint32(downValue * 1024 * 1024 / 8)
	}
	transport, err := parseTransportConfig(mapping)
	if err != nil {
		return nil, fmt.Errorf("parse transport: %w", err)
	}
	hysteriaConfig.TlsConfig = transport.GetTls()

	outbound := &configs.OutboundHandlerConfig{
		Tag:      name,
		Address:  server,
		Ports:    ports,
		Protocol: serial.ToTypedMessage(hysteriaConfig),
	}
	return outbound, nil
}

// getPort extracts port number from various types
func getPort(portVal any) uint32 {
	switch v := portVal.(type) {
	case int:
		return uint32(v)
	case float64:
		return uint32(v)
	case string:
		port, _ := strconv.Atoi(v)
		return uint32(port)
	default:
		return 0
	}
}

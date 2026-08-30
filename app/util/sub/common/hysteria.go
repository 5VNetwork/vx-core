// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/util/sub"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/transport/security/tls"
)

// TODO: pinSHA256
// hysteria2://letmein@example.com:123,5000-6000/?insecure=1&obfs=
// salamander&obfs-password=gawrgura&pinSHA256=deadbeef&sni=real.example.com
//
// hysteria2+realm://token@rendezvous-host[:port]/realm-name?auth=...&sni=...&insecure=1&pinSHA256=...
func ParseHysteriaFromLink(link string) (*configs.OutboundHandlerConfig, error) {
	if strings.HasPrefix(link, "hysteria2+realm://") || strings.HasPrefix(link, "hysteria2+realm+http://") {
		u, err := url.Parse(link)
		if err != nil {
			return nil, err
		}
		return parseHysteriaRealmFromLink(u)
	}

	port := extractHysteriaPortFromURL(link)
	if port != "" {
		link = strings.Replace(link, port, "", 1)
	}

	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "hysteria2" && u.Scheme != "hy2" {
		return nil, fmt.Errorf("not a valid hysteria2 link")
	}

	config := &configs.OutboundHandlerConfig{
		Tag: u.Fragment,
	}

	query := u.Query()
	config.Address = u.Hostname()

	if port != "" {
		portRanges := sub.TryParsePorts(port)
		if len(portRanges) == 0 {
			return nil, fmt.Errorf("invalid port range")
		}
		config.Ports = portRanges
	} else if query.Get("mport") != "" {
		portRanges := sub.TryParsePorts(query.Get("mport"))
		if len(portRanges) == 0 {
			return nil, fmt.Errorf("invalid port range")
		}
		config.Ports = portRanges
	} else {
		if u.Port() != "" {
			port, err := mynet.PortFromString(u.Port())
			if err != nil {
				return nil, err
			}
			config.Ports = []*mynet.PortRange{
				{
					From: uint32(port),
					To:   uint32(port),
				},
			}
		} else {
			config.Ports = []*mynet.PortRange{
				{
					From: 443,
					To:   443,
				},
			}
		}
	}

	serverName := query.Get("sni")
	if serverName == "" {
		serverName = u.Hostname()
	}

	hysteriaConfig := &configs.Hysteria2ClientConfig{
		Auth: u.User.String(),
		TlsConfig: &tls.TlsConfig{
			ServerName: serverName,
		},
		Bandwidth: &configs.BandwidthConfig{},
	}
	if query.Get("echConfig") != "" {
		echConfig, err := base64.StdEncoding.DecodeString(query.Get("echConfig"))
		if err != nil {
			return nil, err
		}
		hysteriaConfig.TlsConfig.EchConfig = echConfig
	}
	if query.Get("insecure") == "1" {
		hysteriaConfig.TlsConfig.AllowInsecure = true
	}
	if query.Get("obfs") != "" {
		hysteriaConfig.Obfs = &configs.ObfsConfig{
			Obfs: &configs.ObfsConfig_Salamander{
				Salamander: &configs.SalamanderConfig{
					Password: query.Get("obfs-password"),
				},
			},
		}
	}
	if query.Get("pinSHA256") != "" {
		pinSHA256, err := hex.DecodeString(query.Get("pinSHA256"))
		if err != nil {
			return nil, err
		}
		hysteriaConfig.TlsConfig.PinnedPeerCertificateChainSha256 = [][]byte{
			pinSHA256,
		}
		hysteriaConfig.TlsConfig.AllowInsecure = true
	}
	if query.Get("tx") != "" {
		tx, err := strconv.Atoi(query.Get("tx"))
		if err == nil {
			hysteriaConfig.Bandwidth.MaxTx = uint32(tx)
		}
	}
	if query.Get("rx") != "" {
		rx, err := strconv.Atoi(query.Get("rx"))
		if err == nil {
			hysteriaConfig.Bandwidth.MaxRx = uint32(rx)
		}
	}
	config.Protocol = serial.ToTypedMessage(hysteriaConfig)
	return config, nil
}

func parseHysteriaRealmFromLink(u *url.URL) (*configs.OutboundHandlerConfig, error) {
	if u.Scheme != "hysteria2+realm" && u.Scheme != "hysteria2+realm+http" {
		return nil, fmt.Errorf("not a valid hysteria2+realm link")
	}
	if u.User == nil || u.User.Username() == "" {
		return nil, fmt.Errorf("missing realm token")
	}

	realmName := strings.TrimPrefix(u.Path, "/")
	if realmName == "" || strings.Contains(realmName, "/") {
		return nil, fmt.Errorf("invalid realm name")
	}

	query := u.Query()
	auth := query.Get("auth")
	if auth == "" {
		return nil, fmt.Errorf("missing auth parameter")
	}

	realmScheme := "realm"
	if u.Scheme == "hysteria2+realm+http" {
		realmScheme = "realm+http"
	}
	realmAddr := (&url.URL{
		Scheme: realmScheme,
		User:   u.User,
		Host:   u.Host,
		Path:   u.Path,
	}).String()

	config := &configs.OutboundHandlerConfig{
		Tag:     u.Fragment,
		Address: u.Hostname(),
	}
	if u.Port() != "" {
		port, err := mynet.PortFromString(u.Port())
		if err != nil {
			return nil, err
		}
		config.Ports = []*mynet.PortRange{{From: uint32(port), To: uint32(port)}}
	} else {
		config.Ports = []*mynet.PortRange{{From: 443, To: 443}}
	}

	realmCfg := &configs.RealmConfig{RealmAddr: realmAddr}
	for _, stun := range query["stun"] {
		if stun != "" {
			realmCfg.StunServers = append(realmCfg.StunServers, stun)
		}
	}
	if lport := query.Get("lport"); lport != "" {
		port, err := strconv.Atoi(lport)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid lport parameter")
		}
		realmCfg.LocalPort = uint32(port)
	}

	hysteriaConfig := &configs.Hysteria2ClientConfig{
		Auth: auth,
		TlsConfig: &tls.TlsConfig{
			ServerName: query.Get("sni"),
		},
		Bandwidth: &configs.BandwidthConfig{},
		Realm:     realmCfg,
	}
	hysteriaConfig = applyHysteriaQueryParams(hysteriaConfig, query)
	config.Protocol = serial.ToTypedMessage(hysteriaConfig)
	return config, nil
}

func applyHysteriaQueryParams(hysteriaConfig *configs.Hysteria2ClientConfig, query url.Values) *configs.Hysteria2ClientConfig {
	if query.Get("echConfig") != "" {
		echConfig, err := base64.StdEncoding.DecodeString(query.Get("echConfig"))
		if err == nil {
			hysteriaConfig.TlsConfig.EchConfig = echConfig
		}
	}
	if query.Get("insecure") == "1" {
		hysteriaConfig.TlsConfig.AllowInsecure = true
	}
	if query.Get("obfs") != "" {
		hysteriaConfig.Obfs = &configs.ObfsConfig{
			Obfs: &configs.ObfsConfig_Salamander{
				Salamander: &configs.SalamanderConfig{
					Password: query.Get("obfs-password"),
				},
			},
		}
	}
	if query.Get("pinSHA256") != "" {
		pinSHA256, err := hex.DecodeString(query.Get("pinSHA256"))
		if err == nil {
			hysteriaConfig.TlsConfig.PinnedPeerCertificateChainSha256 = [][]byte{pinSHA256}
		}
	}
	if query.Get("tx") != "" {
		tx, err := strconv.Atoi(query.Get("tx"))
		if err == nil {
			hysteriaConfig.Bandwidth.MaxTx = uint32(tx)
		}
	}
	if query.Get("rx") != "" {
		rx, err := strconv.Atoi(query.Get("rx"))
		if err == nil {
			hysteriaConfig.Bandwidth.MaxRx = uint32(rx)
		}
	}
	return hysteriaConfig
}

// extractHysteriaPortFromURL extracts the port part if it contains "," or "-"
func extractHysteriaPortFromURL(urlStr string) string {
	// remove scheme
	slice := strings.Split(urlStr, "://")
	if len(slice) > 1 {
		urlStr = slice[1]
	}

	// Find the last @ symbol to handle URLs with authentication
	atIndex := strings.LastIndex(urlStr, "@")
	if atIndex == -1 {
		atIndex = 0
	} else {
		urlStr = urlStr[atIndex+1:]
	}

	// Find the first colon after @
	colonIndex := strings.Index(urlStr, ":")
	if colonIndex == -1 {
		return "" // No port found
	}
	urlStr = urlStr[colonIndex+1:]

	// Find the first slash or question mark after the colon
	slashIndex := strings.IndexAny(urlStr, "/?")

	var port string
	if slashIndex == -1 {
		// No slash or question mark found, port extends to end of string
		port = urlStr
	} else {
		port = urlStr[:slashIndex]
	}

	// Extract port part
	if strings.Contains(port, ",") || strings.Contains(port, "-") {
		return port
	}
	return ""
}

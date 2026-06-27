// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package uri

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	bufnet "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/common/net"
	"github.com/5vnetwork/vx-core/app/configs"
)

func toHysteria(outboundConfig *configs.OutboundHandlerConfig) (string, error) {
	config, err := outboundConfig.Protocol.UnmarshalNew()
	if err != nil {
		return "", err
	}
	hysteriaConfig, _ := config.(*configs.Hysteria2ClientConfig)

	// Realm mode: delegate to the dedicated builder.
	if hysteriaConfig.GetRealm().GetRealmAddr() != "" {
		return toHysteriaRealm(outboundConfig, hysteriaConfig)
	}

	queryParameters := url.Values{}
	if tlsConfig := hysteriaConfig.GetTlsConfig(); tlsConfig != nil {
		// queryParameters.Add("security", "tls")
		if tlsConfig.GetServerName() != "" {
			queryParameters.Add("sni", tlsConfig.GetServerName())
		}
		allowInsecure := 0
		if tlsConfig.GetAllowInsecure() {
			allowInsecure = 1
		}
		queryParameters.Set("insecure", strconv.Itoa(allowInsecure))
		if len(tlsConfig.PinnedPeerCertificateChainSha256) > 0 {
			queryParameters.Add("pinSHA256", hex.EncodeToString(tlsConfig.PinnedPeerCertificateChainSha256[0]))
		}
		if len(tlsConfig.EchConfig) > 0 {
			queryParameters.Add("echConfig", base64.StdEncoding.EncodeToString(tlsConfig.EchConfig))
		}
	}
	if hysteriaConfig.Obfs.GetSalamander().GetPassword() != "" {
		queryParameters.Add("obfs", "salamander")
		queryParameters.Add("obfs-password", hysteriaConfig.Obfs.GetSalamander().GetPassword())
	}
	if hysteriaConfig.GetBandwidth().GetMaxRx() != 0 {
		queryParameters.Add("rx", strconv.Itoa(int(hysteriaConfig.Bandwidth.MaxRx/1024/1024)))
	}
	if hysteriaConfig.GetBandwidth().GetMaxTx() != 0 {
		queryParameters.Add("tx", strconv.Itoa(int(hysteriaConfig.Bandwidth.MaxTx/1024/1024)))
	}

	u := &url.URL{
		Scheme:   "hysteria2",
		User:     url.User(hysteriaConfig.GetAuth()),
		RawQuery: queryParameters.Encode(),
		Fragment: outboundConfig.GetTag(),
	}
	if ports := outboundConfig.GetPorts(); len(ports) > 0 {
		u.Host = net.JoinHostPort(outboundConfig.GetAddress(), PortRangesToString(ports))
	} else {
		u.Host = net.JoinHostPort(outboundConfig.GetAddress(), strconv.Itoa(int(outboundConfig.GetPort())))
	}
	return u.String(), nil
}

// toHysteriaRealm builds a hysteria2+realm:// URI following the official URI scheme:
// https://v2.hysteria.network/docs/developers/URI-Scheme/#realm-mode
//
//	hysteria2+realm://token@rendezvous-host[:port]/realm-name?auth=...&[sni=...&insecure=1&pinSHA256=...]
func toHysteriaRealm(outboundConfig *configs.OutboundHandlerConfig, hysteriaConfig *configs.Hysteria2ClientConfig) (string, error) {
	realmAddr := hysteriaConfig.GetRealm().GetRealmAddr()
	realmURL, err := url.Parse(realmAddr)
	if err != nil {
		return "", fmt.Errorf("failed to parse realm addr %q: %w", realmAddr, err)
	}

	// Map the realm:// scheme to the official hysteria2+realm:// scheme.
	scheme := "hysteria2+realm"
	if realmURL.Scheme == "realm+http" {
		scheme = "hysteria2+realm+http"
	}

	q := url.Values{}

	// auth is the Hysteria2 password (userinfo slot is taken by the rendezvous token).
	if auth := hysteriaConfig.GetAuth(); auth != "" {
		q.Set("auth", auth)
	}

	// TLS parameters for the Hysteria2 connection over the punched hole.
	if tlsCfg := hysteriaConfig.GetTlsConfig(); tlsCfg != nil {
		if tlsCfg.GetServerName() != "" {
			q.Set("sni", tlsCfg.GetServerName())
		}
		if tlsCfg.GetAllowInsecure() {
			q.Set("insecure", "1")
		}
		if len(tlsCfg.PinnedPeerCertificateChainSha256) > 0 {
			q.Set("pinSHA256", hex.EncodeToString(tlsCfg.PinnedPeerCertificateChainSha256[0]))
		}
	}

	// Obfuscation.
	if pw := hysteriaConfig.GetObfs().GetSalamander().GetPassword(); pw != "" {
		q.Set("obfs", "salamander")
		q.Set("obfs-password", pw)
	}

	// Realm-specific parameters.
	realmCfg := hysteriaConfig.GetRealm()
	for _, stun := range realmCfg.GetStunServers() {
		q.Add("stun", stun)
	}
	if lport := realmCfg.GetLocalPort(); lport > 0 {
		q.Set("lport", strconv.Itoa(int(lport)))
	}

	u := &url.URL{
		Scheme:   scheme,
		User:     url.User(realmURL.User.Username()), // rendezvous token
		Host:     realmURL.Host,                      // rendezvous host[:port]
		Path:     realmURL.Path,                      // /realm-name
		RawQuery: q.Encode(),
		Fragment: outboundConfig.GetTag(),
	}
	return u.String(), nil
}

// PortRangesToString converts a slice of PortRange back to string format
// This is the reverse of TryParsePorts function
func PortRangesToString(portRanges []*bufnet.PortRange) string {
	if len(portRanges) == 0 {
		return ""
	}

	var parts []string
	for _, pr := range portRanges {
		from, to := pr.GetFrom(), pr.GetTo()
		if from == to {
			parts = append(parts, strconv.Itoa(int(from)))
		} else {
			parts = append(parts, fmt.Sprintf("%d-%d", from, to))
		}
	}

	return strings.Join(parts, ",")
}

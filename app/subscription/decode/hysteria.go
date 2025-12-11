package decode

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/transport/security/tls"
)

// TODO: pinSHA256
// hysteria2://letmein@example.com:123,5000-6000/?insecure=1&obfs=
// salamander&obfs-password=gawrgura&pinSHA256=deadbeef&sni=real.example.com
func ParseHysteriaFromLink(link string) (*configs.OutboundHandlerConfig, error) {
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
		portRanges := TryParsePorts(port)
		if len(portRanges) == 0 {
			return nil, fmt.Errorf("invalid port range")
		}
		config.Ports = portRanges
	} else if query.Get("mport") != "" {
		portRanges := TryParsePorts(query.Get("mport"))
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

	hysteriaConfig := &proxy.Hysteria2ClientConfig{
		Auth: u.User.String(),
		TlsConfig: &tls.TlsConfig{
			ServerName: serverName,
		},
		Bandwidth: &proxy.BandwidthConfig{},
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
		hysteriaConfig.Obfs = &proxy.ObfsConfig{
			Obfs: &proxy.ObfsConfig_Salamander{
				Salamander: &proxy.SalamanderConfig{
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
	}
	if query.Get("tx") != "" {
		tx, err := strconv.Atoi(query.Get("tx"))
		if err == nil {
			hysteriaConfig.Bandwidth.MaxTx = uint32(tx)
		}
	}
	if query.Get("tx") != "" {
		rx, err := strconv.Atoi(query.Get("rx"))
		if err == nil {
			hysteriaConfig.Bandwidth.MaxRx = uint32(rx)
		}
	}
	config.Protocol = serial.ToTypedMessage(hysteriaConfig)
	return config, nil
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

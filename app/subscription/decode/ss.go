package decode

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
	// Add required imports for transport protocols and headers
	// for http headers
)

// SsConfig represents a Shadowsocks configuration
type SsConfig struct {
	Cipher   string
	Password string
	Address  string
	Port     string
	Remark   string
}

// ParseSsFromLink parses a Shadowsocks configuration from a URI link
func ParseSsFromLink(link string) (*configs.OutboundHandlerConfig, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "ss" {
		return nil, fmt.Errorf("not a valid shadowsocks link, got %s", u.Scheme)
	}

	var cipher, password string
	// some ss link is such format: ss://BASE64-ENCODED-STRING-WITHOUT-PADDING#TAG
	if u.User == nil {
		decoded, err := DecodeBase64(u.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to decode host: %v", err)
		}
		link = strings.Replace(link, u.Host, decoded, 1)
		u, err = url.Parse(link)
		if err != nil {
			return nil, err
		}
		cipher = u.User.Username()
		password, _ = u.User.Password()
	} else {
		cipherPasswordBase64 := u.User.Username()
		cipherPasswordBytes, err := DecodeBase64(cipherPasswordBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode cipher:password: %v", err)
		}
		cipherPassword := string(cipherPasswordBytes)
		indexOfSeperator := strings.Index(cipherPassword, ":")
		if indexOfSeperator <= 0 || indexOfSeperator == len(cipherPassword)-1 {
			return nil, fmt.Errorf("invalid cipher:password format")
		}
		cipher = cipherPassword[:indexOfSeperator]
		password = cipherPassword[indexOfSeperator+1:]
	}

	portStr := u.Port()
	ports := TryParsePorts(portStr)
	if len(ports) == 0 {
		return nil, errors.New("port invalid: " + portStr)
	}

	ssConfig := &proxy.ShadowsocksClientConfig{
		Password: password,
	}
	switch cipher {
	case "aes-128-gcm":
		ssConfig.CipherType = proxy.ShadowsocksCipherType_AES_128_GCM
	case "aes-256-gcm":
		ssConfig.CipherType = proxy.ShadowsocksCipherType_AES_256_GCM
	case "chacha20-ietf-poly1305":
		ssConfig.CipherType = proxy.ShadowsocksCipherType_CHACHA20_POLY1305
	case "none":
		ssConfig.CipherType = proxy.ShadowsocksCipherType_NONE
	default:
		return nil, fmt.Errorf("unsupported cipher type: %s", cipher)
	}

	outboundConfig := &configs.OutboundHandlerConfig{
		Address:  u.Hostname(),
		Tag:      u.Fragment,
		Ports:    ports,
		Protocol: serial.ToTypedMessage(ssConfig),
	}

	return outboundConfig, nil
}

// ParseSsFromLink parses a Shadowsocks configuration from a URI link
func ParseSsFromLink0(link string) (*SsConfig, error) {
	if !strings.HasPrefix(link, "ss://") {
		return nil, fmt.Errorf("not a valid shadowsocks link")
	}

	content := strings.Split(link, "://")[1]

	atIndex := strings.Index(content, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("invalid shadowsocks link format")
	}

	cipherPasswordBase64 := content[:atIndex]
	cipherPasswordBytes, err := DecodeBase64(cipherPasswordBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cipher:password: %v", err)
	}

	cipherPassword := string(cipherPasswordBytes)
	parts := strings.Split(cipherPassword, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cipher:password format")
	}

	cipher := parts[0]
	password := parts[1]

	rest := content[atIndex+1:]
	colonIndex := strings.Index(rest, ":")
	if colonIndex == -1 {
		return nil, fmt.Errorf("invalid address:port format")
	}

	address := rest[:colonIndex]

	sharpIndex := strings.Index(rest, "#")
	if sharpIndex == -1 {
		sharpIndex = len(rest)
	}

	portStr := rest[colonIndex+1 : sharpIndex]

	var remark string
	if sharpIndex < len(rest) {
		remarkRawUrlEncoded := rest[sharpIndex+1:]
		remark, err = url.QueryUnescape(remarkRawUrlEncoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode remark: %v", err)
		}
	}

	return &SsConfig{
		Cipher:   cipher,
		Password: password,
		Address:  address,
		Port:     portStr,
		Remark:   remark,
	}, nil
}

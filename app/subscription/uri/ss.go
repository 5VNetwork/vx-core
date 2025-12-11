package uri

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
)

// ss://ciper:password@addr:port#name\r\n
// https://shadowsocks.org/doc/sip002.html
func toShadowSocks0(outboundConfig *configs.OutboundHandlerConfig) (string, error) {
	config, err := outboundConfig.Protocol.UnmarshalNew()
	if err != nil {
		return "", err
	}
	ssConfig, _ := config.(*proxy.ShadowsocksClientConfig)

	ret := "ss://"
	// ciper+password
	var ciper string
	switch ssConfig.CipherType {
	case proxy.ShadowsocksCipherType_AES_128_GCM:
		ciper = "aes-128-gcm"
	case proxy.ShadowsocksCipherType_AES_256_GCM:
		ciper = "aes-256-gcm"
	case proxy.ShadowsocksCipherType_CHACHA20_POLY1305:
		ciper = "chacha20-ietf-poly1305"
	case proxy.ShadowsocksCipherType_NONE:
		ciper = "none"
	default:
		return "", fmt.Errorf("unknown ciper type: %v", ssConfig.CipherType)
	}
	ciperPassword := base64.URLEncoding.EncodeToString([]byte(ciper + ":" + ssConfig.Password))
	ret += ciperPassword + "@"
	// addr:port
	ret += outboundConfig.Address + ":" + fmt.Sprint(outboundConfig.Port) + "#"
	// name
	ret += url.QueryEscape(outboundConfig.Tag)
	return ret, nil
}

func toShadowSocks(outboundConfig *configs.OutboundHandlerConfig) (string, error) {
	config, err := outboundConfig.Protocol.UnmarshalNew()
	if err != nil {
		return "", err
	}
	ssConfig, _ := config.(*proxy.ShadowsocksClientConfig)
	var ciper string
	switch ssConfig.CipherType {
	case proxy.ShadowsocksCipherType_AES_128_GCM:
		ciper = "aes-128-gcm"
	case proxy.ShadowsocksCipherType_AES_256_GCM:
		ciper = "aes-256-gcm"
	case proxy.ShadowsocksCipherType_CHACHA20_POLY1305:
		ciper = "chacha20-ietf-poly1305"
	case proxy.ShadowsocksCipherType_NONE:
		ciper = "none"
	default:
		return "", fmt.Errorf("unknown ciper type: %v", ssConfig.CipherType)
	}

	u := &url.URL{
		Scheme:   "ss",
		User:     url.User(strings.TrimSuffix(base64.URLEncoding.EncodeToString([]byte(ciper+":"+ssConfig.Password)), "=")),
		Host:     net.JoinHostPort(outboundConfig.Address, strconv.Itoa(getSinglePort(outboundConfig))),
		Fragment: outboundConfig.Tag,
	}
	return u.String(), nil
}

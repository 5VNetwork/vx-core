// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/util/sub"
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
		decoded, err := sub.DecodeBase64(u.Host)
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
		pass, hasPassword := u.User.Password()
		if hasPassword {
			// ss://2022-blake3-aes-256-gcm:YctPZ6U7xPPcU%2Bgp3u%2B0tx%2FtRizJN9K8y%2BuKlW2qjlI%3D@192.168.100.1:8888#Example3
			password = pass
			cipher = cipherPasswordBase64
		} else {
			cipherPasswordBytes, err := sub.DecodeBase64(cipherPasswordBase64)
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
	}

	portStr := u.Port()
	ports := sub.TryParsePorts(portStr)
	if len(ports) == 0 {
		return nil, errors.New("port invalid: " + portStr)
	}

	ssConfig := &configs.ShadowsocksClientConfig{
		Password: password,
	}
	switch cipher {
	case "aes-128-gcm":
		ssConfig.CipherType = configs.ShadowsocksCipherType_AES_128_GCM
	case "aes-256-gcm":
		ssConfig.CipherType = configs.ShadowsocksCipherType_AES_256_GCM
	case "chacha20-ietf-poly1305":
		ssConfig.CipherType = configs.ShadowsocksCipherType_CHACHA20_POLY1305
	case "none":
		ssConfig.CipherType = configs.ShadowsocksCipherType_NONE
	default:
		ss2022Config := &configs.Shadowsocks2022ClientConfig{
			Key: password,
		}
		switch cipher {
		case "2022-blake3-aes-128-gcm":
			ss2022Config.Method = "2022-blake3-aes-128-gcm"
		case "2022-blake3-aes-256-gcm":
			ss2022Config.Method = "2022-blake3-aes-256-gcm"
		case "2022-blake3-chacha20-poly1305":
			ss2022Config.Method = "2022-blake3-chacha20-poly1305"
		default:
			return nil, fmt.Errorf("unsupported cipher type: %s", cipher)
		}
		ssAny, err := serial.ToTypedMessage0(ss2022Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ss2022 config: %v", err)
		}
		outboundConfig := &configs.OutboundHandlerConfig{
			Address:  u.Hostname(),
			Tag:      u.Fragment,
			Ports:    ports,
			Protocol: ssAny,
		}
		return outboundConfig, nil
	}

	ssAny, err := serial.ToTypedMessage0(ssConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ss config: %v", err)
	}
	outboundConfig := &configs.OutboundHandlerConfig{
		Address:  u.Hostname(),
		Tag:      u.Fragment,
		Ports:    ports,
		Protocol: ssAny,
	}

	return outboundConfig, nil
}

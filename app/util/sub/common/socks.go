// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"net/url"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/util/sub"
	"github.com/5vnetwork/vx-core/common/serial"
)

func ParseSocks5FromLink(link string) (*configs.OutboundHandlerConfig, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "socks5" {
		return nil, fmt.Errorf("not a valid socks link, got %s", u.Scheme)
	}

	password, _ := u.User.Password()

	return &configs.OutboundHandlerConfig{
		Address: u.Hostname(),
		Tag:     u.Fragment,
		Ports:   sub.TryParsePorts(u.Port()),
		Protocol: serial.ToTypedMessage(&configs.SocksClientConfig{
			Name:     u.User.Username(),
			Password: password,
		}),
	}, nil
}

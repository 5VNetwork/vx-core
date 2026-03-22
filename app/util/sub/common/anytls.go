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

func ParseAnytls(link string) (*configs.OutboundHandlerConfig, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "anytls" {
		return nil, fmt.Errorf("not a valid anytls link")
	}

	anytlsConfig := &configs.AnytlsClientConfig{
		Password: u.User.Username(),
	}

	q := u.Query()
	if q.Get("security") == "" {
		q.Add("security", "tls")
	}
	transportConfig, err := getTransportConfig(q)
	if err != nil {
		return nil, err
	}

	ports := sub.TryParsePorts(u.Port())
	if len(ports) == 0 {
		return nil, fmt.Errorf("invalid port: %s", u.Port())
	}

	outboundConfig := &configs.OutboundHandlerConfig{
		Address:   u.Hostname(),
		Tag:       u.Fragment,
		Ports:     ports,
		Protocol:  serial.ToTypedMessage(anytlsConfig),
		Transport: transportConfig,
	}

	return outboundConfig, nil
}

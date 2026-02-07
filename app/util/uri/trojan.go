// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package uri

import (
	"net"
	"net/url"
	"strconv"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
)

func toTrojan(outboundConfig *configs.OutboundHandlerConfig) (string, error) {
	config, err := outboundConfig.Protocol.UnmarshalNew()
	if err != nil {
		return "", err
	}
	trojanConfig, _ := config.(*proxy.TrojanClientConfig)

	queryParameters := url.Values{}
	addQueryParameters(queryParameters, outboundConfig)

	u := &url.URL{
		Scheme:   "trojan",
		User:     url.User(trojanConfig.Password),
		Host:     net.JoinHostPort(outboundConfig.Address, strconv.Itoa(getSinglePort(outboundConfig))),
		RawQuery: queryParameters.Encode(),
		Fragment: outboundConfig.Tag,
	}

	return u.String(), nil
}

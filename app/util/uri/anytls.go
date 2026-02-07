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

func toAnytls(outboundConfig *configs.OutboundHandlerConfig) (string, error) {
	config, err := outboundConfig.Protocol.UnmarshalNew()
	if err != nil {
		return "", err
	}
	anytlsConfig, _ := config.(*proxy.AnytlsClientConfig)

	queryParameters := url.Values{}
	addQueryParameters(queryParameters, outboundConfig)

	u := &url.URL{
		Scheme:   "anytls",
		User:     url.User(anytlsConfig.Password),
		Host:     net.JoinHostPort(outboundConfig.Address, strconv.Itoa(getSinglePort(outboundConfig))),
		RawQuery: queryParameters.Encode(),
		Fragment: outboundConfig.Tag,
	}

	return u.String(), nil
}

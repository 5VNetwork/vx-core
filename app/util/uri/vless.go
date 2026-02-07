// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package uri

import (
	"fmt"
	"net/url"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
)

func toVless0(outboundConfig *configs.OutboundHandlerConfig) (string, error) {
	config, err := outboundConfig.Protocol.UnmarshalNew()
	if err != nil {
		return "", err
	}
	vlessConfig, _ := config.(*proxy.VlessClientConfig)

	uuidAddrPort := fmt.Sprintf("%s@%s:%d", vlessConfig.Id,
		outboundConfig.Address, getSinglePort(outboundConfig))
	queryParameters := url.Values{}
	queryParameters.Add("encryption", vlessConfig.Encryption)
	queryParameters.Add("flow", vlessConfig.Flow)
	addQueryParameters(queryParameters, outboundConfig)
	query := queryParameters.Encode()
	remark := url.QueryEscape(outboundConfig.Tag)
	return fmt.Sprintf("vless://%s?%s#%s", uuidAddrPort, query, remark), nil
}

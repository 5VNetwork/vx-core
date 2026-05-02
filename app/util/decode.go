// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"github.com/5vnetwork/vx-core/app/util/sub"
	"github.com/5vnetwork/vx-core/app/util/sub/clash"
	"github.com/5vnetwork/vx-core/app/util/sub/common"
)

func Decode(content string, shareLinkQueryExtra map[string]string) (*sub.DecodeResult, error) {
	result, err := clash.ParseClashConfig([]byte(content), shareLinkQueryExtra)
	if err != nil {
		return common.DecodeCommon(content, shareLinkQueryExtra)
	}
	return result, nil
}

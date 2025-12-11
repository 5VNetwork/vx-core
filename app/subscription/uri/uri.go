package uri

import (
	"fmt"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
)

func ToUrl(outboundConfig *configs.OutboundHandlerConfig) (string, error) {
	proxyClientConfig, err := serial.GetInstanceOf(outboundConfig.Protocol)
	if err != nil {
		return "", fmt.Errorf("failed to get proxy client config: %w", err)
	}
	switch pc := proxyClientConfig.(type) {
	case *proxy.VmessClientConfig:
		return toVmess(outboundConfig)
	case *proxy.ShadowsocksClientConfig:
		return toShadowSocks(outboundConfig)
	case *proxy.TrojanClientConfig:
		return toTrojan(outboundConfig)
	case *proxy.VlessClientConfig:
		return toVless0(outboundConfig)
	case *proxy.Hysteria2ClientConfig:
		return toHysteria(outboundConfig)
	default:
		return "", fmt.Errorf("unsupported proxy client config: %T", pc)
	}
}

func getSinglePort(outboundConfig *configs.OutboundHandlerConfig) int {
	if len(outboundConfig.Ports) > 0 {
		return int(outboundConfig.Ports[0].From)
	}
	return int(outboundConfig.Port)
}

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
	queryParameters.Add("encryption", "none")
	queryParameters.Add("flow", "xtls-rprx-vision")
	tlsConfig := outboundConfig.GetTransport().GetTls()
	if tlsConfig != nil {
		queryParameters.Add("security", "tls")
		queryParameters.Add("sni", tlsConfig.GetServerName())
		queryParameters.Add("fp", tlsConfig.GetImitate())
	}
	wsConfig := outboundConfig.GetTransport().GetWebsocket()
	if wsConfig != nil {
		queryParameters.Add("network", "ws")
		if wsConfig.Path != "" {
			queryParameters.Add("path", wsConfig.Path)
		}
		if wsConfig.Host != "" {
			queryParameters.Add("host", wsConfig.Host)
		}
	}
	grpcConfig := outboundConfig.GetTransport().GetGrpc()
	if grpcConfig != nil {
		queryParameters.Add("network", "grpc")
		if grpcConfig.ServiceName != "" {
			queryParameters.Add("serviceName", grpcConfig.ServiceName)
		}
	}
	query := queryParameters.Encode()
	remark := url.QueryEscape(outboundConfig.Tag)
	return fmt.Sprintf("vless://%s?%s#%s", uuidAddrPort, query, remark), nil
}

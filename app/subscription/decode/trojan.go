package decode

import (
	"fmt"
	"net/url"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
)

// // TrojanConfig represents a Trojan protocol configuration
// type TrojanConfig struct {
// 	Secret  string
// 	Address string
// 	Port    string
// 	Remark  string
// 	Query   map[string]string
// }

func ParseTrojan(link string) (*configs.OutboundHandlerConfig, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "trojan" {
		return nil, fmt.Errorf("not a valid trojan link")
	}

	trojanConfig := &proxy.TrojanClientConfig{
		Password: u.User.Username(),
	}

	q := u.Query()
	q.Add("security", "tls")
	transportConfig, err := getTransportConfig(q)
	if err != nil {
		return nil, err
	}

	ports := TryParsePorts(u.Port())
	if len(ports) == 0 {
		return nil, fmt.Errorf("invalid port: %s", u.Port())
	}

	outboundConfig := &configs.OutboundHandlerConfig{
		Address:   u.Hostname(),
		Tag:       u.Fragment,
		Ports:     ports,
		Protocol:  serial.ToTypedMessage(trojanConfig),
		Transport: transportConfig,
	}

	return outboundConfig, nil
}

// // ParseTrojanFromLink parses a Trojan configuration from a URI link
// func ParseTrojanFromLink(link string) (*TrojanConfig, error) {
// 	if !strings.HasPrefix(link, "trojan://") {
// 		return nil, fmt.Errorf("not a valid trojan link")
// 	}

// 	rest := strings.Split(link, "://")[1]

// 	atIndex := strings.Index(rest, "@")
// 	if atIndex == -1 {
// 		return nil, fmt.Errorf("invalid trojan link format")
// 	}

// 	secret := rest[:atIndex]
// 	rest = rest[atIndex+1:]

// 	colonIndex := strings.Index(rest, ":")
// 	if colonIndex == -1 {
// 		return nil, fmt.Errorf("invalid address:port format")
// 	}

// 	address := rest[:colonIndex]
// 	rest = rest[colonIndex+1:]

// 	markIndex := strings.Index(rest, "?")
// 	if markIndex == -1 {
// 		markIndex = len(rest)
// 	}

// 	portStr := rest[:markIndex]

// 	var queryMap map[string]string
// 	var sharpIndex int

// 	if markIndex < len(rest) {
// 		queryPart := rest[markIndex+1:]
// 		sharpIndex = strings.Index(queryPart, "#")
// 		if sharpIndex == -1 {
// 			sharpIndex = len(queryPart)
// 		}

// 		queryStr := queryPart[:sharpIndex]
// 		values, err := url.ParseQuery(queryStr)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to parse query parameters: %v", err)
// 		}

// 		queryMap = make(map[string]string)
// 		for k, v := range values {
// 			if len(v) > 0 {
// 				queryMap[k] = v[0]
// 			}
// 		}

// 		sharpIndex += markIndex + 1
// 	} else {
// 		sharpIndex = markIndex
// 		queryMap = make(map[string]string)
// 	}

// 	var remark string
// 	var err error
// 	if sharpIndex < len(rest) {
// 		remarkUrlEncoded := rest[sharpIndex+1:]
// 		remark, err = url.QueryUnescape(remarkUrlEncoded)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to decode remark: %v", err)
// 		}
// 	}

// 	return &TrojanConfig{
// 		Secret:  secret,
// 		Address: address,
// 		Port:    portStr,
// 		Query:   queryMap,
// 		Remark:  remark,
// 	}, nil
// }

// // ToProxyHandlerConfig converts TrojanConfig to OutboundHandlerConfig
// func (t *TrojanConfig) ToProxyHandlerConfig() (*configs.OutboundHandlerConfig, error) {
// 	trojanConfig := &proxy.TrojanClientConfig{
// 		Password: t.Secret,
// 	}

// 	trojanAny := serial.ToTypedMessage(trojanConfig)

// 	// Create TLS config
// 	tlsConfig := &tls.TlsConfig{
// 		ServerName: t.Query["sni"],
// 	}

// 	// Set allowInsecure if specified
// 	if t.Query["allowInsecure"] == "1" {
// 		tlsConfig.AllowInsecure = true
// 	}

// 	// Set ALPN protocols if specified
// 	if alpn, ok := t.Query["alpn"]; ok && alpn != "" {
// 		tlsConfig.NextProtocol = strings.Split(alpn, ",")
// 	}

// 	// Create transport config
// 	transportConfig := &configs.TransportConfig{}
// 	transportConfig.Security = &configs.TransportConfig_Tls{Tls: tlsConfig}

// 	// Set transport protocol
// 	switch t.Query["type"] {
// 	case "ws":
// 		host := t.Query["host"]
// 		if host == "" {
// 			host = t.Address
// 		}
// 		wsConfig := &websocket.WebsocketConfig{
// 			Path: t.Query["path"],
// 			Host: host,
// 		}
// 		transportConfig.Protocol = &configs.TransportConfig_Websocket{Websocket: wsConfig}
// 	case "grpc":
// 		grpcConfig := &grpc.GrpcConfig{
// 			ServiceName: t.Query["serviceName"],
// 		}
// 		transportConfig.Protocol = &configs.TransportConfig_Grpc{Grpc: grpcConfig}
// 	}

// 	ports := TryParsePorts(t.Port)
// 	if len(ports) == 0 {
// 		return nil, fmt.Errorf("invalid port: %s", t.Port)
// 	}

// 	return &configs.OutboundHandlerConfig{
// 		Tag:       t.Remark,
// 		Address:   t.Address,
// 		Ports:     ports,
// 		Protocol:  trojanAny,
// 		Transport: transportConfig,
// 	}, nil
// }

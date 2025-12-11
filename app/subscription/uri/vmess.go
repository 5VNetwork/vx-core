package uri

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
)

type vmessConfig struct {
	V    string `json:"v"`
	Ps   string `json:"ps"`  // remark
	Add  string `json:"add"` // addr
	Port string `json:"port"`
	ID   string `json:"id"`
	Aid  string `json:"aid"`            // alterId
	Scy  string `json:"scy"`            // security
	Net  string `json:"net"`            // 传输协议(tcp\kcp\ws\h2\quic)
	Type string `json:"type,omitempty"` // 伪装类型(none\http\srtp\utp\wechat-video) *tcp or kcp or QUIC
	Host string `json:"host,omitempty"`
	Path string `json:"path,omitempty"`
	TLS  string `json:"tls,omitempty"`
	Sni  string `json:"sni,omitempty"`
	Alpn string `json:"alpn,omitempty"`
	Fp   string `json:"fp,omitempty"` // fingerprint
}

func toVmess(outboundConfig *configs.OutboundHandlerConfig) (string, error) {
	config, err := outboundConfig.Protocol.UnmarshalNew()
	if err != nil {
		return "", err
	}
	vmessClientConfig, _ := config.(*proxy.VmessClientConfig)

	vc := &vmessConfig{
		V:    "2",
		Ps:   outboundConfig.Tag,
		Add:  outboundConfig.Address,
		Port: strconv.Itoa(getSinglePort(outboundConfig)),
		ID:   vmessClientConfig.Id,
		Aid:  "0",
		Scy:  "auto",
		Type: "none",
	}

	switch p := outboundConfig.GetTransport().GetProtocol().(type) {
	case *configs.TransportConfig_Tcp:
		vc.Net = "tcp"
	case *configs.TransportConfig_Kcp:
		vc.Net = "kcp"
	case *configs.TransportConfig_Websocket:
		vc.Net = "ws"
		vc.Path = p.Websocket.GetPath()
	case *configs.TransportConfig_Http:
		vc.Net = "h2"
	default:
		vc.Net = "tcp"
	}

	switch s := outboundConfig.GetTransport().GetSecurity().(type) {
	case *configs.TransportConfig_Tls:
		vc.TLS = "tls"
		vc.Sni = s.Tls.GetServerName()
		vc.Alpn = strings.Join(s.Tls.GetNextProtocol(), ",")
		vc.Fp = s.Tls.GetImitate()
	}
	jsonBytes, err := json.Marshal(vc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal vmess config: %w", err)
	}
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonBytes), nil
}

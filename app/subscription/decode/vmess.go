package decode

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/transport/security/tls"
)

// VmessConfig represents a Vmess protocol configuration
type VmessConfig struct {
	V    string      `json:"v,omitempty"`
	PS   string      `json:"ps,omitempty"`
	Add  string      `json:"add"`
	Port json.Number `json:"port"`
	ID   string      `json:"id,omitempty"`
	Aid  json.Number `json:"aid,omitempty"`
	Scy  string      `json:"scy,omitempty"`
	Net  string      `json:"net,omitempty"`
	Type string      `json:"type,omitempty"`
	Host string      `json:"host,omitempty"`
	Path string      `json:"path,omitempty"`
	TLS  string      `json:"tls,omitempty"`
	SNI  string      `json:"sni,omitempty"`
	ALPN string      `json:"alpn,omitempty"`
	FP   string      `json:"fp,omitempty"`
}

// ParseVmessFromJSON parses a VmessConfig from JSON
func ParseVmessFromJSON(data []byte) (*VmessConfig, error) {
	var config VmessConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// // Handle aid conversion if it's an integer in the JSON
	// if config.Aid == "" {
	// 	var jsonMap map[string]interface{}
	// 	if err := json.Unmarshal(data, &jsonMap); err == nil {
	// 		if aid, ok := jsonMap["aid"]; ok {
	// 			switch v := aid.(type) {
	// 			case float64:
	// 				config.Aid = strconv.FormatInt(int64(v), 10)
	// 			case int:
	// 				config.Aid = strconv.Itoa(v)
	// 			}
	// 		}
	// 	}
	// }

	return &config, nil
}

// ToProxyHandlerConfig converts VmessConfig to OutboundHandlerConfig
func (v *VmessConfig) ToProxyHandlerConfig() (*configs.OutboundHandlerConfig, error) {
	port := TryParsePorts(v.Port.String())
	if len(port) == 0 {
		return nil, fmt.Errorf("invalid port: %s", v.Port)
	}

	c := &configs.OutboundHandlerConfig{
		Address: v.Add,
		Ports:   port,
		Tag:     v.PS,
	}

	// Create Vmess config
	vmessConfig := &proxy.VmessClientConfig{
		Id: v.ID,
	}

	// Parse AlterId
	if v.Aid != "" {
		alterId, err := strconv.Atoi(v.Aid.String())
		if err != nil {
			return nil, fmt.Errorf("invalid alterId: %s", v.Aid)
		}
		vmessConfig.AlterId = uint32(alterId)
	}

	// Set security type
	switch v.Scy {
	case "auto":
		vmessConfig.Security = proxy.SecurityType_SecurityType_AUTO
	case "aes-128-gcm":
		vmessConfig.Security = proxy.SecurityType_SecurityType_AES128_GCM
	case "chacha20-poly1305":
		vmessConfig.Security = proxy.SecurityType_SecurityType_CHACHA20_POLY1305
	case "none":
		vmessConfig.Security = proxy.SecurityType_SecurityType_NONE
	case "zero":
		vmessConfig.Security = proxy.SecurityType_SecurityType_ZERO
	default:
		vmessConfig.Security = proxy.SecurityType_SecurityType_AUTO
	}

	// Pack the Vmess config into Any
	vmessAny := serial.ToTypedMessage(vmessConfig)
	c.Protocol = vmessAny

	// Create transport config
	transportConfig := &configs.TransportConfig{}

	// Set header settings based on type
	err := setProtocol(transportConfig, url.Values{
		"type":   {v.Net},
		"header": {v.Type},
		"path":   {v.Path},
		"host":   {v.Host},
	})
	if err != nil {
		return nil, err
	}

	// Set TLS config if enabled
	if v.TLS == "tls" {
		var alpnProtos []string
		if v.ALPN != "" {
			alpnProtos = strings.Split(v.ALPN, ",")
		}

		tlsConfig := &tls.TlsConfig{
			ServerName:   v.SNI,
			Imitate:      v.FP,
			NextProtocol: alpnProtos,
		}
		transportConfig.Security = &configs.TransportConfig_Tls{Tls: tlsConfig}
	}

	c.Transport = transportConfig
	return c, nil
}

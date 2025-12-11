package decode

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
)

func ParseVlessFromLink(link string) (*configs.OutboundHandlerConfig, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "vless" {
		return nil, fmt.Errorf("not a valid vless link")
	}
	query := u.Query()

	vlessConfig := &proxy.VlessClientConfig{
		Id:         u.User.Username(),
		Encryption: "none",
		Flow:       query.Get("flow"),
	}

	transportConfig, err := getTransportConfig(query)
	if err != nil {
		return nil, err
	}

	addr := u.Hostname()
	port := u.Port()

	ports := TryParsePorts(port)
	if len(ports) == 0 {
		return nil, errors.New("port invalid: " + port)
	}

	outboundConfig := &configs.OutboundHandlerConfig{
		Address:   addr,
		Tag:       u.Fragment,
		Ports:     ports,
		Protocol:  serial.ToTypedMessage(vlessConfig),
		Transport: transportConfig,
	}

	return outboundConfig, nil
}

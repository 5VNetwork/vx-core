package decode

import (
	"fmt"
	"net/url"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
)

func ParseSocks5FromLink(link string) (*configs.OutboundHandlerConfig, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "socks5" {
		return nil, fmt.Errorf("not a valid socks link, got %s", u.Scheme)
	}

	password, _ := u.User.Password()

	return &configs.OutboundHandlerConfig{
		Address: u.Hostname(),
		Tag:     u.Fragment,
		Ports:   TryParsePorts(u.Port()),
		Protocol: serial.ToTypedMessage(&proxy.SocksClientConfig{
			Name:     u.User.Username(),
			Password: password,
		}),
	}, nil
}

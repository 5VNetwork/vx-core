//go:build android

package x_android

import "github.com/5vnetwork/vx-core/app/configs"

type TunConfig interface {
	GetName() string
	GetCidr4() string
	GetCidr6() string
	GetMtu() int32
	GetDns4() StringList
	GetDns6() StringList
	GetRoutes4() StringList
	GetRoutes6() StringList
	GetWhiteListApps() StringList
	GetBlackListApps() StringList
}

func ToTunConfig(config *configs.TunDeviceConfig, enable6 bool) TunConfig {
	t := &tunConfig{
		Name:          config.Name,
		Cidr4:         config.Cidr4,
		Mtu:           int32(config.Mtu),
		Dns4:          config.Dns4,
		Routes4:       config.Routes4,
		WhiteListApps: config.WhiteListApps,
		BlackListApps: config.BlackListApps,
	}
	if enable6 {
		t.Cidr6 = config.Cidr6
		t.Dns6 = config.Dns6
		t.Routes6 = config.Routes6
	}
	return t
}

// implement TunConfig
type tunConfig struct {
	Name          string
	Cidr4         string
	Cidr6         string
	Mtu           int32
	Dns4          []string
	Dns6          []string
	Routes4       []string
	Routes6       []string
	WhiteListApps []string
	BlackListApps []string
}

func (t *tunConfig) GetName() string {
	return t.Name
}

func (t *tunConfig) GetCidr4() string {
	return t.Cidr4
}

func (t *tunConfig) GetCidr6() string {
	return t.Cidr6
}

func (t *tunConfig) GetMtu() int32 {
	return t.Mtu
}

func (t *tunConfig) GetDns4() StringList {
	return &stringList{strings: t.Dns4}
}

func (t *tunConfig) GetDns6() StringList {
	return &stringList{strings: t.Dns6}
}

func (t *tunConfig) GetRoutes4() StringList {
	return &stringList{strings: t.Routes4}
}

func (t *tunConfig) GetRoutes6() StringList {
	return &stringList{strings: t.Routes6}
}

func (t *tunConfig) GetWhiteListApps() StringList {
	return &stringList{strings: t.WhiteListApps}
}

func (t *tunConfig) GetBlackListApps() StringList {
	return &stringList{strings: t.BlackListApps}
}

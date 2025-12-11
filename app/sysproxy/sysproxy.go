package sysproxy

type ProxySetting struct {
	*SocksProxySetting
	*HttpProxySetting
	*HttpsProxySetting
}

type SocksProxySetting struct {
	Address string
	Port    uint16
}

type HttpProxySetting struct {
	Address string
	Port    uint16
}

type HttpsProxySetting struct {
	Address string
	Port    uint16
}

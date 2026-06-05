package wireguard

import (
	"errors"
	"net/netip"
	"strings"

	"buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/wireguard"
	"github.com/rs/zerolog/log"
	"golang.zx2c4.com/wireguard/device"
)

var wgLogger = &device.Logger{
	Verbosef: func(format string, args ...any) {
		log.Debug().Msgf(format, args...)
	},
	Errorf: func(format string, args ...any) {
		log.Error().Msgf(format, args...)
	},
}

// convert endpoint string to netip.Addr
func parseEndpoints(conf *wireguard.DeviceConfig) ([]netip.Addr, bool, bool, error) {
	var hasIPv4, hasIPv6 bool

	endpoints := make([]netip.Addr, len(conf.Endpoint))
	for i, str := range conf.Endpoint {
		var addr netip.Addr
		if strings.Contains(str, "/") {
			prefix, err := netip.ParsePrefix(str)
			if err != nil {
				return nil, false, false, err
			}
			addr = prefix.Addr()
			if prefix.Bits() != addr.BitLen() {
				return nil, false, false, errors.New("interface address subnet should be /32 for IPv4 and /128 for IPv6")
			}
		} else {
			var err error
			addr, err = netip.ParseAddr(str)
			if err != nil {
				return nil, false, false, err
			}
		}
		endpoints[i] = addr

		if addr.Is4() {
			hasIPv4 = true
		} else if addr.Is6() {
			hasIPv6 = true
		}
	}

	return endpoints, hasIPv4, hasIPv6, nil
}

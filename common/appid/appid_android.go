//go:build android

package appid

import (
	"context"
	"errors"

	"github.com/5vnetwork/vx-core/common/net"
	"golang.org/x/sys/unix"
)

type getPackageNameFunc func(protocol int, source string, sourcePort int, destination string, destinationPort int) (string, error)

var GetPackageName getPackageNameFunc

func GetAppId(ctx context.Context, src net.Destination, dst *net.Destination) (string, error) {
	if GetPackageName == nil {
		return "", errors.New("getPackageName is not set")
	}

	if dst == nil {
		return "", errors.New("dst is nil")
	}

	if dst.Address.Family().IsDomain() {
		return "", nil
	}

	if src.Address == nil {
		return "", nil
	}

	protocol := unix.IPPROTO_TCP
	if src.Network == net.Network_UDP {
		protocol = unix.IPPROTO_UDP
	}

	return GetPackageName(protocol, src.Address.String(), int(src.Port), dst.Address.String(), int(dst.Port))
}

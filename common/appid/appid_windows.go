package appid

import (
	"context"
	"errors"
	"fmt"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/rs/zerolog/log"

	gnet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

func GetAppId(ctx context.Context, src net.Destination, dst *net.Destination) (string, error) {
	if !src.IsValid() {
		return "", errors.New("invalid source destination")
	}
	var kind string
	ipv4 := src.Address.Family().IsIPv4()
	isTCP := src.Network == net.Network_TCP
	isUDP := src.Network == net.Network_UDP
	if isTCP {
		if ipv4 {
			kind = "tcp4"
		} else {
			kind = "tcp6"
		}
	} else if isUDP {
		if ipv4 {
			kind = "udp4"
		} else {
			kind = "udp6"
		}
	} else {
		return "", errors.New("unsupported network")
	}

	conns, err := gnet.ConnectionsWithContext(ctx, kind)
	if err != nil {
		return "", err
	}

	for _, connStat := range conns {
		// src port
		if connStat.Laddr.Port != uint32(src.Port) {
			continue
		}
		// src ip
		srcIpA := net.ParseIP(connStat.Laddr.IP)
		srcIpB := net.IP(src.Address.IP()).To16()
		if srcIpA != nil && !srcIpA.IsUnspecified() && !srcIpA.Equal(srcIpB) {
			log.Ctx(ctx).Debug().Str("srcIpA", srcIpA.String()).Str("srcIpB", srcIpB.String()).Msg("src ip not equal")
			continue
		}

		// // dst port
		// if dst != nil && isTCP {
		// 	if connStat.Raddr.Port != uint32(dst.Port) {
		// 		continue
		// 	}
		// }
		p, err := process.NewProcessWithContext(ctx, connStat.Pid)
		if err != nil {
			return "", fmt.Errorf("failed to get process: %w", err)
		}
		name, err := p.Exe()
		if err != nil {
			return "", fmt.Errorf("failed to get process name: %w", err)
		}
		return name, nil
	}
	return "", errors.New("not found")
}

// a is ip string in gopsutil
func ipEqual(a, b string) bool {
	if a == b {
		return true
	}
	if a == "*" && (b == "0.0.0.0" || b == "::") {
		return true
	}
	return false
}

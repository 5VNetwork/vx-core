//go:build darwin && !ios
// +build darwin,!ios

package appid

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/rs/zerolog/log"

	"golang.org/x/sys/unix"
)

func GetAppId(ctx context.Context, src net.Destination, dst *net.Destination) (string, error) {
	sysctlName := ""
	// Choose PCB list based on protocol
	if src.Network == net.Network_TCP {
		sysctlName = "net.inet.tcp.pcblist_n"
	} else if src.Network == net.Network_UDP {
		sysctlName = "net.inet.udp.pcblist_n"
	} else {
		return "", errors.New("unsupported network")
	}

	data, err := unix.SysctlRaw(sysctlName)
	if err != nil {
		return "", fmt.Errorf("failed to get PCB list: %w", err)
	}

	// Calculate structure sizes
	itemLen := structSize1 // base size for UDP
	if src.Network == net.Network_TCP {
		itemLen += 208 // additional TCP structure size
	}

	// isIPv4 := src.Address.Family().IsIPv4()
	srcAddr, ok := netip.AddrFromSlice(src.Address.IP())
	if !ok {
		return "", errors.New("invalid source address")
	}
	srcAddr = srcAddr.Unmap()

	// Skip first 24 bytes (xinpgen header)
	for offset := 24; offset+itemLen <= len(data); offset += itemLen {
		// Check port first (offset 18, 2 bytes)
		srcPort := binary.BigEndian.Uint16(data[offset+18 : offset+20])
		if srcPort != uint16(src.Port) {
			continue
		}

		log.Ctx(ctx).Debug().Msg("same src port found")

		// Check IP version flag (offset 44, 1 byte)
		flag := data[offset+44]

		var srcIP netip.Addr
		switch {
		case flag&0x1 > 0:
			// IPv4 address at offset 76
			var ipv4 [4]byte
			copy(ipv4[:], data[offset+76:offset+80])
			srcIP = netip.AddrFrom4(ipv4)
		case flag&0x2 > 0:
			// IPv6 address at offset 64
			var ipv6 [16]byte
			copy(ipv6[:], data[offset+64:offset+80])
			srcIP = netip.AddrFrom16(ipv6)
		default:
			log.Ctx(ctx).Debug().Msg("no ip found")
			continue
		}

		// Compare src addresses
		if !srcIP.IsUnspecified() {
			if srcAddr != srcIP {
				log.Ctx(ctx).Debug().Str("srcAddr", srcAddr.String()).Str("srcIP", srcIP.String()).Msg("srcAddr != srcIP")
				continue
			}
		}

		// Get PID from socket structure (offset +104, then +68)
		socketOffset := offset + 104
		pid := *(*uint32)(unsafe.Pointer(&data[socketOffset+68]))
		if pid == 0 {
			continue // Skip entries with no associated process
		}

		return GetProcessPath(pid)
	}

	return "", errors.New("no process found for socket")
}

func GetProcessPath(pid uint32) (string, error) {
	const maxPathLen = 1024
	// Buffer for path (MAXPATHLEN is typically 1024 on Darwin)
	buf := make([]byte, maxPathLen)

	// PROC_PIDPATHINFO (0xb) - get path of process
	_, _, errno := syscall.Syscall6(
		syscall.SYS_PROC_INFO,
		2,            // PROC_INFO_CALL_PIDINFO
		uintptr(pid), // PID
		0xb,          // PROC_PIDPATHINFO
		0,
		uintptr(unsafe.Pointer(&buf[0])), // buffer
		uintptr(maxPathLen))              // buffer size

	if errno != 0 {
		return "", errno
	}

	return unix.ByteSliceToString(buf), nil
}

var structSize1 int

func init() {
	value, _ := syscall.Sysctl("kern.osrelease")
	major, _, _ := strings.Cut(value, ".")
	n, _ := strconv.ParseInt(major, 10, 64)
	switch true {
	case n >= 22:
		structSize1 = 408
	default:
		structSize1 = 384
	}
}

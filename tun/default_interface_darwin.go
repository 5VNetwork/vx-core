// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package tun

import (
	"errors"
	"fmt"
	"net"
	"strings"
	sync "sync"
	"syscall"
	"unsafe"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
	"tailscale.com/util/mak"
)

// ErrNoGatewayIndexFound is returned by DefaultRouteInterfaceIndex when no
// default route is found.
var ErrNoGatewayIndexFound = errors.New("no gateway index found")

// DefaultRouteInterfaceIndex returns the index of the network interface that
// owns the default route. It returns the first IPv4 or IPv6 default route it
// finds (it does not prefer one or the other).
// This one is different from original one in that it could return a interface with
// RTF_IFSCOPE to be a DefaultRouteInterface
func DefaultRouteInterfaceIndex() (int, error) {
	// $ netstat -nr
	// Routing tables
	// Internet:
	// Destination        Gateway            Flags        Netif Expire
	// default            10.0.0.1           UGSc           en0         <-- want this one
	// default            10.0.0.1           UGScI          en1

	// From man netstat:
	// U       RTF_UP           Route usable
	// G       RTF_GATEWAY      Destination requires forwarding by intermediary
	// S       RTF_STATIC       Manually added
	// c       RTF_PRCLONING    Protocol-specified generate new routes on use
	// I       RTF_IFSCOPE      Route is associated with an interface scope

	rib, err := fetchRoutingTable()
	if err != nil {
		return 0, fmt.Errorf("route.FetchRIB: %w", err)
	}
	msgs, err := parseRoutingTable(rib)
	if err != nil {
		return 0, fmt.Errorf("route.ParseRIB: %w", err)
	}
	for _, m := range msgs {
		rm, ok := m.(*route.RouteMessage)
		if !ok {
			continue
		}

		if zerolog.GlobalLevel() == zerolog.DebugLevel {
			printRoute(rm)
		}

		if isDefaultGateway(rm) {
			if delegatedIndex, err := getDelegatedInterface(rm.Index); err == nil && delegatedIndex != 0 {
				return delegatedIndex, nil
			} else if err != nil {
				log.Printf("interfaces_bsd: could not get delegated interface: %v", err)
			}
			return rm.Index, nil
		}
	}
	return 0, ErrNoGatewayIndexFound
}

// Print the RouteMessage in a way similar to "netstat -nr" using zerolog
func printRoute(rm *route.RouteMessage) {
	// Get interface name
	ifNames.Lock()
	ifName, ok := ifNames.m[rm.Index]
	if !ok {
		iface, err := net.InterfaceByIndex(rm.Index)
		if err != nil {
			ifName = fmt.Sprintf("if%d", rm.Index)
		} else {
			ifName = iface.Name
			mak.Set(&ifNames.m, rm.Index, ifName)
		}
	}
	ifNames.Unlock()

	// Format destination
	var dst string
	if len(rm.Addrs) > unix.RTAX_DST && rm.Addrs[unix.RTAX_DST] != nil {
		addr := rm.Addrs[unix.RTAX_DST]
		if addr.Family() == syscall.AF_INET {
			if dstAddr, ok := addr.(*route.Inet4Addr); ok {
				dst = net.IP(dstAddr.IP[:]).String()
			}
		} else if addr.Family() == syscall.AF_INET6 {
			if dstAddr, ok := addr.(*route.Inet6Addr); ok {
				dst = net.IP(dstAddr.IP[:]).String()
			}
		}
		if dst == "" {
			dst = "default"
		}
	} else {
		dst = "default"
	}

	// Format gateway
	var gateway string
	if len(rm.Addrs) > unix.RTAX_GATEWAY && rm.Addrs[unix.RTAX_GATEWAY] != nil {
		addr := rm.Addrs[unix.RTAX_GATEWAY]
		if addr.Family() == syscall.AF_INET {
			if gwAddr, ok := addr.(*route.Inet4Addr); ok {
				gateway = net.IP(gwAddr.IP[:]).String()
			}
		} else if addr.Family() == syscall.AF_INET6 {
			if gwAddr, ok := addr.(*route.Inet6Addr); ok {
				gateway = net.IP(gwAddr.IP[:]).String()
			}
		}
		if gateway == "" {
			gateway = "-"
		}
	} else {
		gateway = "-"
	}

	// Format flags
	var flags string
	if rm.Flags&unix.RTF_UP != 0 {
		flags += "U"
	}
	if rm.Flags&unix.RTF_GATEWAY != 0 {
		flags += "G"
	}
	if rm.Flags&unix.RTF_STATIC != 0 {
		flags += "S"
	}
	if rm.Flags&unix.RTF_PRCLONING != 0 {
		flags += "c"
	}
	// RTF_IFSCOPE is not defined in unix package, but we can check for it
	// const RTF_IFSCOPE = 0x1000000
	if rm.Flags&0x1000000 != 0 {
		flags += "I"
	}

	// Print in netstat format: Destination Gateway Flags Netif
	log.Debug().Str("dst", dst).Str("gateway", gateway).
		Str("flags", flags).Str("ifName", ifName).Msg("netstat -nr route")
}

// fetchRoutingTable calls route.FetchRIB, fetching NET_RT_DUMP2.
func fetchRoutingTable() (rib []byte, err error) {
	return route.FetchRIB(syscall.AF_UNSPEC, syscall.NET_RT_DUMP2, 0)
}

func parseRoutingTable(rib []byte) ([]route.Message, error) {
	return route.ParseRIB(syscall.NET_RT_IFLIST2, rib)
}

var ifNames struct {
	sync.Mutex
	m map[int]string // ifindex => name
}

var v4default = [4]byte{0, 0, 0, 0}
var v6default = [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func isDefaultGateway(rm *route.RouteMessage) bool {
	if rm.Flags&unix.RTF_GATEWAY == 0 {
		return false
	}
	// Defined locally because FreeBSD does not have unix.RTF_IFSCOPE.
	// const RTF_IFSCOPE = 0x1000000
	// if rm.Flags&RTF_IFSCOPE != 0 {
	// 	return false
	// }

	// Addrs is [RTAX_DST, RTAX_GATEWAY, RTAX_NETMASK, ...]
	if len(rm.Addrs) <= unix.RTAX_NETMASK {
		return false
	}

	dst := rm.Addrs[unix.RTAX_DST]
	netmask := rm.Addrs[unix.RTAX_NETMASK]

	if dst == nil || netmask == nil {
		return false
	}

	if dst.Family() == syscall.AF_INET && netmask.Family() == syscall.AF_INET {
		dstAddr, dstOk := dst.(*route.Inet4Addr)
		nmAddr, nmOk := netmask.(*route.Inet4Addr)
		if dstOk && nmOk && dstAddr.IP == v4default && nmAddr.IP == v4default {
			return true
		}
	}

	if dst.Family() == syscall.AF_INET6 && netmask.Family() == syscall.AF_INET6 {
		dstAddr, dstOk := dst.(*route.Inet6Addr)
		nmAddr, nmOk := netmask.(*route.Inet6Addr)

		if dstOk && nmOk && dstAddr.IP == v6default && nmAddr.IP == v6default {
			return true
		}
	}

	return false
}

// getDelegatedInterface returns the interface index of the underlying interface
// for the given interface index. 0 is returned if the interface does not
// delegate.
func getDelegatedInterface(ifIndex int) (int, error) {
	ifNames.Lock()
	defer ifNames.Unlock()

	// To get the delegated interface, we do what ifconfig does and use the
	// SIOCGIFDELEGATE ioctl. It operates in term of a ifreq struct, which
	// has to be populated with a interface name. To avoid having to do a
	// interface index -> name lookup every time, we cache interface names
	// (since indexes and names are stable after boot).
	ifName, ok := ifNames.m[ifIndex]
	if !ok {
		iface, err := net.InterfaceByIndex(ifIndex)
		if err != nil {
			return 0, err
		}
		ifName = iface.Name
		mak.Set(&ifNames.m, ifIndex, ifName)
	}

	// Only tunnels (like Tailscale itself) have a delegated interface, avoid
	// the ioctl if we can.
	if !strings.HasPrefix(ifName, "utun") {
		return 0, nil
	}

	// We don't cache the result of the ioctl, since the delegated interface can
	// change, e.g. if the user changes the preferred service order in the
	// network preference pane.
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return 0, err
	}
	defer unix.Close(fd)

	// Match the ifreq struct/union from the bsd/net/if.h header in the Darwin
	// open source release.
	var ifr struct {
		ifr_name      [unix.IFNAMSIZ]byte
		ifr_delegated uint32
	}
	copy(ifr.ifr_name[:], ifName)

	// SIOCGIFDELEGATE is not in the Go x/sys package or in the public macOS
	// <sys/sockio.h> headers. However, it is in the Darwin/xnu open source
	// release (and is used by ifconfig, see
	// https://github.com/apple-oss-distributions/network_cmds/blob/6ccdc225ad5aa0d23ea5e7d374956245d2462427/ifconfig.tproj/ifconfig.c#L2183-L2187).
	// We generate its value by evaluating the `_IOWR('i', 157, struct ifreq)`
	// macro, which is how it's defined in
	// https://github.com/apple/darwin-xnu/blob/2ff845c2e033bd0ff64b5b6aa6063a1f8f65aa32/bsd/sys/sockio.h#L264
	const SIOCGIFDELEGATE = 0xc020699d

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(SIOCGIFDELEGATE),
		uintptr(unsafe.Pointer(&ifr)))
	if errno != 0 {
		return 0, errno
	}
	return int(ifr.ifr_delegated), nil
}

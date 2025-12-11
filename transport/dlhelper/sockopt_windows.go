package dlhelper

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"syscall"
	"unsafe"

	"github.com/5vnetwork/vx-core/common/errors"

	"github.com/rs/zerolog/log"

	"golang.org/x/sys/windows"
)

const (
	TCP_FASTOPEN    = 15 // nolint: revive,stylecheck
	IP_UNICAST_IF   = 31 // nolint: revive,stylecheck
	IPV6_UNICAST_IF = 31 // nolint: revive,stylecheck
)

func setTFO(fd syscall.Handle, settings SocketConfig_TCPFastOpenState) error {
	switch settings {
	case SocketConfig_Enable:
		if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, TCP_FASTOPEN, 1); err != nil {
			return err
		}
	case SocketConfig_Disable:
		if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, TCP_FASTOPEN, 0); err != nil {
			return err
		}
	}
	return nil
}

// [address] is the address parameter passed to dialer.Control. In most cases (maybe all cases), it is an ip address.
// [address] is the destination of dial.
func applyOutboundSocketOptions(ctx context.Context, network string, address string, fd uintptr, config *SocketSetting) error {
	log.Ctx(ctx).Debug().Str("network", network).Str("address", address).
		Uint32("bd4", config.GetBindToDevice4()).Uint32("bd6", config.GetBindToDevice6()).Msg("applyOutboundSocketOptions")

	if isTCPSocket(network) {
		if err := setTFO(syscall.Handle(fd), config.GetTfo()); err != nil {
			return err
		}
		if config.GetTcpKeepAliveIdle() > 0 {
			if err := syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
				return errors.New("failed to set SO_KEEPALIVE", err)
			}
		}
	}

	host, _, _ := net.SplitHostPort(address)
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		log.Ctx(ctx).Warn().Msg("address is loopback")
		return nil
	}

	var device4 uint32
	/*
		if udp6 and (ip is nil or zero), the conn might be used to send ipv4 packets too (This happens when
		the network in the top level listen is udp, and address is unspecified or zero ip).
		so if udp6 and (ip is nil or zero), bind to device4, if bind failed (this could happen when the top level listen has "udp6"
		as its "network" parameter and address is unspecified or zero), the err will be ignored
	*/
	if strings.Contains(network, "4") || (network == "udp6" && (ip == nil || ip.Equal(net.IPv4zero))) {
		if config.GetBindToDevice4() != 0 {
			device4 = config.GetBindToDevice4()
		}
	}

	var device6 uint32
	if strings.Contains(network, "6") {
		if config.GetBindToDevice6() != 0 {
			device6 = config.GetBindToDevice6()
		}
	}

	isUdp := strings.HasPrefix(network, "udp")

	if device4 != 0 {
		var bytes [4]byte
		binary.BigEndian.PutUint32(bytes[:], device4)
		index := *(*uint32)(unsafe.Pointer(&bytes[0]))
		if err := windows.SetsockoptInt(
			windows.Handle(fd), windows.IPPROTO_IP, IP_UNICAST_IF, int(index)); err != nil {
			if network != "udp6" {
				return fmt.Errorf("failed to set IP_UNICAST_IF, %w", err)
			}
		}
		if isUdp {
			if err := windows.SetsockoptInt(
				windows.Handle(fd), windows.IPPROTO_IP, syscall.IP_MULTICAST_IF, int(index)); err != nil {
				if network != "udp6" {
					return fmt.Errorf("failed to set IP_MULTICAST_IF, %w", err)
				}
			}
		}
	}
	if device6 != 0 {
		if err := windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IPV6,
			IPV6_UNICAST_IF, int(device6)); err != nil {
			return fmt.Errorf("failed to set IPV6_UNICAST_IF, %w", err)
		}
		if isUdp {
			if err := windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IPV6,
				syscall.IPV6_MULTICAST_IF, int(device6)); err != nil {
				return fmt.Errorf("failed to set IPV6_MULTICAST_IF, %w", err)
			}
		}
	}

	if config.GetTxBufSize() != 0 {
		if err := windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_SNDBUF, int(config.TxBufSize)); err != nil {
			return errors.New("failed to set SO_SNDBUF").Base(err)
		}
	}

	if config.GetRxBufSize() != 0 {
		if err := windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_RCVBUF, int(config.TxBufSize)); err != nil {
			return errors.New("failed to set SO_RCVBUF").Base(err)
		}
	}

	return nil
}

func applyInboundSocketOptions(ctx context.Context, network, address string, fd uintptr, config *SocketSetting) error {
	return applyOutboundSocketOptions(ctx, network, address, fd, config)
}

func bindAddr(fd uintptr, ip []byte, port uint32) error {
	return nil
}

func setReuseAddr(fd uintptr) error {
	return nil
}

func setReusePort(fd uintptr) error {
	return nil
}

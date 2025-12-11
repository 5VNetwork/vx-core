package dlhelper

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/rs/zerolog/log"

	"golang.org/x/sys/unix"
)

const (
	// For incoming connections.
	TCP_FASTOPEN = 23 // nolint: revive,stylecheck
	// For out-going connections.
	TCP_FASTOPEN_CONNECT = 30 // nolint: revive,stylecheck
)

func bindAddr(fd uintptr, ip []byte, port uint32) error {
	setReuseAddr(fd)
	setReusePort(fd)

	var sockaddr syscall.Sockaddr

	switch len(ip) {
	case net.IPv4len:
		a4 := &syscall.SockaddrInet4{
			Port: int(port),
		}
		copy(a4.Addr[:], ip)
		sockaddr = a4
	case net.IPv6len:
		a6 := &syscall.SockaddrInet6{
			Port: int(port),
		}
		copy(a6.Addr[:], ip)
		sockaddr = a6
	default:
		return errors.New("unexpected length of ip")
	}

	return syscall.Bind(int(fd), sockaddr)
}

func applyOutboundSocketOptions(ctx context.Context, network string, address string, fd uintptr, config *SocketSetting) error {
	log.Ctx(ctx).Debug().Str("network", network).Str("address", address).
		Uint32("bd4", config.GetBindToDevice4()).Uint32("bd6", config.GetBindToDevice6()).
		Str("bdn", config.BindToDeviceName).Msg("applyOutboundSocketOptions")

	if config.FdFunc != nil {
		if err := config.FdFunc(fd); err != nil {
			return fmt.Errorf("failed to set FdFunc: %w", err)
		}
	}

	if config.Mark != 0 {
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_MARK, int(config.Mark)); err != nil {
			return errors.New("failed to set SO_MARK").Base(err)
		}
	}

	if isTCPSocket(network) {
		switch config.Tfo {
		case SocketConfig_Enable:
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, TCP_FASTOPEN_CONNECT, 1); err != nil {
				return errors.New("failed to set TCP_FASTOPEN_CONNECT=1").Base(err)
			}
		case SocketConfig_Disable:
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, TCP_FASTOPEN_CONNECT, 0); err != nil {
				return errors.New("failed to set TCP_FASTOPEN_CONNECT=0").Base(err)
			}
		}

		if config.TcpKeepAliveInterval > 0 {
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, int(config.TcpKeepAliveInterval)); err != nil {
				return errors.New("failed to set TCP_KEEPINTVL", err)
			}
		}
		if config.TcpKeepAliveIdle > 0 {
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, int(config.TcpKeepAliveIdle)); err != nil {
				return errors.New("failed to set TCP_KEEPIDLE", err)
			}
		}
		if config.TcpKeepAliveInterval > 0 || config.TcpKeepAliveIdle > 0 {
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
				return errors.New("failed to set SO_KEEPALIVE").Base(err)
			}
		}
	}

	if config.Tproxy.IsEnabled() {
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
			return errors.New("failed to set IP_TRANSPARENT").Base(err)
		}
	}

	//TODO
	if config.BindToDeviceName != "" {
		if err := unix.BindToDevice(int(fd), config.BindToDeviceName); err != nil {
			return errors.New("failed to set SO_BINDTODEVICE").Base(err)
		}
	}

	if config.TxBufSize != 0 {
		syscallTarget := unix.SO_SNDBUF
		if config.ForceBufSize {
			syscallTarget = unix.SO_SNDBUFFORCE
		}
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, syscallTarget, int(config.TxBufSize)); err != nil {
			return errors.New("failed to set SO_SNDBUF/SO_SNDBUFFORCE").Base(err)
		}
	}

	if config.RxBufSize != 0 {
		syscallTarget := unix.SO_RCVBUF
		if config.ForceBufSize {
			syscallTarget = unix.SO_RCVBUFFORCE
		}
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, syscallTarget, int(config.RxBufSize)); err != nil {
			return errors.New("failed to set SO_RCVBUF/SO_RCVBUFFORCE").Base(err)
		}
	}

	return nil
}

func applyInboundSocketOptions(ctx context.Context, network, address string, fd uintptr, config *SocketSetting) error {
	log.Ctx(ctx).Debug().Str("network", network).Str("address", address).
		Uint32("bd4", config.GetBindToDevice4()).Uint32("bd6", config.GetBindToDevice6()).
		Str("bdn", config.BindToDeviceName).Msg("applyInboundSocketOptions")

	if config.FdFunc != nil {
		if err := config.FdFunc(fd); err != nil {
			return fmt.Errorf("failed to set FdFunc: %w", err)
		}
	}

	if config.Mark != 0 {
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_MARK, int(config.Mark)); err != nil {
			return errors.New("failed to set SO_MARK").Base(err)
		}
	}
	if isTCPSocket(network) {
		switch config.Tfo {
		case SocketConfig_Enable:
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, TCP_FASTOPEN, int(config.TfoQueueLength)); err != nil {
				return errors.New("failed to set TCP_FASTOPEN=", config.TfoQueueLength).Base(err)
			}
		case SocketConfig_Disable:
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, TCP_FASTOPEN, 0); err != nil {
				return errors.New("failed to set TCP_FASTOPEN=0").Base(err)
			}
		}

		if config.TcpKeepAliveInterval > 0 {
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, int(config.TcpKeepAliveInterval)); err != nil {
				return errors.New("failed to set TCP_KEEPINTVL", err)
			}
		}
		if config.TcpKeepAliveIdle > 0 {
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, int(config.TcpKeepAliveIdle)); err != nil {
				return errors.New("failed to set TCP_KEEPIDLE", err)
			}
		}
		if config.TcpKeepAliveInterval > 0 || config.TcpKeepAliveIdle > 0 {
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
				return errors.New("failed to set SO_KEEPALIVE", err)
			}
		}
	}

	if config.Tproxy.IsEnabled() {
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
			return errors.New("failed to set IP_TRANSPARENT").Base(err)
		}
	}

	if config.ReceiveOriginalDestAddress && isUDPSocket(network) {
		err1 := syscall.SetsockoptInt(int(fd), syscall.SOL_IPV6, unix.IPV6_RECVORIGDSTADDR, 1)
		err2 := syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_RECVORIGDSTADDR, 1)
		if err1 != nil && err2 != nil {
			return err1
		}
	}
	//TODO
	if config.BindToDeviceName != "" {
		if err := unix.BindToDevice(int(fd), config.BindToDeviceName); err != nil {
			return errors.New("failed to set SO_BINDTODEVICE").Base(err)
		}
	}

	if config.TxBufSize != 0 {
		syscallTarget := unix.SO_SNDBUF
		if config.ForceBufSize {
			syscallTarget = unix.SO_SNDBUFFORCE
		}
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, syscallTarget, int(config.TxBufSize)); err != nil {
			return errors.New("failed to set SO_SNDBUF/SO_SNDBUFFORCE").Base(err)
		}
	}

	if config.RxBufSize != 0 {
		syscallTarget := unix.SO_RCVBUF
		if config.ForceBufSize {
			syscallTarget = unix.SO_RCVBUFFORCE
		}
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, syscallTarget, int(config.RxBufSize)); err != nil {
			return errors.New("failed to set SO_RCVBUF/SO_RCVBUFFORCE").Base(err)
		}
	}
	return nil
}

func setReuseAddr(fd uintptr) error {
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return errors.New("failed to set SO_REUSEADDR").Base(err)
	}
	return nil
}

func setReusePort(fd uintptr) error {
	if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return errors.New("failed to set SO_REUSEPORT").Base(err)
	}
	return nil
}

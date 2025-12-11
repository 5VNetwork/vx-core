package dlhelper

import (
	"context"
	"strings"

	"github.com/5vnetwork/vx-core/common/errors"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

const (
	// TCP_FASTOPEN_SERVER is the value to enable TCP fast open on darwin for server connections.
	TCP_FASTOPEN_SERVER = 0x01 // nolint: revive,stylecheck
	// TCP_FASTOPEN_CLIENT is the value to enable TCP fast open on darwin for client connections.
	TCP_FASTOPEN_CLIENT = 0x02 // nolint: revive,stylecheck
)

func applyOutboundSocketOptions(ctx context.Context, network string, address string, fd uintptr, config *SocketSetting) error {
	log.Ctx(ctx).Debug().Str("network", network).Str("address", address).
		Uint32("bd", config.GetBindToDevice4()).Msg("applyOutboundSocketOptions")
	if isTCPSocket(network) {
		if config.TcpKeepAliveInterval > 0 {
			if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, int(config.TcpKeepAliveInterval)); err != nil {
				return errors.New("failed to set TCP_KEEPINTVL", err)
			}
		}
		if config.TcpKeepAliveIdle > 0 {
			if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPALIVE, int(config.TcpKeepAliveIdle)); err != nil {
				return errors.New("failed to set TCP_KEEPALIVE (TCP keepalive idle time on Darwin)", err)
			}
		}

		if config.TcpKeepAliveInterval > 0 || config.TcpKeepAliveIdle > 0 {
			if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil {
				return errors.New("failed to set SO_KEEPALIVE", err)
			}
		}
	}

	if strings.Contains(network, "6") && config.GetBindToDevice6() != 0 {
		if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_BOUND_IF, int(config.GetBindToDevice6())); err != nil {
			return errors.New("failed to set IPV6_BOUND_IF", err)
		}
	}
	if strings.Contains(network, "4") && config.GetBindToDevice4() != 0 {
		if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF, int(config.GetBindToDevice4())); err != nil {
			return errors.New("failed to set IP_BOUND_IF", err)
		}
	}

	if config.TxBufSize != 0 {
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_SNDBUF, int(config.TxBufSize)); err != nil {
			return errors.New("failed to set SO_SNDBUF", err)
		}
	}

	if config.RxBufSize != 0 {
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_RCVBUF, int(config.RxBufSize)); err != nil {
			return errors.New("failed to set SO_RCVBUF", err)
		}
	}

	return nil
}

func applyInboundSocketOptions(ctx context.Context, network, address string, fd uintptr, config *SocketSetting) error {
	log.Ctx(ctx).Debug().Str("network", network).Str("address", address).
		Uint32("bd", config.GetBindToDevice4()).Msg("applyInboundSocketOptions")

	if isTCPSocket(network) {
		if config.TcpKeepAliveInterval > 0 {
			if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, int(config.TcpKeepAliveInterval)); err != nil {
				return errors.New("failed to set TCP_KEEPINTVL", err)
			}
		}
		if config.TcpKeepAliveIdle > 0 {
			if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPALIVE, int(config.TcpKeepAliveIdle)); err != nil {
				return errors.New("failed to set TCP_KEEPALIVE (TCP keepalive idle time on Darwin)", err)
			}
		}
		if config.TcpKeepAliveInterval > 0 || config.TcpKeepAliveIdle > 0 {
			if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil {
				return errors.New("failed to set SO_KEEPALIVE", err)
			}
		}
	}

	if strings.Contains(network, "4") && config.GetBindToDevice4() != 0 {
		if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF,
			int(config.GetBindToDevice4())); err != nil {
			return errors.New("failed to set IP_BOUND_IF", err)
		}
	}
	if strings.Contains(network, "6") && config.GetBindToDevice6() != 0 {
		if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_BOUND_IF,
			int(config.GetBindToDevice6())); err != nil {
			return errors.New("failed to set IPV6_BOUND_IF", err)
		}
	}

	if config.TxBufSize != 0 {
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_SNDBUF, int(config.TxBufSize)); err != nil {
			return errors.New("failed to set SO_SNDBUF/SO_SNDBUFFORCE", err)
		}
	}

	if config.RxBufSize != 0 {
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_RCVBUF, int(config.RxBufSize)); err != nil {
			return errors.New("failed to set SO_RCVBUF/SO_RCVBUFFORCE", err)
		}
	}
	return nil
}

func bindAddr(fd uintptr, address []byte, port uint32) error {
	return nil
}

func setReuseAddr(fd uintptr) error {
	return nil
}

func setReusePort(fd uintptr) error {
	return nil
}

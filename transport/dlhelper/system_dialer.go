package dlhelper

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"time"

	net1 "github.com/5vnetwork/vx-core/common/net"

	"github.com/rs/zerolog/log"
)

type DefaultSystemDialer struct {
	controllers []controller
}

// src is outbound.gateway
func (d *DefaultSystemDialer) DialConn(ctx context.Context, raddr net1.Destination, sockopt *SocketSetting) (net.Conn, error) {
	// determine laddr
	var laddr net1.Address
	if sockopt != nil && (sockopt.LocalAddr4 != "" || sockopt.LocalAddr6 != "") {
		// confirm which local address to use: 4 or 6
		if raddr.Address.Family().IsDomain() {
			ips, err := net.LookupIP(raddr.Address.Domain())
			if err != nil || len(ips) == 0 {
				return nil, fmt.Errorf("failed to get IP address for domain %s, %w", raddr.Address.Domain(), err)
			}
			if sockopt.LocalAddr4 != "" {
				for _, ip := range ips {
					// ipv4
					if ip.To4() != nil {
						laddr = net1.ParseAddress(sockopt.LocalAddr4)
						break
					}
				}
			}
			if sockopt.LocalAddr6 != "" {
				for _, ip := range ips {
					// ipv6
					if ip.To4() == nil {
						laddr = net1.ParseAddress(sockopt.LocalAddr6)
						break
					}
				}
			}
		} else {
			if raddr.Address.Family().IsIPv4() && sockopt.LocalAddr4 != "" {
				laddr = net1.ParseAddress(sockopt.LocalAddr4)
			} else if raddr.Address.Family().IsIPv6() && sockopt.LocalAddr6 != "" {
				laddr = net1.ParseAddress(sockopt.LocalAddr6)
			}
		}
	}

	goStdKeepAlive := time.Duration(0)
	if sockopt != nil && (sockopt.TcpKeepAliveInterval != 0 || sockopt.TcpKeepAliveIdle != 0) {
		goStdKeepAlive = time.Duration(-1)
	}
	dialer := &net.Dialer{
		Timeout:   time.Second * 16,
		LocalAddr: resolveSrcAddr(raddr.Network, laddr),
		KeepAlive: goStdKeepAlive,
	}
	if sockopt != nil && sockopt.Resolver != nil {
		dialer.Resolver = sockopt.Resolver
	}

	if sockopt != nil || len(d.controllers) > 0 {
		dialer.Control = func(network, address string, c syscall.RawConn) error {
			var controlErr error
			err := c.Control(func(fd uintptr) {
				if sockopt != nil {
					if err := applyOutboundSocketOptions(ctx, network, address, fd, sockopt); err != nil {
						controlErr = fmt.Errorf("failed to apply socket options: %w", err)
						log.Ctx(ctx).Error().Err(err).Msg("failed to apply socket options")
						return
					}
					if sockopt != nil && raddr.Network == net1.Network_UDP && hasBindAddr(sockopt) {
						if err := bindAddr(fd, sockopt.BindAddress, sockopt.BindPort); err != nil {
							controlErr = fmt.Errorf("failed to bind source address: %w", err)
							log.Ctx(ctx).Error().Err(err).Msg("failed to bind source address")
							return
						}
					}
				}

				for _, ctl := range d.controllers {
					if err := ctl(network, address, fd); err != nil {
						controlErr = fmt.Errorf("failed to apply external controller: %w", err)
						log.Ctx(ctx).Error().Err(err).Msg("failed to apply external controller")
						return
					}
				}
			})
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("failed to apply Control")
				return err
			}
			if controlErr != nil {
				log.Ctx(ctx).Error().Err(controlErr).Msg("controlErr")
				return controlErr
			}
			return nil
		}
	}

	conn, err := dialer.DialContext(ctx, raddr.Network.SystemString(), raddr.NetAddr())
	if err != nil {
		return nil, &DialError{
			error: err,
		}
	}
	if sockopt != nil && (sockopt.StatsReadCounter != nil || sockopt.StatsWriteCounter != nil) {
		conn = net1.NewStatsConn(conn, sockopt.StatsReadCounter, sockopt.StatsWriteCounter)
	}
	return conn, nil
}

type DialError struct {
	error
}

func hasBindAddr(sockopt *SocketSetting) bool {
	return sockopt != nil && len(sockopt.BindAddress) > 0 && sockopt.BindPort > 0
}

// from a nethelper.Address to a net.Addr with port set to 0
func resolveSrcAddr(network net1.Network, src net1.Address) net.Addr {
	if src == nil {
		return nil
	}

	if network == net1.Network_TCP {
		return &net.TCPAddr{
			IP:   src.IP(),
			Port: 0,
		}
	}

	return &net.UDPAddr{
		IP:   src.IP(),
		Port: 0,
	}
}

package buildclient

import (
	"fmt"
	"net/netip"

	"github.com/5vnetwork/vx-core/app/client"
	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/inbound/reject"
	"github.com/5vnetwork/vx-core/app/inbound/system"
	"github.com/5vnetwork/vx-core/app/userlogger"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/tun"
	"github.com/rs/zerolog/log"
)

func Tun(config *configs.TmConfig, fc *Builder, client *client.Client) error {
	if config.Tun != nil {
		rejector := &reject.TCPReject{
			InboundTag:  config.Tun.Tag,
			FakeDnsPool: client.AllFakeDns,
		}
		udpRejector := &reject.UdpReject{
			InboundTag:  config.Tun.Tag,
			FakeDnsPool: client.AllFakeDns,
		}
		fc.requireFeature(func(r i.Router, ul *userlogger.UserLogger) {
			rejector.Router = r
			udpRejector.Router = r
			rejector.UserLogger = ul
			udpRejector.UserLogger = ul
		})
		// tun inbound
		if config.Tun.GetMode() == configs.Mode_MODE_SYSTEM {
			return newTunSystemInbound(config.Tun, fc, rejector, udpRejector, client)
		} else if config.Tun.GetMode() == configs.Mode_MODE_GVISOR {
			return NewTunGvisorInbound(config.Tun, fc, rejector, udpRejector, client)
		} else {
			return fmt.Errorf("invalid tun mode: %d", config.Tun.GetMode())
		}
	}
	return nil
}

func newTunSystemInbound(
	config *configs.TunConfig, fc *Builder,
	rejector *reject.TCPReject, udpRejector *reject.UdpReject, client *client.Client) error {

	device, err := getTun(config, fc)
	if err != nil {
		return fmt.Errorf("failed to create tun device: %w", err)
	}
	return fc.requireFeature(func(h *dispatcher.Dispatcher, dnsConns *dns.Dns) error {
		tunInbound, err := NewTunSystemInbound(device, config.Tag, h, dnsConns,
			rejector, udpRejector)
		if err != nil {
			return fmt.Errorf("failed to create tun system inbound: %w", err)
		}
		client.Inbounds = append(client.Inbounds, tunInbound)
		return nil
	})
}

func NewTunSystemInbound(
	device tun.TunDeviceWithInfo, tag string, handler i.Handler, dnsConns *dns.Dns,
	rejector *reject.TCPReject, udpRejector *reject.UdpReject) (*system.TunSystemInbound, error) {
	dnsAddress := make([]net.Destination, 0)
	for _, dns := range device.DnsServers() {
		dnsAddress = append(dnsAddress, net.UDPDestination(net.IPAddress(dns.AsSlice()), 53))
	}

	opts := []system.Option{
		system.WithTag(tag),
		system.WithTun(device),
		system.With4(device.IP4().Next().AsSlice(), device.IP4().AsSlice(), 0),
		system.WithHandler(handler),
		system.WithDns(dnsConns, dnsAddress),
	}

	if device.IP6().IsValid() {
		log.Info().Msg("666666666666666666666")
		napIp6 := device.IP6().Next().AsSlice()
		rejector.NatIp6 = napIp6
		opts = append(opts, system.WithRejector(rejector))
		opts = append(opts, system.With6(napIp6, device.IP6().AsSlice(), 0))
		opts = append(opts, system.WithUdpRejector(udpRejector))
	}
	tunInbound := system.New(opts...)

	return tunInbound, nil
}

func TunConfigToTunOption(config *configs.TunDeviceConfig) (*tun.TunOption, error) {
	option := &tun.TunOption{
		Name:   config.Name,
		Mtu:    config.Mtu,
		Path:   config.Path,
		Offset: offset,
		FD:     int(config.Fd),
	}

	var err error
	if config.Cidr4 != "" {
		if option.Ip4, err = netip.ParsePrefix(config.Cidr4); err != nil {
			return nil, err
		}
		if len(config.Dns4) > 0 {
			option.Dns4 = make([]netip.Addr, len(config.Dns4))
			for i, dns := range config.Dns4 {
				if option.Dns4[i], err = netip.ParseAddr(dns); err != nil {
					return nil, err
				}
			}
		}
		if len(config.Routes4) > 0 {
			option.Route4 = make([]netip.Prefix, len(config.Routes4))
			for i, route := range config.Routes4 {
				if option.Route4[i], err = netip.ParsePrefix(route); err != nil {
					return nil, err
				}
			}
		}
	}
	if config.Cidr6 != "" {
		if option.Ip6, err = netip.ParsePrefix(config.Cidr6); err != nil {
			return nil, err
		}
		if len(config.Dns6) > 0 {
			option.Dns6 = make([]netip.Addr, len(config.Dns6))
			for i, dns := range config.Dns6 {
				if option.Dns6[i], err = netip.ParseAddr(dns); err != nil {
					return nil, err
				}
			}
		}
		if len(config.Routes6) > 0 {
			option.Route6 = make([]netip.Prefix, len(config.Routes6))
			for i, route := range config.Routes6 {
				if option.Route6[i], err = netip.ParsePrefix(route); err != nil {
					return nil, err
				}
			}
		}
	}

	return option, nil
}

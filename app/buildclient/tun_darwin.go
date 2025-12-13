package buildclient

import (
	"fmt"
	"os"
	"reflect"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/inbound/gvisor"
	"github.com/5vnetwork/vx-core/app/inbound/reject"
	"github.com/5vnetwork/vx-core/app/inbound/system"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/tun"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const offset int32 = 4

func NewTunGvisorInbound(config *configs.TunConfig, f *Builder,
	rejector *reject.TCPReject, udpRejector *reject.UdpReject, client *client.Client) error {
	var le stack.LinkEndpoint
	if config.Device.Fd == 0 {
		tun, err := getTun(config, f)
		if err != nil {
			return err
		}
		le = gvisor.NewTunLinkEndpoint(tun, 1500, gvisor.TunLinkEndpointWithRejector(rejector))
	} else {
		log.Info().Int("fd", int(config.Device.Fd)).Send()
		newFd, err := unix.Dup(int(config.Device.Fd))
		if err != nil {
			return err
		}
		err = unix.SetNonblock(newFd, true)
		if err != nil {
			unix.Close(newFd)
			return err
		}
		log.Info().Int("newFd", newFd).Send()
		rw := os.NewFile(uintptr(newFd), "/dev/tun")
		udpRw := gvisor.NewReadWriteCloserSplitUdp(rw, offset)
		le = gvisor.NewIOLinkEndpoint(udpRw,
			gvisor.IOLinkEndpointWithOffset(4),
			gvisor.IOLinkEndpointWithRejector(rejector),
			gvisor.IOLinkEndpointWithMtu(config.Device.Mtu),
		)

		opts := []system.Option{
			system.WithTag(config.Tag),
			system.WithTun(udpRw),
			system.WithUdpRejector(udpRejector),
		}
		tunInbound := system.New(opts...)
		dnsAddress := make([]net.Destination, 0)
		for _, dns := range config.Device.Dns4 {
			dnsAddress = append(dnsAddress, net.UDPDestination(net.ParseAddress(dns), 53))
		}
		for _, dns := range config.Device.Dns6 {
			dnsAddress = append(dnsAddress, net.UDPDestination(net.ParseAddress(dns), 53))
		}
		f.requireFeature(func(h *dispatcher.Dispatcher, dnsConns *dns.Dns) {
			system.WithHandler(h)(tunInbound)
			system.WithDns(dnsConns, dnsAddress)(tunInbound)
		})
		client.Inbounds = append(client.Inbounds, tunInbound)
	}
	o, err := gvisor.NewGvisorInbound(&gvisor.GvisorInboundOption{
		Tag:          config.Tag,
		LinkEndpoint: le,
	})
	if err != nil {
		return fmt.Errorf("failed to create tun gvisor inbound: %w", err)
	}
	f.requireFeature(func(h *dispatcher.Dispatcher) {
		o.WithHandler(h)
	})
	client.Inbounds = append(client.Inbounds, o)
	return nil
}

func NewInterfaceMonotor(name string, f *Builder) (i.DefaultInterfaceInfo, error) {
	return tun.NewInterfaceMonitor(name)
}

func getTun(config *configs.TunConfig, fc *Builder) (tun.TunDeviceWithInfo, error) {
	if device := fc.getFeature(reflect.TypeOf((*tun.TunDeviceWithInfo)(nil)).Elem()); device != nil {
		return device.(tun.TunDeviceWithInfo), nil
	}
	panic("not implemented")
	// option, err := TunConfigToTunOption(config.GetDevice())
	// if err != nil {
	// 	return nil, err
	// }

	// // dynamic ipv6 and current nic does not support ipv6
	// if config.Tun46Setting == configs.TunConfig_FOUR_ONLY ||
	// 	config.Tun46Setting == configs.TunConfig_DYNAMIC  {
	// 	// option.Ip6 = netip.Prefix{}
	// 	option.Route6 = []netip.Prefix{}
	// 	// option.Dns6 = []netip.Addr{}
	// }

	// return tun.NewTun(option)
}

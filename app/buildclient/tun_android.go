package buildclient

import (
	"errors"
	"fmt"
	"reflect"
	"syscall"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/inbound/gvisor"
	"github.com/5vnetwork/vx-core/app/inbound/reject"
	"github.com/5vnetwork/vx-core/app/inbound/system"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/tun"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"

	"github.com/rs/zerolog/log"
)

const offset int32 = 0

func prepareTunSetter(config *configs.TunConfig, fc *Builder) error {
	return nil
}

func getTun(config *configs.TunConfig, fc *Builder) (tun.TunDeviceWithInfo, error) {
	if device := fc.getFeature(reflect.TypeOf((*tun.TunDeviceWithInfo)(nil)).Elem()); device != nil {
		return device.(tun.TunDeviceWithInfo), nil
	}
	return nil, errors.New("tun device not found")
}

func NewInterfaceMonotor(name string, f *Builder) (i.DefaultInterfaceInfo, error) {
	panic("not implemented")
}

func NewTunGvisorInbound(config *configs.TunConfig, f *Builder,
	rejector *reject.TCPReject, udpReject *reject.UdpReject, client *client.Client) error {
	log.Info().Int("fd", int(config.Device.Fd)).Send()
	newFd, err := syscall.Dup(int(config.Device.Fd))
	if err != nil {
		return err
	}
	log.Info().Int("oldFd", int(config.Device.Fd)).Int("newFd", newFd).Send()
	ep, err := fdbased.New(&fdbased.Options{
		FDs: []int{newFd},
		MTU: 1500,
		//TODO offload
	})
	if err != nil {
		return fmt.Errorf("failed to create fdbased endpoint: %w", err)
	}
	ep = gvisor.NewFilterLinkEndpoint(ep, &reject.CombineRejector{
		TCPReject: rejector,
		UDPReject: udpReject,
	}, true)

	opts := []system.Option{
		system.WithTag(config.Tag),
		system.WithTun(gvisor.NewFilterLinkEndpointToRunnable(ep.(*gvisor.FilterLinkEndpoint), "")),
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
	common.Must(f.addComponent(tunInbound))

	f.requireFeature(func(h *dispatcher.Dispatcher) error {
		o, err := gvisor.NewGvisorInbound(&gvisor.GvisorInboundOption{
			Tag:          config.Tag,
			LinkEndpoint: ep,
			Handler:      h,
			OnClose: func() {
				unix.Close(newFd)
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create tun gvisor inbound: %w", err)
		}
		common.Must(f.addComponent(o))
		return nil
	})
	return nil
}

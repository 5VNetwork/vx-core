//go:build !android && !darwin && !windows

package buildclient

import (
	"fmt"
	"net/netip"
	"reflect"
	"strings"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/inbound/gvisor"
	"github.com/5vnetwork/vx-core/app/inbound/reject"
	"github.com/5vnetwork/vx-core/app/tunmanage"
	"github.com/5vnetwork/vx-core/app/tunset"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/tun"
	"github.com/rs/zerolog/log"
)

const offset int32 = 0

func prepareTunSetter(config *configs.TunConfig, fc *Builder) error {
	return nil
}

func NewTunGvisorInbound(config *configs.TunConfig, f *Builder,
	rejector *reject.TCPReject, udpReject *reject.UdpReject, client *client.Client) error {
	tun, err := getTun(config, f)
	if err != nil {
		return err
	}
	o, err := gvisor.NewGvisorInbound(&gvisor.GvisorInboundOption{
		Tag:          config.Tag,
		LinkEndpoint: gvisor.NewTunLinkEndpoint(tun, 1500),
	})
	if err != nil {
		return fmt.Errorf("failed to create tun gvisor inbound: %w", err)
	}
	f.requireFeature(func(h *dispatcher.Dispatcher) {
		o.WithHandler(h)
	})
	common.Must(f.addComponent(o))
	return nil
}

func NewInterfaceMonotor(name string, f *Builder) (i.DefaultInterfaceInfo, error) {
	return tun.NewInterfaceMonitor(name)
}

func getTun(config *configs.TunConfig, fc *Builder) (
	tun.TunDeviceWithInfo, error) {
	if device := fc.getFeature(reflect.
		TypeOf((*tun.TunDeviceWithInfo)(nil)).Elem()); device != nil {
		log.Print("using recommended tun!!!")
		return device.(tun.TunDeviceWithInfo), nil
	}

	option, err := TunConfigToTunOption(config.GetDevice())
	if err != nil {
		return nil, err
	}

	dynamic := config.Tun46Setting == configs.TunConfig_DYNAMIC

	defaultNicInfo := fc.getFeature(reflect.
		TypeOf((*i.DefaultInterfaceInfo)(nil)).Elem()).(i.DefaultInterfaceInfo)
	defaultNicHas6 := true
	if dynamic {
		yes, err := defaultNicInfo.HasGlobalIPv6()
		if err != nil {
			log.Error().Err(err).Msg("failed to check global ipv6")
		} else {
			defaultNicHas6 = yes
		}
	}

	fourOnly := config.Tun46Setting == configs.TunConfig_FOUR_ONLY ||
		(config.Tun46Setting == configs.TunConfig_DYNAMIC && !defaultNicHas6)
	if fourOnly {
		option.Route6 = []netip.Prefix{}
	}

	d, err := tun.NewTun(option.Name)
	if err != nil {
		if strings.Contains(err.Error(), "device or resource busy") {
			// delete the tun device and retry
			common.Must(tunmanage.DeleteTunDevice(option.Name))
			d, err = tun.NewTun(option.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to create tun: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create tun: %w", err)
		}
	}
	tunDeviceWithInfo := tun.NewTunDeviceWithInfo(d,
		option.Ip4.Addr(), option.Ip6.Addr(), append(option.Dns4, option.Dns6...))
	if err = fc.addFeature(tunDeviceWithInfo); err != nil {
		return nil, fmt.Errorf("failed to add tun device with info: %w", err)
	}

	tm := tunmanage.NewTunManager(option, !fourOnly)
	if err = fc.addComponent(tm); err != nil {
		return nil, fmt.Errorf("failed to add tun manager: %w", err)
	}

	if dynamic {
		ts := tunset.NewTun6FollowsDefaultNIC(defaultNicInfo,
			!fourOnly, tm)
		if err = fc.addComponent(ts); err != nil {
			return nil, fmt.Errorf("failed to add tun set: %w", err)
		}
	}

	return tunDeviceWithInfo, nil
}

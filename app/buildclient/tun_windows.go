package buildclient

import (
	"fmt"
	"net/netip"
	"reflect"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/inbound/gvisor"
	"github.com/5vnetwork/vx-core/app/inbound/reject"
	"github.com/5vnetwork/vx-core/app/tunset"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/tun"
	"github.com/rs/zerolog/log"
)

const offset int32 = 0

func NewInterfaceMonotor(name string, f *Builder) (i.DefaultInterfaceInfo, error) {
	return tun.NewInterfaceMonitor(name)
}

func NewTunGvisorInbound(config *configs.TunConfig, f *Builder,
	rejector *reject.TCPReject, udpRejector *reject.UdpReject, client *client.Client) error {
	tun, err := getTun(config, f)
	if err != nil {
		return err
	}
	o, err := gvisor.NewGvisorInbound(&gvisor.GvisorInboundOption{
		Tag:          config.Tag,
		LinkEndpoint: gvisor.NewTunLinkEndpoint(tun, 1500),
		TcpOnly:      true,
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

func getTun(config *configs.TunConfig, fc *Builder) (tun.TunDeviceWithInfo, error) {
	if device := fc.getFeature(reflect.TypeOf((*tun.TunDeviceWithInfo)(nil)).Elem()); device != nil {
		return device.(tun.TunDeviceWithInfo), nil
	}

	option, err := TunConfigToTunOption(config.GetDevice())
	if err != nil {
		return nil, err
	}

	dynamic := config.Tun46Setting == configs.TunConfig_DYNAMIC

	defaultNicInfo := fc.getFeature(reflect.TypeOf((*i.DefaultInterfaceInfo)(nil)).Elem()).(i.DefaultInterfaceInfo)
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

	t, err := tun.NewTun(option)
	if err != nil {
		return nil, fmt.Errorf("failed to create tun: %w", err)
	}
	if err = fc.addFeature(t); err != nil {
		return nil, fmt.Errorf("failed to add tun device: %w", err)
	}

	if dynamic {
		tm, err := tun.NewTunManager(option, t.(*tun.NativeTun))
		if err != nil {
			return nil, fmt.Errorf("failed to create tun manager: %w", err)
		}
		ts := tunset.NewTun6FollowsDefaultNIC(defaultNicInfo, !fourOnly, tm)
		if err = fc.addComponent(ts); err != nil {
			return nil, fmt.Errorf("failed to add tun set: %w", err)
		}
	}

	return t, nil
}

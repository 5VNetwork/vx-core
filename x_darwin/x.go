//go:build darwin

package x_darwin

import (
	"encoding/base64"
	"fmt"
	"net/netip"
	"runtime"
	"time"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/client"
	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/tunset"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/5vnetwork/vx-core/tun"
	"github.com/5vnetwork/vx-core/tun/netmon"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

func New(configBytes []byte, in Interface,
	tunSupport6 bool, geoConfig []byte) (*X, error) {
	var opts []buildclient.Option

	config := &configs.TmConfig{}
	err := proto.Unmarshal(configBytes, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config, %w", err)
	}

	// dmg version only
	if len(geoConfig) > 0 {
		var gc configs.GeoConfig
		err := proto.Unmarshal(geoConfig, &gc)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal geo config, %w", err)
		}
		config.Geo = &gc
	}

	if config.RedirectStdErr != "" {
		err := RedirectStderr(config.RedirectStdErr)
		if err != nil {
			return nil, fmt.Errorf("failed to redirect stderr: %w", err)
		}
	}

	tun.IosInterfaceGetter = func(s string) ([]netip.Addr, error) {
		servers := in.GetDnsServersForInterface(s)
		if servers == nil {
			return nil, nil
		}
		dnsServers := make([]netip.Addr, 0, servers.Len())
		for i := 0; i < servers.Len(); i++ {
			addr, err := netip.ParseAddr(servers.Get(i))
			if err != nil {
				log.Debug().Msgf("failed to parse dns server %s, %v", servers.Get(i), err)
				continue
			}
			dnsServers = append(dnsServers, addr)
		}
		log.Debug().Str("interface", s).Any("dns servers", dnsServers).Msg("dns servers for default nic")
		return dnsServers, nil
	}

	nicMon, err := tun.NewInterfaceMonitor(in.GetTunName())
	if err != nil {
		return nil, fmt.Errorf("failed to create interface monitor: %w", err)
	}
	opts = append(opts, buildclient.WithComponents(nicMon))

	// tun
	tunConfig := config.GetTun()
	if tunConfig != nil {
		// tun setter
		if tunConfig.Tun46Setting == configs.TunConfig_DYNAMIC {
			tunFollowDefaultNic := tunset.NewTun6FollowsDefaultNIC(nicMon, tunSupport6, in)
			opts = append(opts, buildclient.WithComponents(tunFollowDefaultNic))
		}

		tunConfig.Device.Name = in.GetTunName()
		tunDeviceConfig := tunConfig.GetDevice()
		if tunDeviceConfig != nil {
			fd, err := in.GetFd()
			if err != nil {
				return nil, fmt.Errorf("failed to get fd, %w", err)
			}
			tunDeviceConfig.Fd = uint32(fd)
		}

	}

	instance, err := buildclient.NewX(config, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance, %w", err)
	}

	config = nil
	runtime.GC()

	return &X{instance: instance}, nil
}

func GenerateTls() ([]byte, error) {
	crt, err := cert.Generate(nil, cert.NotBefore(time.Now().Add(-time.Hour*24*365)),
		cert.NotAfter(time.Now().Add(time.Hour*24*365)))
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %s", err)
	}
	// Convert certificate to PEM
	certPEM := "-----BEGIN CERTIFICATE-----\n" +
		base64.StdEncoding.EncodeToString(crt.Certificate) +
		"\n-----END CERTIFICATE-----\n"
	// Convert private key to PEM
	keyPEM := "-----BEGIN PRIVATE KEY-----\n" +
		base64.StdEncoding.EncodeToString(crt.PrivateKey) +
		"\n-----END PRIVATE KEY-----\n"
	bytes, err := proto.Marshal(&tls.Certificate{
		Certificate: []byte(certPEM),
		Key:         []byte(keyPEM),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal certificate: %s", err)
	}
	return bytes, nil
}

func HasNICHavingGlobalIPv6Address() (bool, error) {
	info, err := tun.GetPrimaryPhysicalInterface()
	if err != nil {
		return false, err
	}
	return util.NICHasGlobalIPv6Address(uint32(info.Index))
}

func UpdateDefaultRouteInterface(ifName string) {
	log.Info().Str("default route nic", ifName).Msg("UpdateDefaultRouteInterface")
	netmon.UpdateLastKnownDefaultRouteInterface(ifName)
}

type X struct {
	instance *client.Client
}

func (x *X) Start() error {
	return x.instance.Start()
}

func (x *X) Close() error {
	if x.instance == nil {
		return nil
	}
	err := x.instance.Close()
	x.instance = nil
	return err
}

type Interface interface {
	GetTunName() string
	GetFd() (int32, error)
	GetDnsServersForInterface(string) Strings
	SetTunSupport6(support6 bool) error
}

type Strings interface {
	Get(index int) string
	Len() int
}

//go:build android

package x_android

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"reflect"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/client"
	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/inbound/reject"
	"github.com/5vnetwork/vx-core/app/tunset"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/appid"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/5vnetwork/vx-core/tun"
	"github.com/rs/zerolog/log"

	"google.golang.org/protobuf/proto"
)

// var f int32

func New(configBytes []byte, aai AndroidApiInterface) (Tm, error) {
	var config configs.TmConfig
	// f = fd
	err := proto.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config, %w", err)
	}

	if config.RedirectStdErr != "" {
		err := RedirectStderr(config.RedirectStdErr)
		if err != nil {
			log.Error().Err(err).Msg("failed to redirect stderr")
		}
	}

	defaultNicHasGlobal6, _ := monitor.HasGlobalIPv6()
	enable6 := (config.Tun.Tun46Setting == configs.TunConfig_BOTH) ||
		(config.Tun.Tun46Setting == configs.TunConfig_DYNAMIC && defaultNicHasGlobal6)
	tunConfig := ToTunConfig(config.Tun.Device, enable6)
	fd, err := aai.GetTun(tunConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get tun: %w", err)
	}
	tunDeviceWithInfo, err := getTunDeviceWithInfo(fd, config.Tun, enable6)
	if err != nil {
		return nil, fmt.Errorf("failed to get tun device with info: %w", err)
	}

	tm := &tm{}

	opts := []buildclient.Option{
		buildclient.WithComponents(&monitor),
		buildclient.WithFeatures(tunDeviceWithInfo),
		buildclient.WithFeatures(transport.FdFunc(func(fd uintptr) error {
			return aai.ProtectFd(int32(fd))
		})),
	}

	if config.Tun.Tun46Setting == configs.TunConfig_DYNAMIC {
		tunSetter := &tunSetter{
			tunConfig: config.Tun,
			aa:        aai,
			tm:        tm,
		}
		ts := tunset.NewTun6FollowsDefaultNIC(&monitor, enable6, tunSetter)
		opts = append(opts, buildclient.WithComponents(ts))
	}

	appid.GetPackageName = aai.GetPackageName
	client, err := buildclient.NewX(&config, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	tm.Client = client

	log.Debug().Bool("enable6", enable6).Msg("tun enable6")

	return tm, nil
}

func GenerateTls() ([]byte, error) {
	crt, err := cert.Generate(nil)
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

type Tm interface {
	Start() error
	Close() error
}

type tm struct {
	*client.Client
}

func (t *tm) CloseInbound() error {
	c := t.Client
	if c == nil {
		return errors.New("client is nil")
	}

	// exsitingSystemInbound := c.Components.GetComponent(reflect.TypeOf(&system.TunSystemInbound{}))
	// if exsitingSystemInbound != nil {
	// 	exsitingSystemInbound.(*system.TunSystemInbound).Close()
	// 	c.Components.RemoveComponent(exsitingSystemInbound)
	// }

	// exsitingGvisorInbound := c.Components.GetComponent(reflect.TypeOf(&gvisor.TunGvisorInbound{}))
	// if exsitingGvisorInbound != nil {
	// 	exsitingGvisorInbound.(*gvisor.TunGvisorInbound).CloseBlocking()
	// 	c.Components.RemoveComponent(exsitingGvisorInbound)
	// }

	for _, inbound := range c.Inbounds {
		common.Must(common.Close(inbound))
	}

	log.Debug().Msg("close inbound success")
	return nil
}

func getTunDeviceWithInfo(fd int32, config *configs.TunConfig, enable6 bool) (tun.TunDeviceWithInfo, error) {
	tunDevice, err := tun.NewTun(int(fd), int(config.GetDevice().GetMtu()))
	if err != nil {
		return nil, fmt.Errorf("failed to create tun: %w", err)
	}
	to, err := buildclient.TunConfigToTunOption(config.Device)
	if err != nil {
		return nil, fmt.Errorf("failed to create tun option: %w", err)
	}
	if !enable6 {
		to.Ip6 = netip.Prefix{}
		to.Route6 = []netip.Prefix{}
		to.Dns6 = []netip.Addr{}
	}
	tunDeviceWithInfo := tun.NewTunDeviceWithInfo(tunDevice,
		to.Ip4.Addr(), to.Ip6.Addr(), append(to.Dns4, to.Dns6...))
	if err != nil {
		return nil, fmt.Errorf("failed to create tun device with info: %w", err)
	}
	return tunDeviceWithInfo, nil
}

func (t *tm) CreateInbound(tunConfig *configs.TunConfig, fd int32, support6 bool) error {
	c := t.Client
	if c == nil {
		return errors.New("client is nil")
	}

	rejector := &reject.TCPReject{
		InboundTag:  tunConfig.Tag,
		Router:      c.Router,
		FakeDnsPool: c.AllFakeDns,
		UserLogger:  c.UserLogger,
	}
	udpRejector := &reject.UdpReject{
		InboundTag:  tunConfig.Tag,
		Router:      c.Router,
		FakeDnsPool: c.AllFakeDns,
		UserLogger:  c.UserLogger,
	}

	tunDeviceWithInfo, err := getTunDeviceWithInfo(fd, tunConfig, support6)
	if err != nil {
		return fmt.Errorf("failed to get tun device with info: %w", err)
	}
	dnsConn := c.Components.GetComponent(reflect.TypeOf(&dns.Dns{})).(*dns.Dns)

	tunInbound, err := buildclient.NewTunSystemInbound(
		tunDeviceWithInfo, tunConfig.Tag,
		c.Dispatcher, dnsConn, rejector, udpRejector)
	if err != nil {
		return fmt.Errorf("failed to create tun system inbound: %w", err)
	}
	if err := tunInbound.Start(); err != nil {
		return fmt.Errorf("failed to start tun system inbound: %w", err)
	}
	c.Inbounds = append(c.Inbounds, tunInbound)
	log.Debug().Msg("reset tun success")
	return nil
}

func (t *tm) Close() error {
	if t.Client == nil {
		return nil
	}

	err := t.Client.Close()
	t.Client = nil
	return err
}

type AndroidApiInterface interface {
	GetTun(tunConfig TunConfig) (int32, error)
	// close current tun and create a new one
	ResetTun(tunConfig TunConfig) (int32, error)
	GetPackageName(protocol int, source string, sourcePort int,
		destination string, destinationPort int) (string, error)
	Restart()

	ProtectFd(fd int32) error
}

type tunSetter struct {
	tunConfig *configs.TunConfig
	aa        AndroidApiInterface
	tm        *tm
}

func (t *tunSetter) SetTunSupport6(support6 bool) error {
	err := t.tm.CloseInbound()
	if err != nil {
		defer t.aa.Restart()
		log.Error().Err(err).Msg("failed to close inbound")
		return nil
	}

	tunConfig := ToTunConfig(t.tunConfig.Device, support6)
	fd, err := t.aa.ResetTun(tunConfig)
	if err != nil {
		defer t.aa.Restart()
		log.Error().Err(err).Msg("failed to reset tun")
		return nil
	}

	err = t.tm.CreateInbound(t.tunConfig, fd, support6)
	if err != nil {
		defer t.aa.Restart()
		log.Error().Err(err).Msg("failed to create inbound")
		return nil
	}

	return nil
}

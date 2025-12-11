package tunset

import (
	"sync"

	i "github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

// make sure tun ipv6 follows default nic ipv6, i.e.,
// when default nic has global ipv6, tun should also support ipv6(a
// default route, have ipv6 address and dns server)
type Tun6FollowsDefaultNIC struct {
	sync.Mutex
	DefaultNICMon i.DefaultInterfaceInfo
	// whether current tun support ipv6
	TunSupport6 bool
	TunSetter   TunSetter

	closed bool
}

type TunSetter interface {
	SetTunSupport6(support6 bool) error
}

func NewTun6FollowsDefaultNIC(defaultNICMon i.DefaultInterfaceInfo, tunSupport6 bool, tunSetter TunSetter) *Tun6FollowsDefaultNIC {
	return &Tun6FollowsDefaultNIC{DefaultNICMon: defaultNICMon, TunSupport6: tunSupport6, TunSetter: tunSetter}
}

func (t *Tun6FollowsDefaultNIC) Start() error {
	t.DefaultNICMon.Register(t)
	if has, err := t.DefaultNICMon.HasGlobalIPv6(); err == nil && has != t.TunSupport6 {
		log.Warn().Bool("tun_support6", t.TunSupport6).Bool("defaultNic6", has).Msg("Tun does not follow default nic ipv6")
		defer t.OnDefaultInterfaceChanged()
	}
	return nil
}

func (t *Tun6FollowsDefaultNIC) Close() error {
	t.closed = true
	t.DefaultNICMon.Unregister(t)
	return nil
}

func (t *Tun6FollowsDefaultNIC) OnDefaultInterfaceChanged() {
	if t.closed {
		return
	}

	t.Lock()
	defer t.Unlock()

	if t.closed {
		return
	}

	defaultNICHasGlobalIPv6, err := t.DefaultNICMon.HasGlobalIPv6()
	if err != nil {
		log.Err(err).Msg("HasGlobalIPv6")
		// in this case, make tun support 6
		if !t.TunSupport6 {
			log.Info().Msg("make tun support 6")
			err := t.TunSetter.SetTunSupport6(defaultNICHasGlobalIPv6)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to set tun support6")
				return
			}
			t.TunSupport6 = true
		}
		return
	}

	if defaultNICHasGlobalIPv6 != t.TunSupport6 {
		log.Debug().Bool("defaultNICHasGlobalIPv6", defaultNICHasGlobalIPv6).
			Bool("tunSupport6", t.TunSupport6).Msg("SetTunSupport6")
		err := t.TunSetter.SetTunSupport6(defaultNICHasGlobalIPv6)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to set tun support6")
			return
		}
		t.TunSupport6 = defaultNICHasGlobalIPv6
	}
}

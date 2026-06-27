package hysteria2

import (
	"context"
	"net"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/apernet/hysteria/extras/v2/realm"
	"github.com/rs/zerolog/log"
)

func newRealmPortMapper(ctx context.Context, realmID string, localPort int, config RealmPortMappingConfig) *realm.PortMapper {
	if localPort <= 0 {
		return nil
	}
	log.Debug().Str("realm", realmID).Int("port", localPort).Msg("realm port mapping started")
	start := time.Now()
	mapper, err := realm.NewPortMapper(ctx, localPort, realm.PortMapConfig{
		Timeout:  config.Timeout,
		Lifetime: config.Lifetime,
	})
	if err != nil {
		log.Warn().Err(err).Str("realm", realmID).Int("port", localPort).
			Msg("realm port mapping failed; continuing without it")
		return nil
	}
	log.Debug().Str("realm", realmID).Str("gateway", mapper.GatewayType()).
		Int("port", localPort).Str("external", mapper.ExternalAddr().String()).
		Str("duration", formatLogDuration(time.Since(start))).Msg("realm port mapping added")
	return mapper
}

func realmPortMapLoop(ctx context.Context, realmID string, mapper *realm.PortMapper) {
	defer func() {
		if err := mapper.Close(); err != nil {
			log.Debug().Err(err).Str("realm", realmID).Msg("realm port mapping removal failed")
		} else {
			log.Debug().Str("realm", realmID).Msg("realm port mapping removed")
		}
	}()
	interval := mapper.Lifetime() / 2
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	failing := false
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			changed, err := mapper.Renew(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				if !failing {
					log.Warn().Err(err).Str("realm", realmID).Msg("realm port mapping renewal failed")
					failing = true
				}
				continue
			}
			if failing {
				log.Info().Str("realm", realmID).Str("external", mapper.ExternalAddr().String()).
					Msg("realm port mapping recovered")
				failing = false
			}
			log.Debug().Str("realm", realmID).Str("external", mapper.ExternalAddr().String()).
				Bool("changed", changed).Msg("realm port mapping renewed")
		}
	}
}

type cleanupPacketConn struct {
	net.PacketConn
	cleanup func()
}

func (c *cleanupPacketConn) Close() error {
	c.cleanup()
	return c.PacketConn.Close()
}

func withMappedAddr(addrs []netip.AddrPort, mapper *realm.PortMapper) []netip.AddrPort {
	if mapper == nil {
		return addrs
	}
	return mergeMappedAddr(addrs, mapper.ExternalAddr())
}

func mergeMappedAddr(addrs []netip.AddrPort, addr netip.AddrPort) []netip.AddrPort {
	if !addr.IsValid() {
		return addrs
	}
	out := append([]netip.AddrPort(nil), addrs...)
	i, found := slices.BinarySearchFunc(out, addr, func(a, b netip.AddrPort) int {
		return strings.Compare(a.String(), b.String())
	})
	if found {
		return out
	}
	return slices.Insert(out, i, addr)
}

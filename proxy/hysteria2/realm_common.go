package hysteria2

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/apernet/hysteria/extras/v2/realm"
)

// peerAddrMapKey is the canonical peer address key for srcAddrMap and punchedPeers.
// Unmap() collapses IPv4-mapped IPv6 (::ffff:x.x.x.x) to plain IPv4 so punch and QUIC
// paths match regardless of socket family (udp4, udp6, or dual-stack udp).
func peerAddrMapKey(ap netip.AddrPort) netip.AddrPort {
	return netip.AddrPortFrom(ap.Addr().Unmap(), ap.Port())
}

func peerAddrMapKeyFrom(addr net.Addr) netip.AddrPort {
	return peerAddrMapKey(addr.(*net.UDPAddr).AddrPort())
}

func realmIPMode(mode string) (realm.AddrFamily, string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "dual":
		return realm.AddrFamilyAny, "udp", nil
	case "v4":
		return realm.AddrFamilyIPv4, "udp4", nil
	case "v6":
		return realm.AddrFamilyIPv6, "udp6", nil
	default:
		return realm.AddrFamilyAny, "", fmt.Errorf("invalid ipMode %q (expected v4, v6, or dual)", mode)
	}
}

func realmHTTPClient(cfg RealmConfig) *http.Client {
	if !cfg.Insecure {
		return nil
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{Transport: tr}
}

func realmSTUNServers(addr *realm.Addr, realmCfg RealmConfig) []string {
	if stunServers := addr.Params["stun"]; len(stunServers) > 0 {
		return append([]string(nil), stunServers...)
	}
	if len(realmCfg.STUNServers) > 0 {
		return append([]string(nil), realmCfg.STUNServers...)
	}
	return append([]string(nil), defaultRealmSTUNServers...)
}

func sessionTTLDuration(ttl int) time.Duration {
	if ttl <= 0 {
		return time.Minute
	}
	return time.Duration(ttl) * time.Second
}

func isRealmSessionInvalid(err error) bool {
	var statusErr *realm.StatusError
	return errors.As(err, &statusErr) &&
		(statusErr.StatusCode == http.StatusUnauthorized || statusErr.StatusCode == http.StatusNotFound)
}

func isRealmRegisterFatal(err error) bool {
	var statusErr *realm.StatusError
	return errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusBadRequest
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func shortAttempt(nonce string) string {
	if len(nonce) <= 8 {
		return nonce
	}
	return nonce[:8]
}

func punchPacketTypeString(t realm.PunchPacketType) string {
	switch t {
	case realm.PunchPacketHello:
		return "hello"
	case realm.PunchPacketAck:
		return "ack"
	default:
		return "unknown"
	}
}

func formatLogDuration(d time.Duration) string {
	return d.Round(time.Millisecond).String()
}

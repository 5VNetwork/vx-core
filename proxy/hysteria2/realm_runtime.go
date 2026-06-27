package hysteria2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/i"
	"github.com/apernet/hysteria/extras/v2/realm"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

const realmConnectSTUNCacheTTL = 10 * time.Second

type realmServerRuntime struct {
	cancel      context.CancelFunc
	client      *realm.Client
	realmID     string
	punchConn   *realm.PunchPacketConn
	stunServers []string
	puncher     *realm.ServerPuncher
	config      RealmConfig
	family      realm.AddrFamily
	mapper      *realm.PortMapper

	mu      sync.Mutex
	session realmSession
	addrs   []netip.AddrPort
	addrsAt time.Time

	connectSF singleflight.Group

	peerMu       sync.Mutex
	punchedPeers map[netip.AddrPort]struct{}
}

type realmSession struct {
	id  string
	ttl int
}

var (
	errRealmSessionInvalid = errors.New("realm session invalid")
	errRealmSessionLost    = errors.New("realm session lost")
)

func startRealmServerRuntime(ctx context.Context, cancel context.CancelFunc, addr *realm.Addr,
	punchConn *realm.PunchPacketConn, family realm.AddrFamily, realmCfg RealmConfig) (*realmServerRuntime, error) {
	stunServers := realmSTUNServers(addr, realmCfg)
	rClient, err := realm.NewClientFromAddr(addr, realmHTTPClient(realmCfg))
	if err != nil {
		return nil, fmt.Errorf("failed to create realm client: %w", err)
	}
	puncher, err := realm.NewServerPuncher(ctx, punchConn)
	if err != nil {
		return nil, fmt.Errorf("failed to create realm server puncher: %w", err)
	}
	rt := &realmServerRuntime{
		cancel:       cancel,
		client:       rClient,
		realmID:      addr.RealmID,
		punchConn:    punchConn,
		stunServers:  stunServers,
		puncher:      puncher,
		config:       realmCfg,
		family:       family,
		punchedPeers: make(map[netip.AddrPort]struct{}),
	}
	if realmCfg.PortMapping.Enabled {
		localPort := 0
		if udpAddr, ok := punchConn.LocalAddr().(*net.UDPAddr); ok {
			localPort = udpAddr.Port
		}
		rt.mapper = newRealmPortMapper(ctx, addr.RealmID, localPort, realmCfg.PortMapping)
	}
	cleanupMapper := func() {
		if rt.mapper != nil {
			_ = rt.mapper.Close()
		}
	}
	if _, _, err := rt.refreshAddrsDirect(ctx); err != nil {
		cleanupMapper()
		return nil, fmt.Errorf("failed to refresh addrs: %w", err)
	}
	initialSession, err := rt.register(ctx)
	if err != nil {
		cleanupMapper()
		return nil, fmt.Errorf("failed to register: %w", err)
	}
	rt.setSession(initialSession)
	if rt.mapper != nil {
		go realmPortMapLoop(ctx, addr.RealmID, rt.mapper)
	}
	go rt.run(ctx, initialSession)
	return rt, nil
}

func (r *realmServerRuntime) run(ctx context.Context, sess realmSession) {
	for ctx.Err() == nil {
		if err := r.runSession(ctx, sess); err != nil && ctx.Err() == nil {
			log.Warn().Err(err).Str("realm", r.realmID).Msg("realm session lost")
		}
		sess = r.registerWithBackoff(ctx)
		if sess.id == "" {
			return
		}
	}
}

func (r *realmServerRuntime) registerWithBackoff(ctx context.Context) realmSession {
	backoff := time.Second
	for ctx.Err() == nil {
		if _, _, err := r.refreshAddrs(ctx); err != nil {
			log.Warn().Err(err).Str("realm", r.realmID).Msg("realm STUN refresh before re-register failed")
		}
		sess, err := r.register(ctx)
		if err == nil {
			r.setSession(sess)
			return sess
		}
		if isRealmRegisterFatal(err) {
			log.Error().Err(err).Str("realm", r.realmID).Msg("realm re-register rejected; giving up")
			return realmSession{}
		}
		log.Warn().Err(err).Str("realm", r.realmID).Msg("realm re-register failed")
		log.Debug().Str("realm", r.realmID).Str("backoff", formatLogDuration(backoff)).Msg("realm re-register scheduled")
		if !sleepContext(ctx, backoff) {
			return realmSession{}
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
	return realmSession{}
}

func (r *realmServerRuntime) register(ctx context.Context) (realmSession, error) {
	localAddrs := r.currentAddrs()
	log.Debug().Str("realm", r.realmID).Strs("addresses", addrPortStrings(localAddrs)).Msg("realm registration started")
	start := time.Now()
	registerResp, err := r.client.Register(ctx, r.realmID, addrPortStrings(localAddrs))
	if err != nil {
		return realmSession{}, err
	}
	sess := realmSession{id: registerResp.SessionID, ttl: registerResp.TTL}
	log.Debug().Str("realm", r.realmID).Int("ttl", sess.ttl).Str("duration", formatLogDuration(time.Since(start))).Msg("realm registration completed")
	log.Info().Str("realm", r.realmID).Strs("addresses", addrPortStrings(localAddrs)).Int("ttl", sess.ttl).Msg("realm registered")
	return sess, nil
}

func (r *realmServerRuntime) runSession(ctx context.Context, sess realmSession) error {
	sessionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error, 2)
	go func() { errCh <- r.heartbeatLoop(sessionCtx, sess) }()
	go func() { errCh <- r.eventsLoop(sessionCtx, sess) }()
	err := <-errCh
	cancel()
	return err
}

// keeps the realm server's rendezvous session alive and optionally updates
// its published public addresses on the rendezvous server.
func (r *realmServerRuntime) heartbeatLoop(ctx context.Context, sess realmSession) error {
	interval := r.config.HeartbeatInterval
	if interval == 0 {
		interval = sessionTTLDuration(sess.ttl) / 2
		if interval <= 0 {
			interval = 15 * time.Second
		}
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	lastOK := time.Now()
	lastPublished := r.currentAddrs()
	for {
		select {
		case <-ctx.Done():
			log.Debug().Str("realm", r.realmID).Msg("realm heartbeat loop stopped")
			return ctx.Err()
		case <-t.C:
			log.Debug().Str("realm", r.realmID).Msg("realm heartbeat started")
			start := time.Now()
			req := realm.HeartbeatRequest{}
			if current := r.currentAddrs(); !slices.Equal(current, lastPublished) {
				req.Addresses = addrPortStrings(current)
				lastPublished = current
				log.Debug().Str("realm", r.realmID).Strs("addresses", req.Addresses).Msg("realm addresses changed")
			}
			resp, err := r.client.Heartbeat(ctx, r.realmID, sess.id, req)
			if err != nil {
				if isRealmSessionInvalid(err) {
					return errRealmSessionInvalid
				}
				log.Warn().Err(err).Str("realm", r.realmID).Msg("realm heartbeat failed")
				if time.Since(lastOK) > sessionTTLDuration(sess.ttl) {
					return errRealmSessionLost
				}
				continue
			}
			lastOK = time.Now()
			log.Debug().Str("realm", r.realmID).Int("ttl", resp.TTL).Bool("addressesUpdated", len(req.Addresses) > 0).
				Str("duration", formatLogDuration(time.Since(start))).Msg("realm heartbeat completed")
			if r.config.HeartbeatInterval == 0 && resp.TTL > 0 {
				next := time.Duration(resp.TTL) * time.Second / 2
				if next > 0 && next != interval {
					interval = next
					t.Reset(interval)
				}
			}
		}
	}
}

// processes punch events from the rendezvous server.
func (r *realmServerRuntime) eventsLoop(ctx context.Context, sess realmSession) error {
	backoff := time.Second
	lastOK := time.Now()
	for {
		if ctx.Err() != nil {
			log.Debug().Str("realm", r.realmID).Msg("realm events loop stopped")
			return ctx.Err()
		}
		log.Debug().Str("realm", r.realmID).Msg("realm events stream connecting")
		stream, err := r.client.Events(ctx, r.realmID, sess.id)
		if err != nil {
			if isRealmSessionInvalid(err) {
				return errRealmSessionInvalid
			}
			log.Warn().Err(err).Str("realm", r.realmID).Msg("realm events stream failed")
			if time.Since(lastOK) > sessionTTLDuration(sess.ttl) {
				return errRealmSessionLost
			}
			log.Debug().Str("realm", r.realmID).Str("backoff", formatLogDuration(backoff)).Msg("realm events stream reconnect scheduled")
			if !sleepContext(ctx, backoff) {
				return ctx.Err()
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}
		lastOK = time.Now()
		log.Debug().Str("realm", r.realmID).Msg("realm events stream connected")
		backoff = time.Second
		for {
			ev, err := stream.Next()
			if err != nil {
				_ = stream.Close()
				if ctx.Err() == nil {
					log.Warn().Err(err).Str("realm", r.realmID).Msg("realm events stream dropped")
				}
				break
			}
			lastOK = time.Now()
			log.Debug().Str("realm", r.realmID).Str("attempt", shortAttempt(ev.Nonce)).Strs("addresses", ev.Addresses).Msg("realm punch event received")
			go r.respond(ctx, ev)
		}
	}
}

func (r *realmServerRuntime) connectAddrs(ctx context.Context) ([]netip.AddrPort, error) {
	if cached := r.cachedAddrs(); cached != nil {
		return cached, nil
	}
	v, err, _ := r.connectSF.Do("stun", func() (any, error) {
		if cached := r.cachedAddrs(); cached != nil {
			return cached, nil
		}
		addrs, _, err := r.refreshAddrs(ctx)
		if err != nil {
			return nil, err
		}
		return addrs, nil
	})
	if err != nil {
		if fallback := r.currentAddrs(); len(fallback) > 0 {
			return fallback, err
		}
		return nil, err
	}
	return v.([]netip.AddrPort), nil
}

func (r *realmServerRuntime) cachedAddrs() []netip.AddrPort {
	r.mu.Lock()
	if r.addrs == nil || time.Since(r.addrsAt) >= realmConnectSTUNCacheTTL {
		r.mu.Unlock()
		return nil
	}
	addrs := append([]netip.AddrPort(nil), r.addrs...)
	r.mu.Unlock()
	return withMappedAddr(addrs, r.mapper)
}

func (r *realmServerRuntime) respond(ctx context.Context, ev *realm.PunchEvent) {
	attempt := shortAttempt(ev.Nonce)
	peerAddrs, err := parseRealmAddrPorts(ev.Addresses)
	if err != nil {
		log.Warn().Err(err).Str("realm", r.realmID).Str("attempt", attempt).Msg("invalid realm punch addresses")
		return
	}

	freshAddrs, stunErr := r.connectAddrs(ctx)
	if stunErr != nil {
		log.Warn().Err(stunErr).Str("realm", r.realmID).Str("attempt", attempt).Msg("realm connect STUN failed; using last-known addresses")
	}

	if sess := r.currentSession(); sess.id != "" && len(freshAddrs) > 0 {
		postCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		err := r.client.ConnectResponse(postCtx, r.realmID, sess.id, ev.Nonce, addrPortStrings(freshAddrs))
		cancel()
		if err != nil {
			log.Warn().Err(err).Str("realm", r.realmID).Str("attempt", attempt).Msg("realm connect-response post failed")
		}
	}

	log.Debug().Str("realm", r.realmID).Str("attempt", attempt).Strs("candidates", ev.Addresses).Msg("realm punch response started")
	start := time.Now()
	result, err := r.puncher.Respond(ctx, ev.Nonce, freshAddrs, peerAddrs, ev.PunchMetadata, realm.PunchConfig{
		Timeout: r.config.PunchTimeout,
		Family:  r.family,
	})
	if err != nil {
		log.Warn().Err(err).Str("realm", r.realmID).Str("attempt", attempt).Msg("realm punch failed")
		return
	}
	r.notePunchedPeer(result.PeerAddr)
	log.Debug().Str("realm", r.realmID).Str("attempt", attempt).Str("peer", result.PeerAddr.String()).
		Str("packet", punchPacketTypeString(result.Packet.Type)).
		Str("duration", formatLogDuration(time.Since(start))).Msg("realm punch completed")
}

func (r *realmServerRuntime) Close() error {
	r.cancel()
	sess := r.currentSession()
	if sess.id == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	log.Debug().Str("realm", r.realmID).Msg("realm deregister started")
	if err := r.client.Deregister(ctx, r.realmID, sess.id); err != nil {
		return err
	}
	log.Info().Str("realm", r.realmID).Msg("realm deregistered")
	return nil
}

func (r *realmServerRuntime) setSession(sess realmSession) {
	r.mu.Lock()
	r.session = sess
	r.mu.Unlock()
}

func (r *realmServerRuntime) currentSession() realmSession {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.session
}

func (r *realmServerRuntime) refreshAddrs(ctx context.Context) ([]netip.AddrPort, bool, error) {
	return r.refreshAddrsWith(ctx, func(ctx context.Context, config realm.STUNConfig) ([]netip.AddrPort, error) {
		return realm.DiscoverWithDemux(ctx, r.punchConn, config)
	})
}

func (r *realmServerRuntime) refreshAddrsDirect(ctx context.Context) ([]netip.AddrPort, bool, error) {
	return r.refreshAddrsWith(ctx, func(ctx context.Context, config realm.STUNConfig) ([]netip.AddrPort, error) {
		return realm.Discover(ctx, r.punchConn.PacketConn, config)
	})
}

func (r *realmServerRuntime) refreshAddrsWith(ctx context.Context,
	discover func(context.Context, realm.STUNConfig) ([]netip.AddrPort, error)) (
	[]netip.AddrPort, bool, error) {
	log.Debug().Str("realm", r.realmID).Strs("stunServers", r.stunServers).Msg("realm server STUN discovery started")
	start := time.Now()
	addrs, err := discover(ctx, realm.STUNConfig{
		Servers:  r.stunServers,
		Timeout:  r.config.STUNTimeout,
		Family:   r.family,
		Resolver: &ipResolverAdapter{resolver: r.config.Resolver},
	})
	if err != nil {
		return nil, false, err
	}
	r.mu.Lock()
	changed := !slices.Equal(r.addrs, addrs)
	if changed {
		r.addrs = append([]netip.AddrPort(nil), addrs...)
	}
	r.addrsAt = time.Now()
	current := append([]netip.AddrPort(nil), r.addrs...)
	r.mu.Unlock()
	log.Debug().Str("realm", r.realmID).Strs("addresses", addrPortStrings(current)).Bool("changed", changed).
		Str("duration", formatLogDuration(time.Since(start))).Msg("realm server STUN discovery completed")
	return withMappedAddr(current, r.mapper), changed, nil
}

func (r *realmServerRuntime) currentAddrs() []netip.AddrPort {
	r.mu.Lock()
	addrs := append([]netip.AddrPort(nil), r.addrs...)
	r.mu.Unlock()
	return withMappedAddr(addrs, r.mapper)
}

func (r *realmServerRuntime) notePunchedPeer(addr netip.AddrPort) {
	r.peerMu.Lock()
	r.punchedPeers[peerAddrMapKey(addr)] = struct{}{}
	r.peerMu.Unlock()
}

// activePunchedPeers returns QUIC peers that completed a realm STUN punch.
// Both active and punchedPeers use peerAddrMapKey, so keys compare directly.
func (r *realmServerRuntime) activePunchedPeers(active map[netip.AddrPort]*srcAddrInfo) int {
	r.peerMu.Lock()
	defer r.peerMu.Unlock()
	n := 0
	for addr := range active {
		if _, ok := r.punchedPeers[addr]; ok {
			n++
		}
	}
	for addr := range r.punchedPeers {
		if _, ok := active[addr]; !ok {
			delete(r.punchedPeers, addr)
		}
	}
	return n
}

type ipResolverAdapter struct {
	resolver i.IPResolver
}

func (r *ipResolverAdapter) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	addrs, err := r.resolver.LookupIP(ctx, host)
	if err != nil {
		return nil, err
	}
	out := make([]net.IPAddr, 0, len(addrs))
	for _, addr := range addrs {
		out = append(out, net.IPAddr{IP: addr})
	}
	return out, nil
}

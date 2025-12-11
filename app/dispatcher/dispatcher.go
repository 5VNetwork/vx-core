package dispatcher

import (
	"context"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/inbound/monitor"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/app/sniff"
	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/app/userlogger"
	"github.com/5vnetwork/vx-core/common/appid"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/mux"
	"github.com/5vnetwork/vx-core/common/net"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/retry"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/common/strmatcher"
	"github.com/5vnetwork/vx-core/common/units"
	"github.com/5vnetwork/vx-core/common/uot"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	vless_out "github.com/5vnetwork/vx-core/proxy/vless/outbound"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO: improve error handling
// TODO: a udp session should only be closed when idle
type Dispatcher struct {
	// when direct failed, fallback to proxy. use proxy selector first, if not found, use any selector or node in outbound.manager
	FallbackToProxy bool
	// when fallback to domain for connections go proxy and use ip targets
	FallbackToDomain bool
	// Fallback to ipv4
	Ipv6FallbackToDomain bool
	Om                   i.OutboundManager
	Sm                   *selector.Selectors

	// rewrite destination
	Sniff               bool
	Sniffer             *sniff.Sniffer
	DestinationOverride []string
	FakeDns             i.FakeDnsPool
	Dns                 i.IPResolver

	TimeoutPolicy i.TimeoutSetting

	// stats
	StatsPolicy  i.StatsSetting
	Um           *user.Manager
	LinkStats    sync.Map              //key is prefix string, value is *LinkStats
	InboundStats *monitor.InboundStats //key is inbound tag, value is *InboundStats
	OutStats     *outbound.OutStats

	UserLogger *userlogger.UserLogger
	Router     i.Router //Router

	observerLock          sync.Mutex
	HandlerErrorObservers []i.HandlerErrorObserver
	Flows                 atomic.Int32
	PacketConns           atomic.Int32
}

func (p *Dispatcher) AddHandlerErrorObserver(observer i.HandlerErrorObserver) {
	p.observerLock.Lock()
	defer p.observerLock.Unlock()
	p.HandlerErrorObservers = append(p.HandlerErrorObservers, observer)
}

func (p *Dispatcher) RemoveHandlerErrorObserver(observer i.HandlerErrorObserver) {
	p.observerLock.Lock()
	defer p.observerLock.Unlock()
	for i, o := range p.HandlerErrorObservers {
		if o == observer {
			p.HandlerErrorObservers = append(p.HandlerErrorObservers[:i], p.HandlerErrorObservers[i+1:]...)
			break
		}
	}
}

type OnHandlerErrorFunc func(tag string, err error)

func New() *Dispatcher {
	d := &Dispatcher{
		Sniffer: sniff.NewSniffer(),
	}
	return d
}

func infoFromContext(ctx context.Context, dst net.Destination) *session.Info {
	info := session.Info{
		Target: dst,
	}
	info.InboundTag, _ = inbound.InboundTagFromContext(ctx)
	info.Source, _ = inbound.SrcFromContext(ctx)
	info.Gateway, _ = inbound.GatewayFromContext(ctx)
	info.User, _ = proxy.UserFromContext(ctx)
	id, _ := inbound.IDFromContext(ctx)
	info.ID = session.ID(id)
	return &info
}

func (p *Dispatcher) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	if dst.Address == mux.MuxCoolAddressDst {
		return mux.Serve(ctx, rw, p)
	}
	if dst.Address == uot.Addr {
		return uot.Serve(ctx, rw, p)
	}

	info := infoFromContext(ctx, dst)
	ctx = session.ContextWithInfo(ctx, info)

	p.Flows.Add(1)
	defer p.Flows.Add(-1)

	if !info.Target.IsValid() {
		return errors.New("dispatcher: Invalid destination")
	}

	p.populateAppId(ctx, info)

	rw = p.rewriteDest(ctx, info, rw).(buf.ReaderWriter)

	/* determine which outbound */
	rw0, handler, err := p.determineOutbound(ctx, info, rw)
	p.logRoute(ctx, info, handler, true)
	if err != nil {
		p.logUserError(ctx, info, err)
		return err
	}
	rw = rw0.(buf.ReaderWriter)

	ctx, idleChecker, rw0 := p.idle(ctx, info, rw)
	if idleChecker != nil {
		defer idleChecker.Cancel()
	}
	rw = rw0.(buf.ReaderWriter)
	rw = p.stats(ctx, info, rw, handler).(buf.ReaderWriter)

	if p.Ipv6FallbackToDomain && info.Target.Address.Family().IsIP() && info.Target.Address.Family().IsIPv6() {
		if handlerSupport6, ok := handler.(i.HandlerWith6Info); ok && !handlerSupport6.Support6() && info.SniffedDomain != "" {
			log.Ctx(ctx).Debug().Str("handler", handler.Tag()).Str("dst", info.Target.String()).Msg("ipv6 not supported, replace it with the domain")
			info.Target.Address = mynet.DomainAddress(info.SniffedDomain)
		}
	}

	// if dialer, ok := handler.(i.ProxyDialer); ok {
	// 	var conn i.FlowConn
	// 	if ddlRw, ok := rw.(buf.DdlReaderWriter); ok &&
	// 		ddlRw.SetReadDeadline(time.Now().Add(time.Millisecond*10)) == nil {
	// 		var mb buf.MultiBuffer
	// 		mb, err = ddlRw.ReadMultiBuffer()
	// 		ddlRw.SetReadDeadline(time.Time{})
	// 		if err == nil {
	// 			conn, err = dialer.ProxyDial(ctx, info.Target, mb)
	// 		}
	// 	} else {
	// 		conn, err = dialer.ProxyDial(ctx, info.Target, nil)
	// 	}
	// 	if conn != nil {
	// 		defer conn.Close()
	// 		err = Relay(ctx, info, rw, conn)
	// 	}
	// } else {
	err = handler.HandleFlow(ctx, info.Target, rw)
	// }
	if err != nil {
		p.onHandlerError(ctx, info, handler.Tag(), err)
		if p.canFallback(ctx, info, err) {
			err = p.fallback(ctx, info, rw, err, handler)
		}
	}
	if err != nil {
		p.logUserError(ctx, info, err)
	}

	log.Ctx(ctx).Debug().Uint64("up", info.SessionUpCounter.Load()).
		Uint64("down", info.SessionDownCounter.Load()).Msg("flow session end")

	return err
}

func (p *Dispatcher) HandlePacketConn(ctx context.Context, dst net.Destination, pc udp.PacketReaderWriter) error {
	info := infoFromContext(ctx, dst)
	ctx = session.ContextWithInfo(ctx, info)

	defer p.PacketConns.Add(1)
	defer p.PacketConns.Add(-1)

	p.populateAppId(ctx, info)

	pc = p.rewriteDest(ctx, info, pc).(udp.PacketReaderWriter)

	pc0, handler, err := p.determineOutbound(ctx, info, pc)
	if err != nil {
		p.logUserError(ctx, info, err)
		return err
	}
	pc = pc0.(udp.PacketReaderWriter)

	pc = p.stats(ctx, info, pc, handler).(udp.PacketReaderWriter)
	ctx, idleChecker, pc0 := p.idle(ctx, info, pc)
	if idleChecker != nil {
		defer idleChecker.Cancel()
	}
	pc = pc0.(udp.PacketReaderWriter)

	p.logRoute(ctx, info, handler, false)

	// if listener, ok := handler.(i.ProxyPacketListener); ok {
	// 	var udpConn udp.UdpConn
	// 	udpConn, err = listener.ListenPacket(ctx, info.Target)
	// 	if err != nil {
	// 		p.logError(ctx, info, err)
	// 		return err
	// 	}
	// 	defer udpConn.Close()
	// 	err = helper.RelayUDPPacketConn(ctx, pc, udpConn)
	// } else {
	err = handler.HandlePacketConn(ctx, info.Target, pc)
	// }
	if err != nil {
		p.onHandlerError(ctx, info, handler.Tag(), err)
		p.logUserError(ctx, info, err)
	}

	log.Ctx(ctx).Debug().Uint64("up", info.SessionUpCounter.Load()).
		Uint64("down", info.SessionDownCounter.Load()).Msg("packet conn session end")

	return err
}

func (p *Dispatcher) fallback(ctx context.Context, info *session.Info, rw buf.ReaderWriter, err error, handler i.Outbound) error {
	// routing matches default direct but ip is banned.
	// TODO: might be broken in future since net.errTimeout might be changed
	// TODO: as long as no application data has been read, fallback is okay
	// TODO: inform user about the fallback. log
	if handler.Tag() == "direct" && p.FallbackToProxy {
		log.Ctx(ctx).Warn().Str("dst", info.Target.String()).Str("domain", info.SniffedDomain).Msg("fallback to proxy")
		// since ip might be polluted, replace it with the domain
		if info.Target.Address.Family().IsIP() && info.SniffedDomain != "" {
			info.Target.Address = mynet.DomainAddress(info.SniffedDomain)
		}
		proxySelector := p.Sm.GetSelector("代理")
		var handler i.Outbound
		if proxySelector != nil {
			handler = proxySelector.GetHandler(info)
		} else {
			for _, selector := range p.Sm.GetAllSelectors() {
				handler = selector.GetHandler(info)
				if handler != nil {
					break
				}
			}
			for _, h := range p.Om.GetAllHandlers() {
				if h != nil && h.Tag() != "direct" && h.Tag() != "dns" {
					handler = h
					break
				}
			}
		}
		if handler != nil {
			if p.UserLogger != nil {
				p.UserLogger.LogFallback(info, handler.Tag())
			}
			err = handler.HandleFlow(ctx, info.Target, rw)
		}
	} else if p.FallbackToDomain && handler.Tag() != "direct" && info.Target.Address.Family().IsIP() &&
		(info.GetTargetDomain() != "") && strings.Contains(err.Error(), "i/o timeout") {
		// This might due to polluted ip
		log.Ctx(ctx).Warn().Str("dst", info.Target.String()).Str("domain", info.GetTargetDomain()).Msg("retry domain")
		info.Target.Address = mynet.DomainAddress(info.GetTargetDomain())
		err = handler.HandleFlow(ctx, info.Target, rw)
	}
	return err
}

func (p *Dispatcher) idle(ctx context.Context, info *session.Info, rw interface{}) (context.Context, *signal.ActivityChecker, interface{}) {
	// idle
	idleTimeout := p.getTimeout(info)
	if idleTimeout != 0 {
		var cancelCause context.CancelCauseFunc
		ctx, cancelCause = context.WithCancelCause(ctx)
		idleChecker := signal.NewActivityChecker(func() {
			cancelCause(errors.ErrIdle)
			log.Ctx(ctx).Debug().Msg("flow idle timeout")
		}, idleTimeout)
		if r, ok := rw.(i.DeadlineRW); ok {
			rw = &TimeoutDeadlineRW{
				timeout:    p.TimeoutPolicy,
				idle:       idleChecker,
				DeadlineRW: r,
				upOnly:     info.Target.Network == mynet.Network_UDP,
			}
		} else if r, ok := rw.(buf.ReaderWriter); ok {
			rw = &TimeoutReaderWriter{
				timeout:      p.TimeoutPolicy,
				idle:         idleChecker,
				ReaderWriter: r,
				upOnly:       info.Target.Network == mynet.Network_UDP,
			}
		} else if pc, ok := rw.(udp.PacketReaderWriter); ok {
			rw = &TimeoutPacketConn{
				idle:               idleChecker,
				PacketReaderWriter: pc,
			}
		}
		info.ActivityChecker = idleChecker
		return ctx, idleChecker, rw
	} else {
		log.Ctx(ctx).Debug().Msg("no idle timeout")
		return ctx, nil, rw
	}
}

func (p *Dispatcher) canFallback(ctx context.Context, info *session.Info, err error) bool {
	if !errors.Is(ctx.Err(), context.Canceled) && !errors.Is(err, context.Canceled) &&
		!errors.Is(err, errors.LeftToRightError{}) {
		return p.StatsPolicy.CalculateSessionStats() && info.SessionUpCounter.Load() == 0
	}
	return false
}

func (p *Dispatcher) logUserError(ctx context.Context, info *session.Info, err error) {
	if p.UserLogger != nil {
		if err == router.ErrBlocked {
			p.UserLogger.LogReject(info, err.Error())
		} else if err == router.ErrNoHandler {
			p.UserLogger.LogRoute(info, "")
		} else {
			if p.StatsPolicy.CalculateSessionStats() &&
				(info.SessionDownCounter.Load() == 0 || info.SessionUpCounter.Load() == 0) &&
				!errors.As(err, &errors.LeftToRightError{}) {
				p.UserLogger.LogSessionError(info, err)
			}
		}
	}
}

func (p *Dispatcher) logRoute(ctx context.Context, info *session.Info, handler i.Outbound, flow bool) {
	if p.UserLogger != nil {
		p.UserLogger.LogRoute(info, handler.Tag())
	}
	if flow {
		log.Ctx(ctx).Debug().Str("dst", info.Target.String()).Str("out_tag", handler.Tag()).
			Str("net", info.Target.Network.String()).Str("in_tag", info.InboundTag).
			Str("src", info.Source.String()).Str("sniffed_domain", info.SniffedDomain).
			Str("app", info.AppId).Str("protocol", info.Protocol).Msg("flow info")
	} else {
		log.Ctx(ctx).Debug().Str("udp src", info.Source.String()).Str("inbound", info.InboundTag).
			Str("sniff", info.SniffedDomain).Str("dst", info.Target.String()).Str("outbound", handler.Tag()).
			Str("app", info.AppId).Str("protocol", info.Protocol).Msg("packetconn info")
	}
}

func (p *Dispatcher) populateAppId(ctx context.Context, info *session.Info) {
	if info.AppId == "" {
		if (zerolog.GlobalLevel() == zerolog.DebugLevel) &&
			(!strings.Contains(info.InboundTag, "dns")) &&
			!strings.Contains(info.InboundTag, "DNS") {
			appId, err := appid.GetAppId(ctx, info.Source, &info.Target)
			if err != nil {
				log.Ctx(ctx).Debug().Err(err).Msg("failed to get appId")
			}
			info.AppId = appId
		}
	}
}

func (p *Dispatcher) getTimeout(info *session.Info) time.Duration {
	if info.Target.Port == 22 {
		return p.TimeoutPolicy.SshIdleTimeout()
	}
	if info.Target.Port == 53 {
		return p.TimeoutPolicy.DnsIdleTimeout()
	}
	if info.Target.Network == mynet.Network_TCP {
		return p.TimeoutPolicy.TcpIdleTimeout()
	}
	return p.TimeoutPolicy.UdpIdleTimeout()
}

func udpShouldReconnect(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, errors.ErrIdle) {
		return false
	}
	if errors.Is(err, vless_out.ErrRejectQuic) {
		return false
	}
	if errors.Is(err, retry.ErrRetryFailed) {
		return false
	}
	if errors.Is(err, outbound.ErrIpv6NotSupported) {
		return false
	}
	if errors.Is(err, errors.ErrClosed) {
		return false
	}
	if strings.Contains(err.Error(), "connection was refused") {
		return false
	}
	return true
}

func (p *Dispatcher) onHandlerError(ctx context.Context, info *session.Info, tag string, err error) {
	if tag == "dns" || tag == "direct" {
		return
	}
	if errors.Is(err, context.Canceled) {
		return
	}

	var closeError *websocket.CloseError
	if errors.As(err, &closeError) && closeError.Code == websocket.CloseNormalClosure {
		return
	}

	if errors.Is(err, io.EOF) {
		return
	}

	if info.SessionDownCounter.Load() != 0 {
		return
	}

	p.observerLock.Lock()
	defer p.observerLock.Unlock()
	if len(p.HandlerErrorObservers) == 0 {
		return
	}

	errStr := err.Error()
	// this error occurs if the src closes the connection, x continues to write response data, and src send rst.
	if strings.Contains(errStr, "endpoint is closed for send") {
		return
	}

	if strings.Contains(errStr, "An established connection was aborted by the software in your host machine.") {
		return
	}

	if strings.Contains(errStr, "write: broken pipe") {
		return
	}

	if strings.Contains(errStr, "connection reset by peer") {
		return
	}

	if strings.Contains(errStr, "reject quic over hysteria2") {
		return
	}

	if strings.Contains(errStr, "XTLS rejected QUIC traffic") {
		return
	}

	log.Ctx(ctx).Debug().Str("tag", tag).Err(err).Msg("handler error")

	for _, observer := range p.HandlerErrorObservers {
		go observer.OnHandlerError(tag, err)
	}
}

// TODO: Rewrite private ip to real one
func (p *Dispatcher) rewriteDest(ctx context.Context, si *session.Info, rw interface{}) interface{} {
	var fakeButNotFound bool
	// change fake ip to real domain
	fd := p.FakeDns
	if fd != nil && si.Target.IsValid() && si.Target.Address.Family().IsIP() &&
		fd.IsIPInIPPool(si.Target.Address) {
		if s := fd.GetDomainFromFakeDNS(si.Target.Address); s != "" {
			log.Ctx(ctx).Debug().Str("domain", s).Msg("fake ip found")
			domain, err := strmatcher.ToDomain(s)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Str("domain", s).Msg("failed to convert to domain")
				domain = s
			}
			si.FakeIP = si.Target.Address.IP()
			si.Target = mynet.Destination{
				Address: mynet.ParseAddress(domain),
				Port:    si.Target.Port,
				Network: si.Target.Network,
			}
		} else {
			log.Ctx(ctx).Debug().IPAddr("ip", si.Target.Address.IP()).Msg("fake ip but domain not found")
			fakeButNotFound = true
		}
	}

	shouldSniff := !si.Sniffed && ((si.Target.IsValid() && si.Target.Address.Family().IsIP() && len(p.DestinationOverride) > 0) ||
		fakeButNotFound || (p.Sniff && si.Target.IsValid() && si.Target.Address.Family().IsIP())) && si.Target.Port != 53
	// sniff
	if shouldSniff {
		rw0, err := p.Sniffer.Sniff(ctx, si, rw)
		if err == nil && si.Protocol != "" {
			log.Ctx(ctx).Debug().Str("protocol", si.Protocol).Str("domain", si.SniffedDomain).Msg("sniff result")
		} else {
			log.Ctx(ctx).Debug().Err(err).Msg("sniff failed")
		}
		rw = rw0
	}
	if shouldOverride(si, p.DestinationOverride, fakeButNotFound) {
		log.Ctx(ctx).Debug().Str("original dest", si.Target.String()).
			Str("sniff domain", si.SniffedDomain).Msg("replace destination")
		si.Target.Address = mynet.ParseAddress(si.SniffedDomain)
	}

	if pc, ok := rw.(udp.PacketReaderWriter); ok && fd != nil && p.Dns != nil {
		rw = &RealIpPacketConn{
			m:                  map[mynet.Address]mynet.Address{},
			PacketReaderWriter: pc,
			fakeDns:            fd,
			dns:                p.Dns,
			ctx:                ctx,
		}
	}
	return rw
}

func (p *Dispatcher) determineOutbound(ctx context.Context, si *session.Info, rw interface{}) (interface{}, i.Outbound, error) {
	rw0, handler, err := p.Router.PickHandlerWithData(ctx, si, rw)
	rw = rw0
	if err != nil {
		return rw, nil, err
	}
	if handler == nil {
		return rw, nil, errors.New("dispatcher: no handler found")
	}
	return rw, handler, nil
}

func (p *Dispatcher) stats(ctx context.Context, info *session.Info, rw interface{}, handler i.Outbound) interface{} {
	var ups session.UpCounters
	var downs session.DownCounters

	if p.StatsPolicy.CalculateOutboundLinkStats() || p.StatsPolicy.CalculateInboundLinkStats() {
		var throughputAdder linkStatsAdder
		if p.StatsPolicy.CalculateInboundLinkStats() &&
			info.Source.Address != nil && info.Source.Address.Family().IsIP() {
			network := mynet.PrefixStringFromIP(info.Source.Address.IP())
			stats, ok := p.LinkStats.Load(network)
			if !ok {
				stats = &LinkStats{}
				p.LinkStats.Store(network, stats)
			}
			throughputAdder = stats.(*LinkStats)
		} else if p.StatsPolicy.CalculateOutboundLinkStats() && p.OutStats != nil {
			stats := p.OutStats.Get(handler.Tag())
			throughputAdder = stats
		}
		if throughputAdder != nil {
			ls := &linkStats{
				ctx:     ctx,
				ohStats: throughputAdder,
			}
			ups = append(ups, ls)
			downs = append(downs, ls)
		}
	}

	// server
	if p.StatsPolicy.CalculateInboundStats() && p.InboundStats != nil {
		inboundStats := p.InboundStats.Get(info.InboundTag)
		ups = append(ups, atomicCounter{
			counter: &inboundStats.Traffic,
		})
		downs = append(downs, atomicCounter{
			counter: &inboundStats.Traffic,
		})
	}
	if p.StatsPolicy.CalculateUserStats() && p.Um != nil {
		us := p.Um.GetUser(info.User)
		if us != nil {
			ups = append(ups, atomicCounter{
				counter: us.Counter(),
			})
			downs = append(downs, atomicCounter{
				counter: us.Counter(),
			})
			us.AddPrefix(net.PrefixStringFromIP(info.Source.Address.IP()))
		} else {
			log.Warn().Str("uid", info.User).Msg("no user stats found")
		}
	}
	if p.StatsPolicy.CalculateSessionStats() {
		ups = append(ups, atomicCounter{
			counter: &info.SessionUpCounter,
		})
		downs = append(downs, atomicCounter{
			counter: &info.SessionDownCounter,
		})
	}
	if p.StatsPolicy.CalculateOutboundLinkStats() && p.OutStats != nil {
		stats := p.OutStats.Get(handler.Tag())
		ups = append(ups, atomicCounter{
			counter: &stats.UpCounter,
		})
		downs = append(downs, atomicCounter{
			counter: &stats.DownCounter,
		})
	}
	if len(ups) > 0 || len(downs) > 0 {
		if r, ok := rw.(i.DeadlineRW); ok {
			rw = &StatsDeadlineRW{
				DeadlineRW:  r,
				upCounter:   ups,
				downCounter: downs,
			}
		} else if r, ok := rw.(buf.ReaderWriter); ok {
			rw = &StatsReaderWriter{
				ReaderWriter: r,
				upCounter:    ups,
				downCounter:  downs,
			}
		} else {
			rw = &StatsPacketConn{
				PacketReaderWriter: rw.(udp.PacketReaderWriter),
				upCounter:          ups,
				downCounter:        downs,
			}
		}
	}
	return rw
}

func shouldOverride(info *session.Info, domainOverride []string, fakeIPNotFound bool) bool {
	if info.SniffedDomain == "" {
		return false
	}
	if fakeIPNotFound {
		return true
	}
	protocolString := info.Protocol
	if protocolString == "" {
		return false
	}
	for _, p := range domainOverride {
		if strings.HasPrefix(protocolString, p) || strings.HasSuffix(protocolString, p) {
			return true
		}
	}
	return false
}

type LinkStats struct {
	sync.Mutex
	Num       uint32
	BWTotal   uint32 //MBps
	PingTotal uint32 //ms
}

func (l *LinkStats) AddPing(pingMs uint64) {
	l.Lock()
	defer l.Unlock()
	l.PingTotal += uint32(pingMs)
}

func (l *LinkStats) AddThroughput(bytesPerSec uint64) {
	l.Lock()
	defer l.Unlock()
	l.Num++
	l.BWTotal += uint32(units.BytesToMB(bytesPerSec))
}

package dns

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand/v2"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// reuse conn for dns query
type DnsServerConcurrent struct {
	*dnsConnImpl

	tag     string
	rrCache *rrCache

	destsLock sync.RWMutex
	dests     []net.Destination
	// tcp
	t2s    []*t2
	useTls bool

	idLock sync.RWMutex
	nextId uint16

	// dispatcher                   atomic.Value
	// periodicallyUpdateDispatcher *task.Periodic
	handler    atomic.Value //i.FlowHandler
	dispatcher packetDispatcher

	ipToDomain *IPToDomain

	udpWaitingLock sync.RWMutex
	udpWaiting     map[uint16]*request

	tcpWaitingLock sync.RWMutex
	tcpWaiting     map[uint16]*request

	clientIp net.IP

	startOnce sync.Once
	closeOnce sync.Once
}

type packetDispatcher interface {
	DispatchPacket(destination net.Destination, payload *buf.Buffer) error
	SetResponseCallback(callback func(packet *udp.Packet))
}

type DnsServerConcurrentOption struct {
	Name            string
	NameserverAddrs []net.AddressPort
	Handler         i.FlowHandler
	IPToDomain      *IPToDomain
	Tls             bool
	ClientIp        net.IP
	Dispatcher      packetDispatcher
}

func NewDnsServerConcurrent(opts DnsServerConcurrentOption) *DnsServerConcurrent {
	cache := NewRrCache()
	ns := &DnsServerConcurrent{
		udpWaiting: make(map[uint16]*request),
		tcpWaiting: make(map[uint16]*request),
		tag:        opts.Name,
		rrCache:    cache,
		ipToDomain: opts.IPToDomain,
		useTls:     opts.Tls,
		clientIp:   opts.ClientIp,
		nextId:     1,
		dispatcher: opts.Dispatcher,
	}
	ns.dnsConnImpl = NewDnsConnImpl(opts.Name, cache,
		func(msg *dns.Msg) error {
			if len(ns.clientIp) > 0 {
				addClientIP(msg, ns.clientIp)
			}
			if ns.useTls {
				return ns.sendTcp(msg)
			} else {
				return ns.send(msg)
			}
		})
	ns.handler.Store(opts.Handler)
	ns.dispatcher.SetResponseCallback(ns.handleReply)
	ns.SetDests(opts.NameserverAddrs)

	return ns
}

func (d *DnsServerConcurrent) SetDests(dests []net.AddressPort) {
	d.destsLock.Lock()
	defer d.destsLock.Unlock()

	d.setDests(dests)
}

func (d *DnsServerConcurrent) setDests(dests []net.AddressPort) {
	var dsts []net.Destination
	var m []*t2

	for _, dest := range dests {
		dsts = append(dsts, net.UDPDestination(dest.Address, dest.Port))
		m = append(m, &t2{
			dst: net.TCPDestination(dest.Address, dest.Port),
			b:   make([]byte, 2048),
		})
	}

	d.dests = dsts
	d.t2s = m
}

func (d *DnsServerConcurrent) RemoveDest(remove net.AddressPort, fallback []net.AddressPort) {
	d.destsLock.Lock()
	defer d.destsLock.Unlock()

	newDests := make([]net.Destination, 0, len(d.dests))
	newT2s := make([]*t2, 0, len(d.dests))

	for i, dest := range d.dests {
		if dest.Address == remove.Address && dest.Port == remove.Port {
			continue
		}
		newDests = append(newDests, dest)
		newT2s = append(newT2s, d.t2s[i])
	}
	d.dests = newDests
	d.t2s = newT2s

	if len(newDests) == 0 {
		d.setDests(fallback)
	}
}

func (ns *DnsServerConcurrent) Start() error {
	ns.startOnce.Do(func() {
		ns.rrCache.Start()
		// ns.periodicallyUpdateDispatcher.Start()
		ns.dnsConnImpl.Start()
	})
	return nil
}

func (ns *DnsServerConcurrent) Close() error {
	ns.closeOnce.Do(func() {
		ns.dnsConnImpl.Close()
		ns.rrCache.Close()
		for _, t2 := range ns.t2s {
			t2.lock.Lock()
			if t2.conn != nil {
				t2.conn.Close()
			}
			t2.lock.Unlock()
		}
	})
	return nil
}

var defaultTimeout = time.Second * 3

type msgAndResolver struct {
	*dns.Msg
	src net.Destination
}

// msg should be standard dns query message: has only one question,
// opcode is QUERY
func (ns *DnsServerConcurrent) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	log.Ctx(ctx).Debug().Str("tag", ns.tag).Str("domain", msg.Question[0].Name).
		Uint16("id", msg.Id).Msg("dns handle query")

	question := msg.Question[0]

	cachedMsg, ok := ns.rrCache.Get(&question)
	if ok {
		log.Ctx(ctx).Debug().Str("question", question.String()).Msg("cache hit")
		return makeReply(msg, cachedMsg), nil
	}

	log.Ctx(ctx).Debug().Str("question", question.String()).Msg("cache miss")

	lock := &ns.udpWaitingLock
	waiting := ns.udpWaiting
	send := ns.send
	oldId := msg.Id
	if tcp || ns.useTls {
		lock = &ns.tcpWaitingLock
		waiting = ns.tcpWaiting
		send = ns.sendTcp

		// to prevent id conflict
		msg = msg.Copy() // copy the msg to avoid modifying the original msg
		ns.idLock.Lock()
		newId := ns.nextId
		ns.nextId++
		if ns.nextId == 65535 {
			ns.nextId = 1
		}
		ns.idLock.Unlock()
		msg.Id = newId
		log.Ctx(ctx).Debug().Uint16("new_id", newId).Msg("new id assigned")
	}

	ch := make(chan msgAndResolver, 1)

	lock.Lock()
	req, ok := waiting[msg.Id]
	if ok {
		req.channels = append(req.channels, ch)
	} else {
		req = &request{
			channels: []chan msgAndResolver{ch},
		}
		waiting[msg.Id] = req
	}
	lock.Unlock()

	defer func() {
		lock.Lock()
		req.channels = slices.DeleteFunc(req.channels, func(c chan msgAndResolver) bool {
			return c == ch
		})
		if len(req.channels) == 0 {
			delete(waiting, msg.Id)
		}
		lock.Unlock()
	}()

	err := send(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send dns query: %w", err)
	}

	startTime := time.Now()

	select {
	case <-ctx.Done():
		log.Ctx(ctx).Debug().
			Uint16("id", msg.Id).
			Str("domain", question.Name).
			Str("type", dns.TypeToString[question.Qtype]).
			Msg("dns ctx done")
		return nil, ctx.Err()
	case reply := <-ch:
		log.Ctx(ctx).Debug().
			Uint16("id", msg.Id).
			Str("domain", question.Name).
			Str("resolver", reply.src.String()).
			Dur("duration", time.Since(startTime)).
			Int("rcode", int(reply.Msg.Rcode)).
			Str("type", dns.TypeToString[question.Qtype]).
			Str("reply", reply.String()).Msg("dnsConcurrent reply")
		if reply.Rcode == dns.RcodeSuccess && !reply.Truncated {
			ns.rrCache.Set(reply.Msg)
		}
		if tcp || ns.useTls {
			reply.Msg.Id = oldId
		}
		return reply.Msg, nil
	}
}

// make a msg that is a reply to the request, and has the content of the cache
func makeReply(request *dns.Msg, cache *dns.Msg) *dns.Msg {
	ret := new(dns.Msg)
	ret = ret.SetReply(request)
	ret.Answer = cache.Answer
	ret.Ns = cache.Ns
	ret.Extra = cache.Extra
	ret.Rcode = cache.Rcode
	ret.RecursionAvailable = true
	// reply.Compress = true
	return ret
}

func (w *DnsServerConcurrent) send(msg *dns.Msg) error {
	w.destsLock.RLock()
	dests := w.dests
	w.destsLock.RUnlock()

	for _, dest := range dests {
		b := buf.New()
		by, err := msg.PackBuffer(b.BytesTo(b.Cap()))
		if err != nil {
			b.Release()
			return err
		}
		b.Resize(0, int32(len(by)))
		if err = w.dispatcher.DispatchPacket(dest, b); err != nil {
			b.Release()
			return err
		}
	}
	log.Debug().Uint16("id", msg.Id).Msg("dns msg sent")
	return nil
}

type t2 struct {
	dst  net.Destination
	lock sync.Mutex
	conn net.Conn
	b    []byte
}

func (t2 *t2) recreateConn(dest net.Destination, w *DnsServerConcurrent) error {
	ctx := log.Logger.With().
		Uint32("conn_id", rand.Uint32()).
		Str("name", w.tag).
		Logger().WithContext(context.Background())
	ctx = inbound.ContextWithInboundTag(ctx, w.tag)
	conn, err := (&util.FlowHandlerToDialer{
		FlowHandler: w.handler.Load().(i.Handler),
	}).Dial(ctx, dest)
	if err != nil {
		return err
	}
	log.Ctx(ctx).Debug().Msg("dnsConn created")
	if w.useTls {
		conn = tls.Client(conn, &tls.Config{
			ServerName: dest.Address.String(),
		})
	}
	t2.conn = conn
	go w.handleTcpReply(ctx, t2, t2.conn, dest)
	return nil
}

func (w *DnsServerConcurrent) sendTcp(msg *dns.Msg) error {
	w.destsLock.RLock()
	t2s := w.t2s
	w.destsLock.RUnlock()

	for _, t2 := range t2s {
		t2.lock.Lock()
		if t2.conn == nil {
			err := t2.recreateConn(t2.dst, w)
			if err != nil {
				log.Err(err).Msg("DnsServerConcurrent recreateConn")
				t2.lock.Unlock()
				continue
			}
		}
		b := t2.b
		// write length first
		binary.BigEndian.PutUint16(b[:2], uint16(msg.Len()))
		l, err := msg.PackBuffer(b[2:])
		if err != nil || len(l) != msg.Len() {
			log.Err(err).Msg("failed to pack")
			t2.lock.Unlock()
			continue
		}
		b = b[:len(l)+2]

		t2.conn.SetWriteDeadline(time.Now().Add(defaultTimeout))
		_, err = t2.conn.Write(b)
		if err != nil {
			log.Err(err).Msg("failed to write")
			t2.conn.Close()
			t2.conn = nil
		}
		t2.lock.Unlock()
	}

	log.Debug().Msg("dns msg sent by tcp")

	return nil
}

func (w *DnsServerConcurrent) handleTcpReply(ctx context.Context, t2 *t2,
	conn net.Conn, ns net.Destination) {
	defer func() {
		log.Ctx(ctx).Debug().Msg("dns conn closed")
		conn.Close()

		t2.lock.Lock()
		if t2.conn == conn {
			t2.conn = nil
		}
		t2.lock.Unlock()
	}()

	b := make([]byte, 2048)
	for {
		length, err := serial.ReadUint16(conn)
		if err != nil {
			log.Err(err).Msg("failed to read length")
			return
		}
		if int(length) > len(b) {
			b = make([]byte, length)
		}
		l, err := io.ReadFull(conn, b[:length])
		if err != nil {
			log.Err(err).Msg("failed to read full")
			return
		}
		msg := new(dns.Msg)
		if err := msg.Unpack(b[:l]); err != nil {
			log.Err(err).Msg("failed to unpack")
			return
		}
		if len(msg.Question) != 1 {
			log.Warn().Int("question_length", len(msg.Question)).Str("msg", msg.String()).Msg("question length is not 1")
		}
		w.tcpWaitingLock.Lock()
		pending, ok := w.tcpWaiting[msg.Id]
		if ok {
			for _, ch := range pending.channels {
				select {
				case ch <- msgAndResolver{
					Msg: msg,
					src: ns,
				}:
				default:
					log.Debug().Msg("channel is blocked")
				}
			}
			delete(w.tcpWaiting, msg.Id)
		}
		w.tcpWaitingLock.Unlock()
		if !ok {
			w.dnsConnImpl.handlerReply(msg)
		}
		if w.ipToDomain != nil {
			w.ipToDomain.SetDomain(msg, ns.String())
		}
	}
}

func (w *DnsServerConcurrent) handleReply(b *udp.Packet) {
	defer b.Release()
	msg := new(dns.Msg)
	if err := msg.Unpack(b.Payload.Bytes()); err != nil {
		log.Err(err).Msg("failed to unpack")
		return
	}

	if len(msg.Question) != 1 {
		log.Debug().Int("question_length", len(msg.Question)).Str("msg", msg.String()).Msg("question length is not 1")
	}

	w.udpWaitingLock.Lock()
	pending, ok := w.udpWaiting[msg.Id]
	if ok {
		for _, ch := range pending.channels {
			select {
			case ch <- msgAndResolver{
				Msg: msg,
				src: b.Source,
			}:
			default:
				log.Debug().Msg("channel is blocked")
			}
		}
		delete(w.udpWaiting, msg.Id)
	}
	w.udpWaitingLock.Unlock()
	if !ok {
		w.dnsConnImpl.handlerReply(msg)
	}
	if w.ipToDomain != nil {
		w.ipToDomain.SetDomain(msg, b.Source.String())
	}
}

type request struct {
	channels []chan msgAndResolver
}

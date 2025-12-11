package dns

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type DnsConn interface {
	common.Runnable
	// packet contains a dns message.
	// WritePacket takes ownership of the packet. Caller should not
	// use Packet after WritePacket.
	WritePacket(*udp.Packet) error
	// packet contains a dns message
	ReadPacket() (*udp.Packet, error)
}

type dnsInfo struct {
	id           uint16
	src          net.Destination
	dst          net.Destination
	addedAt      time.Time
	firstAddedAt time.Time
}

func msgIntoPacket(msg *dns.Msg, p *udp.Packet) error {
	p.Payload.Clear()
	bCap := p.Payload.Cap()
	by, err := msg.PackBuffer(p.Payload.BytesTo(bCap))
	if err != nil {
		return err
	}
	if len(by) > int(bCap) {
		log.Debug().Msg("PackBuffer allocate new buffer")
		p.Payload.Release()
		p.Payload = buf.FromBytes(by)
	} else {
		p.Payload.Extend(int32(len(by)))
	}
	return nil
}

type dnsConnImpl struct {
	tag            string
	rrCache        *rrCache
	dnsLock        sync.RWMutex
	dnsMsgIds      map[uint16]*dnsInfo
	requests       chan *udp.Packet
	responses      chan *udp.Packet
	dnsMsgIdsClean *task.Periodic
	done           *done.Instance
	send           func(*dns.Msg) error
}

func NewDnsConnImpl(tag string, rrCache *rrCache,
	send func(*dns.Msg) error) *dnsConnImpl {
	d := &dnsConnImpl{
		tag:       tag,
		rrCache:   rrCache,
		send:      send,
		dnsMsgIds: make(map[uint16]*dnsInfo),
		requests:  make(chan *udp.Packet, 100),
		responses: make(chan *udp.Packet, 100),
		done:      done.New(),
	}
	d.dnsMsgIdsClean = &task.Periodic{
		Interval: time.Second * 10,
		Execute:  d.dnsCleanup,
	}
	return d
}

func (t *dnsConnImpl) Start() error {
	t.dnsMsgIdsClean.Start()
	go t.handleRequestLoop()
	return nil
}

func (t *dnsConnImpl) Close() error {
	t.dnsMsgIdsClean.Close()
	t.done.Close()
	return nil
}

func (t *dnsConnImpl) dnsCleanup() error {
	t.dnsLock.Lock()
	for _, info := range t.dnsMsgIds {
		if info.addedAt.Before(time.Now().Add(-time.Second * 10)) {
			log.Debug().Uint16("id", info.id).
				Str("src", info.src.String()).
				Str("dst", info.dst.String()).Msg("dns cleanup")
			delete(t.dnsMsgIds, info.id)
		}
	}
	t.dnsLock.Unlock()
	return nil
}

func (ns *dnsConnImpl) handleRequestLoop() {
	for {
		select {
		case <-ns.done.Wait():
			return
		case p := <-ns.requests:
			ns.handleRequest(p)
		}
	}
}

func (ns *dnsConnImpl) handleRequest(p *udp.Packet) {
	msg := dns.Msg{}
	if err := msg.Unpack(p.Payload.Bytes()); err != nil {
		p.Release()
		log.Err(err).Msg("failed to unpack DNS message")
		return
	}

	if len(msg.Question) == 0 {
		p.Release()
		return
	}

	log.Debug().Str("question", msg.Question[0].String()).
		Str("type", dns.TypeToString[msg.Question[0].Qtype]).
		Str("name", ns.tag).
		Uint16("id", msg.Id).Msg("dns request")

	question := msg.Question[0]
	cachedMsg, ok := ns.rrCache.Get(&question)
	if ok {
		log.Debug().Str("question", question.String()).Msg("cache hit")
		rspMsg := makeReply(&msg, cachedMsg)
		err := msgIntoPacket(rspMsg, p)
		if err != nil {
			log.Err(err).Msg("failed to pack DNS message")
			p.Release()
			return
		}
		p.Source, p.Target = p.Target, p.Source
		ns.writeResponse(p)
		return
	}

	ns.dnsLock.Lock()
	existing, ok := ns.dnsMsgIds[msg.Id]
	if ok {
		existing.addedAt = time.Now()
	} else {
		ns.dnsMsgIds[msg.Id] = &dnsInfo{
			id:           msg.Id,
			src:          p.Source,
			dst:          p.Target,
			addedAt:      time.Now(),
			firstAddedAt: time.Now(),
		}
	}
	ns.dnsLock.Unlock()

	err := ns.send(&msg)
	if err != nil {
		log.Err(err).Msg("failed to send DNS message")
	}
	p.Release()
}

// msg should be standard dns query message: has only one question,
// opcode is QUERY, no other fancy stuff.
func (ns *dnsConnImpl) WritePacket(p *udp.Packet) error {
	select {
	case <-ns.done.Wait():
		p.Release()
		return errors.New("closed")
	case ns.requests <- p:
		return nil
	default:
		return errors.New("requests channel is blocked")
	}
}

func (ns *dnsConnImpl) writeResponse(p *udp.Packet) {
	if !ns.done.Done() {
		select {
		case ns.responses <- p:
			return
		default:
			log.Warn().Msg("responses channel is blocked")
		}
	}
	p.Release()
}

func (ns *dnsConnImpl) ReadPacket() (*udp.Packet, error) {
	select {
	case <-ns.done.Wait():
		return nil, io.EOF
	case p := <-ns.responses:
		return p, nil
	}
}

func (w *dnsConnImpl) handlerReply(msg *dns.Msg) {
	w.dnsLock.Lock()
	info, ok := w.dnsMsgIds[msg.Id]
	delete(w.dnsMsgIds, msg.Id)
	w.dnsLock.Unlock()
	if ok {
		log.Debug().Uint16("id", msg.Id).
			// Str("domain", msg.Question[0].Name).
			Interface("reply", msg).
			Str("name", w.tag).
			Dur("duration", time.Since(info.firstAddedAt)).
			Msg("dns conn reply")
		if msg.Rcode == dns.RcodeSuccess && !msg.Truncated {
			w.rrCache.Set(msg)
		}
		p := &udp.Packet{
			Source:  info.dst,
			Target:  info.src,
			Payload: buf.New(),
		}
		err := msgIntoPacket(msg, p)
		if err != nil {
			p.Release()
			log.Err(err).Msg("failed to pack DNS message")
			return
		}
		w.writeResponse(p)
		// packedMsg, err := msg.PackBuffer(p.Payload.BytesTo(p.Payload.Cap()))
		// if err != nil {
		// 	p.Release()
		// 	log.Err(err).Msg("msg pack buffer")
		// } else {
		// 	if len(packedMsg) > int(p.Payload.Cap()) {
		// 		p.Payload.Release()
		// 		p.Payload = buf.FromBytes(packedMsg)
		// 	} else {
		// 		p.Payload.Resize(0, int32(len(packedMsg)))
		// 	}
		// 	w.writeResponse(p)
		// }
	}
}

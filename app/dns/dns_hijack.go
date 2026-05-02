package dns

import (
	"context"
	"errors"
	"io"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type HijackDns struct {
	local    *StaticDnsServer
	dnsRules []*DnsRule

	done          *done.Instance
	requests      chan *udp.Packet
	responses     chan *udp.Packet
	enableFakeDns atomic.Bool
}

func NewHijackDns(local *StaticDnsServer, rules []*DnsRule, enableDns bool) *HijackDns {
	d := &HijackDns{
		local:     local,
		dnsRules:  rules,
		done:      done.New(),
		requests:  make(chan *udp.Packet, 100),
		responses: make(chan *udp.Packet, 100),
	}
	d.enableFakeDns.Store(enableDns)
	return d
}

func (dsp *HijackDns) Start() error {
	go dsp.dispatchWorker()
	for _, rule := range dsp.dnsRules {
		if dnsConn, ok := rule.dnsServer.(DnsConn); ok {
			go dsp.handleConnResponse(dnsConn)
		}
	}
	return nil
}

func (dsp *HijackDns) Close() error {
	dsp.done.Close()
	return nil
}

func (dsp *HijackDns) SetFakeDnsEnabled(enabled bool) {
	dsp.enableFakeDns.Store(enabled)
}

func (dsp *HijackDns) GetFakeDnsEnabled() bool {
	return dsp.enableFakeDns.Load()
}

type DnsRule struct {
	conditions []Condition
	dnsServer  DnsServer
}

func NewDnsRule(dnsServer DnsServer, conditions ...Condition) *DnsRule {
	return &DnsRule{
		conditions: conditions,
		dnsServer:  dnsServer,
	}
}

func (d *DnsRule) match(msg *DnsMsgMeta) bool {
	for _, condition := range d.conditions {
		if !condition.Match(msg) {
			return false
		}
	}
	return true
}

type DnsMsgMeta struct {
	*dns.Msg
	Src *net.Destination
}

func (d *HijackDns) HandleQuery(ctx context.Context, msg *DnsMsgMeta, tcp bool) (*dns.Msg, error) {
	if len(msg.Question) == 0 {
		return nil, ErrNoQuestion
	}
	if d.local != nil {
		if rp, ok := d.local.ReplyFor(msg.Msg); ok {
			return rp, nil
		}
	}

	for _, dnsRule := range d.dnsRules {
		if isFakeDns(dnsRule.dnsServer) && !d.enableFakeDns.Load() {
			continue
		}
		if dnsRule.match(msg) {
			return dnsRule.dnsServer.HandleQuery(ctx, msg.Msg, tcp)
		}
	}
	return nil, ErrAllServersFailed
}

func (dsp *HijackDns) WritePacket(p *udp.Packet) error {
	if dsp.done.Done() {
		p.Release()
		return nil
	}
	select {
	case dsp.requests <- p:
		return nil
	default:
		p.Release()
		return errors.New("requests channel is blocked")
	}
}

func (dsp *HijackDns) ReadPacket() (*udp.Packet, error) {
	select {
	case <-dsp.done.Wait():
		return nil, io.EOF
	case p, open := <-dsp.responses:
		if !open {
			return nil, io.EOF
		}
		return p, nil
	}
}

func (dsp *HijackDns) writeReply(p *udp.Packet, reply *dns.Msg) {
	err := msgIntoPacket(reply, p)
	if err != nil {
		p.Release()
		log.Err(err).Msg("msgIntoPacket")
		return
	}
	p.Source, p.Target = p.Target, p.Source
	dsp.writeResponse(p)
}

func (dsp *HijackDns) dispatchWorker() {
	var msg dns.Msg
	for {
		select {
		case <-dsp.done.Wait():
			return
		case p, open := <-dsp.requests:
			if !open {
				return
			}

			if err := msg.Unpack(p.Payload.Bytes()); err != nil {
				p.Release()
				continue
			}

			if len(msg.Question) == 0 {
				p.Release()
				continue
			}

			if dsp.local != nil {
				if rp, ok := dsp.local.ReplyFor(&msg); ok {
					dsp.writeReply(p, rp)
					continue
				}
			}
			found := false
			for _, rule := range dsp.dnsRules {
				if isFakeDns(rule.dnsServer) && !dsp.enableFakeDns.Load() {
					continue
				}
				if rule.match(&DnsMsgMeta{Msg: &msg, Src: &p.Source}) {
					found = true
					if dnsConn, ok := rule.dnsServer.(DnsConn); ok {
						err := dnsConn.WritePacket(p)
						if err != nil {
							log.Error().Err(err).Msg("failed to write packet to dns conn")
						}
					} else if fakeDns, ok := rule.dnsServer.(*FakeDns); ok {
						rply, err := fakeDns.HandleQuery(context.Background(), &msg, false)
						if err != nil {
							log.Debug().Err(err).Msg("fakeDns.HandleQuery")
							p.Release()
						} else {
							dsp.writeReply(p, rply)
						}
					} else {
						go func() {
							msg := msg.Copy()
							ctx := log.Logger.WithContext(context.Background())
							reply, err := rule.dnsServer.HandleQuery(ctx, msg, false)
							if err != nil {
								p.Release()
								log.Err(err).Msg("DnsServer.HandleQuery")
								return
							}
							dsp.writeReply(p, reply)
						}()
					}
					break
				}
			}
			if !found {
				reply := emptyReply(&msg)
				dsp.writeReply(p, reply)
			}
		}
	}
}

func (t *HijackDns) handleConnResponse(conn DnsConn) {
	for {
		if t.done.Done() {
			return
		}
		p, err := conn.ReadPacket()
		if err != nil {
			return
		}
		t.writeResponse(p)
	}
}

func (dsp *HijackDns) writeResponse(p *udp.Packet) {
	if !dsp.done.Done() {
		select {
		case dsp.responses <- p:
			return
		default:
			log.Warn().Msg("responses channel is blocked")
		}
	}
	p.Release()
}

type HijackDnsToDnsServer struct {
	*HijackDns
}

func (d *HijackDnsToDnsServer) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	return d.HijackDns.HandleQuery(ctx, &DnsMsgMeta{Msg: msg}, tcp)
}

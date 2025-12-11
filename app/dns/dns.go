package dns

import (
	"context"
	"errors"
	"io"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type Dns struct {
	local      *StaticDnsServer
	dnsRules   []*DnsRule
	dnsServers []DnsServer

	done      *done.Instance
	requests  chan *udp.Packet
	responses chan *udp.Packet
}

func NewDns(local *StaticDnsServer, rules []*DnsRule, dnsServers []DnsServer) *Dns {
	return &Dns{
		local:      local,
		dnsRules:   rules,
		dnsServers: dnsServers,
		done:       done.New(),
		requests:   make(chan *udp.Packet, 100),
		responses:  make(chan *udp.Packet, 100),
	}
}

func (dsp *Dns) Start() error {
	for _, client := range dsp.dnsServers {
		if err := client.Start(); err != nil {
			return err
		}
	}
	go dsp.dispatchWorker()
	for _, client := range dsp.dnsServers {
		if dnsConn, ok := client.(DnsConn); ok {
			go dsp.handleConnResponse(dnsConn)
		}
	}
	return nil
}

func (dsp *Dns) Close() error {
	for _, dnsServer := range dsp.dnsServers {
		if err := dnsServer.Close(); err != nil {
			return err
		}
	}
	dsp.done.Close()
	return nil
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

func (d *DnsMsgMeta) Tcp() bool {
	return d.Src != nil && d.Src.Network == net.Network_TCP
}

func (d *Dns) HandleQuery(ctx context.Context, msg *DnsMsgMeta) (*dns.Msg, error) {
	if len(msg.Question) == 0 {
		return nil, ErrNoQuestion
	}
	if d.local != nil {
		if rp, ok := d.local.ReplyFor(msg.Msg); ok {
			return rp, nil
		}
	}

	for _, dnsRule := range d.dnsRules {
		if dnsRule.match(msg) {
			return dnsRule.dnsServer.HandleQuery(ctx, msg.Msg, msg.Tcp())
		}
	}
	return nil, ErrAllServersFailed
}

func (dsp *Dns) WritePacket(p *udp.Packet) error {
	if dsp.done.Done() {
		p.Release()
		return nil
	}
	select {
	case dsp.requests <- p:
	default:
		p.Release()
		return errors.New("requests channel is blocked")
	}
	return nil
}

func (dsp *Dns) ReadPacket() (*udp.Packet, error) {
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

func (dsp *Dns) dispatchWorker() {
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
					err := msgIntoPacket(rp, p)
					if err != nil {
						p.Release()
						log.Err(err).Msg("msgIntoPacket")
						continue
					}
					p.Source, p.Target = p.Target, p.Source
					dsp.writeResponse(p)
					continue
				}
			}

			found := false
			for _, rule := range dsp.dnsRules {
				if rule.match(&DnsMsgMeta{Msg: &msg, Src: &p.Source}) {
					found = true
					if dnsConn, ok := rule.dnsServer.(DnsConn); ok {
						err := dnsConn.WritePacket(p)
						if err != nil {
							log.Error().Err(err).Msg("failed to write packet to dns conn")
							break
						}
					} else if fakeDns, ok := rule.dnsServer.(*FakeDns); ok {
						rply, err := fakeDns.HandleQuery(context.Background(), &msg, false)
						if err != nil {
							log.Err(err).Msg("fakeDns.HandleQuery")
							p.Release()
							break
						}
						err = msgIntoPacket(rply, p)
						if err != nil {
							log.Err(err).Msg("msgIntoPacket")
							p.Release()
							break
						}
						p.Source, p.Target = p.Target, p.Source
						dsp.writeResponse(p)
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
							err = msgIntoPacket(reply, p)
							if err != nil {
								p.Release()
								log.Err(err).Msg("msgIntoPacket")
								return
							}
							p.Source, p.Target = p.Target, p.Source
							dsp.writeResponse(p)
						}()
					}
					break
				}
			}
			if !found {
				p.Release()
			}
		}
	}
}

func (t *Dns) handleConnResponse(conn DnsConn) {
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

func (dsp *Dns) writeResponse(p *udp.Packet) {
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

func addClientIP(msg *dns.Msg, clientIp net.IP) {
	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT

	subnet := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Family:        1, // IPv4
		SourceNetmask: 24,
		SourceScope:   0,
		Address:       clientIp.To4(),
	}
	if clientIp.To4() == nil && clientIp.To16() != nil {
		subnet.Family = 2 // IPv6
		subnet.SourceNetmask = 64
		subnet.Address = clientIp.To16()
	}
	o.Option = append(o.Option, subnet)
	msg.Extra = append(msg.Extra, o)
}

type DnsToDnsServer struct {
	*Dns
}

func (d *DnsToDnsServer) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	return d.Dns.HandleQuery(ctx, &DnsMsgMeta{Msg: msg})
}

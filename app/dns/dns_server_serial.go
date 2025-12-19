package dns

import (
	"context"
	"errors"
	sync "sync"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/i"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// does not reuse conn for dns query
type DnsServerSerial struct {
	sync.RWMutex
	doen       *done.Instance
	tcpDests   []net.Destination
	udpDests   []net.Destination
	dialer     i.Dialer
	tcpClient  *dns.Client
	udpClient  *dns.Client
	cache      *rrCache
	ipToDomain *IPToDomain

	startOnce sync.Once
	closeOnce sync.Once
}

func NewDnsServerSerial(dests []net.AddressPort, dialer i.Dialer, ipToDomain *IPToDomain) *DnsServerSerial {
	s := &DnsServerSerial{
		tcpClient: &dns.Client{
			Net: "tcp",
		},
		udpClient: &dns.Client{
			Net: "udp",
		},
		cache:      NewRrCache(),
		doen:       done.New(),
		ipToDomain: ipToDomain,
	}
	s.dialer = dialer
	for _, dest := range dests {
		s.tcpDests = append(s.tcpDests, net.Destination{
			Address: dest.Address,
			Port:    dest.Port,
			Network: net.Network_TCP,
		})
		s.udpDests = append(s.udpDests, net.Destination{
			Address: dest.Address,
			Port:    dest.Port,
			Network: net.Network_UDP,
		})
	}
	return s
}

func (d *DnsServerSerial) ReplaceDests(dests []net.AddressPort) {
	d.Lock()
	defer d.Unlock()
	d.setDests(dests)
}

func (d *DnsServerSerial) setDests(dests []net.AddressPort) {
	d.tcpDests = []net.Destination{}
	d.udpDests = []net.Destination{}
	for _, dest := range dests {
		d.tcpDests = append(d.tcpDests, net.Destination{
			Address: dest.Address,
			Port:    dest.Port,
			Network: net.Network_TCP,
		})
		d.udpDests = append(d.udpDests, net.Destination{
			Address: dest.Address,
			Port:    dest.Port,
			Network: net.Network_UDP,
		})
	}
}

func (d *DnsServerSerial) RemoveDest(toBeRemoved net.AddressPort, fallback []net.AddressPort) {
	d.Lock()
	defer d.Unlock()

	var dests []net.AddressPort

	for _, dest := range d.tcpDests {
		if dest.Address == toBeRemoved.Address && dest.Port == toBeRemoved.Port {
			continue
		}
		dests = append(dests, net.AddressPort{
			Address: dest.Address,
			Port:    dest.Port,
		})
	}
	if len(dests) == 0 {
		dests = fallback
	}
	d.setDests(dests)
}

func (d *DnsServerSerial) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	question := msg.Question[0]
	cachedMsg, ok := d.cache.Get(&question)
	if ok {
		log.Ctx(ctx).Debug().Str("domain", question.Name).Str("type", dns.TypeToString[question.Qtype]).
			Any("reply", cachedMsg).Msg("dns1 cache hit")
		return makeReply(msg, cachedMsg), nil
	}

	d.RLock()
	dests := d.udpDests
	client := d.udpClient
	if tcp {
		dests = d.tcpDests
		client = d.tcpClient
	}
	d.RUnlock()
	for _, dest := range dests {
		conn, err := d.dialer.Dial(ctx, dest)
		if err != nil {
			log.Ctx(ctx).Err(err).Str("dest", dest.String()).Msg("dns server1 dial failed")
			continue
		}
		defer conn.Close()
		dnsConn := &dns.Conn{
			Conn:    conn,
			UDPSize: client.UDPSize,
		}
		if msg == nil {
			log.Ctx(ctx).Fatal().Msg("msg is nil")
		}
		rspMsg, time, err := client.ExchangeWithConnContext(ctx, msg, dnsConn)
		if err != nil {
			log.Ctx(ctx).Err(err).Str("dest", dest.String()).Msg("dns server1 failed")
			continue
		}

		log.Ctx(ctx).Debug().Uint16("id", msg.Id).Str("domain", msg.Question[0].Name).
			Dur("time", time).Str("type", dns.TypeToString[msg.Question[0].Qtype]).
			Str("ns", dest.String()).Str("reply", rspMsg.String()).Msg("dns1 reply")
		if rspMsg.Rcode == dns.RcodeSuccess && !rspMsg.Truncated {
			d.cache.Set(rspMsg)
		}
		if d.ipToDomain != nil {
			d.ipToDomain.SetDomain(rspMsg, dest.Address)
		}
		return rspMsg, nil
	}
	log.Ctx(ctx).Debug().Str("domain", msg.Question[0].Name).Str("type", dns.TypeToString[msg.Question[0].Qtype]).
		Msg("direct dns failed")
	return nil, errors.New("all servers failed")
}

func (d *DnsServerSerial) Start() error {
	d.startOnce.Do(func() {
		d.cache.Start()
	})
	return nil
}

func (d *DnsServerSerial) Close() error {
	d.closeOnce.Do(func() {
		d.cache.Close()
	})
	return nil
}

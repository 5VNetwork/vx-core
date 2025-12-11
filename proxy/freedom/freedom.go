package freedom

import (
	"context"
	"fmt"
	"sync"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/helper"

	"github.com/rs/zerolog/log"
)

type FreedomHandler struct {
	tag            string
	Dialer         i.Dialer
	PacketListener i.PacketListener
	Dns            i.IPResolver
}

func New(dialer i.Dialer, pl i.PacketListener, tag string, dns i.IPResolver) *FreedomHandler {
	c := &FreedomHandler{
		Dialer:         dialer,
		PacketListener: pl,
		tag:            tag,
		Dns:            dns,
	}
	return c
}

func (c *FreedomHandler) Tag() string {
	return c.tag
}

func (c *FreedomHandler) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	target := dst
	if !target.IsValid() {
		return errors.New("freedom.Client does not know the request destination")
	}

	conn, err := c.Dialer.Dial(ctx, target)
	if err != nil {
		return fmt.Errorf("failed to open connection to %s, %w", target.String(), err)
	}
	defer conn.Close()

	log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("freedom.Client: dial ok")

	writer := buf.NewWriter(conn)
	reader := buf.NewReader(conn)
	return helper.Relay(ctx, rw, rw, reader, writer)
}

func (c *FreedomHandler) HandlePacketConn(ctx context.Context, dst net.Destination, rw udp.PacketReaderWriter) error {
	pc, err := c.PacketListener.ListenPacket(ctx, dst.ToUdpNetwork(), "")
	if err != nil {
		return fmt.Errorf("failed to dial packet conn: %w", err)
	}
	defer pc.Close()
	log.Ctx(ctx).Debug().Str("laddr", pc.LocalAddr().String()).Msg("freedom listen pc succ")

	return helper.RelayUDPPacketConn(ctx, rw,
		&domainToIpPacketConn{
			UdpConn: &udp.PacketRW{
				PacketReader: &udp.ReaderFromerToPacketReader{
					ReadFromer: pc,
				},
				PacketWriter: &udp.WriteToerToPacketWriter{
					WriteToer: pc,
				},
			},
			domainToIp: make(map[net.Address]net.Address),
			ipToDomain: make(map[net.Address]net.Address),
			dns:        c.Dns,
			ctx:        ctx,
		})
}

// It changes all packets' target address to ip address
type domainToIpPacketConn struct {
	udp.UdpConn
	domainToIp map[net.Address]net.Address
	mapLocks   sync.RWMutex
	ipToDomain map[net.Address]net.Address
	dns        i.IPResolver
	ctx        context.Context
}

func (d *domainToIpPacketConn) WritePacket(p *udp.Packet) error {
	if p.Target.Address.Family().IsDomain() {
		ipAddress, ok := d.domainToIp[p.Target.Address]
		if ok {
			p.Target.Address = ipAddress
		} else {
			ips, err := d.dns.LookupIPv4(d.ctx, p.Target.Address.Domain())
			if err != nil {
				return fmt.Errorf("failed to resolve domain %s, %w", p.Target.Address.Domain(), err)
			}
			if len(ips) == 0 {
				ips, err = d.dns.LookupIPv6(d.ctx, p.Target.Address.Domain())
				if err != nil {
					return fmt.Errorf("failed to resolve domain %s, %w", p.Target.Address.Domain(), err)
				}
				if len(ips) == 0 {
					return errors.New("no ip found for domain %s", p.Target.Address.Domain())
				}
			}
			ip := net.IPAddress(ips[0])
			domain := p.Target.Address
			d.domainToIp[domain] = ip
			d.mapLocks.Lock()
			d.ipToDomain[ip] = domain
			d.mapLocks.Unlock()
			p.Target.Address = ip
		}
	}
	return d.UdpConn.WritePacket(p)
}

func (d *domainToIpPacketConn) ReadPacket() (*udp.Packet, error) {
	p, err := d.UdpConn.ReadPacket()
	if err != nil {
		return nil, err
	}

	// d.mapLocks.RLock()
	// domain, found := d.ipToDomain[p.Source.Address]
	// if found {
	// 	p.Source.Address = domain
	// }
	// d.mapLocks.RUnlock()

	return p, err
}

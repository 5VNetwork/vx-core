// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"context"
	"fmt"
	"io"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/dns"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/task"

	d "github.com/miekg/dns"

	"github.com/rs/zerolog/log"
)

// intercept dns queries and send back dns responses
type DnsHandler struct {
	tag string
	dnsServer
}

type dnsServer interface {
	HandleQuery(ctx context.Context, msg *DnsMsgMeta, tcp bool) (*d.Msg, error)
}

func NewHandlerV() *DnsHandler {
	hander := &DnsHandler{}
	return hander
}

func (s *DnsHandler) WithDns(dns dnsServer) *DnsHandler {
	s.dnsServer = dns
	return s
}

func (s *DnsHandler) WithTag(tag string) *DnsHandler {
	s.tag = tag
	return s
}

func (s *DnsHandler) Tag() string {
	return s.tag
}

// implements i.FlowHandler
func (s *DnsHandler) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	network := dst.Network
	var reader dns.MessageReader //leftReader
	var writer dns.MessageWriter //leftWriter
	if network == net.Network_TCP {
		reader = dns.NewTCPReader(rw)
		writer = &dns.TCPWriter{
			Writer: rw,
		}
	} else {
		reader = &dns.UDPReader{
			Reader: rw,
		}
		writer = &dns.UDPWriter{
			Writer: rw,
		}
	}
	return s.handle(ctx, dst, reader, writer, network)
}

func (s *DnsHandler) HandlePacketConn(ctx context.Context, dst net.Destination, rw udp.PacketReaderWriter) error {
	reader := &dns.PacketReaderToMessageReader{
		PacketReader: rw,
	}
	writer := &dns.PackerWriterToMessageWriter{
		PacketWriter: rw,
		Src:          dst,
	}
	return s.handle(ctx, dst, reader, writer, net.Network_UDP)
}

func (s *DnsHandler) handle(ctx context.Context, dst net.Destination,
	reader dns.MessageReader, writer dns.MessageWriter, network net.Network) error {
	src, _ := inbound.SrcFromContext(ctx)
	return task.Run(ctx, func() error {
		for {
			b, err := reader.ReadMessage()
			if err != nil {
				b.Release()
				if err == io.EOF {
					return nil
				}
				return err
			}

			m := &d.Msg{}
			if err := m.Unpack(b.Bytes()); err != nil {
				b.Release()
				return fmt.Errorf("failed to unpack DNS message: %w", err)
			}

			go s.handleDnsQuery(ctx, b, m, writer, src)
		}
	})
}

func MsgToBuffer(msg *d.Msg) (*buf.Buffer, error) {
	b := buf.New()
	by, err := msg.PackBuffer(b.BytesTo(b.Cap()))
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Extend(int32(len(by)))
	return b, nil
}

func (s *DnsHandler) handleDnsQuery(ctx context.Context, b *buf.Buffer, msg *d.Msg, writer dns.MessageWriter, src net.Destination) {
	reply, err := s.dnsServer.HandleQuery(ctx, &DnsMsgMeta{Msg: msg, Src: &src}, src.Network == net.Network_TCP)
	if err != nil {
		b.Release()
		log.Ctx(ctx).Err(err).Msg("failed to handle DNS message")
		return
	}
	b.Clear()
	bCap := b.Cap()

	bp, err := reply.PackBuffer(b.BytesTo(bCap))
	if err != nil {
		b.Release()
		log.Ctx(ctx).Err(err).Msg("failed to pack DNS message")
		return
	}
	// a new buffer is allocated, so we need to release the old one
	if bCap < int32(len(bp)) {
		b.Release()
		b = buf.FromBytes(bp)
	} else {
		b.Extend(int32(len(bp)))
	}

	if err = writer.WriteMessage(b); err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to write reply DNS message back")
		return
	}
}

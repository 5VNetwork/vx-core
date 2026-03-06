package shadowsocks_2022

import (
	"context"
	"fmt"
	"strings"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/singbridge"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/rs/zerolog/log"
	shadowsocks "github.com/sagernet/sing-shadowsocks"
	"github.com/sagernet/sing-shadowsocks/shadowaead_2022"
	C "github.com/sagernet/sing/common"
	B "github.com/sagernet/sing/common/buf"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/common/network"
	N "github.com/sagernet/sing/common/network"
)

type Inbound struct {
	networks []net.Network
	service  shadowsocks.Service
	user     i.User
	handler  i.Handler
}

type ServerConfig struct {
	Method  string
	User    i.User
	Network []net.Network
	Handler i.Handler
}

func NewServer(config *ServerConfig) (*Inbound, error) {
	networks := config.Network
	if len(networks) == 0 {
		networks = []net.Network{
			net.Network_TCP,
			net.Network_UDP,
		}
	}
	inbound := &Inbound{
		networks: networks,
		user:     config.User,
		handler:  config.Handler,
	}
	if !C.Contains(shadowaead_2022.List, config.Method) {
		return nil, fmt.Errorf("unsupported method %s", config.Method)
	}
	keySize := keySizeForMethod(config.Method)
	key, err := toBase64PSK(strings.TrimSpace(config.User.Secret()), keySize)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}
	service, err := shadowaead_2022.NewServiceWithPassword(config.Method,
		key, 500, inbound, nil)
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}
	inbound.service = service
	return inbound, nil
}

func (i *Inbound) Network() []net.Network {
	return i.networks
}

func (i *Inbound) Process(ctx context.Context, conn net.Conn) error {
	ctx = proxy.ContextWithInboundProxyProtocol(ctx, "shadowsocks-2022")

	var metadata M.Metadata
	metadata.Source = M.ParseSocksaddr(conn.RemoteAddr().String())

	network := net.NetworkFromAddr(conn.LocalAddr())
	if network == net.Network_TCP {
		return i.service.NewConnection(ctx, conn, metadata)
	} else {
		reader := buf.NewReader(conn)
		pc := &natPacketConn{conn}
		for {
			mb, err := reader.ReadMultiBuffer()
			if err != nil {
				buf.ReleaseMulti(mb)
				return err
			}
			for _, buffer := range mb {
				packet := B.As(buffer.Bytes()).ToOwned()
				buffer.Release()
				err = i.service.NewPacket(ctx, pc, packet, metadata)
				if err != nil {
					packet.Release()
					buf.ReleaseMulti(mb)
					return err
				}
			}
		}
	}
}

func (i *Inbound) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	ctx = proxy.ContextWithUser(ctx, i.user)

	return i.handler.HandleFlow(ctx,
		singbridge.ToDestination(metadata.Destination, net.Network_TCP),
		buf.NewRWD(buf.NewReader(conn), buf.NewWriter(conn), conn))
}

func (i *Inbound) NewPacketConnection(ctx context.Context,
	conn N.PacketConn, metadata M.Metadata) error {
	ctx = proxy.ContextWithUser(ctx, i.user)
	destination := singbridge.ToDestination(metadata.Destination, net.Network_UDP)

	return i.handler.HandlePacketConn(ctx, destination,
		udp.PacketRW{
			PacketReader: &PacketReaderAdaptor{PacketReader: conn},
			PacketWriter: &PacketWriterAdaptor{PacketWriter: conn},
			OnClose: func() error {
				return conn.Close()
			},
		})
}

func (i *Inbound) NewError(ctx context.Context, err error) {
	if E.IsClosed(err) {
		return
	}
	log.Ctx(ctx).Err(err).Msg("shadowsocks-2022 inbound error")
}

type natPacketConn struct {
	net.Conn
}

func (c *natPacketConn) ReadPacket(buffer *B.Buffer) (addr M.Socksaddr, err error) {
	_, err = buffer.ReadFrom(c)
	return
}

func (c *natPacketConn) WritePacket(buffer *B.Buffer, addr M.Socksaddr) error {
	_, err := buffer.WriteTo(c)
	return err
}

type PacketReaderAdaptor struct {
	network.PacketReader
}

func (a *PacketReaderAdaptor) ReadPacket() (*udp.Packet, error) {
	buffer := buf.New()
	singBuffer := B.With(buffer.BytesTo(buffer.Cap()))
	addr, err := a.PacketReader.ReadPacket(singBuffer)
	if err != nil {
		return nil, err
	}
	buffer.Extend(int32(singBuffer.Len()))
	return &udp.Packet{
		Payload: buffer,
		Target:  singbridge.ToDestination(addr, net.Network_UDP),
	}, nil
}

type PacketWriterAdaptor struct {
	network.PacketWriter
}

func (rw *PacketWriterAdaptor) WritePacket(p *udp.Packet) error {
	defer p.Release()
	payload := p.Payload.Bytes()
	frontHeadroom := network.CalculateFrontHeadroom(rw.PacketWriter)
	rearHeadroom := network.CalculateRearHeadroom(rw.PacketWriter)
	singBuf := B.NewSize(frontHeadroom + len(payload) + rearHeadroom)
	if frontHeadroom > 0 {
		singBuf.Resize(frontHeadroom, 0)
	}
	_, err := singBuf.Write(payload)
	if err != nil {
		return err
	}
	return rw.PacketWriter.WritePacket(singBuf, singbridge.ToSocksaddr(p.Source))
}

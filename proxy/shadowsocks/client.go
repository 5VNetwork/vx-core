package shadowsocks

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/helper"

	"github.com/rs/zerolog/log"
)

// Client is a inbound handler for Shadowsocks protocol
type Client struct {
	ClientSettings
}

type ClientSettings struct {
	Address    net.Address
	PortPicker i.PortSelector
	Account    *MemoryAccount
	Dialer     i.Dialer
}

// NewClient create a new Shadowsocks client.
func NewClient(settings *ClientSettings) *Client {
	client := &Client{
		ClientSettings: *settings,
	}
	return client
}

func (c *Client) dial(ctx context.Context, network net.Network) (net.Conn, *MemoryAccount, error) {
	port := c.PortPicker.SelectPort()
	dest := net.TCPDestination(c.Address, net.Port(port))
	dest.Network = network
	conn, err := c.Dialer.Dial(ctx, dest)
	return conn, c.Account, err
}

func (c *Client) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	request := &protocol.RequestHeader{
		Version: Version,
		Address: dst.Address,
		Port:    dst.Port,
	}
	if dst.Network == net.Network_TCP {
		request.Command = protocol.RequestCommandTCP
	} else {
		request.Command = protocol.RequestCommandUDP
	}

	if request.Command == protocol.RequestCommandTCP {
		conn, memoryAccount, err := c.dial(ctx, dst.Network)
		if err != nil {
			return fmt.Errorf("failed to dial to %s, %w", dst.String(), err)
		}
		defer conn.Close()

		log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("shadowsocks dial ok")

		request.Account = memoryAccount
		requestDone := func() error {
			bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
			bodyWriter, err := WriteTCPRequest(request, bufferedWriter)
			if err != nil {
				return errors.New("failed to write request").Base(err)
			}
			if err = buf.CopyOnceTimeout(rw, bodyWriter, proxy.FirstPayloadTimeout); err != nil &&
				err != buf.ErrNotTimeoutReader && err != buf.ErrReadTimeout {
				return errors.New("failed to write A request payload").Base(err)
			}
			if err := bufferedWriter.SetBuffered(false); err != nil {
				return err
			}
			if err := buf.Copy(rw, bodyWriter); err != nil {
				err = fmt.Errorf("failed to transport all TCP request: %w", err)
				log.Ctx(ctx).Debug().Err(err).Send()
				return err
			}
			bodyWriter.CloseWrite()
			return nil
		}
		responseDone := func() error {
			responseReader, err := ReadTCPResponse(memoryAccount, conn)
			if err != nil {
				return err
			}
			if err := buf.Copy(responseReader, rw); err != nil {
				err = fmt.Errorf("failed to transport all TCP response: %w", err)
				log.Ctx(ctx).Debug().Err(err).Send()
				return err
			}
			rw.CloseWrite()
			return nil
		}
		if err := task.Run(ctx, requestDone, responseDone); err != nil {
			return fmt.Errorf("connnection ends with error: %w", err)
		}
		return nil
	}

	if request.Command == protocol.RequestCommandUDP {
		conn, a, err := c.dial(ctx, net.Network_UDP)
		if err != nil {
			return fmt.Errorf("failed to dial to %s, %w", dst.String(), err)
		}
		defer conn.Close()

		log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("shadowsocks dial ok")

		request.Account = a
		writer := &buf.SequentialWriter{Writer: &UDPWriter{
			Writer:  conn,
			Request: request,
		}}
		reader := &UDPReader{
			Reader: conn,
			User:   a,
		}
		return helper.Relay(ctx, rw, rw, reader, writer)
	}

	return nil
}

func (c *Client) HandlePacketConn(ctx context.Context, dst net.Destination, p udp.PacketReaderWriter) error {
	conn, a, err := c.dial(ctx, dst.Network)
	if err != nil {
		return err
	}
	defer conn.Close()

	request := &protocol.RequestHeader{
		Version: Version,
		Address: dst.Address,
		Port:    dst.Port,
		Account: a,
		Command: protocol.RequestCommandUDP,
	}

	return helper.RelayUDPPacketConn(ctx, p, &udp.PacketRW{
		PacketWriter: &UDPWriter{
			Writer:  conn,
			Request: request,
		},
		PacketReader: &UDPReader{
			Reader: conn,
			User:   a,
		},
	})
}

func (c *Client) ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error) {
	conn, a, err := c.dial(ctx, net.Network_UDP)
	if err != nil {
		return nil, err
	}

	request := &protocol.RequestHeader{
		Version: Version,
		Address: dst.Address,
		Port:    dst.Port,
		Account: a,
		Command: protocol.RequestCommandUDP,
	}

	return &udp.PacketRW{
		PacketWriter: &UDPWriter{
			Writer:  conn,
			Request: request,
		},
		PacketReader: &UDPReader{
			Reader: conn,
			User:   a,
		},
		OnClose: conn.Close,
	}, nil
}

func (c *Client) ProxyDial(ctx context.Context, dst net.Destination,
	initialData buf.MultiBuffer) (i.FlowConn, error) {
	target := dst
	if !target.IsValid() {
		return nil, errors.New("target not specified")
	}

	request := &protocol.RequestHeader{
		Version: Version,
		Address: target.Address,
		Port:    target.Port,
	}
	if target.Network == net.Network_TCP {
		request.Command = protocol.RequestCommandTCP
	} else {
		request.Command = protocol.RequestCommandUDP
	}

	if request.Command == protocol.RequestCommandTCP {
		conn, memoryAccount, err := c.dial(ctx, target.Network)
		if err != nil {
			return nil, fmt.Errorf("failed to dial to %s, %w", target.String(), err)
		}

		log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("shadowsocks dial ok")

		request.Account = memoryAccount
		bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
		bodyWriter, err := WriteTCPRequest(request, bufferedWriter)
		if err != nil {
			conn.Close()
			return nil, errors.New("failed to write request").Base(err)
		}
		if initialData.Len() > 0 {
			err := bodyWriter.WriteMultiBuffer(initialData)
			if err != nil {
				conn.Close()
				return nil, errors.New("failed to write initial data").Base(err)
			}
		}
		if err := bufferedWriter.SetBuffered(false); err != nil {
			conn.Close()
			return nil, err
		}

		return &clientBuffConn{
			conn:          conn,
			Writer:        bodyWriter,
			memoryAccount: memoryAccount,
		}, nil
	}

	if request.Command == protocol.RequestCommandUDP {
		conn, a, err := c.dial(ctx, net.Network_UDP)
		if err != nil {
			return nil, fmt.Errorf("failed to dial to %s, %w", target.String(), err)
		}

		log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("shadowsocks dial ok")

		request.Account = a
		writer := &buf.SequentialWriter{Writer: &UDPWriter{
			Writer:  conn,
			Request: request,
		}}
		reader := &UDPReader{
			Reader: conn,
			User:   a,
		}
		return &clientBuffConn{
			conn:   conn,
			Reader: reader,
			Writer: writer,
		}, nil
	}

	return nil, errors.New("invalid command")
}

type clientBuffConn struct {
	conn net.Conn
	buf.Writer
	buf.Reader
	memoryAccount *MemoryAccount
}

func (c *clientBuffConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if c.Reader == nil {
		responseReader, err := ReadTCPResponse(c.memoryAccount, c.conn)
		if err != nil {
			return nil, err
		}
		c.Reader = responseReader
	}
	return c.Reader.ReadMultiBuffer()
}

func (c *clientBuffConn) Close() error {
	return c.conn.Close()
}

// func (c *Client) Dial(ctx context.Context, dst net.Destination) (net.Conn, error) {
// 	return c.DialWithInitialData(ctx, dst, nil)
// }

// func (c *Client) DialWithInitialData(ctx context.Context, dst net.Destination, initialData []byte) (net.Conn, error) {
// 	target := dst
// 	if !target.IsValid() {
// 		return nil, errors.New("target not specified")
// 	}

// 	request := &protocol.RequestHeader{
// 		Version: Version,
// 		Address: target.Address,
// 		Port:    target.Port,
// 	}
// 	if target.Network == net.Network_TCP {
// 		request.Command = protocol.RequestCommandTCP
// 	} else {
// 		request.Command = protocol.RequestCommandUDP
// 	}

// 	if request.Command == protocol.RequestCommandTCP {
// 		conn, memoryAccount, err := c.dial(ctx, target.Network)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to dial to %s, %w", target.String(), err)
// 		}

// 		log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("shadowsocks dial ok")

// 		request.Account = memoryAccount
// 		bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
// 		bodyWriter, err := WriteTCPRequestIO(request, bufferedWriter)
// 		if err != nil {
// 			conn.Close()
// 			return nil, errors.New("failed to write request").Base(err)
// 		}
// 		if initialData != nil {
// 			_, err := bodyWriter.Write(initialData)
// 			if err != nil {
// 				conn.Close()
// 				return nil, errors.New("failed to write initial data").Base(err)
// 			}
// 		}
// 		if err := bufferedWriter.SetBuffered(false); err != nil {
// 			conn.Close()
// 			return nil, err
// 		}

// 		return &clientConn{
// 			Conn:          conn,
// 			Writer:        bodyWriter,
// 			memoryAccount: memoryAccount,
// 		}, nil
// 	}

// 	if request.Command == protocol.RequestCommandUDP {
// 		conn, a, err := c.dial(ctx, net.Network_UDP)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to dial to %s, %w", target.String(), err)
// 		}

// 		log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("shadowsocks dial ok")

// 		request.Account = a
// 		writer := &UDPWriter{
// 			Writer:  conn,
// 			Request: request,
// 		}
// 		reader := &UDPReader{
// 			Reader: conn,
// 			User:   a,
// 		}
// 		return proxy.NewProxyConn(proxy.CustomConnOption{
// 			Conn:   conn,
// 			Reader: reader,
// 			Writer: writer,
// 		}), nil
// 	}

// 	return nil, errors.New("invalid command")
// }

// type clientConn struct {
// 	net.Conn
// 	io.Writer
// 	io.Reader
// 	memoryAccount *MemoryAccount
// }

// func (c *clientConn) Write(b []byte) (int, error) {
// 	return c.Writer.Write(b)
// }

// func (c *clientConn) Read(b []byte) (int, error) {
// 	if c.Reader == nil {
// 		responseReader, err := ReadTCPResponseIO(c.memoryAccount, c.Conn)
// 		if err != nil {
// 			return 0, err
// 		}
// 		c.Reader = responseReader
// 	}
// 	return c.Reader.Read(b)
// }

// func (c *clientConn) CloseWrite() error {
// 	if closeWriter, ok := c.Conn.(buf.CloseWriter); ok {
// 		return closeWriter.CloseWrite()
// 	}
// 	return nil
// }

package tcp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	gonet "net"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
)

type Server struct {
	Port         net.Port
	MsgProcessor func(msg []byte) []byte
	ShouldClose  bool
	SendFirst    []byte
	Listen       net.Address
	listener     gonet.Listener
	TlsConfig    *tls.Config
}

func (server *Server) Start() (net.Destination, error) {
	return server.StartContext(context.Background(), nil)
}

func (server *Server) StartContext(ctx context.Context, sockopt *dlhelper.SocketSetting) (net.Destination, error) {
	listenerAddr := server.Listen
	if listenerAddr == nil {
		listenerAddr = net.LocalHostIP
	}
	listener, err := dlhelper.ListenSystem(ctx, &gonet.TCPAddr{
		IP:   listenerAddr.IP(),
		Port: int(server.Port),
	}, sockopt)
	if err != nil {
		return net.Destination{}, err
	}

	localAddr := listener.Addr().(*gonet.TCPAddr)
	server.Port = net.Port(localAddr.Port)
	server.listener = listener
	go server.acceptConnections(listener.(*gonet.TCPListener))

	return net.TCPDestination(net.IPAddress(localAddr.IP), net.Port(localAddr.Port)), nil
}

func (server *Server) acceptConnections(listener *gonet.TCPListener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed accept TCP connection: %v\n", err)
			return
		}

		if server.TlsConfig != nil {
			conn = tls.Server(conn, server.TlsConfig)
		}

		go server.handleConnection(conn)
	}
}

func (server *Server) handleConnection(conn gonet.Conn) {
	if len(server.SendFirst) > 0 {
		conn.Write(server.SendFirst)
	}

	p := pipe.NewPipe(-1, false)
	err := task.Run(context.Background(),
		// read from conn, and write them to pipe
		func() error {
			defer p.Close()
			for {
				b := buf.New()
				if _, err := b.ReadOnce(conn); err != nil {
					if err == io.EOF {
						return nil
					}
					fmt.Println("failed to read from conn: ", err.Error())
					return err
				}
				// fmt.Printf("received data: %d\n", b.Len())
				copy(b.Bytes(), server.MsgProcessor(b.Bytes()))
				if err := p.WriteMultiBuffer(buf.MultiBuffer{b}); err != nil {
					return err
				}
			}
		},
		// read from pipe, and write them to conn
		func() error {
			defer p.Interrupt(nil)
			w := buf.NewWriter(conn)
			for {
				mb, err := p.ReadMultiBuffer()
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
				if err := w.WriteMultiBuffer(mb); err != nil {
					fmt.Println("failed to write to conn: ", err.Error())
					return err
				}
			}
		})
	if err != nil {
		fmt.Println("failed to transfer data: ", err.Error())
	}
	err = conn.Close()
	// if err != nil {
	// 	fmt.Println("failed to close connection: ", err.Error())
	// }
}

func (server *Server) Close() error {
	return server.listener.Close()
}

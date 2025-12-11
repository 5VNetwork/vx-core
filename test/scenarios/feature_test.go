//go:build test

package scenarios

import (
	"context"
	"testing"

	"github.com/5vnetwork/vx-core/app/buildserver"
	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
)

func TestPassiveConnection(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
		SendFirst:    []byte("send first"),
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)

	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.DokodemoConfig{
						Address:  dest.Address.String(),
						Port:     uint32(dest.Port),
						Networks: []net.Network{net.Network_TCP},
					},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: int(serverPort),
	})
	common.Must(err)

	{
		response := make([]byte, 1024)
		nBytes, err := conn.Read(response)
		common.Must(err)
		if string(response[:nBytes]) != "send first" {
			t.Error("unexpected first response: ", string(response[:nBytes]))
		}
	}

	if err := WriteToConn(conn, 1024, Timeout)(); err != nil {
		t.Error(err)
	}
}

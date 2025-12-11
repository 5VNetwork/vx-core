//go:build test

package scenarios

import (
	"testing"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/test/servers/udp"

	"golang.org/x/sync/errgroup"
)

func TestFreedomUDP(t *testing.T) {
	clientPort := udp.PickPort()
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  "127.0.0.1",
							Port:     uint32(udpDest.Port),
							Networks: []net.Network{net.Network_UDP},
						},
					),
				},
			},
		},
	}
	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group

	for i := 0; i < 5; i++ {
		errg.Go(TestUDPConnN(clientPort, 1024, Timeout, 1024))
	}

	if err := errg.Wait(); err != nil {
		t.Fatal(err)
	}
}

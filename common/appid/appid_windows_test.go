package appid_test

import (
	"context"
	"log"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/appid"
	"github.com/5vnetwork/vx-core/common/net"

	"github.com/5vnetwork/vx-core/test/servers/tcp"
)

func TestGetAppId(t *testing.T) {
	tcpServer := tcp.Server{}
	tcpDest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()
	// tcp4
	tcpConn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{
		IP:   net.LocalHostIP.IP(),
		Port: int(tcpDest.Port),
	})
	common.Must(err)
	defer tcpConn.Close()
	appId, err := appid.GetAppId(context.Background(),
		net.DestinationFromAddr(tcpConn.LocalAddr()), nil)
	common.Must(err)
	t.Logf("appId: %s", appId)
	// udp4
	udpConn, err := net.ListenUDP("udp4", nil)
	common.Must(err)
	defer udpConn.Close()
	appId, err = appid.GetAppId(context.Background(),
		net.DestinationFromAddr(udpConn.LocalAddr()), nil)
	common.Must(err)
	t.Logf("appId: %s", appId)
}

func TestGetAppIdUdp6(t *testing.T) {
	udpConn, err := net.DialUDP("udp6", nil, &net.UDPAddr{
		IP:   net.LocalHostIPv6.IP(),
		Port: 53,
	})
	common.Must(err)
	defer udpConn.Close()
	log.Print("udpConn: ", udpConn.LocalAddr())
	appId, err := appid.GetAppId(context.Background(),
		net.DestinationFromAddr(udpConn.LocalAddr()), nil)
	common.Must(err)
	t.Logf("appId: %s", appId)

	udpConn, err = net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP("1.1.1.1"),
		Port: 53,
	})
	common.Must(err)
	defer udpConn.Close()
	log.Print("udpConn: ", udpConn.LocalAddr())
	appId, err = appid.GetAppId(context.Background(),
		net.DestinationFromAddr(udpConn.LocalAddr()), nil)
	common.Must(err)
	t.Logf("appId: %s", appId)
}

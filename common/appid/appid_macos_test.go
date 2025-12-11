//go:build darwin && !ios
// +build darwin,!ios

package appid_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/appid"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
)

// func TestGetAppId11(t *testing.T) {
// 	appId, err := appid.GetAppId(context.Background(), net.Destination{
// 		Address: net.ParseAddress("172.23.27.1"),
// 		Port:    52910,
// 		Network: net.Network_UDP,
// 	}, nil)
// 	common.Must(err)
// 	t.Logf("appId: %s", appId)
// }

// func TestGetAppId(t *testing.T) {
// 	currentProcessName := GetProcessName()
// 	tcpServer := tcp.Server{}
// 	tcpDest, err := tcpServer.Start()
// 	common.Must(err)
// 	defer tcpServer.Close()
// 	// tcp4
// 	tcpConn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{
// 		IP:   net.LocalHostIP.IP(),
// 		Port: int(tcpDest.Port),
// 	})
// 	common.Must(err)
// 	defer tcpConn.Close()
// 	appId, err := appid.GetAppId(context.Background(),
// 		net.DestinationFromAddr(tcpConn.LocalAddr()), nil)
// 	common.Must(err)
// 	if appId != currentProcessName {
// 		t.Fatalf("appId: %s", appId)
// 	}
// 	// udp4
// 	udpConn, err := net.ListenUDP("udp4", nil)
// 	common.Must(err)
// 	defer udpConn.Close()
// 	appId, err = appid.GetAppId(context.Background(),
// 		net.DestinationFromAddr(udpConn.LocalAddr()), nil)
// 	common.Must(err)
// 	if appId != currentProcessName {
// 		t.Fatalf("appId: %s", appId)
// 	}
// }

// GetProcessName returns the name of the current process
func GetProcessName() string {
	// Get the full path of the executable
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	// Return just the base name of the executable
	return filepath.Base(executable)
}

// func BenchmarkGetAppId(b *testing.B) {
// 	tcpServer := tcp.Server{}
// 	tcpDest, err := tcpServer.Start()
// 	common.Must(err)
// 	defer tcpServer.Close()
// 	tcpConn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{
// 		IP:   net.LocalHostIP.IP(),
// 		Port: int(tcpDest.Port),
// 	})
// 	common.Must(err)
// 	defer tcpConn.Close()
// 	for i := 0; i < b.N; i++ {
// 		appid.GetAppId(context.Background(), net.DestinationFromAddr(tcpConn.LocalAddr()), nil)
// 	}
// }

func BenchmarkGetAppIdSB(b *testing.B) {
	tcpServer := tcp.Server{}
	tcpDest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()
	tcpConn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{
		IP:   net.LocalHostIP.IP(),
		Port: int(tcpDest.Port),
	})
	common.Must(err)
	defer tcpConn.Close()
	for i := 0; i < b.N; i++ {
		appid.GetAppId(context.Background(), net.DestinationFromAddr(tcpConn.LocalAddr()), nil)
	}
}

func TestGetAppId(t *testing.T) {
	currentProcessName := GetProcessName()
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
	if filepath.Base(appId) != currentProcessName {
		t.Fatalf("appId: %s", appId)
	}
	// udp4
	udpConn, err := net.ListenUDP("udp4", nil)
	common.Must(err)
	defer udpConn.Close()
	appId, err = appid.GetAppId(context.Background(),
		net.DestinationFromAddr(udpConn.LocalAddr()), nil)
	common.Must(err)
	if filepath.Base(appId) != currentProcessName {
		t.Fatalf("appId: %s", appId)
	}
	// udp6
	udpConn, err = net.ListenUDP("udp6", nil)
	common.Must(err)
	defer udpConn.Close()
	appId, err = appid.GetAppId(context.Background(),
		net.DestinationFromAddr(udpConn.LocalAddr()), nil)
	common.Must(err)
	if filepath.Base(appId) != currentProcessName {
		t.Fatalf("appId: %s", appId)
	}
	// udp
	udpConn, err = net.ListenUDP("udp", nil)
	common.Must(err)
	defer udpConn.Close()
	appId, err = appid.GetAppId(context.Background(),
		net.DestinationFromAddr(udpConn.LocalAddr()), nil)
	common.Must(err)
	if filepath.Base(appId) != currentProcessName {
		t.Fatalf("appId: %s", appId)
	}
}

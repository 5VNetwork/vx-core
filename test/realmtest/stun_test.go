//go:build test

package realmtest

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/apernet/hysteria/extras/v2/correctnet"
	"github.com/apernet/hysteria/extras/v2/realm"
	"github.com/stretchr/testify/require"
)

func TestReflectiveSTUNWithPunchConn(t *testing.T) {
	stunAddr, stop := StartReflectiveSTUN(t)
	defer stop()

	baseConn, err := correctnet.ListenUDP("udp4", &net.UDPAddr{Port: 0})
	require.NoError(t, err)
	defer baseConn.Close()

	punchConn, err := realm.NewPunchPacketConn(baseConn, 0)
	require.NoError(t, err)

	addrs, err := realm.Discover(context.Background(), punchConn.PacketConn, realm.STUNConfig{
		Servers: []string{stunAddr},
		Timeout: time.Second,
	})
	require.NoError(t, err)
	require.NotEmpty(t, addrs)
	t.Logf("discovered: %v", addrs)
}

func TestReflectiveSTUNServerPort(t *testing.T) {
	stunAddr, stop := StartReflectiveSTUN(t)
	defer stop()

	baseConn, err := correctnet.ListenUDP("udp", &net.UDPAddr{Port: 0})
	require.NoError(t, err)
	defer baseConn.Close()

	addrs, err := realm.Discover(context.Background(), baseConn, realm.STUNConfig{
		Servers: []string{stunAddr},
		Timeout: time.Second,
	})
	require.NoError(t, err)
	require.NotEmpty(t, addrs)
	t.Logf("discovered: %v local %v", addrs, baseConn.LocalAddr())
}

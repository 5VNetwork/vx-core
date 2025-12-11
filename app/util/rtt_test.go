package util_test

import (
	"context"
	"testing"

	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/transport"
)

func TestRttTest(t *testing.T) {
	dest := net.AddressPort{
		Address: net.ParseAddress("dbfnjsnc.kobvlq.vvvvv.cloud"),
		Port:    net.Port(443),
	}

	rtt, err := util.RttTest(context.Background(), dest,
		transport.DefaultDialer, transport.DefaultPacketListener,
		dns.NewGoIpResolver())
	if err != nil {
		t.Error(err)
	}
	t.Logf("RTT: %v", rtt)
}

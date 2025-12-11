package dns_test

import (
	"context"
	"net"
	"testing"

	configs "github.com/5vnetwork/vx-core/app/configs"
	d "github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/common"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/strmatcher"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/rs/zerolog/log"

	"github.com/google/go-cmp/cmp"
)

func TestUDPServer(t *testing.T) {
	dnsServer1 := d.NewDnsServerSerial([]mynet.AddressPort{
		{
			Address: mynet.ParseAddress("127.0.0.1"),
			Port:    mynet.Port(port),
		},
	}, transport.DefaultDialer, nil)
	dnsServer1.Start()
	defer dnsServer1.Close()

	dns := d.DnsServerToResolver{
		DnsServers: []d.DnsServer{dnsServer1},
	}

	logger := log.With().Str("dns", "test").Logger()
	ctx := logger.WithContext(context.Background())

	{
		ips, err := dns.LookupIP(ctx, "google.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}

		if r := cmp.Diff(ips, []net.IP{{8, 8, 8, 8}}); r != "" {
			t.Fatal(r)
		}
	}

	{
		ips, err := dns.LookupIP(ctx, "facebook.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}

		if r := cmp.Diff(ips, []net.IP{{9, 9, 9, 9}}); r != "" {
			t.Fatal(r)
		}
	}
	{
		ips, _ := dns.LookupIP(ctx, "notexist.google.com")
		if len(ips) != 0 {
			t.Fatal("more than 0 ips: ", ips)
		}
	}

	{
		ips, _ := dns.LookupIPv6(ctx, "ipv4only.google.com")
		if len(ips) != 0 {
			t.Fatal("ips: ", ips)
		}
	}

	{
		ips, err := dns.LookupIP(ctx, common.Must2(strmatcher.ToDomain("🍕.ws")).(string))
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}

		if r := cmp.Diff(ips, []net.IP{{208, 100, 42, 200}}); r != "" {
			t.Fatal(r)
		}
	}

	{
		ips, err := dns.LookupIP(ctx, common.Must2(strmatcher.ToDomain("ああああ.com")).(string))
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}

		if r := cmp.Diff(ips, []net.IP{{0, 0xa, 0, 0xa, 0, 0xa, 0, 0xa, 0, 0, 0, 0, 0, 0, 0xaa, 0xaa}}); r != "" {
			t.Fatal(r)
		}
	}

	{
		ips, err := dns.LookupIP(ctx, "google.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}

		if r := cmp.Diff(ips, []net.IP{{8, 8, 8, 8}}); r != "" {
			t.Fatal(r)
		}
	}
}

func TestUDPServerIPv6(t *testing.T) {
	dnsServer1 := d.NewDnsServerSerial([]mynet.AddressPort{
		{
			Address: mynet.ParseAddress("127.0.0.1"),
			Port:    mynet.Port(port),
		},
	}, transport.DefaultDialer, nil)

	dns := d.DnsServerToResolver{
		DnsServers: []d.DnsServer{dnsServer1},
	}
	dnsServer1.Start()
	defer dnsServer1.Close()

	{
		ips, err := dns.LookupIPv6(context.Background(), "ipv6.google.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}

		if r := cmp.Diff(ips, []net.IP{{32, 1, 72, 96, 72, 96, 0, 0, 0, 0, 0, 0, 0, 0, 136, 136}}); r != "" {
			t.Fatal(r)
		}
	}
}

func TestStaticHostDomain(t *testing.T) {
	staticDnsServer := d.NewStaticDnsServer([]*configs.Record{
		{
			Domain: "example.com",
			Ip:     []string{"8.8.8.8"},
		},
	})

	dnsServer1 := d.NewDnsServerSerial([]mynet.AddressPort{
		{
			Address: mynet.ParseAddress("127.0.0.1"),
			Port:    mynet.Port(port),
		},
	}, transport.DefaultDialer, nil)

	dns := d.NewInternalDns(staticDnsServer, dnsServer1)
	dnsServer1.Start()
	defer dnsServer1.Close()

	{
		ips, err := dns.LookupIP(context.Background(), "example.com")
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}

		if r := cmp.Diff(ips, []net.IP{{8, 8, 8, 8}}); r != "" {
			t.Fatal(r)
		}
	}

}

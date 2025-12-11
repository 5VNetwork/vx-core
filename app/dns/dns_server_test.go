package dns_test

import (
	"context"
	"testing"
	"time"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"

	d "github.com/miekg/dns"
)

func TestDnsServerConcurrentUdp(t *testing.T) {
	freedom := freedom.New(
		transport.DefaultDialer,
		transport.DefaultPacketListener,
		"", nil,
	)
	ns := dns.NewDnsServerConcurrent(dns.DnsServerConcurrentOption{
		Name: "test",
		NameserverAddrs: []net.AddressPort{
			{
				Address: net.LocalHostIP,
				Port:    net.Port(port),
			}},
		Handler:    freedom,
		Dispatcher: dispatcher.NewPacketDispatcher(context.Background(), freedom),
	})

	ctx := context.Background()
	// A
	m := new(d.Msg)
	m.SetQuestion("www.apple.com.", d.TypeA)
	reply, err := ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	rr, ok := reply.Answer[0].(*d.A)
	if !ok {
		t.Fatalf("expected A record, got %T", reply.Answer[0])
	}
	if rr.A.String() != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %s", rr.A.String())
	}
	// AAAA
	m.SetQuestion("ipv6.google.com.", d.TypeAAAA)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	rrr, ok := reply.Answer[0].(*d.AAAA)
	if !ok {
		t.Fatalf("expected AAAA record, got %T", reply.Answer[0])
	}
	if rrr.AAAA.String() != "2001:4860:4860::8888" {
		t.Fatalf("expected 2001:4860:4860::8888, got %s", rrr.AAAA.String())
	}
	// rcode error
	m.SetQuestion("notexist.google.com.", d.TypeAAAA)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if reply.Rcode != d.RcodeNameError {
		t.Fatalf("expected rcode name error, got %d", reply.Rcode)
	}
	// MX
	m.SetQuestion("mx.google.com.", d.TypeMX)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	mxr, ok := reply.Answer[0].(*d.MX)
	if !ok {
		t.Fatalf("expected MX record, got %T", reply.Answer[0])
	}
	if mxr.Mx != "google.com." {
		t.Fatalf("expected google.com., got %s", mxr.Mx)
	}
	if mxr.Preference != 10 {
		t.Fatalf("expected preference 10, got %d", mxr.Preference)
	}
	// NS
	m.SetQuestion("ns.google.com.", d.TypeNS)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	nsr, ok := reply.Answer[0].(*d.NS)
	if !ok {
		t.Fatalf("expected NS record, got %T", reply.Answer[0])
	}
	if nsr.Ns != "ns1.google.com." {
		t.Fatalf("expected ns1.google.com., got %s", nsr.Ns)
	}
	// CNAME
	m.SetQuestion("cname.google.com.", d.TypeCNAME)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	cnr, ok := reply.Answer[0].(*d.CNAME)
	if !ok {
		t.Fatalf("expected CNAME record, got %T", reply.Answer[0])
	}
	if cnr.Target != "google.com." {
		t.Fatalf("expected google.com., got %s", cnr.Target)
	}
}

// TODO
func TestDnsServerConcurrentTcp(t *testing.T) {
	freedom := freedom.New(
		transport.DefaultDialer,
		transport.DefaultPacketListener,
		"", nil,
	)
	ns := dns.NewDnsServerConcurrent(dns.DnsServerConcurrentOption{
		Name: "test",
		NameserverAddrs: []net.AddressPort{
			{
				Address: net.LocalHostIP,
				Port:    net.Port(port),
			},
		},
		Handler:    freedom,
		Dispatcher: dispatcher.NewPacketDispatcher(context.Background(), freedom),
	})

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	// A
	m := new(d.Msg)
	m.SetQuestion("www.apple.com.", d.TypeA)
	reply, err := ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	rr, ok := reply.Answer[0].(*d.A)
	if !ok {
		t.Fatalf("expected A record, got %T", reply.Answer[0])
	}
	if rr.A.String() != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %s", rr.A.String())
	}
	// AAAA
	m.SetQuestion("ipv6.google.com.", d.TypeAAAA)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	rrr, ok := reply.Answer[0].(*d.AAAA)
	if !ok {
		t.Fatalf("expected AAAA record, got %T", reply.Answer[0])
	}
	if rrr.AAAA.String() != "2001:4860:4860::8888" {
		t.Fatalf("expected 2001:4860:4860::8888, got %s", rrr.AAAA.String())
	}
	// rcode error
	m.SetQuestion("notexist.google.com.", d.TypeAAAA)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if reply.Rcode != d.RcodeNameError {
		t.Fatalf("expected rcode name error, got %d", reply.Rcode)
	}
	// MX
	m.SetQuestion("mx.google.com.", d.TypeMX)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	mxr, ok := reply.Answer[0].(*d.MX)
	if !ok {
		t.Fatalf("expected MX record, got %T", reply.Answer[0])
	}
	if mxr.Mx != "google.com." {
		t.Fatalf("expected google.com., got %s", mxr.Mx)
	}
	if mxr.Preference != 10 {
		t.Fatalf("expected preference 10, got %d", mxr.Preference)
	}
	// NS
	m.SetQuestion("ns.google.com.", d.TypeNS)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	nsr, ok := reply.Answer[0].(*d.NS)
	if !ok {
		t.Fatalf("expected NS record, got %T", reply.Answer[0])
	}
	if nsr.Ns != "ns1.google.com." {
		t.Fatalf("expected ns1.google.com., got %s", nsr.Ns)
	}
	// CNAME
	m.SetQuestion("cname.google.com.", d.TypeCNAME)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	cnr, ok := reply.Answer[0].(*d.CNAME)
	if !ok {
		t.Fatalf("expected CNAME record, got %T", reply.Answer[0])
	}
	if cnr.Target != "google.com." {
		t.Fatalf("expected google.com., got %s", cnr.Target)
	}
}

func TestNameserverVLocalRecords(t *testing.T) {
	ns := dns.NewStaticDnsServer(
		[]*configs.Record{
			{Domain: "www.apple.com", Ip: []string{"127.0.0.1"}},
			{Domain: "www.ipv6.com", Ip: []string{"2001:4860:4860::8888"}},
			{Domain: "www.baidu.com", ProxiedDomain: "baidu.com"},
		},
	)
	// apple
	m := new(d.Msg)
	m.SetQuestion("www.apple.com.", d.TypeA)
	reply, ok := ns.ReplyFor(m)
	if !ok {
		t.Fatal("expected reply, got nil")
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	r, ok := reply.Answer[0].(*d.A)
	if !ok {
		t.Fatalf("expected AAAA record, got %T", reply.Answer[0])
	}
	if r.A.String() != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %s", r.A.String())
	}
	// ipv6
	m.SetQuestion("www.ipv6.com.", d.TypeAAAA)
	reply, ok = ns.ReplyFor(m)
	if !ok {
		t.Fatal("expected reply, got nil")
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	rrr, ok := reply.Answer[0].(*d.AAAA)
	if !ok {
		t.Fatalf("expected AAAA record, got %T", reply.Answer[0])
	}
	if rrr.AAAA.String() != "2001:4860:4860::8888" {
		t.Fatalf("expected 2001:4860:4860::8888, got %s", rrr.AAAA.String())
	}
	// baidu
	m.SetQuestion("www.baidu.com.", d.TypeCNAME)
	reply, ok = ns.ReplyFor(m)
	if !ok {
		t.Fatal("expected reply, got nil")
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	cname, ok := reply.Answer[0].(*d.CNAME)
	if !ok {
		t.Fatalf("expected CNAME record, got %T", reply.Answer[0])
	}
	if cname.Target != "baidu.com." {
		t.Fatalf("expected baidu.com., got %s", cname.Target)
	}
}

func TestDnsServer1(t *testing.T) {
	ns := dns.NewDnsServerSerial(
		[]net.AddressPort{
			{Address: net.LocalHostIP, Port: net.Port(port)},
		},
		transport.DefaultDialer,
		nil,
	)

	ctx := context.Background()
	// A
	m := new(d.Msg)
	m.SetQuestion("www.apple.com.", d.TypeA)
	reply, err := ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	rr, ok := reply.Answer[0].(*d.A)
	if !ok {
		t.Fatalf("expected A record, got %T", reply.Answer[0])
	}
	if rr.A.String() != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %s", rr.A.String())
	}
	// AAAA
	m.SetQuestion("ipv6.google.com.", d.TypeAAAA)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	rrr, ok := reply.Answer[0].(*d.AAAA)
	if !ok {
		t.Fatalf("expected AAAA record, got %T", reply.Answer[0])
	}
	if rrr.AAAA.String() != "2001:4860:4860::8888" {
		t.Fatalf("expected 2001:4860:4860::8888, got %s", rrr.AAAA.String())
	}
	// rcode error
	m.SetQuestion("notexist.google.com.", d.TypeAAAA)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if reply.Rcode != d.RcodeNameError {
		t.Fatalf("expected rcode name error, got %d", reply.Rcode)
	}
	// MX
	m.SetQuestion("mx.google.com.", d.TypeMX)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	mxr, ok := reply.Answer[0].(*d.MX)
	if !ok {
		t.Fatalf("expected MX record, got %T", reply.Answer[0])
	}
	if mxr.Mx != "google.com." {
		t.Fatalf("expected google.com., got %s", mxr.Mx)
	}
	if mxr.Preference != 10 {
		t.Fatalf("expected preference 10, got %d", mxr.Preference)
	}
	// NS
	m.SetQuestion("ns.google.com.", d.TypeNS)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	nsr, ok := reply.Answer[0].(*d.NS)
	if !ok {
		t.Fatalf("expected NS record, got %T", reply.Answer[0])
	}
	if nsr.Ns != "ns1.google.com." {
		t.Fatalf("expected ns1.google.com., got %s", nsr.Ns)
	}
	// CNAME
	m.SetQuestion("cname.google.com.", d.TypeCNAME)
	reply, err = ns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	cnr, ok := reply.Answer[0].(*d.CNAME)
	if !ok {
		t.Fatalf("expected CNAME record, got %T", reply.Answer[0])
	}
	if cnr.Target != "google.com." {
		t.Fatalf("expected google.com., got %s", cnr.Target)
	}
}

func TestDnsServer1Tcp(t *testing.T) {
	ns := dns.NewDnsServerSerial(
		[]net.AddressPort{
			{Address: net.LocalHostIP, Port: net.Port(port)},
		},
		transport.DefaultDialer,
		nil,
	)

	ctx := context.Background()
	// A
	m := new(d.Msg)
	m.SetQuestion("www.apple.com.", d.TypeA)
	reply, err := ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	rr, ok := reply.Answer[0].(*d.A)
	if !ok {
		t.Fatalf("expected A record, got %T", reply.Answer[0])
	}
	if rr.A.String() != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %s", rr.A.String())
	}
	// AAAA
	m.SetQuestion("ipv6.google.com.", d.TypeAAAA)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	rrr, ok := reply.Answer[0].(*d.AAAA)
	if !ok {
		t.Fatalf("expected AAAA record, got %T", reply.Answer[0])
	}
	if rrr.AAAA.String() != "2001:4860:4860::8888" {
		t.Fatalf("expected 2001:4860:4860::8888, got %s", rrr.AAAA.String())
	}
	// rcode error
	m.SetQuestion("notexist.google.com.", d.TypeAAAA)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if reply.Rcode != d.RcodeNameError {
		t.Fatalf("expected rcode name error, got %d", reply.Rcode)
	}
	// MX
	m.SetQuestion("mx.google.com.", d.TypeMX)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	mxr, ok := reply.Answer[0].(*d.MX)
	if !ok {
		t.Fatalf("expected MX record, got %T", reply.Answer[0])
	}
	if mxr.Mx != "google.com." {
		t.Fatalf("expected google.com., got %s", mxr.Mx)
	}
	if mxr.Preference != 10 {
		t.Fatalf("expected preference 10, got %d", mxr.Preference)
	}
	// NS
	m.SetQuestion("ns.google.com.", d.TypeNS)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	nsr, ok := reply.Answer[0].(*d.NS)
	if !ok {
		t.Fatalf("expected NS record, got %T", reply.Answer[0])
	}
	if nsr.Ns != "ns1.google.com." {
		t.Fatalf("expected ns1.google.com., got %s", nsr.Ns)
	}
	// CNAME
	m.SetQuestion("cname.google.com.", d.TypeCNAME)
	reply, err = ns.HandleQuery(ctx, m, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	cnr, ok := reply.Answer[0].(*d.CNAME)
	if !ok {
		t.Fatalf("expected CNAME record, got %T", reply.Answer[0])
	}
	if cnr.Target != "google.com." {
		t.Fatalf("expected google.com., got %s", cnr.Target)
	}
}

func TestFakeDnsServer(t *testing.T) {
	pools, err := dns.NewPools([]*configs.FakeDnsServer_PoolConfig{
		{
			Cidr:    "192.168.0.0/24",
			LruSize: 100,
		},
		{
			Cidr:    "2001:4860:4860::/64",
			LruSize: 100,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	m := new(d.Msg)

	fakeDns := dns.NewFakeDns(pools)
	// Test A record
	m.SetQuestion("test.com.", d.TypeA)
	reply, err := fakeDns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	a, ok := reply.Answer[0].(*d.A)
	if !ok {
		t.Fatalf("expected A record, got %T", reply.Answer[0])
	}
	if !pools.IsIPInIPPool(net.ParseAddress(a.A.String())) {
		t.Fatalf("IP %v not in pool", a.A)
	}

	// Test AAAA record
	m.SetQuestion("test.com.", d.TypeAAAA)
	reply, err = fakeDns.HandleQuery(ctx, m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(reply.Answer))
	}
	aaaa, ok := reply.Answer[0].(*d.AAAA)
	if !ok {
		t.Fatalf("expected AAAA record, got %T", reply.Answer[0])
	}
	if !pools.IsIPInIPPool(net.ParseAddress(aaaa.AAAA.String())) {
		t.Fatalf("IP %v not in pool", aaaa.AAAA)
	}

	// Test unsupported record type
	m.SetQuestion("test.com.", d.TypeMX)
	_, err = fakeDns.HandleQuery(ctx, m, false)
	if err == nil {
		t.Fatal("expected error for unsupported query type")
	}
}

package dns_test

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/miekg/dns"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/test/nameserver"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
)

var (
	port uint16
)

func TestMain(m *testing.M) {
	// test.InitZeroLog()

	port = net.PickUDPPort()
	dnsServer := dns.Server{
		Addr:    "127.0.0.1:" + common.Uint16ToString(port),
		Net:     "udp",
		Handler: &nameserver.StaticHandler{},
		UDPSize: 1200,
	}
	go dnsServer.ListenAndServe()

	dnsServerTCP := dns.Server{
		Addr:    "127.0.0.1:" + common.Uint16ToString(port),
		Net:     "tcp",
		Handler: &nameserver.StaticHandler{},
	}
	go dnsServerTCP.ListenAndServe()
	log.Println("dns server started: ", port)
	exitVal := m.Run()
	dnsServer.Shutdown()
	dnsServerTCP.Shutdown()
	os.Exit(exitVal)
}

func TestDNSUDPTunnel(t *testing.T) {
	serverPort := udp.PickPort()
	config := &configs.TmConfig{
		DefaultNicMonitor: true,
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Tag:     "dd",
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxy.DokodemoConfig{
						Address:  "114.114.114.114",
						Port:     53,
						Networks: []net.Network{net.Network_UDP},
					}),
				},
			},
		},
		Dns: &configs.DnsConfig{
			DnsRules: []*configs.DnsRuleConfig{
				{
					DnsServerName: "dns",
				},
			},
			DnsServers: []*configs.DnsServerConfig{
				{
					Type: &configs.DnsServerConfig_PlainDnsServer{
						PlainDnsServer: &configs.PlainDnsServer{
							Addresses: []string{"127.0.0.1:" + common.Uint16ToString(port)},
						},
					},
					Name: "dns",
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Tag:      "direct",
					Protocol: serial.ToTypedMessage(&proxy.FreedomConfig{}),
				},
			},
		},
		Router: &configs.RouterConfig{
			Rules: []*configs.RuleConfig{
				{
					InboundTags: []string{"dns"},
					OutboundTag: "direct",
				},
				{
					InboundTags: []string{"dd"},
					OutboundTag: "dns",
				},
			},
		},
	}

	v, err := buildclient.NewX(config)
	common.Must(err)

	common.Must(v.Start())
	defer v.Close()

	{
		m1 := new(dns.Msg)
		m1.Id = dns.Id()
		m1.RecursionDesired = true
		m1.Question = make([]dns.Question, 1)
		m1.Question[0] = dns.Question{Name: "google.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET}

		c := new(dns.Client)
		c.Timeout = 10 * time.Second

		in, _, err := c.Exchange(m1, "127.0.0.1:"+strconv.Itoa(int(serverPort)))
		common.Must(err)

		if len(in.Answer) != 1 {
			t.Fatal("len(answer): ", len(in.Answer))
		}

		rr, ok := in.Answer[0].(*dns.A)
		if !ok {
			t.Fatal("not A record")
		}
		if r := cmp.Diff(rr.A[:], net.IP{8, 8, 8, 8}); r != "" {
			t.Error(r)
		}
	}

	{
		m1 := new(dns.Msg)
		m1.Id = dns.Id()
		m1.RecursionDesired = true
		m1.Question = make([]dns.Question, 1)
		m1.Question[0] = dns.Question{Name: "ipv4only.google.com.", Qtype: dns.TypeAAAA, Qclass: dns.ClassINET}

		c := new(dns.Client)
		c.Timeout = 10 * time.Second
		in, _, err := c.Exchange(m1, "127.0.0.1:"+strconv.Itoa(int(serverPort)))
		common.Must(err)

		if len(in.Answer) != 0 {
			t.Fatal("len(answer): ", len(in.Answer))
		}
	}

	{
		m1 := new(dns.Msg)
		m1.Id = dns.Id()
		m1.RecursionDesired = true
		m1.Question = make([]dns.Question, 1)
		m1.Question[0] = dns.Question{Name: "notexist.google.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET}

		c := new(dns.Client)
		in, _, err := c.Exchange(m1, "127.0.0.1:"+strconv.Itoa(int(serverPort)))
		common.Must(err)

		if in.Rcode != dns.RcodeNameError {
			t.Error("expected NameError, but got ", in.Rcode)
		}
	}
}

func TestDNSTCPTunnel(t *testing.T) {
	serverPort := tcp.PickPort()
	config := &configs.TmConfig{
		DefaultNicMonitor: true,
		Dns: &configs.DnsConfig{
			DnsRules: []*configs.DnsRuleConfig{
				{
					DnsServerName: "dns",
				},
			},
			DnsServers: []*configs.DnsServerConfig{
				{
					Type: &configs.DnsServerConfig_PlainDnsServer{
						PlainDnsServer: &configs.PlainDnsServer{
							Addresses: []string{"127.0.0.1:" + common.Uint16ToString(port)},
						},
					},
					Name: "dns",
				},
			},
		},
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Tag:     "tun",
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxy.DokodemoConfig{
						Address:  "114.114.114.114",
						Port:     53,
						Networks: []net.Network{net.Network_TCP},
					}),
				},
			},
		},
		Router: &configs.RouterConfig{
			Rules: []*configs.RuleConfig{
				{
					InboundTags: []string{"dns"},
					OutboundTag: "direct",
				},
				{
					InboundTags: []string{"tun"},
					OutboundTag: "dns",
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Tag:      "direct",
					Protocol: serial.ToTypedMessage(&proxy.FreedomConfig{}),
				},
			},
		},
	}
	v, err := buildclient.NewX(config)
	common.Must(err)
	common.Must(v.Start())
	defer v.Close()

	{
		m1 := new(dns.Msg)
		m1.Id = dns.Id()
		m1.RecursionDesired = true
		m1.Question = make([]dns.Question, 1)
		m1.Question[0] = dns.Question{Name: "google.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET}

		c := &dns.Client{
			Net: "tcp",
		}
		in, _, err := c.Exchange(m1, "127.0.0.1:"+serverPort.String())
		common.Must(err)

		if len(in.Answer) != 1 {
			t.Fatal("len(answer): ", len(in.Answer))
		}

		rr, ok := in.Answer[0].(*dns.A)
		if !ok {
			t.Fatal("not A record")
		}
		if r := cmp.Diff(rr.A[:], net.IP{8, 8, 8, 8}); r != "" {
			t.Error(r)
		}
	}

	// {
	// 	m1 := new(dns.Msg)
	// 	m1.Id = dns.Id()
	// 	m1.RecursionDesired = true
	// 	m1.Question = make([]dns.Question, 1)
	// 	m1.Question[0] = dns.Question{Name: "www.baidu.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET}

	// 	c := &dns.Client{
	// 		Net:     "tcp",
	// 		Timeout: 100 * time.Second,
	// 	}

	// 	in, _, err := c.Exchange(m1, "127.0.0.1:"+strconv.Itoa(int(serverPort)))
	// 	common.Must(err)

	// 	if len(in.Answer) == 0 {
	// 		t.Fatal("len(answer): ", len(in.Answer))
	// 	}

	// 	log.Println(in.Answer)
	// }
}

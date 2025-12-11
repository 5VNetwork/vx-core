package nameserver

import (
	"net"

	"github.com/5vnetwork/vx-core/common"

	"github.com/miekg/dns"
)

type StaticHandler struct{}

func (*StaticHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	ans := new(dns.Msg)
	ans.SetReply(r)
	// log.Println("received DNS query from", w.RemoteAddr(), "for", r.Question[0].Name, "id", r.Id)
	var clientIP net.IP

	opt := r.IsEdns0()
	if opt != nil {
		for _, o := range opt.Option {
			if o.Option() == dns.EDNS0SUBNET {
				subnet := o.(*dns.EDNS0_SUBNET)
				clientIP = subnet.Address
			}
		}
	}

	for _, q := range r.Question {
		switch {
		case q.Name == "www.apple.com." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("www.apple.com. IN A 127.0.0.1")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "google.com." && q.Qtype == dns.TypeA:
			if clientIP == nil {
				rr, _ := dns.NewRR("google.com. IN A 8.8.8.8")
				ans.Answer = append(ans.Answer, rr)
			} else {
				rr, _ := dns.NewRR("google.com. IN A 8.8.4.4")
				ans.Answer = append(ans.Answer, rr)
			}

		case q.Name == "api.google.com." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("api.google.com. IN A 8.8.7.7")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "v2.api.google.com." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("v2.api.google.com. IN A 8.8.7.8")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "facebook.com." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("facebook.com. IN A 9.9.9.9")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "ipv6.google.com." && q.Qtype == dns.TypeA:
			rr, err := dns.NewRR("ipv6.google.com. IN A 8.8.8.7")
			common.Must(err)
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "ipv6.google.com." && q.Qtype == dns.TypeAAAA:
			rr, err := dns.NewRR("ipv6.google.com. IN AAAA 2001:4860:4860::8888")
			common.Must(err)
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "notexist.google.com." && q.Qtype == dns.TypeAAAA:
			ans.MsgHdr.Rcode = dns.RcodeNameError
		case q.Name == "notexist.google.com." && q.Qtype == dns.TypeA:
			ans.MsgHdr.Rcode = dns.RcodeNameError
		case q.Name == "hostname." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("hostname. IN A 127.0.0.1")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "hostname.local." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("hostname.local. IN A 127.0.0.1")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "hostname.localdomain." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("hostname.localdomain. IN A 127.0.0.1")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "localhost." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("localhost. IN A 127.0.0.2")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "localhost-a." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("localhost-a. IN A 127.0.0.3")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "localhost-b." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("localhost-b. IN A 127.0.0.4")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "Mijia\\ Cloud." && q.Qtype == dns.TypeA:
			rr, _ := dns.NewRR("Mijia\\ Cloud. IN A 127.0.0.1")
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "xn--vi8h.ws." /* 🍕.ws */ && q.Qtype == dns.TypeA:
			rr, err := dns.NewRR("xn--vi8h.ws. IN A 208.100.42.200")
			common.Must(err)
			ans.Answer = append(ans.Answer, rr)

		case q.Name == "xn--l8jaaa.com." /* ああああ.com */ && q.Qtype == dns.TypeAAAA:
			rr, err := dns.NewRR("xn--l8jaaa.com. IN AAAA a:a:a:a::aaaa")
			common.Must(err)
			ans.Answer = append(ans.Answer, rr)

		// MX
		case q.Name == "mx.google.com." && q.Qtype == dns.TypeMX:
			rr, _ := dns.NewRR("mx.google.com. IN MX 10 google.com.")
			ans.Answer = append(ans.Answer, rr)
		// NS
		case q.Name == "ns.google.com." && q.Qtype == dns.TypeNS:
			rr, _ := dns.NewRR("ns.google.com. IN NS ns1.google.com.")
			ans.Answer = append(ans.Answer, rr)
		// CNAME
		case q.Name == "cname.google.com." && q.Qtype == dns.TypeCNAME:
			rr, _ := dns.NewRR("cname.google.com. IN CNAME google.com.")
			ans.Answer = append(ans.Answer, rr)
		}

	}
	w.WriteMsg(ans)
}

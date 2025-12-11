// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Portions of this code are modified from XTLS/RealiTLScanner
// Original source: https://github.com/XTLS/RealiTLScanner

package api

import (
	context "context"
	"crypto/tls"
	"log/slog"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/5vnetwork/vx-core/app/dns"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/rs/zerolog/log"
)

const (
	timeout = 60 * time.Second
)

func (a *Api) RunRealiScanner(ctx context.Context, req *RunRealiScannerRequest) (*RunRealiScannerResponse, error) {
	results, err := RunRealiScanner(ctx, req.Addr)
	if err != nil {
		return nil, err
	}
	return &RunRealiScannerResponse{
		Results: results,
	}, nil
}

// RunRealiScanner is modified from https://github.com/XTLS/RealiTLScanner
// Modifications include:
// - Adapted to use local DNS resolver instead of system resolver
// - Changed return type to use RealiScannerResult struct
// - Modified context handling and timeout logic
func RunRealiScanner(ctx context.Context, addr string) ([]*RealiScannerResult, error) {
	var results []*RealiScannerResult
	outChan := make(chan string)
	hostChan, err := IterateAddr(ctx, addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	go func() {
		for host := range hostChan {
			time.Sleep(1 * time.Millisecond)
			go ScanTLS(ctx, host, outChan)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return results, nil
		case s := <-outChan:
			parts := strings.Split(s, ",")
			results = append(results, &RealiScannerResult{
				Ip:     parts[0],
				Domain: parts[2],
			})
			if len(results) >= 100 {
				return results, nil
			}
		}
	}
}

func ScanTLS(ctx context.Context, host Host, out chan<- string) {
	hostPort := net.JoinHostPort(host.IP.String(), strconv.Itoa(443))
	dialer := net.Dialer{
		Timeout: 10 * time.Second,
	}
	conn, err := dialer.DialContext(ctx, "tcp", hostPort)
	if err != nil {
		log.Debug().Msgf("Cannot dial %s", hostPort)
		return
	}
	defer conn.Close()
	err = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.Error().Msgf("Error setting deadline %s", err)
		return
	}
	tlsCfg := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
		CurvePreferences:   []tls.CurveID{tls.X25519},
	}
	if host.Type == HostTypeDomain {
		tlsCfg.ServerName = host.Origin
	}
	c := tls.Client(conn, tlsCfg)
	err = c.Handshake()
	if err != nil {
		log.Debug().Msgf("TLS handshake failed %s", hostPort)
		return
	}
	state := c.ConnectionState()
	alpn := state.NegotiatedProtocol
	domain := state.PeerCertificates[0].Subject.CommonName
	issuers := strings.Join(state.PeerCertificates[0].Issuer.Organization, " | ")
	feasible := true
	if state.Version != tls.VersionTLS13 || alpn != "h2" || len(domain) == 0 || len(issuers) == 0 {
		// not feasible
		feasible = false
	} else {
		out <- strings.Join([]string{host.IP.String(), host.Origin, domain, "\"" + issuers}, ",") +
			"\n"
	}
	// log("Connected to target", "feasible", feasible, "ip", host.IP.String(),
	// 	"origin", host.Origin,
	// 	"tls", tls.VersionName(state.Version), "alpn", alpn, "cert-domain", domain, "cert-issuer", issuers,
	// )
	log.Debug().Bool("feasible", feasible).
		Str("ip", host.IP.String()).Str("origin", host.Origin).
		Str("tls", tls.VersionName(state.Version)).Str("alpn", alpn).
		Str("cert-domain", domain).Str("cert-issuer", issuers).Msgf("Connected to target %s", hostPort)
}

const (
	_ = iota
	HostTypeIP
	HostTypeCIDR
	HostTypeDomain
)

type HostType int

type Host struct {
	IP     net.IP
	Origin string
	Type   HostType
}

func IterateAddr(ctx context.Context, addr string) (<-chan Host, error) {
	var err error
	hostChan := make(chan Host)
	// _, _, err := net.ParseCIDR(addr)
	// if err == nil {
	// 	// is CIDR
	// 	return Iterate(strings.NewReader(addr))
	// }
	ip := net.ParseIP(addr)
	if ip == nil {
		ip, err = LookupIP(ctx, addr)
		if err != nil {
			return nil, err
		}
	}
	go func() {
		slog.Info("Enable infinite mode", "init", ip.String())
		lowIP := ip
		highIP := ip
		hostChan <- Host{
			IP:     ip,
			Origin: addr,
			Type:   HostTypeIP,
		}

		for i := 0; i < math.MaxInt; i++ {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if i%2 == 0 {
				lowIP = NextIP(lowIP, false)
				hostChan <- Host{
					IP:     lowIP,
					Origin: lowIP.String(),
					Type:   HostTypeIP,
				}
			} else {
				highIP = NextIP(highIP, true)
				hostChan <- Host{
					IP:     highIP,
					Origin: highIP.String(),
					Type:   HostTypeIP,
				}
			}
		}
	}()
	return hostChan, nil
}
func LookupIP(ctx context.Context, addr string) (net.IP, error) {
	dnsServer := dns.NewDnsServerSerial([]mynet.AddressPort{
		{Address: mynet.ParseAddress("1.1.1.1"), Port: 53}}, transport.DefaultDialer, nil)
	dnsServer.Start()
	defer dnsServer.Close()
	ipResolver := dns.NewDnsServerToResolver(dnsServer)

	ips, err := ipResolver.LookupIPv4(ctx, addr)
	if err != nil {
		return nil, err
	}
	return ips[0], nil
}

func NextIP(ip net.IP, increment bool) net.IP {
	// Convert to big.Int and increment
	ipb := big.NewInt(0).SetBytes(ip)
	if increment {
		ipb.Add(ipb, big.NewInt(1))
	} else {
		ipb.Sub(ipb, big.NewInt(1))
	}

	// Add leading zeros
	b := ipb.Bytes()
	b = append(make([]byte, len(ip)-len(b)), b...)
	return b
}

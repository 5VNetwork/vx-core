package dns

// import (
// 	"context"
// 	"errors"
// 	"net"

// 	net1 "github.com/5vnetwork/vx-core/common/net"
// )

// type LocalNameServer struct {
// }

// func NewLocalNameServer() *LocalNameServer {
// 	return &LocalNameServer{}
// }

// // QueryIP implements Server.
// func (s *LocalNameServer) QueryIP(ctx context.Context, domain string, option IPOption) ([]net.IP, error) {
// 	var ips []net.IP
// 	var err error

// 	ips, err = net.LookupIP(domain)
// 	if err != nil {
// 		return nil, err
// 	}

// 	parsedIPs := make([]net.IP, 0, len(ips))
// 	for _, ip := range ips {
// 		parsed := net1.IPAddress(ip)
// 		if parsed != nil {
// 			if option.IPv4Enable && parsed.Family().IsIPv4() {
// 				parsedIPs = append(parsedIPs, parsed.IP())
// 			} else if option.IPv6Enable && parsed.Family().IsIPv6() {
// 				parsedIPs = append(parsedIPs, parsed.IP())
// 			}
// 		}
// 	}
// 	if len(parsedIPs) == 0 {
// 		return nil, errors.New("no ip found")
// 	}
// 	return parsedIPs, nil
// }

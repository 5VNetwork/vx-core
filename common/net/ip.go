package net

import (
	"math/rand"
	"net"
	"net/netip"
	"time"
)

var (
	PrivateAddress = netip.PrefixFrom(netip.AddrFrom4([4]byte{172, 16, 0, 0}), 12)
)

func RandomIP(startIP, endIP net.IP) net.IP {
	start := IpToInt(startIP.To4())
	end := IpToInt(endIP.To4())
	rand.Seed(time.Now().UnixNano())
	randomIP := rand.Intn(int(end-start)) + int(start)
	return IntToIP(randomIP)
}

func IpToInt(ip net.IP) int {
	return int(ip[0])<<24 + int(ip[1])<<16 + int(ip[2])<<8 + int(ip[3])
}

func IntToIP(ip int) net.IP {
	return net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func IsDirectedBoradcast(ip net.IP) bool {
	ip4 := ip.To4()
	return ip4 != nil && ip4.IsPrivate() && ip4[3] == 255
}

func StandardIp(ip net.IP) net.IP {
	if newIP := ip.To4(); newIP != nil {
		return newIP
	} else if newIP := ip.To16(); newIP != nil {
		return newIP
	} else {
		return ip
	}
}

package geo

import (
	"fmt"
	"net"
	"net/netip"

	vxgeo "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/common/geo"
	"github.com/rs/zerolog/log"
	"go4.org/netipx"
)

// a geoIPMathcer corr to one country's IP
type IPMatcher struct {
	ReverseMatch bool //return the opposite result
	ipSetv4      *netipx.IPSet
	ipSetv6      *netipx.IPSet
}

func NewIPMatcherFromGeoCidrs(cidrs []*vxgeo.CIDR, inverse bool) (*IPMatcher, error) {
	pMatcher := &IPMatcher{
		ReverseMatch: inverse,
	}
	if err := pMatcher.init(cidrs); err != nil {
		return nil, err
	}
	return pMatcher, nil
}

func (m *IPMatcher) Match(ip net.IP) bool {
	isMatched := false
	if ip.To4() != nil {
		isMatched = m.match4(ip)
	} else if ip.To16() != nil {
		isMatched = m.match6(ip)
	}
	if m.ReverseMatch {
		return !isMatched
	}
	return isMatched
}

func (m *IPMatcher) match4(ip net.IP) bool {
	nip, ok := netipx.FromStdIP(ip)
	if !ok {
		return false
	}
	return m.ipSetv4.Contains(nip)
}

func (m *IPMatcher) match6(ip net.IP) bool {
	nip, ok := netipx.FromStdIP(ip)
	if !ok {
		return false
	}
	return m.ipSetv6.Contains(nip)
}

func (m *IPMatcher) init(cidrs []*vxgeo.CIDR) error {
	var builder4, builder6 netipx.IPSetBuilder
	for _, cidr := range cidrs {
		netaddrIP, ok := netip.AddrFromSlice(cidr.GetIp())
		if !ok {
			return fmt.Errorf("invalid IP address %s", cidr.String())
		}
		netaddrIP = netaddrIP.Unmap()
		ipPrefix := netip.PrefixFrom(netaddrIP, int(cidr.GetPrefix()))

		if !ipPrefix.IsValid() {
			log.Warn().Str("cidr", cidr.String()).Msg("invalid IP prefix")
			continue
		}

		switch {
		case netaddrIP.Is4():
			builder4.AddPrefix(ipPrefix)
		case netaddrIP.Is6():
			builder6.AddPrefix(ipPrefix)
		}
	}

	var err error
	m.ipSetv4, err = builder4.IPSet()
	if err != nil {
		return err
	}
	m.ipSetv6, err = builder6.IPSet()
	if err != nil {
		return err
	}

	return nil
}

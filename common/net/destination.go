package net

import (
	"net"
	"strings"
)

type AddressPort struct {
	Address Address
	Port    Port
}

func (ap AddressPort) String() string {
	if ap.Address == nil {
		return ""
	}
	return ap.Address.String() + ":" + ap.Port.String()
}

// Destination represents a network destination including address and protocol (tcp / udp).
type Destination struct {
	Address Address
	Port    Port
	Network Network
}

func (t Destination) ToAddressPort() AddressPort {
	return AddressPort{Address: t.Address, Port: t.Port}
}

func (t Destination) ToUdpNetwork() string {
	if t.Address == nil {
		return "udp"
	}
	if t.Address.Family().IsIPv4() {
		return "udp4"
	} else if t.Address.Family().IsIPv6() {
		return "udp6"
	}
	return "udp"
}

var AnyUdpDest = Destination{
	Address: AnyIP,
	Network: Network_UDP,
}

// DestinationFromAddr generates a Destination from a net address.
func DestinationFromAddr(addr net.Addr) Destination {
	switch addr := addr.(type) {
	case *net.TCPAddr:
		return TCPDestination(IPAddress(addr.IP), Port(addr.Port))
	case *net.UDPAddr:
		return UDPDestination(IPAddress(addr.IP), Port(addr.Port))
	case *net.UnixAddr:
		return UnixDestination(DomainAddress(addr.Name))
	default:
		panic("Net: Unknown address type.")
	}
}

func (d *Destination) Addr() net.Addr {
	switch d.Network {
	case Network_TCP:
		return &net.TCPAddr{
			IP:   d.Address.IP(),
			Port: int(d.Port),
		}
	case Network_UDP:
		return &net.UDPAddr{
			IP:   d.Address.IP(),
			Port: int(d.Port),
		}
	case Network_UNIX:
		return &net.UnixAddr{
			Name: d.Address.String(),
			Net:  "unix",
		}
	default:
		panic("Net: Unknown network type.")
	}
}

// ParseDestination converts a destination from its string presentation.
func ParseDestination(dest string) (Destination, error) {
	d := Destination{
		Address: AnyIP,
		Port:    Port(0),
	}

	switch {
	case strings.HasPrefix(dest, "tcp:"):
		d.Network = Network_TCP
		dest = dest[4:]
	case strings.HasPrefix(dest, "udp:"):
		d.Network = Network_UDP
		dest = dest[4:]
	case strings.HasPrefix(dest, "unix:"):
		d = UnixDestination(DomainAddress(dest[5:]))
		return d, nil
	}

	hstr, pstr, err := net.SplitHostPort(dest)
	if err != nil {
		return d, err
	}
	if len(hstr) > 0 {
		d.Address = ParseAddress(hstr)
	}
	if len(pstr) > 0 {
		port, err := PortFromString(pstr)
		if err != nil {
			return d, err
		}
		d.Port = port
	}
	return d, nil
}

// TCPDestination creates a TCP destination with given address
func TCPDestination(address Address, port Port) Destination {
	return Destination{
		Network: Network_TCP,
		Address: address,
		Port:    port,
	}
}

// UDPDestination creates a UDP destination with given address
func UDPDestination(address Address, port Port) Destination {
	return Destination{
		Network: Network_UDP,
		Address: address,
		Port:    port,
	}
}

// UnixDestination creates a Unix destination with given address
func UnixDestination(address Address) Destination {
	return Destination{
		Network: Network_UNIX,
		Address: address,
	}
}

// NetAddr returns the network address in this Destination in string form.
func (d Destination) NetAddr() string {
	addr := ""
	if d.Address == nil {
		return ""
	}
	if d.Network == Network_TCP || d.Network == Network_UDP {
		addr = d.Address.String() + ":" + d.Port.String()
	} else if d.Network == Network_UNIX {
		addr = d.Address.String()
	}
	return addr
}

// String returns the strings form of this Destination.
func (d Destination) String() string {
	return d.NetAddr()
}

// IsValid returns true if this Destination is valid, i.e.
// it is not unknown
func (d Destination) IsValid() bool {
	return d.Address != nil && d.Network != Network_Unknown
}

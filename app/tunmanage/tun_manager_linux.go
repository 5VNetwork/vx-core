package tunmanage

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"strings"

	"github.com/5vnetwork/vx-core/tun"
	"github.com/rs/zerolog/log"
	"github.com/vishvananda/netlink"
)

type TunManager struct {
	option  *tun.TunOption
	enable6 bool
}

func NewTunManager(option *tun.TunOption, enable6 bool) *TunManager {
	return &TunManager{option: option, enable6: enable6}
}

func (t *TunManager) SetTunSupport6(support6 bool) error {
	if support6 != t.enable6 {
		if support6 {
			link, err := netlink.LinkByName(t.option.Name)
			if err != nil {
				return fmt.Errorf("failed to get interface %s: %v", t.option.Name, err)
			}
			for _, route := range t.option.Route6 {
				r := &netlink.Route{
					LinkIndex: link.Attrs().Index,
					Dst: &net.IPNet{
						IP:   route.Addr().AsSlice(),
						Mask: net.CIDRMask(route.Bits(), route.Addr().BitLen()),
					},
					Src: t.option.Ip6.Addr().AsSlice(),
				}
				if err := netlink.RouteAdd(r); err != nil {
					return fmt.Errorf("failed to add IPv6 route through TUN: %v", err)
				}
			}
		} else {
			// delete ipv6 routes
			if err := DeleteRoutes(t.option.Name, netlink.FAMILY_V6); err != nil {
				return fmt.Errorf("failed to delete IPv6 routes: %v", err)
			}
		}
	}
	t.enable6 = support6
	return nil
}

func (t *TunManager) Start() error {
	return t.SetTun(t.enable6)
}

func (t *TunManager) Close() error {
	return DeleteRoutes(t.option.Name, netlink.FAMILY_ALL)
}

func (t *TunManager) SetTun(enable6 bool) error {
	link, err := netlink.LinkByName(t.option.Name)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", t.option.Name, err)
	}

	// set ip address
	if t.option.Ip4.IsValid() {
		if err := SetIPAddress(link, t.option.Ip4); err != nil {
			return fmt.Errorf("failed to set IPv4 address: %v", err)
		}
	}
	if t.option.Ip6.IsValid() {
		if err := SetIPAddress(link, t.option.Ip6); err != nil {
			return fmt.Errorf("failed to set IPv6 address: %v", err)
		}
	}
	// bring interface up
	if err := BringInterfaceUp(t.option.Name); err != nil {
		return fmt.Errorf("failed to bring interface up: %v", err)
	}
	// set dns
	if len(t.option.Dns6) > 0 || len(t.option.Dns4) > 0 {
		if err := SetDNSForInterface(t.option.Name,
			append(t.option.Dns4, t.option.Dns6...)); err != nil {
			return fmt.Errorf("failed to set DNS servers: %v", err)
		}
	}
	for _, route := range t.option.Route4 {
		r := &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst: &net.IPNet{
				IP:   route.Addr().AsSlice(),
				Mask: net.CIDRMask(route.Bits(), route.Addr().BitLen()),
			},
			Src: t.option.Ip4.Addr().AsSlice(),
		}
		if err := netlink.RouteAdd(r); err != nil {
			return fmt.Errorf("failed to add IPv4 route through TUN: %v", err)
		}
	}
	if enable6 {
		for _, route := range t.option.Route6 {
			r := &netlink.Route{
				LinkIndex: link.Attrs().Index,
				Dst: &net.IPNet{
					IP:   route.Addr().AsSlice(),
					Mask: net.CIDRMask(route.Bits(), route.Addr().BitLen()),
				},
			}
			if err := netlink.RouteAdd(r); err != nil {
				return fmt.Errorf("failed to add IPv6 default route through TUN: %v", err)
			}
		}
	}
	return nil
}

// NetworkConfig represents the network configuration for an interface
type NetworkConfig struct {
	InterfaceName string
	IP4           netip.Prefix
	IP6           netip.Prefix
	Gateway4      netip.Addr
	Gateway6      netip.Addr
	DNS4          []netip.Addr
	DNS6          []netip.Addr
	Routes4       []netip.Prefix
	Routes6       []netip.Prefix
	MTU           int
}

// SetIPAddress sets an IP address for a network interface
func SetIPAddress(link netlink.Link, ipAddr netip.Prefix) error {
	// Parse the IP address
	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ipAddr.Addr().AsSlice(),
			Mask: net.CIDRMask(ipAddr.Bits(), ipAddr.Addr().BitLen()),
		},
	}

	// Add the IP address to the interface
	if err := netlink.AddrAdd(link, addr); err != nil {
		if strings.Contains(err.Error(), "file exists") {
			return nil
		}
		return fmt.Errorf("failed to add IP address %s to interface %s: %v", ipAddr,
			link.Attrs().Name, err)
	}

	log.Info().Str("interface", link.Attrs().Name).Str("ip", ipAddr.String()).Msg("IP address set successfully")
	return nil
}

// RemoveIPAddress removes an IP address from a network interface
func RemoveIPAddress(ifaceName string, ipAddr netip.Prefix) error {
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", ifaceName, err)
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ipAddr.Addr().AsSlice(),
			Mask: net.CIDRMask(ipAddr.Bits(), ipAddr.Addr().BitLen()),
		},
	}

	if err := netlink.AddrDel(link, addr); err != nil {
		return fmt.Errorf("failed to remove IP address %s from interface %s: %v", ipAddr, ifaceName, err)
	}

	log.Info().Str("interface", ifaceName).Str("ip", ipAddr.String()).Msg("IP address removed successfully")
	return nil
}

// SetGateway sets the default gateway for an interface
func SetGateway(ifaceName string, gateway netip.Addr) error {
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", ifaceName, err)
	}

	// Create default route (0.0.0.0/0 for IPv4 or ::/0 for IPv6)
	var dst *net.IPNet
	if gateway.Is4() {
		dst = &net.IPNet{IP: net.IPv4zero, Mask: net.IPv4Mask(0, 0, 0, 0)}
	} else {
		dst = &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
		Gw:        gateway.AsSlice(),
	}

	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("failed to add gateway route %s for interface %s: %v", gateway, ifaceName, err)
	}

	log.Info().Str("interface", ifaceName).Str("gateway", gateway.String()).Msg("Gateway set successfully")
	return nil
}

// AddRoute adds a specific route to the routing table
func AddRoute(ifaceName string, destination netip.Prefix, gateway netip.Addr) error {
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", ifaceName, err)
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst: &net.IPNet{
			IP:   destination.Addr().AsSlice(),
			Mask: net.CIDRMask(destination.Bits(), destination.Addr().BitLen()),
		},
		Gw: gateway.AsSlice(),
	}

	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("failed to add route %s via %s for interface %s: %v", destination, gateway, ifaceName, err)
	}

	log.Info().Str("interface", ifaceName).Str("destination", destination.String()).Str("gateway", gateway.String()).Msg("Route added successfully")
	return nil
}

// RemoveRoute removes a specific route from the routing table
func RemoveRoute(ifaceName string, destination netip.Prefix, gateway netip.Addr) error {
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", ifaceName, err)
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst: &net.IPNet{
			IP:   destination.Addr().AsSlice(),
			Mask: net.CIDRMask(destination.Bits(), destination.Addr().BitLen()),
		},
		Gw: gateway.AsSlice(),
	}

	if err := netlink.RouteDel(route); err != nil {
		return fmt.Errorf("failed to remove route %s via %s for interface %s: %v", destination, gateway, ifaceName, err)
	}

	log.Info().Str("interface", ifaceName).Str("destination", destination.String()).Str("gateway", gateway.String()).Msg("Route removed successfully")
	return nil
}

// SetDNS sets DNS servers for the system
func SetDNS(dnsServers []netip.Addr) error {
	var resolvConf strings.Builder

	// Add search domains if needed
	// resolvConf.WriteString("search example.com\n")

	// Add DNS servers
	for _, dns := range dnsServers {
		resolvConf.WriteString(fmt.Sprintf("nameserver %s\n", dns.String()))
	}

	// Write to /etc/resolv.conf
	if err := os.WriteFile("/etc/resolv.conf", []byte(resolvConf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write DNS configuration to /etc/resolv.conf: %v", err)
	}

	log.Info().Any("dns_servers", dnsServers).Msg("DNS servers configured successfully")
	return nil
}

// SetDNSForInterface sets DNS servers for a specific interface using systemd-resolved
func SetDNSForInterface(ifaceName string, dnsServers []netip.Addr) error {
	// Try using systemd-resolved first
	if err := setDNSWithSystemdResolved(ifaceName, dnsServers); err == nil {
		return nil
	} else {
		log.Debug().Err(err).Msg("setDNSWithSystemdResolved")
	}

	// Fallback to NetworkManager
	if err := setDNSWithNetworkManager(ifaceName, dnsServers); err == nil {
		return nil
	} else {
		log.Debug().Err(err).Msg("setDNSWithNetworkManager")
	}

	// Final fallback to /etc/resolv.conf
	return SetDNS(dnsServers)
}

// setDNSWithSystemdResolved sets DNS using systemd-resolved
func setDNSWithSystemdResolved(ifaceName string, dnsServers []netip.Addr) error {
	// Build resolvectl command
	args := []string{"dns", ifaceName}
	for _, dns := range dnsServers {
		args = append(args, dns.String())
	}

	cmd := exec.Command("resolvectl", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("resolvectl failed: %v", err)
	}

	log.Info().Str("interface", ifaceName).Any("dns_servers", dnsServers).Msg("DNS servers set via systemd-resolved")
	return nil
}

// setDNSWithNetworkManager sets DNS using NetworkManager
func setDNSWithNetworkManager(ifaceName string, dnsServers []netip.Addr) error {
	var dns4 []string
	for _, dns := range dnsServers {
		if dns.Is4() {
			dns4 = append(dns4, dns.String())
		}
	}
	cmd := exec.Command("nmcli", "con", "mod", ifaceName, "ipv4.dns", strings.Join(dns4, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nmcli connection modify failed: %v", err)
	}
	// prevent DHCP or RA from overriding it:
	cmd = exec.Command("nmcli", "con", "mod", ifaceName, "ipv4.ignore-auto-dns", "yes")
	if err := cmd.Run(); err != nil {
		log.Debug().Msgf("nmcli connection modify ipv4.ignore-auto-dns failed: %v", err)
	}
	// make tun dns has higher priority than others:
	cmd = exec.Command("nmcli", "con", "mod", ifaceName, "ipv4.dns-priority", "-1")
	if err := cmd.Run(); err != nil {
		log.Debug().Msgf("nmcli connection modify ipv4.dns-priority failed: %v", err)
	}

	var dns6 []string
	for _, dns := range dnsServers {
		if dns.Is6() {
			dns6 = append(dns6, dns.String())
		}
	}
	// without this, the following set will fail with
	// Error: Failed to modify connection 'tun0': ipv6.method: method 'manual' requires at least an address or a route
	cmd = exec.Command("nmcli", "con", "mod", ifaceName, "ipv6.method", "auto")
	if err := cmd.Run(); err != nil {
		log.Debug().Msgf("nmcli connection modify ipv6.method failed: %v", err)
	}
	cmd = exec.Command("nmcli", "con", "mod", ifaceName, "ipv6.dns", strings.Join(dns6, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nmcli connection modify ipv6.dns failed: %v", err)
	}
	cmd = exec.Command("nmcli", "con", "mod", ifaceName, "ipv6.ignore-auto-dns", "yes")
	if err := cmd.Run(); err != nil {
		log.Debug().Msgf("nmcli connection modify ipv6.ignore-auto-dns failed: %v", err)
	}
	cmd = exec.Command("nmcli", "con", "mod", ifaceName, "ipv6.dns-priority", "-1")
	if err := cmd.Run(); err != nil {
		log.Debug().Msgf("nmcli connection modify ipv6.dns-priority failed: %v", err)
	}

	// Restart the connection to apply changes
	cmd = exec.Command("nmcli", "con", "up", ifaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nmcli connection up failed: %v", err)
	}

	log.Info().Str("interface", ifaceName).Any("dns_servers", dnsServers).Msg("DNS servers set via NetworkManager")
	return nil
}

// SetMTU sets the MTU for a network interface
func SetMTU(ifaceName string, mtu int) error {
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", ifaceName, err)
	}

	if err := netlink.LinkSetMTU(link, mtu); err != nil {
		return fmt.Errorf("failed to set MTU %d for interface %s: %v", mtu, ifaceName, err)
	}

	log.Info().Str("interface", ifaceName).Int("mtu", mtu).Msg("MTU set successfully")
	return nil
}

// BringInterfaceUp brings a network interface up
func BringInterfaceUp(ifaceName string) error {
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", ifaceName, err)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to bring interface %s up: %v", ifaceName, err)
	}

	log.Info().Str("interface", ifaceName).Msg("Interface brought up successfully")
	return nil
}

// BringInterfaceDown brings a network interface down
func BringInterfaceDown(ifaceName string) error {
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", ifaceName, err)
	}

	if err := netlink.LinkSetDown(link); err != nil {
		return fmt.Errorf("failed to bring interface %s down: %v", ifaceName, err)
	}

	log.Info().Str("interface", ifaceName).Msg("Interface brought down successfully")
	return nil
}

// setInterfaceMetric sets the routing metric for an interface
func setInterfaceMetric(ifaceName string, metric int) error {
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", ifaceName, err)
	}

	// Set the interface metric
	if err := netlink.LinkSetTxQLen(link, 0); err != nil {
		return fmt.Errorf("failed to set interface metric: %v", err)
	}

	// Update routes with new metric
	routes, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("failed to list routes for interface: %v", err)
	}

	for _, route := range routes {
		route.Priority = metric
		if err := netlink.RouteReplace(&route); err != nil {
			log.Warn().Err(err).Msg("Failed to update route metric")
		}
	}

	return nil
}

// RestoreOriginalRouting restores the original routing configuration
func DeleteRoutes(deviceName string, family int) error {
	// Remove TUN routes
	tunRoutes, err := netlink.RouteList(nil, family)
	if err != nil {
		return fmt.Errorf("failed to list routes: %v", err)
	}

	for _, route := range tunRoutes {
		link, err := netlink.LinkByIndex(route.LinkIndex)
		if err != nil {
			continue
		}

		// Remove routes from TUN interfaces
		if link.Attrs().Name == deviceName {
			if err := netlink.RouteDel(&route); err != nil {
				log.Warn().Err(err).Msg("Failed to remove TUN route")
			}
		}
	}

	log.Info().Msgf("routes deleted for device %s", deviceName)
	return nil
}

// DeleteTunDevice completely removes a TUN device from the system
// Note: The kernel automatically removes all routes and IP addresses when the device is deleted
func DeleteTunDevice(deviceName string) error {
	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %v", deviceName, err)
	}

	// Delete the interface (kernel will automatically clean up routes and addresses)
	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to delete interface %s: %v", deviceName, err)
	}

	log.Info().Str("device", deviceName).Msg("TUN device deleted successfully")
	return nil
}

package sshhelper

import (
	"fmt"
	"strings"
)

// Firewall management for SSH Client
// This package provides methods to manage firewalls on remote servers via SSH.
//
// Supported firewall types:
// - UFW (Uncomplicated Firewall) - Ubuntu/Debian default
// - firewalld - RHEL/CentOS/Fedora default
// - iptables - Legacy, universal
// - nftables - Modern replacement for iptables
//
// Example usage in server preparation:
//
//	client, err := sshhelper.Dial(config)
//	if err != nil {
//		return err
//	}
//	defer client.Close()
//
//	// Enable firewall
//	if err := client.EnableFirewall(); err != nil {
//		log.Warn().Err(err).Msg("failed to enable firewall")
//	}
//
//	// Open SSH port
//	if err := client.OpenPort(server.SshPort, "tcp"); err != nil {
//		return err
//	}
//
//	// Open inbound ports from ServerConfig
//	for _, inbound := range xConfig.Config.Inbounds {
//		if inbound.Port != 0 {
//			client.OpenPort(inbound.Port, "tcp")
//			client.OpenPort(inbound.Port, "udp")
//		}
//		for _, port := range inbound.Ports {
//			client.OpenPort(port, "tcp")
//			client.OpenPort(port, "udp")
//		}
//	}
//
//	// When config changes, delete old port rules:
//	client.DeletePortRule(oldPort, "tcp")
//	client.DeletePortRule(oldPort, "udp")

// FirewallType represents the type of firewall system
type FirewallType string

const (
	FirewallUFW       FirewallType = "ufw"
	FirewallFirewalld FirewallType = "firewalld"
	FirewallIptables  FirewallType = "iptables"
	FirewallNftables  FirewallType = "nftables"
	FirewallUnknown   FirewallType = "unknown"
)

// DetectFirewall determines which firewall system is active on the remote server
func (c *Client) DetectFirewall() (FirewallType, error) {
	// Check UFW first (most user-friendly, default on Ubuntu/Debian)
	if exists, _ := c.CommandExists("ufw", true); exists {
		output, _ := c.Output("ufw status", true)
		if strings.Contains(output, "Status: active") || strings.Contains(output, "Status: inactive") {
			return FirewallUFW, nil
		}
	}

	// Check firewalld (default on RHEL/CentOS/Fedora)
	output, err := c.Output("systemctl is-active firewalld", false)
	if err == nil && strings.TrimSpace(output) == "active" {
		return FirewallFirewalld, nil
	}

	// Check if firewalld is installed but inactive
	if exists, _ := c.CommandExists("firewall-cmd", true); exists {
		return FirewallFirewalld, nil
	}

	// Check nftables
	output, err = c.Output("nft list ruleset", true)
	if err == nil && len(strings.TrimSpace(output)) > 0 {
		return FirewallNftables, nil
	}

	// Fallback to iptables if it exists
	if exists, _ := c.CommandExists("iptables", true); exists {
		return FirewallIptables, nil
	}

	return FirewallUnknown, fmt.Errorf("no firewall detected")
}

// OpenPort allows inbound traffic to a specific port with the given protocol (tcp/udp)
// This operation is idempotent - it checks if the rule already exists before adding
func (c *Client) OpenPort(port uint32, protocol string) error {
	firewallType, err := c.DetectFirewall()
	if err != nil {
		return fmt.Errorf("failed to detect firewall: %w", err)
	}

	switch firewallType {
	case FirewallUFW:
		// Check if rule already exists
		output, _ := c.Output("ufw status numbered", true)
		rulePattern := fmt.Sprintf("%d/%s", port, protocol)
		if strings.Contains(output, rulePattern) {
			// Rule already exists, skip
			return nil
		}

		// Add the rule
		cmd := fmt.Sprintf("ufw allow %d/%s", port, protocol)
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to open port %d/%s with ufw: %w", port, protocol, err)
		}
		return nil

	case FirewallFirewalld:
		// Check if rule already exists
		output, _ := c.Output("firewall-cmd --list-ports", true)
		rulePattern := fmt.Sprintf("%d/%s", port, protocol)
		if strings.Contains(output, rulePattern) {
			// Rule already exists, skip
			return nil
		}

		// Add the rule permanently and reload
		cmd := fmt.Sprintf("firewall-cmd --add-port=%d/%s --permanent && firewall-cmd --reload", port, protocol)
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to open port %d/%s with firewalld: %w", port, protocol, err)
		}
		return nil

	case FirewallIptables:
		// Check if rule already exists
		output, _ := c.Output(fmt.Sprintf("iptables -L INPUT -n | grep -E 'dpt:%d.*%s'", port, strings.ToUpper(protocol)), true)
		if strings.TrimSpace(output) != "" {
			// Rule already exists, skip
			return nil
		}

		// Add the rule and save
		cmd := fmt.Sprintf("iptables -A INPUT -p %s --dport %d -j ACCEPT && iptables-save > /etc/iptables/rules.v4", protocol, port)
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to open port %d/%s with iptables: %w", port, protocol, err)
		}
		return nil

	case FirewallNftables:
		// Add the rule (nftables doesn't have simple duplication check)
		cmd := fmt.Sprintf("nft add rule inet filter INPUT %s dport %d ct state new,established counter accept", protocol, port)
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to open port %d/%s with nftables: %w", port, protocol, err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported firewall type: %s", firewallType)
	}
}

// ClosePort denies/blocks inbound traffic to a specific port with the given protocol (tcp/udp)
func (c *Client) ClosePort(port uint32, protocol string) error {
	firewallType, err := c.DetectFirewall()
	if err != nil {
		return fmt.Errorf("failed to detect firewall: %w", err)
	}

	switch firewallType {
	case FirewallUFW:
		cmd := fmt.Sprintf("ufw deny %d/%s", port, protocol)
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to close port %d/%s with ufw: %w", port, protocol, err)
		}
		return nil

	case FirewallFirewalld:
		cmd := fmt.Sprintf("firewall-cmd --remove-port=%d/%s --permanent && firewall-cmd --reload", port, protocol)
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to close port %d/%s with firewalld: %w", port, protocol, err)
		}
		return nil

	case FirewallIptables:
		cmd := fmt.Sprintf("iptables -A INPUT -p %s --dport %d -j DROP && iptables-save > /etc/iptables/rules.v4", protocol, port)
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to close port %d/%s with iptables: %w", port, protocol, err)
		}
		return nil

	case FirewallNftables:
		cmd := fmt.Sprintf("nft add rule inet filter INPUT %s dport %d counter drop", protocol, port)
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to close port %d/%s with nftables: %w", port, protocol, err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported firewall type: %s", firewallType)
	}
}

// DeletePortRule removes a firewall rule for a specific port
// This operation is idempotent - it won't error if the rule doesn't exist
func (c *Client) DeletePortRule(port uint32, protocol string) error {
	firewallType, err := c.DetectFirewall()
	if err != nil {
		return fmt.Errorf("failed to detect firewall: %w", err)
	}

	switch firewallType {
	case FirewallUFW:
		// Check if rule exists first
		output, _ := c.Output("ufw status numbered", true)
		rulePattern := fmt.Sprintf("%d/%s", port, protocol)
		if !strings.Contains(output, rulePattern) {
			// Rule doesn't exist, nothing to delete
			return nil
		}

		// Delete both allow and deny rules (if they exist)
		c.Run(fmt.Sprintf("ufw delete allow %d/%s", port, protocol), true)
		c.Run(fmt.Sprintf("ufw delete deny %d/%s", port, protocol), true)
		return nil

	case FirewallFirewalld:
		// Check if rule exists first
		output, _ := c.Output("firewall-cmd --list-ports", true)
		rulePattern := fmt.Sprintf("%d/%s", port, protocol)
		if !strings.Contains(output, rulePattern) {
			// Rule doesn't exist, nothing to delete
			return nil
		}

		cmd := fmt.Sprintf("firewall-cmd --remove-port=%d/%s --permanent && firewall-cmd --reload", port, protocol)
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to delete port rule %d/%s with firewalld: %w", port, protocol, err)
		}
		return nil

	case FirewallIptables:
		// Try to delete the rule (ignore errors if it doesn't exist)
		cmd := fmt.Sprintf("iptables -D INPUT -p %s --dport %d -j ACCEPT 2>/dev/null || true", protocol, port)
		c.Run(cmd, true)
		cmd = fmt.Sprintf("iptables -D INPUT -p %s --dport %d -j DROP 2>/dev/null || true", protocol, port)
		c.Run(cmd, true)
		// Save the changes
		c.Run("iptables-save > /etc/iptables/rules.v4", true)
		return nil

	case FirewallNftables:
		// nftables requires handle-based deletion, which is more complex
		// For simplicity, we'll skip automatic deletion for nftables
		// Users would need to manually manage nftables rules or flush/reload config
		return fmt.Errorf("nftables rule deletion requires manual handle-based deletion")

	default:
		return fmt.Errorf("unsupported firewall type: %s", firewallType)
	}
}

// OpenPorts opens multiple ports for the given protocol (tcp/udp)
// This is a bulk operation that calls OpenPort for each port
func (c *Client) OpenPorts(ports []uint32, protocol string) error {
	var errors []string
	for _, port := range ports {
		if err := c.OpenPort(port, protocol); err != nil {
			errors = append(errors, fmt.Sprintf("port %d: %v", port, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to open some ports: %s", strings.Join(errors, "; "))
	}
	return nil
}

// ClosePorts closes multiple ports for the given protocol (tcp/udp)
// This is a bulk operation that calls ClosePort for each port
func (c *Client) ClosePorts(ports []uint32, protocol string) error {
	var errors []string
	for _, port := range ports {
		if err := c.ClosePort(port, protocol); err != nil {
			errors = append(errors, fmt.Sprintf("port %d: %v", port, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to close some ports: %s", strings.Join(errors, "; "))
	}
	return nil
}

// DeletePortRules deletes firewall rules for multiple ports
// This is a bulk operation that calls DeletePortRule for each port
func (c *Client) DeletePortRules(ports []uint32, protocol string) error {
	var errors []string
	for _, port := range ports {
		if err := c.DeletePortRule(port, protocol); err != nil {
			errors = append(errors, fmt.Sprintf("port %d: %v", port, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete some port rules: %s", strings.Join(errors, "; "))
	}
	return nil
}

// EnableFirewall enables the firewall service
func (c *Client) EnableFirewall() error {
	firewallType, err := c.DetectFirewall()
	if err != nil {
		return fmt.Errorf("failed to detect firewall: %w", err)
	}

	switch firewallType {
	case FirewallUFW:
		// Check if already enabled
		output, _ := c.Output("ufw status", true)
		if strings.Contains(output, "Status: active") {
			// Already enabled, skip
			return nil
		}

		// Enable UFW (--force to avoid interactive prompt)
		if err := c.Run("ufw --force enable", true); err != nil {
			return fmt.Errorf("failed to enable ufw: %w", err)
		}
		return nil

	case FirewallFirewalld:
		if err := c.Run("systemctl enable --now firewalld", true); err != nil {
			return fmt.Errorf("failed to enable firewalld: %w", err)
		}
		return nil

	case FirewallNftables:
		if err := c.Run("systemctl enable --now nftables", true); err != nil {
			return fmt.Errorf("failed to enable nftables: %w", err)
		}
		return nil

	case FirewallIptables:
		if err := c.Run("systemctl enable --now iptables", true); err != nil {
			return fmt.Errorf("failed to enable iptables: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported firewall type: %s", firewallType)
	}
}

// DisableFirewall disables the firewall service
func (c *Client) DisableFirewall() error {
	firewallType, err := c.DetectFirewall()
	if err != nil {
		return fmt.Errorf("failed to detect firewall: %w", err)
	}

	switch firewallType {
	case FirewallUFW:
		if err := c.Run("ufw disable", true); err != nil {
			return fmt.Errorf("failed to disable ufw: %w", err)
		}
		return nil

	case FirewallFirewalld:
		if err := c.Run("systemctl disable --now firewalld", true); err != nil {
			return fmt.Errorf("failed to disable firewalld: %w", err)
		}
		return nil

	case FirewallNftables:
		if err := c.Run("systemctl disable --now nftables", true); err != nil {
			return fmt.Errorf("failed to disable nftables: %w", err)
		}
		return nil

	case FirewallIptables:
		if err := c.Run("systemctl disable --now iptables", true); err != nil {
			return fmt.Errorf("failed to disable iptables: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported firewall type: %s", firewallType)
	}
}

// GetFirewallStatus returns the firewall status and rules
func (c *Client) GetFirewallStatus() (string, error) {
	firewallType, err := c.DetectFirewall()
	if err != nil {
		return "", fmt.Errorf("failed to detect firewall: %w", err)
	}

	switch firewallType {
	case FirewallUFW:
		output, err := c.Output("ufw status verbose", true)
		if err != nil {
			return "", fmt.Errorf("failed to get ufw status: %w", err)
		}
		return output, nil

	case FirewallFirewalld:
		output, err := c.Output("firewall-cmd --list-all", true)
		if err != nil {
			return "", fmt.Errorf("failed to get firewalld status: %w", err)
		}
		return output, nil

	case FirewallIptables:
		output, err := c.Output("iptables -L -n -v", true)
		if err != nil {
			return "", fmt.Errorf("failed to get iptables status: %w", err)
		}
		return output, nil

	case FirewallNftables:
		output, err := c.Output("nft list ruleset", true)
		if err != nil {
			return "", fmt.Errorf("failed to get nftables status: %w", err)
		}
		return output, nil

	default:
		return "", fmt.Errorf("unsupported firewall type: %s", firewallType)
	}
}

// AllowAllInbound sets the firewall to allow all inbound connections
// WARNING: This is insecure and should only be used for testing
func (c *Client) AllowAllInbound() error {
	firewallType, err := c.DetectFirewall()
	if err != nil {
		return fmt.Errorf("failed to detect firewall: %w", err)
	}

	switch firewallType {
	case FirewallUFW:
		// Set default policy to allow incoming
		if err := c.Run("ufw default allow incoming", true); err != nil {
			return fmt.Errorf("failed to set ufw default allow: %w", err)
		}
		return nil

	case FirewallFirewalld:
		// Set default zone to trusted (allows all)
		cmd := "firewall-cmd --set-default-zone=trusted && firewall-cmd --reload"
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to set firewalld to trusted: %w", err)
		}
		return nil

	case FirewallIptables:
		// Set INPUT chain policy to ACCEPT and flush rules
		cmd := "iptables -P INPUT ACCEPT && iptables -F INPUT && iptables-save > /etc/iptables/rules.v4"
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to set iptables allow all: %w", err)
		}
		return nil

	case FirewallNftables:
		// Set input policy to accept
		cmd := "nft add table inet filter; nft add chain inet filter input { type filter hook input priority 0 \\; policy accept \\; }"
		if err := c.Run(cmd, true); err != nil {
			return fmt.Errorf("failed to set nftables allow all: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported firewall type: %s", firewallType)
	}
}

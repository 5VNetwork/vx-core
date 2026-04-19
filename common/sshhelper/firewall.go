package sshhelper

import (
	"fmt"
	"strings"
)

// FirewallType represents the type of firewall system
type FirewallType string

const (
	FirewallUFW       FirewallType = "ufw"
	FirewallFirewalld FirewallType = "firewalld"
	FirewallIptables  FirewallType = "iptables"
	FirewallNftables  FirewallType = "nftables"
	FirewallUnknown   FirewallType = "unknown"
)

// netfilter-persistent (Debian/Ubuntu) layout; matches existing iptables IPv4 path.
const (
	iptablesRulesV4Path = "/etc/iptables/rules.v4"
	iptablesRulesV6Path = "/etc/iptables/rules.v6"
	// Loaded by nftables.service on Debian/Ubuntu; RHEL often uses /etc/sysconfig/nftables — write both when present.
	nftablesConfPath       = "/etc/nftables.conf"
	nftablesSysconfigPath  = "/etc/sysconfig/nftables"
	nftablesSysconfigPath2 = "/etc/sysconfig/nftables.conf"
)

// iptablesInputHasPortRule reports whether INPUT contains an ACCEPT rule matching dport/protocol.
// Uses `iptables -C` (the canonical existence check) rather than fragile grep on `-L` output.
func (c *Client) iptablesInputHasPortRule(iptablesBin string, port uint32, protocol string) bool {
	err := c.Run(fmt.Sprintf("%s -C INPUT -p %s --dport %d -j ACCEPT", iptablesBin, protocol, port), true)
	return err == nil
}

// nftRulesetLooksManagedByIptablesNft is true when nft's ruleset reflects the iptables-nft compatibility
// layer (manage with iptables/ip6tables), not a hand-written nft-only configuration.
func nftRulesetLooksManagedByIptablesNft(ruleset string) bool {
	s := strings.ToLower(ruleset)
	if strings.Contains(s, "managed by iptables-nft") {
		return true
	}
	// iptables-nft creates the legacy IPv4 "ip filter" table; typical raw-nft host configs use "inet" only.
	return strings.Contains(s, "table ip filter")
}

// flushIptablesInputPermissive sets INPUT policy to ACCEPT, flushes INPUT chain rules in the kernel,
// and writes the result to persistent rule files (IPv4 + IPv6 when ip6tables exists).
func (c *Client) flushIptablesInputPermissive() error {
	cmd := fmt.Sprintf("iptables -P INPUT ACCEPT && iptables -F INPUT && iptables-save > %s", iptablesRulesV4Path)
	if hasIp6, _ := c.CommandExists("ip6tables", true); hasIp6 {
		cmd += fmt.Sprintf(" && ip6tables -P INPUT ACCEPT && ip6tables -F INPUT && ip6tables-save > %s", iptablesRulesV6Path)
	}
	return c.ShRun(cmd)
}

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

	iptablesExists, _ := c.CommandExists("iptables", true)

	// Non-empty nft ruleset: distinguish raw nft from iptables-nft (Debian/Ubuntu default), which
	// shares the nf_tables backend but must be managed via iptables, not raw "nft add rule ...".
	output, err = c.Output("nft list ruleset", true)
	if err == nil && len(strings.TrimSpace(output)) > 0 {
		if iptablesExists && nftRulesetLooksManagedByIptablesNft(output) {
			return FirewallIptables, nil
		}
		return FirewallNftables, nil
	}

	if iptablesExists {
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
		if err := c.ShRun(cmd); err != nil {
			return fmt.Errorf("failed to open port %d/%s with firewalld: %w", port, protocol, err)
		}
		return nil

	case FirewallIptables:
		v4Done := c.iptablesInputHasPortRule("iptables", port, protocol)
		hasIp6tables, _ := c.CommandExists("ip6tables", true)
		v6Done := !hasIp6tables || c.iptablesInputHasPortRule("ip6tables", port, protocol)
		if v4Done && v6Done {
			return nil
		}

		var parts []string
		if !v4Done {
			parts = append(parts, fmt.Sprintf("iptables -A INPUT -p %s --dport %d -j ACCEPT && iptables-save > %s", protocol, port, iptablesRulesV4Path))
		}
		if hasIp6tables && !v6Done {
			parts = append(parts, fmt.Sprintf("ip6tables -A INPUT -p %s --dport %d -j ACCEPT && ip6tables-save > %s", protocol, port, iptablesRulesV6Path))
		}
		cmd := strings.Join(parts, " && ")
		if err := c.ShRun(cmd); err != nil {
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
		if err := c.ShRun(cmd); err != nil {
			return fmt.Errorf("failed to close port %d/%s with firewalld: %w", port, protocol, err)
		}
		return nil

	case FirewallIptables:
		cmd := fmt.Sprintf("iptables -A INPUT -p %s --dport %d -j DROP && iptables-save > %s", protocol, port, iptablesRulesV4Path)
		if hasIp6, _ := c.CommandExists("ip6tables", true); hasIp6 {
			cmd += fmt.Sprintf(" && ip6tables -A INPUT -p %s --dport %d -j DROP && ip6tables-save > %s", protocol, port, iptablesRulesV6Path)
		}
		if err := c.ShRun(cmd); err != nil {
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
		if err := c.ShRun(cmd); err != nil {
			return fmt.Errorf("failed to delete port rule %d/%s with firewalld: %w", port, protocol, err)
		}
		return nil

	case FirewallIptables:
		// Try to delete the rule (ignore errors if it doesn't exist)
		cmd := fmt.Sprintf("iptables -D INPUT -p %s --dport %d -j ACCEPT 2>/dev/null || true", protocol, port)
		c.Run(cmd, true)
		cmd = fmt.Sprintf("iptables -D INPUT -p %s --dport %d -j DROP 2>/dev/null || true", protocol, port)
		c.Run(cmd, true)
		c.Run(fmt.Sprintf("iptables-save > %s", iptablesRulesV4Path), true)
		if hasIp6, _ := c.CommandExists("ip6tables", true); hasIp6 {
			cmd = fmt.Sprintf("ip6tables -D INPUT -p %s --dport %d -j ACCEPT 2>/dev/null || true", protocol, port)
			c.Run(cmd, true)
			cmd = fmt.Sprintf("ip6tables -D INPUT -p %s --dport %d -j DROP 2>/dev/null || true", protocol, port)
			c.Run(cmd, true)
			c.Run(fmt.Sprintf("ip6tables-save > %s", iptablesRulesV6Path), true)
		}
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

// DisableFirewall disables the firewall service or allow inbound traffic so that incoming connections
// are allowed.
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
		return c.AllowAllInbound()
	case FirewallIptables:
		return c.AllowAllInbound()
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
		if hasIp6, _ := c.CommandExists("ip6tables", true); hasIp6 {
			out6, err := c.Output("ip6tables -L -n -v", true)
			if err != nil {
				return "", fmt.Errorf("failed to get ip6tables status: %w", err)
			}
			output += "\n--- ip6tables ---\n" + out6
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
// The setting is persistent
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
		if err := c.ShRun(cmd); err != nil {
			return fmt.Errorf("failed to set firewalld to trusted: %w", err)
		}
		return nil

	case FirewallIptables:
		if err := c.flushIptablesInputPermissive(); err != nil {
			return fmt.Errorf("failed to set iptables allow all: %w", err)
		}
		return nil

	case FirewallNftables:
		// Flush any pre-existing drop rules so nothing in nft still blocks inbound, rebuild a permissive
		// inet filter/input chain, then save the ruleset so nftables.service restores it after reboot.
		// NOTE: this also removes rules installed by Docker/Kubernetes/fail2ban; those services usually
		// recreate their rules when restarted.
		cmd := fmt.Sprintf(
			"nft flush ruleset && "+
				"nft add table inet filter && "+
				"nft add chain inet filter input { type filter hook input priority 0 \\; policy accept \\; } && "+
				"nft list ruleset > %s && "+
				"(test -d /etc/sysconfig && nft list ruleset > %s 2>/dev/null || true) && "+
				"(test -d /etc/sysconfig && nft list ruleset > %s 2>/dev/null || true) && "+
				"(systemctl enable nftables 2>/dev/null || true)",
			nftablesConfPath, nftablesSysconfigPath, nftablesSysconfigPath2)
		if err := c.ShRun(cmd); err != nil {
			return fmt.Errorf("failed to set nftables allow all: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported firewall type: %s", firewallType)
	}
}

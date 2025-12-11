package sshhelper

import (
	"fmt"
)

func (c *Client) EnableBbr() error {
	var fileToModify string
	if yes, err := c.FileExisted("/etc/sysctl.conf"); err == nil && yes {
		fileToModify = "/etc/sysctl.conf"

		hasQdisc := false
		hasBbr := false

		// Check for existing configuration
		output, _ := c.Output("grep -q 'net.core.default_qdisc=fq' /etc/sysctl.conf && echo found", true)
		if output == "found\n" {
			hasQdisc = true
		}

		output, _ = c.Output("grep -q 'net.ipv4.tcp_congestion_control=bbr' /etc/sysctl.conf && echo found", true)
		if output == "found\n" {
			hasBbr = true
		}

		// Only append if not already present
		if !hasQdisc {
			if err := c.AppendToFile(fileToModify, "net.core.default_qdisc=fq", true); err != nil {
				return fmt.Errorf("failed to append to %s: %w", fileToModify, err)
			}
		}

		if !hasBbr {
			if err := c.AppendToFile(fileToModify, "net.ipv4.tcp_congestion_control=bbr", true); err != nil {
				return fmt.Errorf("failed to append to %s: %w", fileToModify, err)
			}
		}

		// reload sysctl only if we made changes
		if !hasQdisc || !hasBbr {
			if err := c.Run("sysctl -p", true); err != nil {
				return fmt.Errorf("failed to reload sysctl: %w", err)
			}
		}
	} else {
		fileToModify = "/etc/sysctl.d/99-sysctl.conf"

		// Check if BBR is already configured
		hasQdisc := false
		hasBbr := false

		// Check for existing configuration
		output, _ := c.Output(fmt.Sprintf("grep -q 'net.core.default_qdisc=fq' %s && echo found", fileToModify), true)
		if output == "found\n" {
			hasQdisc = true
		}

		output, _ = c.Output(fmt.Sprintf("grep -q 'net.ipv4.tcp_congestion_control=bbr' %s && echo found", fileToModify), true)
		if output == "found\n" {
			hasBbr = true
		}

		// Only append if not already present
		if !hasQdisc {
			if err := c.AppendToFile(fileToModify, "net.core.default_qdisc=fq", true); err != nil {
				return fmt.Errorf("failed to append to %s: %w", fileToModify, err)
			}
		}

		if !hasBbr {
			if err := c.AppendToFile(fileToModify, "net.ipv4.tcp_congestion_control=bbr", true); err != nil {
				return fmt.Errorf("failed to append to %s: %w", fileToModify, err)
			}
		}

		// reload sysctl only if we made changes
		if !hasQdisc || !hasBbr {
			if err := c.Run("sysctl -p /etc/sysctl.d/99-sysctl.conf", true); err != nil {
				return fmt.Errorf("failed to reload sysctl: %w", err)
			}
		}
	}

	// check bbr
	output, err := c.Output("sysctl net.ipv4.tcp_congestion_control", true)
	if err != nil {
		return fmt.Errorf("failed to check bbr: %w", err)
	}
	if output != "net.ipv4.tcp_congestion_control = bbr\n" {
		return fmt.Errorf("bbr check returned %s", output)
	}
	return nil
}

package sshhelper

import (
	"fmt"
	"strconv"
	"strings"
)

type ServiceStatus struct {
	Running bool
	StartAt string  // unix timestamp
	Memory  float32 // MB
}

func (c *Client) ServiceStatus(name string) (*ServiceStatus, error) {
	status := &ServiceStatus{}

	// Check if service is active/running
	isActiveOutput, err := c.Output(fmt.Sprintf("systemctl is-active %s", name), true)
	if err != nil {
		// Service might not be running, but that's not necessarily an error
		status.Running = false
	} else {
		status.Running = strings.TrimSpace(isActiveOutput) == "active"
	}

	if !status.Running {
		return status, nil
	}

	// Get service start time (unix timestamp)
	// Using systemctl show to get ActiveEnterTimestamp
	startTimeOutput, err := c.Output(fmt.Sprintf("systemctl show %s --property=ActiveEnterTimestampMonotonic --value", name), true)
	if err == nil && strings.TrimSpace(startTimeOutput) != "" && strings.TrimSpace(startTimeOutput) != "0" {
		// Get the actual unix timestamp
		timestampOutput, err := c.Output(fmt.Sprintf("systemctl show %s --property=ActiveEnterTimestamp --value | xargs -I{} date '+%%s' -d '{}'", name), true)
		if err == nil {
			status.StartAt = timestampOutput
		}
	}

	// Get memory usage in MB
	// Using systemctl's cgroup-based memory accounting (MemoryCurrent)
	// This matches the Memory field shown in systemctl status
	memOutput, err := c.Output(fmt.Sprintf("systemctl show %s --property=MemoryCurrent --value", name), true)
	if err == nil {
		memBytes := strings.TrimSpace(memOutput)
		// MemoryCurrent returns bytes, or "[not set]" if not available
		if memBytes != "" && memBytes != "[not set]" {
			memBytesInt, err := strconv.ParseUint(memBytes, 10, 64)
			if err == nil {
				// Convert bytes to MB
				status.Memory = float32(memBytesInt) / 1024.0 / 1024.0
			}
		}
	}

	return status, nil
}

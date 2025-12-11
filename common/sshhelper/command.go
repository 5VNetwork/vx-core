package sshhelper

import (
	"fmt"
	"strings"
)

func (c *Client) CommandExists(command string, sudo bool) (bool, error) {
	// Method 1: Using 'which' command
	output, err := c.CombinedOutput(fmt.Sprintf("which %s", command), sudo)
	if err == nil && output != "" {
		return true, nil
	}

	// Method 2: Using 'command -v' (more portable than which)
	output, err = c.CombinedOutput(fmt.Sprintf("command -v %s", command), sudo)
	if err == nil && output != "" {
		return true, nil
	}

	// Method 3: Using 'type' command
	output, err = c.CombinedOutput(fmt.Sprintf("type %s", command), sudo)
	if err == nil && !strings.Contains(output, "not found") {
		return true, nil
	}

	return false, nil
}

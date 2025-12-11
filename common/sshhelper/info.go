package sshhelper

func (c *Client) GetServerOS() (string, error) {
	// Try /etc/os-release first (modern standard)
	output, err := c.CombinedOutput("grep '^PRETTY_NAME=' /etc/os-release | cut -d'=' -f2 | tr -d '\"'", false)
	if err == nil && output != "" {
		return output, nil
	}

	return c.CombinedOutput("uname -s", false)
}

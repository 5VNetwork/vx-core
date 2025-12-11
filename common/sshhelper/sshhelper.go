package sshhelper

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// getSudoCmd changes cmd to a sudo cmd if password is not empty
func (c *Client) getSudoCmd(cmd string) string {
	if c.Password == "" {
		return "sudo " + cmd
	}
	return fmt.Sprintf("echo '%s' | sudo -S %s", c.Password, cmd)
}

// Returns output and error
func (c *Client) CombinedOutput(command string, sudo bool) (string, error) {
	session, err := c.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	if sudo {
		command = c.getSudoCmd(command)
	}

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("failed to run command: %w", err)
	}

	return string(output), nil
}

// Returns error
func (c *Client) Output(command string, sudo bool) (string, error) {
	session, err := c.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	if sudo {
		command = c.getSudoCmd(command)
	}

	output, err := session.Output(command)
	if err != nil {
		return string(output), fmt.Errorf("failed to run command: %w", err)
	}

	return string(output), nil
}

// client.CombinedOutput("sh -c 'curl -fsSL https://get.docker.com | sh'", true)
func (c *Client) ShRun(command string, sudo bool) error {
	command = fmt.Sprintf("sh -c '%s'", command)

	session, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	if sudo {
		command = c.getSudoCmd(command)
	}

	return session.Run(command)
}

func (c *Client) Run(command string, sudo bool) error {
	session, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	if sudo {
		command = c.getSudoCmd(command)
	}

	return session.Run(command)
}

// Returns error
func (c *Client) StdOutAndStdErr(command string, sudo bool) (string, string, error) {
	session, err := c.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	go io.Copy(&stdoutBuf, stdout)
	go io.Copy(&stderrBuf, stderr)

	if sudo {
		command = c.getSudoCmd(command)
	}

	err = session.Start(command)
	if err != nil {
		return "", "", fmt.Errorf("failed to start command: %w", err)
	}

	if err = session.Wait(); err != nil {
		return "", "", fmt.Errorf("failed to wait for command: %w", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}

// download file from url to output using curl
// if path existed, it will be replaced by the new file
func (c *Client) Download(output, url string, sudo bool) error {
	session, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	command := fmt.Sprintf("curl -o %s %s", output, url)
	if sudo {
		command = c.getSudoCmd(command)
	}

	if err = session.Run(command); err != nil {
		return fmt.Errorf("failed to start bin/bash: %w", err)
	}

	return nil
}

func (c *Client) PackageManager() (string, error) {
	output, err := c.Output("type -P apt-get", false)
	if err == nil && output != "" {
		return "apt-get", nil
	}
	output, err = c.Output("type -P yum", false)
	if err == nil && output != "" {
		return "yum", nil
	}
	return "", err
}

func (c *Client) GetOSArch() (os string, arch string, err error) {
	session, err := c.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err = session.Start("/bin/bash"); err != nil {
		return "", "", fmt.Errorf("failed to start bin/bash: %w", err)
	}

	if _, err = stdin.Write([]byte("uname\n")); err != nil {
		return "", "", fmt.Errorf("failed to write [uname] command: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	if ok := scanner.Scan(); ok {
		os = scanner.Text()
	}

	if _, err = stdin.Write([]byte("uname -m\n")); err != nil {
		return "", "", fmt.Errorf("failed to write [uname -m] command: %w", err)
	}
	if ok := scanner.Scan(); ok {
		arch = scanner.Text()
	}

	return
}

func (c *Client) Shutdown() error {
	o, err := c.CombinedOutput("shutdown -h now", true)
	if err != nil {
		return fmt.Errorf("failed to shutdown: %w, output: %s", err, o)
	}
	return nil
}

func (c *Client) Suspend() error {
	o, err := c.CombinedOutput("systemctl suspend", true)
	if err != nil {
		return fmt.Errorf("failed to suspend: %w, output: %s", err, o)
	}
	return nil
}

func (c *Client) Reboot() error {
	o, err := c.CombinedOutput("reboot", true)
	if err != nil {
		return fmt.Errorf("failed to reboot: %w, output: %s", err, o)
	}
	return nil
}

package sshhelper

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

func (c *Client) FileExisted(path string) (bool, error) {
	session, err := c.NewSession()
	if err != nil {
		return false, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.Output(fmt.Sprintf("test -f %s && echo 'yes' || echo 'no'", path))
	if err != nil {
		return false, fmt.Errorf("failed to execute cmd: %w", err)
	}

	result := string(output)
	return result == "yes\n", nil
}

// copy file to tmp folder, then move to dst using sudo
func (c *Client) CopyFileToRemoteSftpSudo(src, dst string, permission uint32) error {
	tmpFile := "/tmp/" + time.Now().Format("20060102150405") + ".tmp"
	err := c.CopyFileToRemoteSftp(src, tmpFile, permission)
	if err != nil {
		return fmt.Errorf("failed to copy file to tmp folder: %w", err)
	}

	o, err := c.CombinedOutput(fmt.Sprintf("mv %s %s", tmpFile, dst), true)
	if err != nil {
		return fmt.Errorf("failed to move file: %w. Output: %s", err, o)
	}

	return nil
}

// src is the local file path, dst is the remote file path.
// permision is fs.FileMode
// if dst file already exists, it will be overwritten.
func (c *Client) CopyFileToRemoteSftp(src, dst string, permission uint32) error {
	client, err := c.GetSftpClient()
	if err != nil {
		return fmt.Errorf("failed to get sftp client: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open src file: %w", err)
	}
	defer srcFile.Close()

	dir := GetParentDir(dst)
	// if dir is not existed, create it
	existed, err := c.DirExist(dir)
	if err != nil {
		return fmt.Errorf("failed to check whether dir exist: %w", err)
	}
	if !existed {
		if err := client.MkdirAll(dir); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	dstFile, err := client.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create dst file: %w", err)
	}
	defer dstFile.Close()

	if _, err = dstFile.ReadFrom(srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if permission != 0 {
		if err = dstFile.Chmod(fs.FileMode(permission)); err != nil {
			return fmt.Errorf("failed to change file permission: %w", err)
		}
	}

	return nil
}

func (c *Client) CopyContentToRemoteSudo(src io.Reader, dst string, permission uint32) error {
	tmpPath := "/tmp/" + time.Now().Format("20060102150405") + ".tmp"
	err := c.CopyContentToRemote(src, tmpPath, permission)
	if err != nil {
		return fmt.Errorf("failed to copy content to remote: %w", err)
	}

	o, err := c.CombinedOutput(fmt.Sprintf("mv %s %s", tmpPath, dst), true)
	if err != nil {
		return fmt.Errorf("failed to move config.json: %w. Output: %s", err, o)
	}

	// remove tmp file
	_, err = c.CombinedOutput(fmt.Sprintf("rm -f %s", tmpPath), true)
	if err != nil {
		log.Debug().Msgf("failed to remove tmp file: %s. Output: %s", tmpPath, o)
	}

	return nil
}

func (c *Client) CopyContentToRemote(src io.Reader, dst string, permission uint32) error {
	client, err := c.GetSftpClient()
	if err != nil {
		return fmt.Errorf("failed to get sftp client: %w", err)
	}

	dir := GetParentDir(dst)
	if err := client.MkdirAll(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	dstFile, err := client.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create dst file: %w", err)
	}

	if _, err = dstFile.ReadFrom(src); err != nil {
		if permission != 0 {
			if err = dstFile.Chmod(fs.FileMode(permission)); err != nil {
				return fmt.Errorf("failed to change file permission: %w", err)
			}
		}
	}

	return nil
}

// src is the local file path, rDir is the remote directory path, rName is the remote file name.
func CopyFileToRemoteScp(client *ssh.Client, src, rDir, rName string, permission uint32) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open src file: %w", err)
	}
	defer srcFile.Close()
	stat, _ := srcFile.Stat()

	wg := sync.WaitGroup{}
	wg.Add(1)

	if permission == 0 {
		permission = 644
	}

	go func() {
		hostIn, _ := session.StdinPipe()
		defer hostIn.Close()
		fmt.Fprintf(hostIn, "C0%d %d %s\n", permission, stat.Size(), rName)
		io.Copy(hostIn, srcFile)
		fmt.Fprint(hostIn, "\x00")
		wg.Done()
	}()

	cmd := "/usr/bin/scp -t /remotedirectory/"
	if o, err := session.CombinedOutput(cmd); err != nil {
		return fmt.Errorf("failed to run scp: %w. Output: %s", err, o)
	}
	wg.Wait()
	return nil
}

// create folder specified by path, if already existed, do nothing.
func (c *Client) CreateDir(path string, sudo bool) error {
	existed, err := c.DirExist(path)
	if err != nil {
		return fmt.Errorf("failed to check whether dir exist: %w", err)
	} else if existed {
		return nil
	}

	session, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	command := fmt.Sprintf("mkdir -p %s", path)
	if sudo {
		command = c.getSudoCmd(command)
	}

	return session.Run(command)
}

func (c *Client) DirExist(path string) (bool, error) {
	session, err := c.NewSession()
	if err != nil {
		return false, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.Output(fmt.Sprintf("test -d %s && echo 'yes' || echo 'no'", path))
	if err != nil {
		return false, fmt.Errorf("failed to execute cmd: %w", err)
	}

	result := string(output)
	return result == "yes\n", nil
}

func (c *Client) PathExisted(path string) (bool, error) {
	session, err := c.NewSession()
	if err != nil {
		return false, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	b, err := session.Output("type -P " + path)
	if err != nil {
		return false, err
	}
	if len(b) == 0 {
		return false, nil
	}
	return true, nil
}

func (c *Client) AppendToFile(path, content string, sudo bool) error {
	session, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	command := fmt.Sprintf("bash -c 'echo \"%s\" >> %s'", content, path)
	if sudo {
		command = c.getSudoCmd(command)
	}

	output, err := session.Output(command)
	log.Debug().Msgf("append to file output: %s", output)
	if err != nil {
		return fmt.Errorf("failed to append to file: %w", err)
	}
	return nil
}

// removeLastPathSegment removes the last segment from a Unix-style path
// a/b/c -> a/b
func removeLastPathSegment(path string) string {
	// Handle empty path
	if path == "" {
		return ""
	}

	// Remove trailing slashes
	path = strings.TrimRight(path, "/")

	// Find the last slash
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return ""
	}

	// Return everything up to the last slash
	return path[:lastSlash]
}

// GetParentDir returns the parent directory of a Unix-style path
func GetParentDir(path string) string {
	return removeLastPathSegment(path)
}

func (c *Client) DownloadRemoteFileToLocal(remotePath, localPath string) error {
	client, err := c.GetSftpClient()
	if err != nil {
		return fmt.Errorf("failed to get sftp client: %w", err)
	}

	srcFile, err := client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

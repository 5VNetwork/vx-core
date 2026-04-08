package sshhelper

import (
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	*ssh.Client
	Password string

	isRoot              bool
	hasPasswordlessSudo bool

	sync.Mutex
	sftpClient *sftp.Client
}

func (c *Client) GetSftpClient() (*sftp.Client, error) {
	c.Lock()
	defer c.Unlock()
	if c.sftpClient == nil {
		sftpClient, err := sftp.NewClient(c.Client)
		if err != nil {
			return nil, fmt.Errorf("failed to create sftp client: %w", err)
		}
		c.sftpClient = sftpClient
	}
	return c.sftpClient, nil
}

type DialConfig struct {
	Addr                 string
	User                 string
	Password             string
	PrivateKey           []byte
	PrivateKeyPassphrase string
	HostKey              []byte
	Timeout              time.Duration
}

func Dial(s *DialConfig) (*Client, []byte, error) {
	// auth
	var authMethods []ssh.AuthMethod

	if len(s.PrivateKey) > 0 {
		if s.PrivateKeyPassphrase != "" {
			signer, err := ssh.ParsePrivateKeyWithPassphrase(s.PrivateKey, []byte(s.PrivateKeyPassphrase))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
			}
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		} else {
			signer, err := ssh.ParsePrivateKey(s.PrivateKey)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
			}
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	} else if s.Password != "" {
		authMethods = append(authMethods, ssh.Password(s.Password))
		// Add both password and keyboard-interactive methods
		// Some servers prefer keyboard-interactive over password
		authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			// Answer all questions with the password
			answers := make([]string, len(questions))
			for i := range answers {
				answers[i] = s.Password
			}
			return answers, nil
		}))
	}

	var hostKeyCallback ssh.HostKeyCallback
	var hostKey []byte
	if s.HostKey != nil {
		pub, err := ParsePublicKeyFromAny(s.HostKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		hostKeyCallback = ssh.FixedHostKey(pub)
	} else {
		hostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if key != nil {
				hostKey = key.Marshal()
				return nil
			}
			log.Warn().Msg("no host key")
			return nil
		}
	}

	conf := &ssh.ClientConfig{
		User:            s.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         s.Timeout,
	}

	client, err := ssh.Dial("tcp", s.Addr, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ssh.Dial: %w", err)
	}
	c := &Client{Client: client, Password: s.Password}

	c.isRoot, err = c.IsRoot()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check if user is root: %w", err)
	}
	if !c.isRoot {
		c.hasPasswordlessSudo, err = c.HasPasswordlessSudo()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to check if user has passwordless sudo: %w", err)
		}
	}

	return c, hostKey, nil
}

func ParsePublicKeyFromAny(keyData []byte) (ssh.PublicKey, error) {
	// Try parsing as raw public key first
	pub, err := ssh.ParsePublicKey(keyData)
	if err == nil {
		return pub, nil
	}

	// Try parsing as authorized key format
	_, _, pub, _, _, err = ssh.ParseKnownHosts(keyData)
	if err == nil {
		return pub, nil
	}

	// Try parsing as PEM
	block, _ := pem.Decode(keyData)
	if block != nil {
		pub, err = ssh.ParsePublicKey(block.Bytes)
		if err == nil {
			return pub, nil
		}
	}

	return nil, fmt.Errorf("failed to parse public key in any known format")
}

func (c *Client) Close() error {
	c.Lock()
	defer c.Unlock()
	var err error
	if c.sftpClient != nil {
		err = c.sftpClient.Close()
	}
	return errors.Join(c.Client.Close(), err)
}

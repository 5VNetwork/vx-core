//go:build !windows && !wasm && server
// +build !windows,!wasm,server

package domainsocket

import (
	"context"
	gotls "crypto/tls"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/transport/security"
)

type Listener struct {
	addr           *net.UnixAddr
	ln             net.Listener
	tlsConfig      *gotls.Config
	config         *DomainSocketConfig
	locker         *fileLocker
	securityConfig security.Engine
}

func Listen(ctx context.Context, address net.Address, port net.Port, config *DomainSocketConfig, securityConfig security.Engine) (*Listener, error) {
	addr, err := config.GetUnixAddr()
	if err != nil {
		return nil, err
	}

	unixListener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on domain socket %s: %w", addr.Name, err)
	}

	ln := &Listener{
		addr:           addr,
		ln:             unixListener,
		config:         config,
		securityConfig: securityConfig,
	}

	if !config.Abstract {
		ln.locker = &fileLocker{
			path: config.Path + ".lock",
		}
		if err := ln.locker.Acquire(); err != nil {
			unixListener.Close()
			return nil, err
		}
	}

	return ln, nil
}

func (ln *Listener) Addr() net.Addr {
	return ln.addr
}

func (ln *Listener) Close() error {
	if ln.locker != nil {
		ln.locker.Release()
	}
	return ln.ln.Close()
}

func (ln *Listener) Accept() (net.Conn, error) {
	conn, err := ln.ln.Accept()
	if err != nil {
		return nil, err
	}

	if ln.securityConfig != nil {
		conn, err = ln.securityConfig.GetClientConn(conn)
	}

	return conn, nil
}

type fileLocker struct {
	path string
	file *os.File
}

func (fl *fileLocker) Acquire() error {
	f, err := os.Create(fl.path)
	if err != nil {
		return err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		f.Close()
		return fmt.Errorf("failed to lock file: %w", err)
	}
	fl.file = f
	return nil
}

func (fl *fileLocker) Release() {
	if err := unix.Flock(int(fl.file.Fd()), unix.LOCK_UN); err != nil {
		log.Err(err).Msg("failed to unlock file")
	}
	if err := fl.file.Close(); err != nil {
		log.Err(err).Msg("failed to close file")
	}
	if err := os.Remove(fl.path); err != nil {
		log.Err(err).Msg("failed to remove file")
	}
}

// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

//go:build android

package tun

import (
	"io"
	"net/netip"
	"os"
	sync "sync"

	"github.com/5vnetwork/vx-core/common/buf"
	"golang.org/x/sys/unix"
)

type ReadWriteCloserTun struct {
	Rw         io.ReadWriteCloser
	once       sync.Once
	mtu        uint32
	name       string
	ip4        netip.Addr
	ip6        netip.Addr
	dnsServers []netip.Addr
}

func NewTun(config *TunOption) (TunDeviceWithInfo, error) {
	dupFd, err := unix.Dup(int(config.FD))
	if err != nil {
		return nil, err
	}
	err = unix.SetNonblock(dupFd, true)
	if err != nil {
		unix.Close(dupFd)
		return nil, err
	}

	file := os.NewFile(uintptr(dupFd), "/dev/tun")

	dnsServers := make([]netip.Addr, 0, len(config.Dns4)+len(config.Dns6))
	dnsServers = append(dnsServers, config.Dns4...)
	dnsServers = append(dnsServers, config.Dns6...)

	return &ReadWriteCloserTun{
		Rw:         file,
		ip4:        config.Ip4.Addr(),
		ip6:        config.Ip6.Addr(),
		dnsServers: dnsServers,
		name:       config.Name,
		mtu:        config.Mtu,
	}, nil
}

func (u *ReadWriteCloserTun) Start() error {
	return nil
}

func (u *ReadWriteCloserTun) Close() error {
	var err error
	u.once.Do(func() {
		err = u.Rw.Close()
	})
	return err
}

func (t *ReadWriteCloserTun) Name() string {
	return t.name
}

func (t *ReadWriteCloserTun) DnsServers() []netip.Addr {
	return t.dnsServers
}

func (t *ReadWriteCloserTun) IP4() netip.Addr {
	return t.ip4
}

func (t *ReadWriteCloserTun) IP6() netip.Addr {
	return t.ip6
}

func (u *ReadWriteCloserTun) ReadPacket() (*buf.Buffer, error) {
	b := buf.NewWithSize(int32(u.mtu))
	n, err := u.Rw.Read(b.BytesTo(b.Cap()))
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Resize(0, int32(n))
	return b, nil
}

var ipv4FourBytes = []byte{0, 0, 0, 2}
var ipv6FourBytes = []byte{0, 0, 0, 30}

func (u *ReadWriteCloserTun) WritePacket(p *buf.Buffer) error {
	defer p.Release()
	_, err := u.Rw.Write(p.Bytes())
	return err
}

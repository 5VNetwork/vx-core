// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package tun

import (
	"net/netip"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
)

const (
	DefaultMtu int = 1500
)

type TunDevice interface {
	common.Runnable
	Name() string
	ReadPacket() (*buf.Buffer, error)
	// WritePacket takes ownership of the buffer.
	// Buffer is released no matter success or not
	WritePacket(*buf.Buffer) error
}

type TunDeviceWithInfo interface {
	TunDevice
	IP4() netip.Addr
	IP6() netip.Addr
	DnsServers() []netip.Addr
}

type TunOption struct {
	Ip4    netip.Prefix
	Ip6    netip.Prefix
	Name   string
	Mtu    uint32
	Dns4   []netip.Addr
	Dns6   []netip.Addr
	Route4 []netip.Prefix
	Route6 []netip.Prefix
	// path of a dir containing wintun.dll
	Path   string
	Metric uint32
	FD     int
	// Offset is the offset of the packet in the tun device.
	Offset int32
}

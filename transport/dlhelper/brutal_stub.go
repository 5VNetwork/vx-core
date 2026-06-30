//go:build !linux

package dlhelper

import (
	"net"

	"github.com/5vnetwork/vx-core/common/errors"
)

const tcpBrutalAvailable = false

func applyBrutalSocketOptions(fd uintptr, sendBPS uint64) error {
	return errors.New("TCP Brutal is only supported on Linux")
}

func SetBrutalOptions(conn net.Conn, sendBPS uint64) error {
	return errors.New("TCP Brutal is only supported on Linux")
}

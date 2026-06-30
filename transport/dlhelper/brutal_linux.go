//go:build linux

package dlhelper

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"syscall"
	"unsafe"
	_ "unsafe"

	"golang.org/x/sys/unix"
)

const (
	tcpBrutalAvailable = true
	tcpBrutalParamsOpt = 23301
	tcpBrutalMinRate   = 65536
	tcpBrutalCwndGain  = 20 // hysteria2 default, tenths (20 = 2.0x)
)

type tcpBrutalParams struct {
	Rate     uint64
	CwndGain uint32
}

//go:linkname setsockopt syscall.setsockopt
func setsockopt(s int, level int, name int, val unsafe.Pointer, vallen uintptr) (err error)

func applyBrutalSocketOptions(fd uintptr, sendBPS uint64) error {
	if sendBPS < tcpBrutalMinRate {
		return fmt.Errorf("tcp brutal send rate must be at least 65536 bytes per second")
	}
	err := unix.SetsockoptString(int(fd), unix.IPPROTO_TCP, unix.TCP_CONGESTION, "brutal")
	if err != nil {
		return fmt.Errorf("failed to set TCP_CONGESTION=brutal; please install the tcp-brutal kernel module, %w", err)
	}
	params := tcpBrutalParams{
		Rate:     sendBPS,
		CwndGain: tcpBrutalCwndGain,
	}
	err = setsockopt(int(fd), unix.IPPROTO_TCP, tcpBrutalParamsOpt, unsafe.Pointer(&params), unsafe.Sizeof(params))
	if err != nil {
		return fmt.Errorf("failed to set TCP_BRUTAL_PARAMS, %w", os.NewSyscallError("setsockopt IPPROTO_TCP TCP_BRUTAL_PARAMS", err))
	}
	return nil
}

func SetBrutalOptions(conn net.Conn, sendBPS uint64) error {
	syscallConn, ok := conn.(syscall.Conn)
	if !ok {
		return fmt.Errorf("tcp brutal: connection does not support syscall.Conn: %s", reflect.TypeOf(conn))
	}
	rawConn, err := syscallConn.SyscallConn()
	if err != nil {
		return fmt.Errorf("tcp brutal: failed to get syscall conn, %w", err)
	}
	var controlErr error
	err = rawConn.Control(func(fd uintptr) {
		controlErr = applyBrutalSocketOptions(fd, sendBPS)
	})
	if err != nil {
		return fmt.Errorf("tcp brutal: failed to apply socket options, %w", err)
	}
	return controlErr
}

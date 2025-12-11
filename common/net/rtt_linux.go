//go:build linux && !android && server
// +build linux,!android,server

package net

import (
	"fmt"
	"net"
	"unsafe"

	"golang.org/x/sys/unix"
)

type LinkMetrics struct {
	Rtt       uint32 // ms
	Bandwidth uint32 //MBps
}

func GetTCPConnectionRTT(conn *net.TCPConn) (*LinkMetrics, error) {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return nil, fmt.Errorf("failed to get syscall conn: %v", err)
	}

	var info unix.TCPInfo
	var rtt float64
	var bandwidth uint32
	err = rawConn.Control(func(fd uintptr) {
		length := uint32(unix.SizeofTCPInfo)
		_, _, e := unix.Syscall6(
			unix.SYS_GETSOCKOPT,
			fd,
			unix.IPPROTO_TCP,
			unix.TCP_INFO,
			uintptr(unsafe.Pointer(&info)),
			uintptr(unsafe.Pointer(&length)),
			0,
		)
		if e != 0 {
			err = fmt.Errorf("syscall error: %v", e)
			return
		}
		rtt = float64(info.Rtt) / 1000 //ms
		rttSec := float64(rtt) / 1000
		if rttSec == 0 {
			return
		}
		bytesPerSec := float64(info.Snd_cwnd*info.Snd_mss) / rttSec
		bandwidth = uint32(bytesPerSec / 1000 / 1000) //MBps
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get TCP_INFO: %v", err)
	}

	// log.Printf("info: %+v", info)

	return &LinkMetrics{
		Rtt:       uint32(rtt),
		Bandwidth: bandwidth,
	}, nil
}

const (
	TCP_CC_INFO  = 26 // Congestion control information
	TCP_BBR_INFO = 24 // BBR-specific information
)

// func GetBBRInfo(conn *net.TCPConn) (inetdiag.BBRInfo, error) {
// 	f, err := conn.File()
// 	if err != nil {
// 		return inetdiag.BBRInfo{}, fmt.Errorf("failed to get fd %v", err)
// 	}

// 	info, err := bbr.GetBBRInfo(f)
// 	return info, err
// }

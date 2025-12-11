package net

// import (
// 	"fmt"
// 	"net"
// 	"time"
// 	"unsafe"

// 	"golang.org/x/sys/windows"
// )

// // TCP_ESTATS_SYN_OPTS_ROS_v0 structure
// type TCP_ESTATS_SYN_OPTS_ROS_v0 struct {
// 	ActiveOpen uint32
// 	MssRcvd    uint32
// 	MssSent    uint32
// }

// // MIB_TCPROW2 structure
// type MIB_TCPROW2 struct {
// 	DwState      uint32
// 	DwLocalAddr  uint32
// 	DwLocalPort  uint32
// 	DwRemoteAddr uint32
// 	DwRemotePort uint32
// }

// const (
// 	TcpConnectionEstatsSynOpts = 2
// 	TCP_ESTATS_TYPE_ROS        = 0
// )

// var (
// 	iphlpapi                      = windows.NewLazySystemDLL("iphlpapi.dll")
// 	procGetPerTcpConnectionEStats = iphlpapi.NewProc("GetPerTcpConnectionEStats")
// )

// func getTCPConnectionRTT(conn *net.TCPConn) (time.Duration, error) {
// 	// Get local and remote addresses
// 	localAddr := conn.LocalAddr().(*net.TCPAddr)
// 	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)

// 	// Prepare MIB_TCPROW2 structure
// 	row := &MIB_TCPROW2{
// 		DwState:      uint32(5),
// 		DwLocalAddr:  uint32(localAddr.IP.To4()[3])<<24 | uint32(localAddr.IP.To4()[2])<<16 | uint32(localAddr.IP.To4()[1])<<8 | uint32(localAddr.IP.To4()[0]),
// 		DwLocalPort:  uint32(localAddr.Port),
// 		DwRemoteAddr: uint32(remoteAddr.IP.To4()[3])<<24 | uint32(remoteAddr.IP.To4()[2])<<16 | uint32(remoteAddr.IP.To4()[1])<<8 | uint32(remoteAddr.IP.To4()[0]),
// 		DwRemotePort: uint32(remoteAddr.Port),
// 	}

// 	var stats TCP_ESTATS_SYN_OPTS_ROS_v0
// 	err := getPerTcpConnectionEStats(row, TcpConnectionEstatsSynOpts, &stats)
// 	if err != nil {
// 		fmt.Println("Error getting TCP stats:", err)
// 	}

// }

// func getPerTcpConnectionEStats(row *MIB_TCPROW2, estatsType uint32, rw *TCP_ESTATS_SYN_OPTS_ROS_v0) error {
// 	ret, _, err := procGetPerTcpConnectionEStats.Call(
// 		uintptr(unsafe.Pointer(row)),
// 		uintptr(estatsType),
// 		uintptr(0),
// 		uintptr(0),
// 		uintptr(0),
// 		uintptr(TCP_ESTATS_TYPE_ROS),
// 		uintptr(0),
// 		uintptr(unsafe.Sizeof(*rw)),
// 		uintptr(unsafe.Pointer(rw)),
// 		uintptr(0),
// 	)
// 	if ret != 0 {
// 		return err
// 	}
// 	return nil
// }

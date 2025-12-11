package dlhelper

import (
	"context"
	"net"
	"sync/atomic"

	net1 "github.com/5vnetwork/vx-core/common/net"
)

// SocketSetting is options to be applied on network sockets.
type SocketSetting struct {
	// Mark of the connection. If non-zero, the value will be set to SO_MARK.
	Mark uint32
	// TFO is the state of TFO settings. Non supported on darwin.
	Tfo SocketConfig_TCPFastOpenState
	// TProxy is for enabling TProxy socket option.
	Tproxy SocketConfig_TProxyMode
	// ReceiveOriginalDestAddress is for enabling IP_RECVORIGDSTADDR socket
	// option. This option is for UDP only.
	ReceiveOriginalDestAddress bool
	BindAddress                []byte
	BindPort                   uint32
	AcceptProxyProtocol        bool
	TcpKeepAliveInterval       int32
	TfoQueueLength             uint32
	TcpKeepAliveIdle           int32
	BindToDevice4              uint32
	BindToDevice6              uint32
	// this field is only considered only on linux now
	BindToDeviceName string
	RxBufSize        int64
	TxBufSize        int64
	ForceBufSize     bool
	// laddr used when dial and listen
	LocalAddr4        string
	LocalAddr6        string
	StatsReadCounter  *atomic.Uint64
	StatsWriteCounter *atomic.Uint64
	Resolver          *net.Resolver
	FdFunc            func(fd uintptr) error
}

func (s *SocketSetting) Dial(ctx context.Context, dest net1.Destination) (net.Conn, error) {
	return DialSystemConn(ctx, dest, s)
}

func (s *SocketSetting) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	return ListenSystemPacket(ctx, network, address, s)
}

func (s *SocketSetting) Copy() *SocketSetting {
	copy := *s
	return &copy
}

func (s *SocketSetting) Listen(ctx context.Context, addr net.Addr) (net.Listener, error) {
	return ListenSystem(ctx, addr, s)
}

type ConstBindToDeviceGet struct {
	BindToDevice uint32
}

func (c *ConstBindToDeviceGet) GetBindToDevice() uint32 {
	return c.BindToDevice
}

type BindToDeviceGet interface {
	GetBindToDevice() uint32
}

type SocketConfig_TCPFastOpenState int32

const (
	// AsIs is to leave the current TFO state as is, unmodified.
	SocketConfig_AsIs SocketConfig_TCPFastOpenState = 0
	// Enable is for enabling TFO explictly.
	SocketConfig_Enable SocketConfig_TCPFastOpenState = 1
	// Disable is for disabling TFO explictly.
	SocketConfig_Disable SocketConfig_TCPFastOpenState = 2
)

// Enum value maps for SocketConfig_TCPFastOpenState.
var (
	SocketConfig_TCPFastOpenState_name = map[int32]string{
		0: "AsIs",
		1: "Enable",
		2: "Disable",
	}
	SocketConfig_TCPFastOpenState_value = map[string]int32{
		"AsIs":    0,
		"Enable":  1,
		"Disable": 2,
	}
)

func (x SocketConfig_TCPFastOpenState) Enum() *SocketConfig_TCPFastOpenState {
	p := new(SocketConfig_TCPFastOpenState)
	*p = x
	return p
}

type SocketConfig_TProxyMode int32

const (
	// TProxy is off.
	SocketConfig_Off SocketConfig_TProxyMode = 0
	// TProxy mode.
	SocketConfig_TProxy SocketConfig_TProxyMode = 1
	// Redirect mode.
	SocketConfig_Redirect SocketConfig_TProxyMode = 2
)

// Enum value maps for SocketConfig_TProxyMode.
var (
	SocketConfig_TProxyMode_name = map[int32]string{
		0: "Off",
		1: "TProxy",
		2: "Redirect",
	}
	SocketConfig_TProxyMode_value = map[string]int32{
		"Off":      0,
		"TProxy":   1,
		"Redirect": 2,
	}
)

func (x SocketConfig_TProxyMode) Enum() *SocketConfig_TProxyMode {
	p := new(SocketConfig_TProxyMode)
	*p = x
	return p
}

func (x *SocketSetting) GetMark() uint32 {
	if x != nil {
		return x.Mark
	}
	return 0
}

func (x *SocketSetting) GetTfo() SocketConfig_TCPFastOpenState {
	if x != nil {
		return x.Tfo
	}
	return SocketConfig_AsIs
}

func (x *SocketSetting) GetTproxy() SocketConfig_TProxyMode {
	if x != nil {
		return x.Tproxy
	}
	return SocketConfig_Off
}

func (x *SocketSetting) GetReceiveOriginalDestAddress() bool {
	if x != nil {
		return x.ReceiveOriginalDestAddress
	}
	return false
}

func (x *SocketSetting) GetBindAddress() []byte {
	if x != nil {
		return x.BindAddress
	}
	return nil
}

func (x *SocketSetting) GetBindPort() uint32 {
	if x != nil {
		return x.BindPort
	}
	return 0
}

func (x *SocketSetting) GetAcceptProxyProtocol() bool {
	if x != nil {
		return x.AcceptProxyProtocol
	}
	return false
}

func (x *SocketSetting) GetTcpKeepAliveInterval() int32 {
	if x != nil {
		return x.TcpKeepAliveInterval
	}
	return 0
}

func (x *SocketSetting) GetTfoQueueLength() uint32 {
	if x != nil {
		return x.TfoQueueLength
	}
	return 0
}

func (x *SocketSetting) GetTcpKeepAliveIdle() int32 {
	if x != nil {
		return x.TcpKeepAliveIdle
	}
	return 0
}

func (x *SocketSetting) GetBindToDevice4() uint32 {
	if x != nil {
		return x.BindToDevice4
	}
	return 0
}

func (x *SocketSetting) GetBindToDevice6() uint32 {
	if x != nil {
		return x.BindToDevice6
	}
	return 0
}

func (x *SocketSetting) GetRxBufSize() int64 {
	if x != nil {
		return x.RxBufSize
	}
	return 0
}

func (x *SocketSetting) GetTxBufSize() int64 {
	if x != nil {
		return x.TxBufSize
	}
	return 0
}

func (x *SocketSetting) GetForceBufSize() bool {
	if x != nil {
		return x.ForceBufSize
	}
	return false
}

func (x *SocketSetting) GetLocalAddr4() string {
	if x != nil {
		return x.LocalAddr4
	}
	return ""
}

func (x *SocketSetting) GetLocalAddr6() string {
	if x != nil {
		return x.LocalAddr6
	}
	return ""
}

func (m SocketConfig_TProxyMode) IsEnabled() bool {
	return m != SocketConfig_Off
}

func (so *SocketSetting) Dulplicate() *SocketSetting {
	if so == nil {
		return nil
	}
	return &SocketSetting{
		Mark:                       so.Mark,
		Tproxy:                     so.Tproxy,
		Tfo:                        so.Tfo,
		ReceiveOriginalDestAddress: so.ReceiveOriginalDestAddress,
		BindAddress:                so.BindAddress,
		BindPort:                   so.BindPort,
		AcceptProxyProtocol:        so.AcceptProxyProtocol,
		TcpKeepAliveInterval:       so.TcpKeepAliveInterval,
		TfoQueueLength:             so.TfoQueueLength,
		TcpKeepAliveIdle:           so.TcpKeepAliveIdle,
		BindToDevice4:              so.BindToDevice4,
		BindToDevice6:              so.BindToDevice6,
		LocalAddr4:                 so.LocalAddr4,
		LocalAddr6:                 so.LocalAddr6,
		RxBufSize:                  so.RxBufSize,
		TxBufSize:                  so.TxBufSize,
		ForceBufSize:               so.ForceBufSize,
		StatsReadCounter:           so.StatsReadCounter,
		StatsWriteCounter:          so.StatsWriteCounter,
		Resolver:                   so.Resolver,
		FdFunc:                     so.FdFunc,
	}
}

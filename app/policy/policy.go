package policy

import (
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
)

type Policy struct {
	Timeout
	UserBufferPolicy
	StatsPolicy
	defaultBufferSize int32
}

type Timeout struct {
	handshakeTimeout         time.Duration
	tcpConnectionIdleTimeout time.Duration
	sshIdleTimeout           time.Duration
	udpIdleTimeout           time.Duration
	upLinkOnlyTimeout        time.Duration
	downLinkOnlyTimeout      time.Duration
	dnsIdleTimeout           time.Duration
}

func (sp *Timeout) HandshakeTimeout() time.Duration {
	return sp.handshakeTimeout
}

func (sp *Timeout) TcpIdleTimeout() time.Duration {
	return sp.tcpConnectionIdleTimeout
}

func (sp *Timeout) UdpIdleTimeout() time.Duration {
	return sp.udpIdleTimeout
}

func (sp *Timeout) SshIdleTimeout() time.Duration {
	return sp.sshIdleTimeout
}

func (sp *Timeout) DnsIdleTimeout() time.Duration {
	return sp.dnsIdleTimeout
}

func (sp *Timeout) DownLinkOnlyTimeout() time.Duration {
	return sp.downLinkOnlyTimeout
}

func (sp *Timeout) UpLinkOnlyTimeout() time.Duration {
	return sp.upLinkOnlyTimeout
}

func (sp *Timeout) SetHandshakeTimeout(d time.Duration) {
	sp.handshakeTimeout = d
}

func (sp *Timeout) SetTcpIdleTimeout(d time.Duration) {
	sp.tcpConnectionIdleTimeout = d
}

func (sp *Timeout) SetUdpIdleTimeout(d time.Duration) {
	sp.udpIdleTimeout = d
}

func (sp *Timeout) SetDownLinkOnlyTimeout(d time.Duration) {
	sp.downLinkOnlyTimeout = d
}

func (sp *Timeout) SetUpLinkOnlyTimeout(d time.Duration) {
	sp.upLinkOnlyTimeout = d
}

var DefaultTimeout = Timeout{
	handshakeTimeout:         time.Second * 4,
	sshIdleTimeout:           0,
	tcpConnectionIdleTimeout: time.Second * 60, // 0 means no timeout
	udpIdleTimeout:           time.Second * 120,
	dnsIdleTimeout:           time.Second * 16,
	upLinkOnlyTimeout:        time.Second * 5,
	downLinkOnlyTimeout:      time.Second * 2,
}

type UserBufferPolicy map[uint32]int32

func (sp UserBufferPolicy) UserBufferSize(level uint32) int32 {
	return sp[level]
}

func (sp UserBufferPolicy) SetUserBufferSize(level uint32, size int32) {
	sp[level] = size
}

var DefaultUserBufferPolicy = UserBufferPolicy{
	0: buf.BufferSize,
	1: buf.BufferSize * 10,
}

var DefaultPolicy = &Policy{
	defaultBufferSize: buf.BufferSize,
	UserBufferPolicy:  DefaultUserBufferPolicy,
	Timeout:           DefaultTimeout,
}

type StatsPolicy struct {
	UserStats         bool
	LinkStats         bool
	OutboundLinkStats bool
	InboundStats      bool
	SessionStats      bool
}

func (sp *StatsPolicy) CalculateUserStats() bool {
	return sp.UserStats
}

func (sp *StatsPolicy) CalculateInboundLinkStats() bool {
	return sp.LinkStats
}

func (sp *StatsPolicy) CalculateOutboundLinkStats() bool {
	return sp.OutboundLinkStats
}

func (sp *StatsPolicy) CalculateInboundStats() bool {
	return sp.InboundStats
}

func (sp *StatsPolicy) CalculateSessionStats() bool {
	return sp.SessionStats
}

func (sp *StatsPolicy) SetInboundStats(b bool) {
	sp.InboundStats = b
}

func (sp *StatsPolicy) SetUserStats(b bool) {
	sp.UserStats = b
}

func (sp *StatsPolicy) SetLinkStats(b bool) {
	sp.LinkStats = b
}

func (sp *StatsPolicy) SetSessionStats(b bool) {
	sp.SessionStats = b
}

func (sp *StatsPolicy) SetOutboundStats(b bool) {
	sp.OutboundLinkStats = b
}
func (sp *Policy) SetDefaultBufferSize(size int32) {
	sp.defaultBufferSize = size
}

func New() *Policy {
	ply := DefaultPolicy
	return ply
}

func (sp *Policy) DefaultBufferSize() int32 {
	return sp.defaultBufferSize
}

func (sp *Policy) UserStats(level uint32) bool {
	return true
}

// Package session provides functions for sessions of incoming requests.
package session

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/common/appid"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/common/uuid"

	"github.com/rs/zerolog/log"
)

// ID of a session. For debugging purpose.
type ID uint32

// NewID generates a new ID. The generated ID is high likely to be unique, but not cryptographically secure.
// The generated ID will never be 0.
func NewID() ID {
	return ID(rand.Uint32())
}

func (id ID) String() string {
	return fmt.Sprintf("[%d]", id)
}

// Sockopt is the settings for socket connection.
type Sockopt struct {
	// Mark of the socket connection.
	Mark           uint32
	InterfaceName4 string
	InterfaceName6 string
	// InterfaceIP4   net.IP
	// InterfaceIP6   net.IP
}

type Info struct {
	ID      ID
	Source  net.Destination
	Gateway net.Destination
	// if not nil, this is the original target address and it is an ip in fake dns pool
	FakeIP          net.IP
	Target          net.Destination
	InboundTag      string
	InboundProtocol string
	UdpUuid         uuid.UUID
	Sockopt         *Sockopt
	User            string

	Sniffed       bool
	Protocol      string
	SniffedDomain string

	AppId       string
	MatchedRule string

	ActivityChecker *signal.ActivityChecker
	// all up traffic should be added to the counter. It typically contains user traffic meter,
	// session traffic meter, and outbound link traffic meter.
	UpCounter   UpCounters
	DownCounter DownCounters
	// for debug purpose
	SessionUpCounter   atomic.Uint64
	SessionDownCounter atomic.Uint64
	// the selector being used initially
	UsedSelector string
}

type UpCounter interface {
	UpTraffic(uint64)
}
type DownCounter interface {
	DownTraffic(uint64)
}
type UpCounters []UpCounter

func (c UpCounters) UpTraffic(n uint64) {
	for _, counter := range c {
		counter.UpTraffic(n)
	}
}

type DownCounters []DownCounter

func (c DownCounters) DownTraffic(n uint64) {
	for _, counter := range c {
		counter.DownTraffic(n)
	}
}

type Option func(*Info)

func NewInfo(options ...Option) *Info {
	info := &Info{
		ID: NewID(),
	}
	for _, opt := range options {
		opt(info)
	}

	return info
}

func NewInfoInbound(options ...Option) (*Info, context.Context, context.CancelCauseFunc) {
	info := &Info{
		ID: NewID(),
	}
	for _, opt := range options {
		opt(info)
	}
	ctx := ContextWithInfo(context.Background(), info)
	ctx, cancelCause := context.WithCancelCause(ctx)
	logger := log.With().Uint32("sid", uint32(info.ID)).Logger()
	ctx = logger.WithContext(ctx)

	event := logger.Debug()
	if info.Source.IsValid() {
		event = event.Str("src", info.Source.String())
		event = event.Str("network", info.Source.Network.String())
	}
	if info.Target.IsValid() {
		event = event.Str("dst", info.Target.String())
	}
	if info.AppId != "" {
		event = event.Str("app", info.AppId)
	}
	if info.UdpUuid.IsSet() {
		event = event.Str("uuid", info.UdpUuid.String())
	}
	if info.InboundTag != "" {
		event = event.Str("inbound", info.InboundTag)
	}
	if info.Gateway.IsValid() {
		event = event.Str("gateway", info.Gateway.String())
	}
	event.Msg("new session")

	return info, ctx, cancelCause
}

func WithSource(src net.Destination) Option {
	return func(info *Info) {
		info.Source = src
	}
}

func WithTarget(tgt net.Destination) Option {
	return func(info *Info) {
		info.Target = tgt
	}
}

func WithInboundTag(tag string) Option {
	return func(info *Info) {
		info.InboundTag = tag
	}
}

func WithGateway(gtw net.Destination) Option {
	return func(info *Info) {
		info.Gateway = gtw
	}
}

func WithUdpUuid(uuid uuid.UUID) Option {
	return func(info *Info) {
		info.UdpUuid = uuid
	}
}

func WithUser(user string) Option {
	return func(info *Info) {
		info.User = user
	}
}

func WithAppId(appId string) Option {
	return func(info *Info) {
		info.AppId = appId
	}
}

func WithInboundProtocol(s string) Option {
	return func(info *Info) {
		info.InboundProtocol = s
	}
}

func Copy(old *Info) *Info {
	return &Info{
		ID:              old.ID,
		Source:          old.Source,
		Gateway:         old.Gateway,
		InboundTag:      old.InboundTag,
		UdpUuid:         old.UdpUuid,
		User:            old.User,
		Target:          old.Target,
		Sockopt:         old.Sockopt,
		Protocol:        old.Protocol,
		InboundProtocol: old.InboundProtocol,
		SniffedDomain:   old.SniffedDomain,
		AppId:           old.AppId,
		ActivityChecker: old.ActivityChecker,
		UpCounter:       old.UpCounter,
		DownCounter:     old.DownCounter,
	}
}

func (c *Info) GetInboundTag() string {
	return c.InboundTag
}

func (c *Info) GetSourceIPs() net.IP {
	src := c.Source
	if src.Address.Family().IsDomain() {
		return nil
	}
	return src.Address.IP()
}

func (c *Info) GetSourceAddr() net.Destination {
	return c.Source
}

func (c *Info) GetSourcePort() net.Port {
	return c.Source.Port
}

func (c *Info) GetTargetIP() net.IP {
	dest := c.Target
	if !dest.IsValid() || dest.Address.Family().IsDomain() {
		return nil
	}
	return dest.Address.IP()
}

func (c *Info) GetTargetPort() net.Port {
	return c.Target.Port
}

func (c *Info) GetTargetDomain() string {
	if c.Target.Address == nil || c.Target.Address.Family().IsIP() {
		if c.SniffedDomain != "" {
			return c.SniffedDomain
		}
		return ""
	}
	return c.Target.Address.Domain()
}

func (c *Info) GetTargetAddr() net.Destination {
	return c.Target
}

func (c *Info) GetNetwork() net.Network {
	return c.Target.Network
}

func (c *Info) GetUser() string {
	return c.User
}

func (c *Info) GetAppId() string {
	if c.AppId != "" {
		return c.AppId
	}
	target := c.Target
	fakeIP := c.FakeIP
	if fakeIP != nil {
		target = net.Destination{
			Address: net.IPAddress(fakeIP),
			Port:    target.Port,
			Network: target.Network,
		}
	}
	appId, err := appid.GetAppId(context.Background(), c.Source, &target)
	if err != nil {
		return ""
	}
	c.AppId = appId
	return appId
}

func (c *Info) GetFakeIP() net.IP {
	return c.FakeIP
}

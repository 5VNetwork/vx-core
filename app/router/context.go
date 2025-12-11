package router

import (
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/uuid"
)

type SessionInfo interface {
	// GetInboundTag returns the tag of the inbound the connection was from.
	GetInboundTag() string

	// GetSourcesIPs returns the source IPs bound to the connection.
	GetSourceIPs() net.IP

	// GetSourcePort returns the source port of the connection.
	GetSourcePort() net.Port

	// GetTargetIPs returns the target IP of the connection or resolved IPs of target domain.
	GetTargetIP() net.IP

	// GetTargetPort returns the target port of the connection.
	GetTargetPort() net.Port

	// GetTargetDomain returns the target domain of the connection, if exists.
	GetTargetDomain() string

	// GetNetwork returns the network type of the connection.
	GetNetwork() net.Network

	// GetProtocol returns the protocol from the connection content, if sniffed out.
	// GetProtocol() string

	// GetUser returns the user email from the connection content, if exists.
	GetUser() uuid.UUID

	GetSourceAddr() net.Destination
	GetTargetAddr() net.Destination
	// GetAttributes returns extra attributes from the connection content.
	// GetAttributes() map[string]string
	GetAppId() string
	GetFakeIP() net.IP
	// GetSkipDNSResolve returns a flag switch for weather skip dns resolve during route pick.
	// GetSkipDNSResolve() bool
}

// type ContextImpl struct {
// 	*session.Info
// }

// func (c *ContextImpl) GetInboundTag() string {
// 	return c.InboundTag
// }

// func (c *ContextImpl) GetSourceIPs() []net.IP {
// 	src := c.Source
// 	if src.Address.Family().IsDomain() {
// 		return nil
// 	}
// 	return []net.IP{src.Address.IP()}
// }

// func (c *ContextImpl) GetSourcePort() net.Port {
// 	return c.Source.Port
// }

// func (c *ContextImpl) GetTargetIPs() []net.IP {
// 	dest := c.Target
// 	if !dest.IsValid() || dest.Address.Family().IsDomain() {
// 		return nil
// 	}
// 	return []net.IP{dest.Address.IP()}
// }

// func (c *ContextImpl) GetTargetPort() net.Port {
// 	return c.Target.Port
// }

// func (c *ContextImpl) GetTargetDomain() string {
// 	dest := c.Target
// 	if !dest.IsValid() || dest.Address.Family().IsIP() {
// 		return ""
// 	}
// 	return dest.Address.Domain()
// }

// func (c *ContextImpl) GetNetwork() net.Network {
// 	return c.Target.Network
// }

// func (c *ContextImpl) GetUser() session.User {
// 	return c.User
// }

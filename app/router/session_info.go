package router

import (
	"context"
	"net"

	"github.com/5vnetwork/vx-core/i"
)

type ContextWithDNS struct {
	SessionInfo
	dnsClient   i.IPResolver
	resolvedIPs []net.IP
}

func NewInfoWithDNS(ctx SessionInfo, dns i.IPResolver) *ContextWithDNS {
	return &ContextWithDNS{SessionInfo: ctx, dnsClient: dns}
}

// GetTargetIPs overrides original routing.Context's implementation.
func (ctx *ContextWithDNS) GetTargetIP() net.IP {
	if ips := ctx.SessionInfo.GetTargetIP(); ips != nil {
		return ips
	}

	if len(ctx.resolvedIPs) > 0 {
		return ctx.resolvedIPs[0]
	}

	if domain := ctx.GetTargetDomain(); len(domain) != 0 {
		ips, err := (ctx.dnsClient).LookupIP(context.Background(), domain)
		if err == nil && len(ips) > 0 {
			ctx.resolvedIPs = ips
			return ips[0]
		}
	}

	return nil
}

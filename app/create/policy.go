package create

import (
	"time"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/policy"
)

func NewPolicy(config *configs.PolicyConfig) *policy.Policy {
	p := policy.New()
	if config.GetConnectionIdleTimeout() != 0 {
		p.SetTcpIdleTimeout(time.Second * time.Duration(config.GetConnectionIdleTimeout()))
	}
	if config.GetUdpIdleTimeout() != 0 {
		p.SetUdpIdleTimeout(time.Second * time.Duration(config.GetUdpIdleTimeout()))
	}
	if config.GetHandshakeTimeout() != 0 {
		p.SetHandshakeTimeout(time.Second * time.Duration(config.GetHandshakeTimeout()))
	}
	if config.GetUpLinkOnlyTimeout() != 0 {
		p.SetUpLinkOnlyTimeout(time.Second * time.Duration(config.GetUpLinkOnlyTimeout()))
	}
	if config.GetDownLinkOnlyTimeout() != 0 {
		p.SetDownLinkOnlyTimeout(time.Second * time.Duration(config.GetDownLinkOnlyTimeout()))
	}
	if config.GetLinkStats() {
		p.SetLinkStats(true)
	}
	if config.GetOutboundStats() {
		p.SetOutboundStats(true)
	}
	if config.GetInboundStats() {
		p.SetInboundStats(true)
	}
	if config.GetUserStats() {
		p.SetUserStats(true)
	}
	if config.GetSessionStats() {
		p.SetSessionStats(true)
	}
	if config.GetUserPolicyMap() != nil {
		for level, size := range config.GetUserPolicyMap() {
			p.SetUserBufferSize(level, size.BufferSize)
		}
	}

	return p
}

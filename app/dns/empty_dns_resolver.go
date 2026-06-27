package dns

import (
	"context"

	"github.com/miekg/dns"
)

type EmptyDnsResolver struct {
	name string
}

func NewEmptyDnsResolver(name string) *EmptyDnsResolver {
	return &EmptyDnsResolver{name: name}
}

func (e *EmptyDnsResolver) Name() string {
	return e.name
}

func (e *EmptyDnsResolver) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	reply := emptyReply(msg)
	return reply, nil
}

func (e *EmptyDnsResolver) Start() error {
	return nil
}

func (e *EmptyDnsResolver) Close() error {
	return nil
}

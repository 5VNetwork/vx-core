package clientgrpc

import (
	"context"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/net"
)

func (s *ClientGrpc) RttTest(ctx context.Context, req *RttTestRequest) (*RttTestResponse, error) {
	dest := net.AddressPort{
		Address: net.ParseAddress(req.Addr),
		Port:    net.Port(req.Port),
	}

	dl, err := s.Client.DialerFactory.GetDialer(nil)
	if err != nil {
		return nil, err
	}
	l, err := s.Client.DialerFactory.GetPacketListener(nil)
	if err != nil {
		return nil, err
	}
	rtt, err := util.RttTest(ctx, dest, dl, l, s.Client.IPResolver)
	if err != nil {
		return nil, err
	}
	return &RttTestResponse{Ping: uint32(rtt)}, nil
}

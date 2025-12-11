package api

import (
	context "context"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/net"
)

func (a *Api) RttTest(ctx context.Context, req *RttTestRequest) (*RttTestResponse, error) {
	dest := net.AddressPort{
		Address: net.ParseAddress(req.Addr),
		Port:    net.Port(req.Port),
	}

	dl, err := a.dialFactory.GetDialer(nil)
	if err != nil {
		return nil, err
	}
	l, err := a.dialFactory.GetPacketListener(nil)
	if err != nil {
		return nil, err
	}
	rtt, err := util.RttTest(ctx, dest, dl, l, a.ipResolver)
	if err != nil {
		return nil, err
	}
	return &RttTestResponse{Ping: uint32(rtt)}, nil
}

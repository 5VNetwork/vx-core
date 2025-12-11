package api

import (
	context "context"

	"github.com/5vnetwork/vx-core/app/util"
)

func (a *Api) GenerateX25519KeyPair(ctx context.Context, req *GenerateX25519KeyPairRequest) (*GenerateX25519KeyPairResponse, error) {
	pub, pri, err := util.Curve25519Genkey(false, "")
	if err != nil {
		return nil, err
	}
	return &GenerateX25519KeyPairResponse{Pub: pub, Pri: pri}, nil
}

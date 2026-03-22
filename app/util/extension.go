package util

import "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/outbound"

func GetTag(handler *outbound.HandlerConfig) string {
	if handler.GetOutbound() != nil {
		return handler.GetOutbound().GetTag()
	}
	if handler.GetChain() != nil {
		return handler.GetChain().GetTag()
	}
	return ""
}

package http

import httppb "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/transport/headers/http"

type (
	Config         = httppb.Config
	RequestConfig  = httppb.RequestConfig
	ResponseConfig = httppb.ResponseConfig
	Header         = httppb.Header
	Status         = httppb.Status
	Version        = httppb.Version
	Method         = httppb.Method
)

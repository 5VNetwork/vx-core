package api

import (
	context "context"

	"github.com/5vnetwork/vx-core/app/subscription/uri"
)

func (a *Api) ToUrl(ctx context.Context, req *ToUrlRequest) (*ToUrlResponse, error) {
	returns := make([]string, 0, len(req.OutboundConfogs))
	var failedNodes []string
	for _, outboundConfig := range req.OutboundConfogs {
		url, err := uri.ToUrl(outboundConfig)
		if err != nil {
			failedNodes = append(failedNodes, outboundConfig.Tag)
			continue
		}
		returns = append(returns, url)
	}
	return &ToUrlResponse{
		Urls:        returns,
		FailedNodes: failedNodes,
	}, nil
}

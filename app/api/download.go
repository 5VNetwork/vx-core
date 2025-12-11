package api

import (
	"bytes"
	"errors"
	"io"
	"log"

	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/transport"
)

// If req.dest is not empty, download data from url to a file. If download failed, the file will be removed.
// If req.dest is empty, download data from url to a buffer.
// TODO: Usage
func (a *Api) ApiDownload(req *DownloadRequest, dialerFactory transport.DialerFactory) (*DownloadResponse, error) {
	response := &DownloadResponse{}
	success := false
	for _, hConfig := range req.Handlers {
		h, err := outbound.NewHandler(&outbound.HandlerConfig{
			HandlerConfig: hConfig,
			DialerFactory: dialerFactory,
			Policy:        policy.New(),
			IPResolver:    a.getIPResolver(),
			DnsServer:     a.dnsServer,
		})
		if err != nil {
			log.Println("failed to create handler", err)
			continue
		}
		httpClient := util.HandlerToHttpClient(h)
		if req.Dest != "" {
			err = util.DownloadToFile(req.Url, httpClient, req.Dest)
			if err != nil {
				log.Println("failed to download to file", err)
				continue
			}
		} else {
			rsp, err := httpClient.Get(req.Url)
			if err != nil {
				log.Println("failed to get", err)
				continue
			}
			buffer := &bytes.Buffer{}
			_, err = io.Copy(buffer, rsp.Body)
			rsp.Body.Close()
			if err != nil {
				log.Println("failed to copy", err)
				continue
			}
			response.Data = buffer.Bytes()
		}
		success = true
		break
	}
	if !success {
		return nil, errors.New("all handlers failed")
	}
	// usage := make(map[string]uint32)
	// st.OutboundStats.Range(func(key, value any) bool {
	// 	usage[key.(string)] = uint32(value.(*outbound.OutboundHandlerStats).DownCounter.Load()) +
	// 		uint32(value.(*outbound.OutboundHandlerStats).UpCounter.Load())
	// 	return true
	// })
	// response.Usage = usage
	return response, nil
}

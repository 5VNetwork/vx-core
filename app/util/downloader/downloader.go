package downloader

import (
	"context"
	"net/url"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
)

type Downloader struct {
	HandlerPicker i.Router
}

func NewDownloader(handlerPicker i.Router) *Downloader {
	return &Downloader{HandlerPicker: handlerPicker}
}

func (d *Downloader) Download(ctx context.Context, u string) ([]byte, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	handler, err := d.HandlerPicker.PickHandler(ctx, &session.Info{
		Target: net.Destination{
			Address: net.ParseAddress(parsedUrl.Host),
			Port:    443,
			Network: net.Network_TCP,
		},
	})
	if err != nil {
		return nil, err
	}
	return util.DownloadToMemoryResty(ctx, u, handler)
}

type Downloader0 struct {
	handlers []i.Outbound
}

func NewDownloader0(handlers []i.Outbound) *Downloader0 {
	return &Downloader0{handlers: handlers}
}

func (d *Downloader0) Download(ctx context.Context, url string) ([]byte, error) {
	return util.DownloadToMemoryResty(ctx, url, d.handlers...)
}

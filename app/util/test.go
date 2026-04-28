// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/rs/zerolog/log"
)

var TraceList = []string{
	"https://blog.cloudflare.com/cdn-cgi/trace",
	"https://developers.cloudflare.com/cdn-cgi/trace",
	"https://hostinger.com/cdn-cgi/trace",
	"https://ahrefs.com/cdn-cgi/trace",
}

const (
	SpeedtestURL1  = "https://speed.cloudflare.com/__down?bytes=1000000"  //1MB
	SpeedtestURL10 = "https://speed.cloudflare.com/__down?bytes=10000000" //10MB
)

var (
	UsableTestUrls = []string{
		"https://www.google.com/generate_204",
		"https://www.gstatic.com/generate_204",
		"https://www.apple.com/library/test/success.html",
		"http://www.msftconnecttest.com/connecttest.txt",
		"http://jsonplaceholder.typicode.com/posts/1",
	}
)

const UsableTestUrlCf = "http://cp.cloudflare.com"

func ApiHandlerUsable1(ctx context.Context, h i.Outbound, url string) (bool, error) {
	httpClient := HandlerToHttpClient(h)
	defer httpClient.CloseIdleConnections()
	httpClient.Timeout = 10 * time.Second
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	rsp, err := httpClient.Do(request)
	if err != nil {
		return false, nil
	} else {
		rsp.Body.Close()
		return true, nil
	}
}

func ApiHandlerPing(ctx context.Context, h i.Outbound, url string) (int, error) {
	httpClient := HandlerToHttpClient(h)
	defer httpClient.CloseIdleConnections()
	httpClient.Timeout = 10 * time.Second
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return -1, err
	}
	start := time.Now()
	rsp, err := httpClient.Do(request)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("ping test err")
		return -1, nil
	} else {
		rsp.Body.Close()
		ping := time.Since(start).Milliseconds()
		return int(ping), nil
	}
}

func Speedtest(ctx context.Context, url string, h i.Outbound) (int64, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("speedtest create request err")
		return -1, err
	}

	httpClient := HandlerToHttpClient(h)
	defer httpClient.CloseIdleConnections()

	httpClient.Timeout = 10 * time.Second
	rsps, err := httpClient.Do(request)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("speedtest get err")
		return -1, nil
	}
	start := time.Now()
	log.Ctx(ctx).Debug().Msg("response got")

	n, err := io.Copy(io.Discard, rsps.Body)
	elapsed := time.Since(start)
	rsps.Body.Close()
	if err != nil && n == 0 {
		log.Ctx(ctx).Debug().Err(err).Msg("speedtest err: failed to read response body")
		return -1, nil
	}

	log.Ctx(ctx).Debug().Int64("n", n/1000).Msg("body read")
	speed := float64(n) / elapsed.Seconds()
	return int64(speed), nil
}

func Ping(ctx context.Context, config *configs.OutboundHandlerConfig, info i.DefaultInterfaceInfo) (uint32, error) {
	// ping
	dest := net.ParseAddress(config.Address)
	if dest.Family().IsDomain() {
		ips, err := net.LookupIP(dest.Domain())
		if err != nil {
			log.Printf("Speedtest lookup ip err: %v", err)
		}
		if len(ips) == 0 {
			return 0, fmt.Errorf("speedtest lookup ip err: %v", err)
		}
		dest = net.IPAddress(ips[0])
	}
	var port uint16
	if config.Port != 0 {
		port = uint16(config.Port)
	} else {
		portList := config.Ports
		if len(portList) == 0 {
			return 0, fmt.Errorf("speedtest port list err")
		}
		port = uint16(portList[0].From)
	}
	socketOption := &dlhelper.SocketSetting{
		BindToDevice4: info.DefaultInterface4(),
		BindToDevice6: info.DefaultInterface6(),
	}
	start := time.Now()
	conn, err := dlhelper.DialSystemConn(ctx, net.TCPDestination(dest, net.Port(port)), socketOption)
	if err != nil {
		return 0, fmt.Errorf("speedtest dial err: %v", err)
	}
	rtt := uint32(time.Since(start).Milliseconds())
	conn.Close()
	return rtt, nil
}

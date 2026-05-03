// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package uri

import (
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
)

func addQueryParameters(queryParameters url.Values, outboundConfig *configs.OutboundHandlerConfig) {
	if t := outboundConfig.Transport.GetTls(); t != nil {
		queryParameters.Add("security", "tls")
		if t.GetServerName() != "" {
			queryParameters.Add("sni", t.GetServerName())
		}
		if t.GetImitate() != "" {
			queryParameters.Add("fp", t.GetImitate())
		}
		if len(t.NextProtocol) > 0 {
			queryParameters.Add("alpn", strings.Join(t.GetNextProtocol(), ","))
		}
		if t.AllowInsecure {
			queryParameters.Add("allowInsecure", "1")
		}
		if len(t.PinnedPeerCertificateChainSha256) > 0 {
			queryParameters.Add("pinSHA256", hex.EncodeToString(t.PinnedPeerCertificateChainSha256[0]))
		}
		if len(t.EchConfig) > 0 {
			queryParameters.Add("echConfig", base64.StdEncoding.EncodeToString(t.EchConfig))
		}
	} else if t := outboundConfig.Transport.GetReality(); t != nil {
		queryParameters.Add("security", "reality")
		if t.GetServerName() != "" {
			queryParameters.Add("sni", t.GetServerName())
		}
		if t.GetSid() != "" {
			queryParameters.Add("sid", t.GetSid())
		}
		if t.GetPbk() != "" {
			queryParameters.Add("pbk", t.GetPbk())
		} else if t.GetPublicKey() != nil {
			queryParameters.Add("pbk", base64.StdEncoding.EncodeToString(t.GetPublicKey()))
		}
		if t.GetFingerprint() != "" {
			queryParameters.Add("fp", t.GetFingerprint())
		}
	}
	if ws := outboundConfig.Transport.GetWebsocket(); ws != nil {
		queryParameters.Add("type", "ws")
		if ws.Path != "" {
			queryParameters.Add("path", ws.GetPath())
		}
		if ws.Host != "" {
			queryParameters.Add("host", ws.Host)
		}
	} else if g := outboundConfig.Transport.GetGrpc(); g != nil {
		queryParameters.Add("type", "grpc")
		queryParameters.Add("serviceName", g.GetServiceName())
	}
	if outboundConfig.EnableMux {
		queryParameters.Add("mux", "1")
		if outboundConfig.GetMuxConfig().GetMaxConcurrency() > 0 {
			queryParameters.Add("mux_max_concurrency",
				strconv.Itoa(int(outboundConfig.GetMuxConfig().GetMaxConcurrency())))
		}
		if outboundConfig.GetMuxConfig().GetMaxConnection() > 0 {
			queryParameters.Add("mux_max_connection",
				strconv.Itoa(int(outboundConfig.GetMuxConfig().GetMaxConnection())))
		}
	}
	if outboundConfig.Uot {
		queryParameters.Add("uot", "1")
	}
}

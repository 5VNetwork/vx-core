package test

import (
	"context"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/app/util/downloader"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/rs/zerolog/log"
)

func TestSub(t *testing.T) {
	t.Skip()
	content, _, err := downloader.DownloadToMemoryResty(context.Background(),
		"",
		map[string]string{
			"User-Agent": "v2ray-core",
		},
		freedom.New(transport.DefaultDialer, transport.DefaultPacketListener, "freedom", nil),
	)
	common.Must(err)
	result, err := util.Decode(string(content))
	common.Must(err)
	log.Print(result)
}

func TestDns(t *testing.T) {
	t.Skip()
	result, err := util.Decode("")
	common.Must(err)
	log.Print(result)
}

func TestUsable(t *testing.T) {
	t.Skip()
	h, err := create.NewOutHandler(&create.Config{
		OutboundHandlerConfig: &configs.OutboundHandlerConfig{
			Address: "",
			Port:    0,
			Protocol: serial.ToTypedMessage(&proxy.VlessClientConfig{
				Id:         "",
				Encryption: "none",
			}),
			Transport: &configs.TransportConfig{
				Protocol: &configs.TransportConfig_Websocket{
					Websocket: &websocket.WebsocketConfig{
						Path:                "/",
						MaxEarlyData:        2560,
						Host:                "",
						EarlyDataHeaderName: "",
					},
				},
				Security: &configs.TransportConfig_Tls{Tls: &tls.TlsConfig{
					ServerName: "",
					Imitate:    "",
				}},
			},
		},
		DialerFactory: transport.DefaultDialerFactory(),
		Policy:        policy.DefaultPolicy,
	})
	common.Must(err)
	speed := util.Speedtest(context.Background(), util.SpeedtestURL1, h)
	log.Print(speed)
}

func TestA(t *testing.T) {
	// t.Skip()
}

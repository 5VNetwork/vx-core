package api

import (
	context "context"
	"log"
	"testing"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"
)

func TestApiHandlerUsable(t *testing.T) {
	// uri := ""
	// decoded, err := util.Decode(uri)
	// common.Must(err)
	// h, err := create.NewOutHandler(&create.Config{
	// 	OutboundHandlerConfig: decoded.Configs[0],
	// 	DialerFactory:         transport.DefaultDialerFactory(),
	// 	IPResolver:            &dns.DnsResolver{},
	// 	Policy:                policy.DefaultPolicy,
	// })
	// common.Must(err)

	freedomHandler := freedom.New(transport.DefaultDialer, transport.DefaultPacketListener, "direct", nil)

	response, err := util.ApiHandlerUsable1(context.Background(), freedomHandler, util.UsableTestUrlCf)
	if err != nil {
		common.Must(err)
		return
	}
	log.Println(response)
}

package dns_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/5vnetwork/vx-core/app/dns"
// 	"github.com/5vnetwork/vx-core/common"
// )

// func TestLocalNameServer(t *testing.T) {
// 	// mockCtrl := gomock.NewController(t)
// 	// mockDefaultInfcMntr := mocks.NewMockInterfaceMonitor(mockCtrl)
// 	s := dns.NewLocalNameServer()
// 	ips, err := s.QueryIP(context.Background(), "www.baidu.com", dns.IPOption{
// 		IPv4Enable: true,
// 		IPv6Enable: true,
// 	})
// 	common.Must(err)
// 	if len(ips) == 0 {
// 		t.Error("expect some ips, but got 0")
// 	}
// }

package freedom

import (
	"context"
	"errors"
	"testing"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/test/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDomainToIpPacketConn_WritePacket(t *testing.T) {
	tests := []struct {
		name           string
		inputPacket    *udp.Packet
		expectedPacket *udp.Packet
		shouldError    bool
		setupMocks     func(*mocks.MockIPResolver, *mocks.MockMyPacketConn)
	}{
		{
			name: "domain target - should resolve to IP",
			inputPacket: &udp.Packet{
				Target: net.Destination{
					Address: net.DomainAddress("example.com"),
				},
			},
			expectedPacket: &udp.Packet{
				Target: net.Destination{
					Address: net.IPAddress([]byte{1, 2, 3, 4}),
				},
			},
			setupMocks: func(dns *mocks.MockIPResolver, pc *mocks.MockMyPacketConn) {
				dns.EXPECT().LookupIPv4(gomock.Any(), "example.com").Return([]net.IP{{1, 2, 3, 4}}, nil)
				pc.EXPECT().WritePacket(gomock.Any()).Return(nil)
			},
		},
		{
			name: "IP target - no resolution needed",
			inputPacket: &udp.Packet{
				Target: net.Destination{
					Address: net.IPAddress([]byte{1, 2, 3, 4}),
				},
			},
			expectedPacket: &udp.Packet{
				Target: net.Destination{
					Address: net.IPAddress([]byte{1, 2, 3, 4}),
				},
			},
			setupMocks: func(dns *mocks.MockIPResolver, pc *mocks.MockMyPacketConn) {
				pc.EXPECT().WritePacket(gomock.Any()).Return(nil)
			},
		},
		{
			name: "domain target - DNS resolution failed",
			inputPacket: &udp.Packet{
				Target: net.Destination{
					Address: net.DomainAddress("example.com"),
				},
			},
			shouldError: true,
			setupMocks: func(dns *mocks.MockIPResolver, pc *mocks.MockMyPacketConn) {
				dns.EXPECT().LookupIPv4(gomock.Any(), "example.com").Return(nil, errors.New("DNS resolution failed"))
			},
		},
		{
			name: "domain target - no IPs returned",
			inputPacket: &udp.Packet{
				Target: net.Destination{
					Address: net.DomainAddress("example.com"),
				},
			},
			shouldError: true,
			setupMocks: func(dns *mocks.MockIPResolver, pc *mocks.MockMyPacketConn) {
				dns.EXPECT().LookupIPv4(gomock.Any(), "example.com").Return([]net.IP{}, nil)
				dns.EXPECT().LookupIPv6(gomock.Any(), "example.com").Return([]net.IP{}, nil)
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDNS := mocks.NewMockIPResolver(ctrl)
			mockPacketConn := mocks.NewMockMyPacketConn(ctrl)
			tt.setupMocks(mockDNS, mockPacketConn)

			conn := &domainToIpPacketConn{
				UdpConn:    mockPacketConn,
				domainToIp: make(map[net.Address]net.Address),
				ipToDomain: make(map[net.Address]net.Address),
				dns:        mockDNS,
				ctx:        context.Background(),
			}

			err := conn.WritePacket(tt.inputPacket)
			if tt.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPacket.Target.Address, tt.inputPacket.Target.Address)
		})
	}
}

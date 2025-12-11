package util

import (
	"context"
	"net"
	"testing"

	commonnet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowHandlerToDialer_Dial_TCP(t *testing.T) {
	// Create a mock flow handler
	mockHandler := &mocks.LoopbackHandler{}

	dialer := &FlowHandlerToDialer{FlowHandler: mockHandler}

	dest := commonnet.TCPDestination(commonnet.ParseAddress("192.168.1.1"), 80)

	conn, err := dialer.Dial(context.Background(), dest)
	require.NoError(t, err)
	require.NotNil(t, conn)

	linkConn, ok := conn.(*pipe.LinkConn)
	require.True(t, ok)

	// Verify addresses are set correctly
	localAddr := linkConn.LocalAddr().(*net.TCPAddr)
	remoteAddr := linkConn.RemoteAddr().(*net.TCPAddr)

	assert.Equal(t, "0.0.0.0", localAddr.IP.String())
	assert.Equal(t, 0, localAddr.Port)
	assert.Equal(t, "192.168.1.1", remoteAddr.IP.String())
	assert.Equal(t, 80, remoteAddr.Port)
}

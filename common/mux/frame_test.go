package mux

import (
	"bytes"
	"testing"

	"github.com/5vnetwork/vx-core/common/bitmask"
	"github.com/5vnetwork/vx-core/common/buf"
	nethelper "github.com/5vnetwork/vx-core/common/net"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameMetadata_MarshalUnmarshal_SessionStatusNew_TCP(t *testing.T) {
	original := FrameMetadata{
		SessionID:     1234,
		SessionStatus: SessionStatusNew,
		Option:        OptionData,
		Target:        nethelper.TCPDestination(nethelper.ParseAddress("192.168.1.1"), nethelper.Port(8080)),
	}

	// Marshal
	b := buf.New()
	defer b.Release()
	err := original.WriteTo(b)
	require.NoError(t, err)

	// Unmarshal
	reader := bytes.NewReader(b.Bytes())
	var decoded FrameMetadata
	err = decoded.Unmarshal(reader)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, original.SessionID, decoded.SessionID)
	assert.Equal(t, original.SessionStatus, decoded.SessionStatus)
	assert.Equal(t, original.Option, decoded.Option)
	assert.Equal(t, original.Target.Network, decoded.Target.Network)
	assert.Equal(t, original.Target.Address.String(), decoded.Target.Address.String())
	assert.Equal(t, original.Target.Port, decoded.Target.Port)
}

func TestFrameMetadata_MarshalUnmarshal_SessionStatusNew_UDP(t *testing.T) {
	original := FrameMetadata{
		SessionID:     5678,
		SessionStatus: SessionStatusNew,
		Option:        OptionData | OptionError,
		Target:        nethelper.UDPDestination(nethelper.ParseAddress("example.com"), nethelper.Port(53)),
	}

	// Marshal
	b := buf.New()
	defer b.Release()
	err := original.WriteTo(b)
	require.NoError(t, err)

	// Unmarshal
	reader := bytes.NewReader(b.Bytes())
	var decoded FrameMetadata
	err = decoded.Unmarshal(reader)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, original.SessionID, decoded.SessionID)
	assert.Equal(t, original.SessionStatus, decoded.SessionStatus)
	assert.Equal(t, original.Option, decoded.Option)
	assert.Equal(t, original.Target.Network, decoded.Target.Network)
	assert.Equal(t, original.Target.Address.String(), decoded.Target.Address.String())
	assert.Equal(t, original.Target.Port, decoded.Target.Port)
}

func TestFrameMetadata_MarshalUnmarshal_SessionStatusKeep(t *testing.T) {
	original := FrameMetadata{
		SessionID:     999,
		SessionStatus: SessionStatusKeep,
		Option:        OptionData,
	}

	// Marshal
	b := buf.New()
	defer b.Release()
	err := original.WriteTo(b)
	require.NoError(t, err)

	// Unmarshal
	reader := bytes.NewReader(b.Bytes())
	var decoded FrameMetadata
	err = decoded.Unmarshal(reader)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, original.SessionID, decoded.SessionID)
	assert.Equal(t, original.SessionStatus, decoded.SessionStatus)
	assert.Equal(t, original.Option, decoded.Option)
}

func TestFrameMetadata_MarshalUnmarshal_SessionStatusEnd(t *testing.T) {
	original := FrameMetadata{
		SessionID:     777,
		SessionStatus: SessionStatusEnd,
		Option:        OptionError,
	}

	// Marshal
	b := buf.New()
	defer b.Release()
	err := original.WriteTo(b)
	require.NoError(t, err)

	// Unmarshal
	reader := bytes.NewReader(b.Bytes())
	var decoded FrameMetadata
	err = decoded.Unmarshal(reader)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, original.SessionID, decoded.SessionID)
	assert.Equal(t, original.SessionStatus, decoded.SessionStatus)
	assert.Equal(t, original.Option, decoded.Option)
}

func TestFrameMetadata_MarshalUnmarshal_SessionStatusKeepAlive(t *testing.T) {
	original := FrameMetadata{
		SessionID:     111,
		SessionStatus: SessionStatusKeepAlive,
		Option:        0,
	}

	// Marshal
	b := buf.New()
	defer b.Release()
	err := original.WriteTo(b)
	require.NoError(t, err)

	// Unmarshal
	reader := bytes.NewReader(b.Bytes())
	var decoded FrameMetadata
	err = decoded.Unmarshal(reader)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, original.SessionID, decoded.SessionID)
	assert.Equal(t, original.SessionStatus, decoded.SessionStatus)
	assert.Equal(t, original.Option, decoded.Option)
}

func TestFrameMetadata_IPv6Address(t *testing.T) {
	original := FrameMetadata{
		SessionID:     2468,
		SessionStatus: SessionStatusNew,
		Option:        OptionData,
		Target:        nethelper.TCPDestination(nethelper.ParseAddress("2001:db8::1"), nethelper.Port(443)),
	}

	// Marshal
	b := buf.New()
	defer b.Release()
	err := original.WriteTo(b)
	require.NoError(t, err)

	// Unmarshal
	reader := bytes.NewReader(b.Bytes())
	var decoded FrameMetadata
	err = decoded.Unmarshal(reader)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, original.SessionID, decoded.SessionID)
	assert.Equal(t, original.Target.Network, decoded.Target.Network)
	assert.Equal(t, original.Target.Address.String(), decoded.Target.Address.String())
	assert.Equal(t, original.Target.Port, decoded.Target.Port)
}

func TestFrameMetadata_OptionFlags(t *testing.T) {
	tests := []struct {
		name   string
		option bitmask.Byte
	}{
		{"NoOptions", 0},
		{"OnlyData", OptionData},
		{"OnlyError", OptionError},
		{"Both", OptionData | OptionError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := FrameMetadata{
				SessionID:     100,
				SessionStatus: SessionStatusKeep,
				Option:        tt.option,
			}

			b := buf.New()
			defer b.Release()
			err := original.WriteTo(b)
			require.NoError(t, err)

			reader := bytes.NewReader(b.Bytes())
			var decoded FrameMetadata
			err = decoded.Unmarshal(reader)
			require.NoError(t, err)

			assert.Equal(t, tt.option, decoded.Option)
			assert.Equal(t, tt.option.Has(OptionData), decoded.Option.Has(OptionData))
			assert.Equal(t, tt.option.Has(OptionError), decoded.Option.Has(OptionError))
		})
	}
}

func TestFrameMetadata_MetadataTooLong(t *testing.T) {
	// Create a frame with a very long domain name to exceed 512 bytes
	longDomain := make([]byte, 300)
	for i := range longDomain {
		longDomain[i] = 'a'
	}

	original := FrameMetadata{
		SessionID:     1,
		SessionStatus: SessionStatusNew,
		Option:        OptionData,
		Target:        nethelper.TCPDestination(nethelper.DomainAddress(string(longDomain)), nethelper.Port(80)),
	}

	b := buf.New()
	defer b.Release()
	err := original.WriteTo(b)

	// Should return error due to long domain
	assert.Error(t, err)
}

func TestFrameMetadata_UnmarshalInvalidMetaLen(t *testing.T) {
	// Create invalid metadata with metaLen > 512
	b := buf.New()
	defer b.Release()

	// Write invalid metaLen (600 > 512)
	b.Extend(2)
	b.Bytes()[0] = 0x02 // 600 in big endian
	b.Bytes()[1] = 0x58

	reader := bytes.NewReader(b.Bytes())
	var meta FrameMetadata
	err := meta.Unmarshal(reader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid metalen")
}

func TestFrameMetadata_UnmarshalInsufficientBuffer(t *testing.T) {
	// Create metadata with insufficient data
	b := buf.New()
	defer b.Release()

	// Write valid metaLen but not enough data
	b.Extend(2)
	b.Bytes()[0] = 0x00
	b.Bytes()[1] = 0x10 // 16 bytes expected

	// Only write 2 bytes of data (need at least 4)
	b.Extend(2)

	reader := bytes.NewReader(b.Bytes())
	var meta FrameMetadata
	err := meta.Unmarshal(reader)

	assert.Error(t, err)
}

func TestFrameMetadata_AllSessionStatuses(t *testing.T) {
	statuses := []SessionStatus{
		SessionStatusNew,
		SessionStatusKeep,
		SessionStatusEnd,
		SessionStatusKeepAlive,
	}

	for _, status := range statuses {
		t.Run(string(rune(status)), func(t *testing.T) {
			original := FrameMetadata{
				SessionID:     123,
				SessionStatus: status,
				Option:        OptionData,
			}

			if status == SessionStatusNew {
				original.Target = nethelper.TCPDestination(nethelper.ParseAddress("1.2.3.4"), nethelper.Port(80))
			}

			b := buf.New()
			defer b.Release()
			err := original.WriteTo(b)
			require.NoError(t, err)

			reader := bytes.NewReader(b.Bytes())
			var decoded FrameMetadata
			err = decoded.Unmarshal(reader)
			require.NoError(t, err)

			assert.Equal(t, status, decoded.SessionStatus)
		})
	}
}

package kcp

import (
	"crypto/cipher"

	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/transport/headers"
)

const protocolName = "mkcp"

// GetMTUValue returns the value of MTU settings.
func GetMTUValue(c *KcpConfig) uint32 {
	if c == nil || c.GetMtu() == 0 {
		return 1350
	}
	return c.GetMtu()
}

// GetTTIValue returns the value of TTI settings.
func GetTTIValue(c *KcpConfig) uint32 {
	if c == nil || c.GetTti() == 0 {
		return 50
	}
	return c.GetTti()
}

// GetUplinkCapacityValue returns the value of UplinkCapacity settings.
func GetUplinkCapacityValue(c *KcpConfig) uint32 {
	if c == nil || c.GetUplinkCapacity() == 0 {
		return 5
	}
	return c.GetUplinkCapacity()
}

// GetDownlinkCapacityValue returns the value of DownlinkCapacity settings.
func GetDownlinkCapacityValue(c *KcpConfig) uint32 {
	if c == nil || c.GetDownlinkCapacity() == 0 {
		return 20
	}
	return c.GetDownlinkCapacity()
}

// GetWriteBufferSize returns the size of WriterBuffer in bytes.
func GetWriteBufferSize(c *KcpConfig) uint32 {
	if c == nil || c.GetWriteBuffer() == 0 {
		return 2 * 1024 * 1024
	}
	return c.GetWriteBuffer()
}

// GetReadBufferSize returns the size of ReadBuffer in bytes.
func GetReadBufferSize(c *KcpConfig) uint32 {
	if c == nil || c.GetReadBuffer() == 0 {
		return 2 * 1024 * 1024
	}
	return c.GetReadBuffer()
}

// GetSecurity returns the security settings.
func GetSecurity(c *KcpConfig) (cipher.AEAD, error) {
	if c.GetSeed() != "" {
		return NewAEADAESGCMBasedOnSeed(c.GetSeed()), nil
	}
	return NewSimpleAuthenticator(), nil
}

func GetPackerHeader(c *KcpConfig) (headers.PacketHeader, error) {
	if c.GetHeaderConfig() != nil {
		rawConfig, err := serial.GetInstanceOf(c.GetHeaderConfig())
		if err != nil {
			return nil, err
		}

		return headers.CreatePacketHeader(rawConfig)
	}
	return nil, nil
}

func GetSendingInFlightSize(c *KcpConfig) uint32 {
	size := GetUplinkCapacityValue(c) * 1024 * 1024 / GetMTUValue(c) / (1000 / GetTTIValue(c))
	if size < 8 {
		size = 8
	}
	return size
}

func GetSendingBufferSize(c *KcpConfig) uint32 {
	return GetWriteBufferSize(c) / GetMTUValue(c)
}

func GetReceivingInFlightSize(c *KcpConfig) uint32 {
	size := GetDownlinkCapacityValue(c) * 1024 * 1024 / GetMTUValue(c) / (1000 / GetTTIValue(c))
	if size < 8 {
		size = 8
	}
	return size
}

func GetReceivingBufferSize(c *KcpConfig) uint32 {
	return GetReadBufferSize(c) / GetMTUValue(c)
}

package kcp

import (
	"crypto/cipher"

	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/transport/headers"
)

const protocolName = "mkcp"

// GetMTUValue returns the value of MTU settings.
func (c *KcpConfig) GetMTUValue() uint32 {
	if c == nil || c.Mtu == 0 {
		return 1350
	}
	return c.Mtu
}

// GetTTIValue returns the value of TTI settings.
func (c *KcpConfig) GetTTIValue() uint32 {
	if c == nil || c.Tti == 0 {
		return 50
	}
	return c.Tti
}

// GetUplinkCapacityValue returns the value of UplinkCapacity settings.
func (c *KcpConfig) GetUplinkCapacityValue() uint32 {
	if c == nil || c.UplinkCapacity == 0 {
		return 5
	}
	return c.UplinkCapacity
}

// GetDownlinkCapacityValue returns the value of DownlinkCapacity settings.
func (c *KcpConfig) GetDownlinkCapacityValue() uint32 {
	if c == nil || c.DownlinkCapacity == 0 {
		return 20
	}
	return c.DownlinkCapacity
}

// GetWriteBufferSize returns the size of WriterBuffer in bytes.
func (c *KcpConfig) GetWriteBufferSize() uint32 {
	if c == nil || c.WriteBuffer == 0 {
		return 2 * 1024 * 1024
	}
	return c.WriteBuffer
}

// GetReadBufferSize returns the size of ReadBuffer in bytes.
func (c *KcpConfig) GetReadBufferSize() uint32 {
	if c == nil || c.ReadBuffer == 0 {
		return 2 * 1024 * 1024
	}
	return c.ReadBuffer
}

// GetSecurity returns the security settings.
func (c *KcpConfig) GetSecurity() (cipher.AEAD, error) {
	if c.Seed != "" {
		return NewAEADAESGCMBasedOnSeed(c.Seed), nil
	}
	return NewSimpleAuthenticator(), nil
}

func (c *KcpConfig) GetPackerHeader() (headers.PacketHeader, error) {
	if c.HeaderConfig != nil {
		rawConfig, err := serial.GetInstanceOf(c.HeaderConfig)
		if err != nil {
			return nil, err
		}

		return headers.CreatePacketHeader(rawConfig)
	}
	return nil, nil
}

func (c *KcpConfig) GetSendingInFlightSize() uint32 {
	size := c.GetUplinkCapacityValue() * 1024 * 1024 / c.GetMTUValue() / (1000 / c.GetTTIValue())
	if size < 8 {
		size = 8
	}
	return size
}

func (c *KcpConfig) GetSendingBufferSize() uint32 {
	return c.GetWriteBufferSize() / c.GetMTUValue()
}

func (c *KcpConfig) GetReceivingInFlightSize() uint32 {
	size := c.GetDownlinkCapacityValue() * 1024 * 1024 / c.GetMTUValue() / (1000 / c.GetTTIValue())
	if size < 8 {
		size = 8
	}
	return size
}

func (c *KcpConfig) GetReceivingBufferSize() uint32 {
	return c.GetReadBufferSize() / c.GetMTUValue()
}

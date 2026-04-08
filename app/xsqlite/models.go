// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package xsqlite

import (
	"fmt"
	"time"

	outbound "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/outbound"
	"github.com/golang/protobuf/proto"
)

type Subscription struct {
	ID                int
	Name              string
	Link              string
	RemainingData     float64
	EndTime           int
	Website           string
	LastUpdate        int
	LastSuccessUpdate int
	Description       string
}

type OutboundHandlerGroup struct {
	Name       string `gorm:"primaryKey"`
	PlaceOnTop bool
}

type OutboundHandlerGroupRelation struct {
	GroupName string `gorm:"primaryKey"`
	HandlerId int    `gorm:"primaryKey"`
}

type GeoDomain struct {
	ID            int `gorm:"primaryKey;autoIncrement"`
	GeoDomain     []byte
	DomainSetName string
}

type AtomicDomainSet struct {
	Name           string `gorm:"primaryKey"`
	GeositeConfig  []byte
	UseBloomFilter bool
	ClashRuleUrls  string
	GeoUrl         string
}

type OutboundHandler struct {
	ID          int
	Selected    bool
	CountryCode string
	Ok          int
	Speed       float64
	// in seconds
	SpeedTestTime    int
	Ping             int
	PingTestTime     int
	SubId            *int
	Config           []byte
	Sni              string
	ServerIp         string
	Support6         int
	Support6TestTime int
}

// if last test time is more than 10 minutes ago, return true
func (h *OutboundHandler) IsSpeedDataOld() bool {
	return time.Now().Unix()-int64(h.SpeedTestTime) > 10*60
}

func (h *OutboundHandler) GetSpeed() float64 {
	if h.IsSpeedDataOld() {
		return 0
	}
	return h.Speed
}

func (h *OutboundHandler) GetSupport6() int {
	if h.IsSupport6DataOld() {
		return 0
	}
	return h.Support6
}

// if last test time is more than 10 minutes ago, return true
func (h *OutboundHandler) IsPingDataOld() bool {
	return time.Now().Unix()-int64(h.PingTestTime) > 10*60
}

func (h *OutboundHandler) GetPing() int {
	if h.IsPingDataOld() {
		return 0
	}
	return h.Ping
}

func (h *OutboundHandler) IsSupport6DataOld() bool {
	return time.Now().Unix()-int64(h.Support6TestTime) > 60*60*6
}

func (h *OutboundHandler) GetTag() string {
	var config outbound.HandlerConfig
	err := proto.Unmarshal(h.Config, &config)
	if err != nil {
		return ""
	}
	if config.GetOutbound() != nil {
		return config.GetOutbound().Tag
	} else if config.GetChain() != nil {
		return config.GetChain().Tag
	}
	return ""
}

func (h *OutboundHandler) ToConfig() *outbound.HandlerConfig {
	var config outbound.HandlerConfig
	err := proto.Unmarshal(h.Config, &config)
	if err != nil {
		return nil
	}
	if config.GetOutbound() != nil {
		config.GetOutbound().Tag = fmt.Sprintf("%d", h.ID)
	} else if config.GetChain() != nil {
		config.GetChain().Tag = fmt.Sprintf("%d", h.ID)
	}
	// config.Ping = uint64(h.Ping)
	// config.Throughput = uint64(h.Speed * 1024 * 1024 / 8)
	return &config
}

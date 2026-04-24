// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package subscription

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	outbound "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/outbound"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/app/util/sub"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type SubscriptionManager struct {
	Running           bool
	Timer             *time.Timer
	Interval          time.Duration
	Downloader        downloader
	Db                *gorm.DB
	OnUpdatedCallback func()
	AutoUpdate        bool
}

type downloader interface {
	Download(ctx context.Context, url string, headers map[string]string) ([]byte, http.Header, error)
}

type SubscriptionOption func(*SubscriptionManager)

type SubscriptionManagerConfig struct {
	Interval          time.Duration
	Db                *gorm.DB
	Downloader        downloader
	OnUpdatedCallback func()
	AutoUpdate        bool
}

func NewSubscriptionManager(config *SubscriptionManagerConfig) *SubscriptionManager {
	s := &SubscriptionManager{
		Db:                config.Db,
		Interval:          config.Interval,
		Downloader:        config.Downloader,
		OnUpdatedCallback: config.OnUpdatedCallback,
		AutoUpdate:        config.AutoUpdate,
	}

	return s
}

func (s *SubscriptionManager) Start() error {
	if s.AutoUpdate && !s.Running {
		log.Debug().Msg("subscription manager start")
		s.Running = true
		s.periodicUpdate()
	}
	return nil
}

func (s *SubscriptionManager) Close() error {
	log.Debug().Msg("subscription manager close")
	if s.Timer != nil {
		s.Timer.Stop()
		s.Timer = nil
	}
	s.Running = false
	return nil
}

func (s *SubscriptionManager) GetLastUpdate() time.Time {
	// find the Subscription with the oldest LastUpdate
	var sub *xsqlite.Subscription
	err := s.Db.Order("last_update ASC").First(&sub).Error
	if err != nil {
		log.Error().Err(err).Msg("failed to get last update")
		return time.Time{}
	}
	return time.UnixMilli(int64(sub.LastUpdate))
}

func (s *SubscriptionManager) periodicUpdate() {
	var count int64
	// if the subscriptions table exists
	if s.Db.Migrator().HasTable(&xsqlite.Subscription{}) {
		s.Db.Model(&xsqlite.Subscription{}).Count(&count)
	}

	var lastUpdate time.Time
	if count != 0 {
		lastUpdate = s.GetLastUpdate()
	} else {
		lastUpdate = time.Now()
	}
	nextUpdateTime := lastUpdate.Add(s.Interval)

	now := time.Now()
	if !nextUpdateTime.After(now.Add(time.Millisecond * 100)) {
		go s.UpdateSubscriptions()
		nextUpdateTime = now.Add(s.Interval)
	}

	log.Debug().Str("next_update", nextUpdateTime.Local().String()).
		Msg("periodic update")

	if s.Timer != nil {
		s.Timer.Stop()
	}
	s.Timer = time.AfterFunc(time.Until(nextUpdateTime), s.periodicUpdate)
}

// just set interval, not start or stop
func (s *SubscriptionManager) SetInterval(interval time.Duration) {
	s.Interval = interval
	log.Debug().Dur("interval", interval).Msg("update interval")
	if s.Running {
		s.periodicUpdate()
	}
}

func (s *SubscriptionManager) SetAutoUpdate(autoUpdate bool) {
	s.AutoUpdate = autoUpdate
	if s.AutoUpdate && !s.Running {
		s.Start()
	} else if !s.AutoUpdate && s.Running {
		s.Close()
	}
}

func (s *SubscriptionManager) UpdateSubscriptions() error {
	UpdateSubscriptions(s.Db, s.Downloader)

	if s.OnUpdatedCallback != nil {
		s.OnUpdatedCallback()
	}
	return nil
}

type UpdateSubscriptionResult struct {
	SuccessSub   int
	SuccessNodes int
	FailedSub    int
	FailedNodes  []string
	ErrorReasons map[string]string
}

func UpdateSubscriptions(db *gorm.DB, downloader downloader) UpdateSubscriptionResult {
	log.Debug().Msg("update subscriptions")
	var subscriptions []*xsqlite.Subscription
	// load all subscriptions from database
	db.Find(&subscriptions)
	var wg sync.WaitGroup
	// s.subscriptions = make(map[int]*futureTask)
	lock := sync.Mutex{}
	result := UpdateSubscriptionResult{
		SuccessSub:   0,
		SuccessNodes: 0,
		FailedSub:    0,
		FailedNodes:  nil,
		ErrorReasons: make(map[string]string),
	}
	for _, sub := range subscriptions {
		wg.Add(1)
		go func(sub *xsqlite.Subscription) {
			defer wg.Done()
			successNodes, failedNodes, err := UpdateSubscription(sub, db, downloader)
			if err != nil {
				lock.Lock()
				result.FailedSub++
				result.ErrorReasons[sub.Name] = err.Error()
				lock.Unlock()
				log.Error().Err(err).Int("id", sub.ID).Str("name", sub.Name).Str("link", sub.Link).
					Msg("update subscription failed")
			} else {
				lock.Lock()
				result.SuccessSub++
				result.SuccessNodes += successNodes
				result.FailedNodes = append(result.FailedNodes, failedNodes...)
				lock.Unlock()
			}
		}(sub)
	}
	wg.Wait()
	return result
}

type FetchSubscriptionResult struct {
	Configs     []*outbound.OutboundHandlerConfig
	FailedNodes []string
	Description string
}

func FetchSubscription(ctx context.Context, link string, downloader downloader) (*FetchSubscriptionResult, error) {
	if parsedUrl, err := url.Parse(link); err == nil {
		q := parsedUrl.Query()
		q.Set("flag", "vx")
		parsedUrl.RawQuery = q.Encode()
		link = parsedUrl.String()
	}

	var uriContent *sub.DecodeResult
	// try no user agent first
	body, header, err := downloader.Download(ctx, link, map[string]string{
		"User-Agent": "shadowrocket",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download subscription: %v", err)
	}
	uriContent, err = util.Decode(string(body))
	// if failed to decode, try again with user agent
	if err != nil || len(uriContent.Configs) == 0 {
		body, header, err = downloader.Download(ctx, link, map[string]string{
			"User-Agent": "v2ray-core",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to download subscription: %v", err)
		}
		uriContent, err = util.Decode(string(body))
		if err != nil {
			return nil, fmt.Errorf("failed to decode subscription: %v", err)
		}
	}

	description := header.Get("subscription-userinfo")
	if description == "" {
		description = uriContent.Description
	}
	// get description
	if description == "" {
		if parsedUrl1, err := url.Parse(link); err == nil {
			q := parsedUrl1.Query()
			q.Set("flag", "shadowrocket")
			parsedUrl1.RawQuery = q.Encode()
			content1, _, err := downloader.Download(ctx, parsedUrl1.String(),
				map[string]string{})
			if err == nil {
				uriContent1, err := util.Decode(string(content1))
				if err == nil {
					description = uriContent1.Description
				}
			}
		}
	}

	return &FetchSubscriptionResult{
		Configs:     uriContent.Configs,
		FailedNodes: uriContent.FailedNodes,
		Description: description,
	}, nil
}

// return success parsed nodes, failed parsed nodes, error
// error means cannot get data from server
func UpdateSubscription(subscription *xsqlite.Subscription, db *gorm.DB, downloader downloader) (int, []string, error) {
	logger := log.With().Int("id", subscription.ID).Str("name", subscription.Name).Str("link", subscription.Link).Logger()
	ctx := logger.WithContext(context.Background())
	logger.Debug().Msg("start")

	subscription.LastUpdate = int(time.Now().UnixMilli())
	db.Model(subscription).Update("last_update", subscription.LastUpdate)

	link := subscription.Link
	// add vx flag
	if parsedUrl, err := url.Parse(link); err == nil {
		q := parsedUrl.Query()
		q.Set("flag", "vx")
		parsedUrl.RawQuery = q.Encode()
		link = parsedUrl.String()
	}

	var uriContent *sub.DecodeResult
	// try no user agent first
	body, header, err := downloader.Download(ctx, link, map[string]string{
		"User-Agent": "shadowrocket",
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to download subscription: %v", err)
	}
	uriContent, err = util.Decode(string(body))
	// if failed to decode, try again with user agent
	if err != nil || len(uriContent.Configs) == 0 {
		body, header, err = downloader.Download(ctx, link, map[string]string{
			"User-Agent": "v2ray-core",
		})
		if err != nil {
			return 0, nil, fmt.Errorf("failed to download subscription: %v", err)
		}
		uriContent, err = util.Decode(string(body))
		if err != nil {
			return 0, nil, fmt.Errorf("failed to decode subscription: %v", err)
		}
	}

	subscription.Description = header.Get("subscription-userinfo")
	if subscription.Description == "" {
		subscription.Description = uriContent.Description
	}
	// get all handlers of current subscription
	var existingHandlers []*xsqlite.OutboundHandler
	db.Where("sub_id = ?", subscription.ID).Find(&existingHandlers)
	var updatedHandlers []*xsqlite.OutboundHandler

	id := rand.Intn(math.MaxInt)

	for _, config := range uriContent.Configs {
		existing := false
		for _, existingHandler := range existingHandlers {
			// try find if there is any existing handler with the same config
			// if bytes.Equal(configBytes, existingHandler.Config) {
			// 	logger.Debug().Str("existing_handler", config.Tag).
			// 		Msg("replace existing handler's config")
			// 	db.Model(&existingHandler).Update("config", configBytes)
			// 	updatedHandlers = append(updatedHandlers, existingHandler)
			// 	existing = true
			// 	break
			// }
			var existingConfig outbound.HandlerConfig
			err := proto.Unmarshal(existingHandler.Config, &existingConfig)
			if err == nil && existingConfig.GetOutbound().GetTag() == config.Tag {
				logger.Debug().Str("existing_handler", existingConfig.GetOutbound().GetTag()).
					Msg("replace existing handler's config")
				// update existing handler's config
				config.EnableMux = existingConfig.GetOutbound().EnableMux
				config.Uot = existingConfig.GetOutbound().Uot
				config.DomainStrategy = existingConfig.GetOutbound().DomainStrategy
				configBytes, err := proto.Marshal(&outbound.HandlerConfig{
					Type: &outbound.HandlerConfig_Outbound{
						Outbound: config,
					},
				})
				if err != nil {
					fmt.Printf("Failed to marshal config: %v\n", err)
					break
				}
				db.Model(&existingHandler).Update("config", configBytes)
				updatedHandlers = append(updatedHandlers, existingHandler)
				existing = true
				break
			}
		}
		if !existing {
			configBytes, err := proto.Marshal(&outbound.HandlerConfig{
				Type: &outbound.HandlerConfig_Outbound{
					Outbound: config,
				},
			})
			if err != nil {
				fmt.Printf("Failed to marshal config: %v\n", err)
				continue
			}
			newHandler := xsqlite.OutboundHandler{
				ID:     id,
				Config: configBytes,
				SubId:  &subscription.ID,
			}
			id++
			// add new handler to database
			db.Create(&newHandler)
		}
	}

	//TODO: make this a preference
	// delete handlers that are not in the new configs
	for _, existingHandler := range existingHandlers {
		if !slices.Contains(updatedHandlers, existingHandler) {
			if err := db.Delete(existingHandler).Error; err != nil {
				log.Error().Err(err).Msg("failed to delete handler")
			}
		}
	}

	// get description
	if subscription.Description == "" {
		if parsedUrl1, err := url.Parse(subscription.Link); err == nil {
			q := parsedUrl1.Query()
			q.Set("flag", "shadowrocket")
			parsedUrl1.RawQuery = q.Encode()
			content1, _, err := downloader.Download(ctx, parsedUrl1.String(),
				map[string]string{})
			if err == nil {
				uriContent1, err := util.Decode(string(content1))
				if err == nil {
					subscription.Description = uriContent1.Description
				}
			}
		}
	}

	subscription.LastSuccessUpdate = subscription.LastUpdate
	if err := db.Model(subscription).Updates(map[string]interface{}{
		"last_success_update": subscription.LastSuccessUpdate,
		"description":         subscription.Description,
	}).Error; err != nil {
		log.Error().Err(err).Msg("failed to update subscription")
	}
	logger.Debug().Msg("done")
	return len(uriContent.Configs), uriContent.FailedNodes, nil
}

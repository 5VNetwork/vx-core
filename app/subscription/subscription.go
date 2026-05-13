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
	"github.com/5vnetwork/vx-core/app/util/uri"
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

func FetchSubscription(ctx context.Context, link string, downloader downloader, shareLinkQueryExtra map[string]string) (*FetchSubscriptionResult, error) {
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
	uriContent, err = util.Decode(string(body), shareLinkQueryExtra)
	// if failed to decode, try again with user agent
	if err != nil || len(uriContent.Configs) == 0 {
		body, header, err = downloader.Download(ctx, link, map[string]string{
			"User-Agent": "v2ray-core",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to download subscription: %v", err)
		}
		uriContent, err = util.Decode(string(body), shareLinkQueryExtra)
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
				uriContent1, err := util.Decode(string(content1), shareLinkQueryExtra)
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
//
// All database writes after the network download happen inside a single
// transaction that first refetches the subscription by ID. This prevents the
// race where the user (Flutter side) deletes a subscription while this update
// is in flight: without the refetch, this function would happily insert new
// outbound_handlers rows whose sub_id points at a row that just got deleted,
// either failing with a foreign-key violation or producing orphan handlers
// that block the user's next delete attempt.
func UpdateSubscription(subscription *xsqlite.Subscription, db *gorm.DB, downloader downloader) (int, []string, error) {
	logger := log.With().Int("id", subscription.ID).Str("name", subscription.Name).Str("link", subscription.Link).Logger()
	ctx := logger.WithContext(context.Background())
	logger.Debug().Msg("start")

	// Mark "we attempted now" before downloading so that periodicUpdate doesn't
	// immediately retry this subscription on the next tick if the download
	// stalls. Use an explicit Where so a zero ID can't accidentally update all
	// rows, and check RowsAffected to detect a sub that's already been deleted.
	now := int(time.Now().UnixMilli())
	preflight := db.Model(&xsqlite.Subscription{}).Where("id = ?", subscription.ID).Update("last_update", now)
	if preflight.Error != nil {
		return 0, nil, fmt.Errorf("failed to mark last_update for subscription %d: %w", subscription.ID, preflight.Error)
	}
	if preflight.RowsAffected == 0 {
		logger.Debug().Msg("subscription no longer exists, skipping update")
		return 0, nil, fmt.Errorf("subscription %d no longer exists", subscription.ID)
	}
	subscription.LastUpdate = now

	link := subscription.Link
	// add vx flag
	if parsedUrl, err := url.Parse(link); err == nil {
		q := parsedUrl.Query()
		q.Set("flag", "vx")
		parsedUrl.RawQuery = q.Encode()
		link = parsedUrl.String()
	}

	// Network download happens OUTSIDE the transaction so we don't hold a
	// SQLite write lock during slow network I/O.
	var uriContent *sub.DecodeResult
	// try no user agent first
	body, header, err := downloader.Download(ctx, link, map[string]string{
		"User-Agent": "shadowrocket",
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to download subscription: %v", err)
	}
	uriContent, err = util.Decode(string(body), sub.ShareLinkQueryExtraFromStored(subscription.ShareLinkQueryExtra))
	// if failed to decode, try again with user agent
	if err != nil || len(uriContent.Configs) == 0 {
		body, header, err = downloader.Download(ctx, link, map[string]string{
			"User-Agent": "v2ray-core",
		})
		if err != nil {
			return 0, nil, fmt.Errorf("failed to download subscription: %v", err)
		}
		uriContent, err = util.Decode(string(body), sub.ShareLinkQueryExtraFromStored(subscription.ShareLinkQueryExtra))
		if err != nil {
			return 0, nil, fmt.Errorf("failed to decode subscription: %v", err)
		}
	}

	description := header.Get("subscription-userinfo")
	if description == "" {
		description = uriContent.Description
	}
	// fallback description fetch is also network I/O; keep it outside the tx
	if description == "" {
		if parsedUrl1, err := url.Parse(subscription.Link); err == nil {
			q := parsedUrl1.Query()
			q.Set("flag", "shadowrocket")
			parsedUrl1.RawQuery = q.Encode()
			content1, _, err := downloader.Download(ctx, parsedUrl1.String(),
				map[string]string{})
			if err == nil {
				uriContent1, err := util.Decode(string(content1), sub.ShareLinkQueryExtraFromStored(subscription.ShareLinkQueryExtra))
				if err == nil {
					description = uriContent1.Description
				}
			}
		}
	}

	txErr := db.Transaction(func(tx *gorm.DB) error {
		// Refetch under the write lock: if the user just deleted the
		// subscription, abort instead of inserting handlers that would
		// either violate FK or become orphans.
		var fresh xsqlite.Subscription
		if err := tx.First(&fresh, subscription.ID).Error; err != nil {
			return fmt.Errorf("subscription %d no longer exists: %w", subscription.ID, err)
		}

		var existingHandlers []*xsqlite.OutboundHandler
		if err := tx.Where("sub_id = ?", fresh.ID).Find(&existingHandlers).Error; err != nil {
			return fmt.Errorf("failed to load existing handlers: %w", err)
		}
		var updatedHandlers []*xsqlite.OutboundHandler

		id := rand.Intn(math.MaxInt)

		for _, config := range uriContent.Configs {
			existing := false
			for _, existingHandler := range existingHandlers {
				var existingConfig outbound.HandlerConfig
				err := proto.Unmarshal(existingHandler.Config, &existingConfig)
				if err == nil && existingConfig.GetOutbound().GetTag() == config.Tag &&
					existingConfig.GetOutbound().GetAddress() == config.Address &&
					existingConfig.GetOutbound().GetPort() == config.Port &&
					uri.PortRangesToString(existingConfig.GetOutbound().GetPorts()) == uri.PortRangesToString(config.GetPorts()) &&
					existingConfig.GetOutbound().GetProtocol().GetTypeUrl() == config.GetProtocol().GetTypeUrl() {
					logger.Debug().Str("existing_handler", existingConfig.GetOutbound().GetTag()).
						Msg("replace existing handler's config")
					config.EnableMux = existingConfig.GetOutbound().EnableMux
					config.Uot = existingConfig.GetOutbound().Uot
					config.DomainStrategy = existingConfig.GetOutbound().DomainStrategy
					configBytes, err := proto.Marshal(&outbound.HandlerConfig{
						Type: &outbound.HandlerConfig_Outbound{
							Outbound: config,
						},
					})
					if err != nil {
						logger.Error().Err(err).Msg("failed to marshal config")
						break
					}
					if err := tx.Model(&existingHandler).Update("config", configBytes).Error; err != nil {
						return fmt.Errorf("failed to update handler %d: %w", existingHandler.ID, err)
					}
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
					logger.Error().Err(err).Msg("failed to marshal config")
					continue
				}
				newHandler := xsqlite.OutboundHandler{
					ID:     id,
					Config: configBytes,
					SubId:  &fresh.ID,
				}
				id++
				if err := tx.Create(&newHandler).Error; err != nil {
					return fmt.Errorf("failed to create handler: %w", err)
				}
			}
		}

		//TODO: make this a preference
		// delete handlers that are not in the new configs
		for _, existingHandler := range existingHandlers {
			if !slices.Contains(updatedHandlers, existingHandler) {
				if err := tx.Delete(existingHandler).Error; err != nil {
					return fmt.Errorf("failed to delete handler %d: %w", existingHandler.ID, err)
				}
			}
		}

		if err := tx.Model(&xsqlite.Subscription{}).Where("id = ?", fresh.ID).Updates(map[string]interface{}{
			"last_success_update": now,
			"description":         description,
		}).Error; err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}

		// keep the caller's struct in sync with what we just wrote.
		subscription.LastSuccessUpdate = now
		subscription.Description = description
		return nil
	})
	if txErr != nil {
		logger.Error().Err(txErr).Msg("update subscription transaction failed")
		return 0, nil, txErr
	}

	logger.Debug().Msg("done")
	return len(uriContent.Configs), uriContent.FailedNodes, nil
}

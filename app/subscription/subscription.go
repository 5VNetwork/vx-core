package subscription

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/subscription/decode"
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
	Download(ctx context.Context, url string) ([]byte, error)
}

type futureTask struct {
	timer *time.Timer
}

type SubscriptionOption func(*SubscriptionManager)

func WithOnUpdatedCallback(callback func()) SubscriptionOption {
	return func(s *SubscriptionManager) {
		s.OnUpdatedCallback = callback
	}
}

func WithDownloader(downloader downloader) SubscriptionOption {
	return func(s *SubscriptionManager) {
		s.Downloader = downloader
	}
}

func WithPeriodicUpdate(periodicUpdate bool) SubscriptionOption {
	return func(s *SubscriptionManager) {
		s.AutoUpdate = periodicUpdate
	}
}

func NewSubscriptionManager(interval time.Duration, db *gorm.DB, downloader downloader,
	opts ...SubscriptionOption) *SubscriptionManager {
	s := &SubscriptionManager{
		Db:         db,
		Interval:   interval,
		Downloader: downloader,
	}
	for _, opt := range opts {
		opt(s)
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

	if nextUpdateTime.Before(time.Now()) || time.Until(nextUpdateTime) < time.Minute {
		go s.UpdateSubscriptions()
		nextUpdateTime = time.Now().Add(s.Interval)
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
				result.SuccessSub++
				result.SuccessNodes += successNodes
				result.FailedNodes = append(result.FailedNodes, failedNodes...)
			}
		}(sub)
	}
	wg.Wait()
	return result
}

// return success parsed nodes, failed parsed nodes, error
// error means cannot get data from server
func UpdateSubscription(sub *xsqlite.Subscription, db *gorm.DB, downloader downloader) (int, []string, error) {
	logger := log.With().Int("id", sub.ID).Str("name", sub.Name).Str("link", sub.Link).Logger()
	ctx := logger.WithContext(context.Background())
	logger.Debug().Msg("start")

	sub.LastUpdate = int(time.Now().UnixMilli())
	db.Model(sub).Update("last_update", sub.LastUpdate)

	link := sub.Link
	// add vx flag
	if parsedUrl, err := url.Parse(link); err == nil {
		q := parsedUrl.Query()
		q.Set("flag", "vx")
		parsedUrl.RawQuery = q.Encode()
		link = parsedUrl.String()
	}

	content, err := downloader.Download(ctx, link)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to download subscription: %v", err)
	}
	uriContent, err := decode.Decode(string(content))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to decode subscription: %v", err)
	}
	sub.Description = uriContent.Description

	// get all handlers of current subscription
	var existingHandlers []*xsqlite.OutboundHandler
	db.Where("sub_id = ?", sub.ID).Find(&existingHandlers)
	var updatedHandlers []*xsqlite.OutboundHandler

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
			var existingConfig configs.HandlerConfig
			err := proto.Unmarshal(existingHandler.Config, &existingConfig)
			if err == nil && existingConfig.GetOutbound().GetTag() == config.Tag {
				logger.Debug().Str("existing_handler", existingConfig.GetOutbound().GetTag()).
					Msg("replace existing handler's config")
				// update existing handler's config
				config.EnableMux = existingConfig.GetOutbound().EnableMux
				config.Uot = existingConfig.GetOutbound().Uot
				config.DomainStrategy = existingConfig.GetOutbound().DomainStrategy
				configBytes, err := proto.Marshal(&configs.HandlerConfig{
					Type: &configs.HandlerConfig_Outbound{
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
			configBytes, err := proto.Marshal(&configs.HandlerConfig{
				Type: &configs.HandlerConfig_Outbound{
					Outbound: config,
				},
			})
			if err != nil {
				fmt.Printf("Failed to marshal config: %v\n", err)
				continue
			}
			newHandler := xsqlite.OutboundHandler{
				ID:     rand.Intn(math.MaxInt),
				Config: configBytes,
				SubId:  &sub.ID,
			}
			// add new handler to database
			db.Create(&newHandler)
		}
	}

	//TODO: make this a preference
	// delete handlers that are not in the new configs
	for _, existingHandler := range existingHandlers {
		if !slices.Contains(updatedHandlers, existingHandler) {
			db.Delete(existingHandler)
		}
	}

	// get description
	if sub.Description == "" {
		if parsedUrl1, err := url.Parse(sub.Link); err == nil {
			q := parsedUrl1.Query()
			q.Set("flag", "shadowrocket")
			parsedUrl1.RawQuery = q.Encode()
			content1, err := downloader.Download(ctx, parsedUrl1.String())
			if err == nil {
				uriContent1, err := decode.Decode(string(content1))
				if err == nil {
					sub.Description = uriContent1.Description
				}
			}
		}
	}

	sub.LastSuccessUpdate = sub.LastUpdate
	db.Model(sub).Updates(map[string]interface{}{
		"last_success_update": sub.LastSuccessUpdate,
		"description":         sub.Description,
	})
	logger.Debug().Msg("done")
	return len(uriContent.Configs), uriContent.FailedNodes, nil
}

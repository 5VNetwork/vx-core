package subscription_test

import (
	"context"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/5vnetwork/vx-core/app/subscription"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockDownloader implements the downloader interface for testing
type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) Download(ctx context.Context, url string, headers map[string]string) ([]byte, http.Header, error) {
	args := m.Called(url, headers)
	return args.Get(0).([]byte), args.Get(1).(http.Header), args.Error(2)
}

type funcDownloader struct {
	download func(ctx context.Context, url string, headers map[string]string) ([]byte, http.Header, error)
}

func (f funcDownloader) Download(ctx context.Context, url string, headers map[string]string) ([]byte, http.Header, error) {
	return f.download(ctx, url, headers)
}

const testSubscriptionContent = "vless://12345678-1234-1234-1234-123456789012@a.b.com:12345?encryption=none&flow=xtls-rprx-vision&fp=chrome&network=ws&path=%2Fvvv&security=tls&sni=a.b.com#test-vless"

func setupTestDB(t *testing.T, name string) *gorm.DB {
	if _, err := os.Stat(name); err == nil {
		os.Remove(name)
	}
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&xsqlite.Subscription{}, &xsqlite.OutboundHandler{})
	assert.NoError(t, err)

	return db
}

func removeTestDB(t *testing.T, name string) {
	os.Remove(name)
}

func TestNewSubscriptionManager(t *testing.T) {
	db := setupTestDB(t, "test1.db")
	defer removeTestDB(t, "test1.db")
	mockDownloader := new(MockDownloader)

	// Test with default options
	manager := NewSubscriptionManager(&SubscriptionManagerConfig{
		Interval:   5 * time.Minute,
		Db:         db,
		Downloader: mockDownloader,
		AutoUpdate: true,
	})
	assert.NotNil(t, manager)
	assert.Equal(t, 5*time.Minute, manager.Interval)
	assert.NotNil(t, manager.Db)
	assert.NotNil(t, manager.Downloader)

	// Test with callback option
	callbackCalled := false
	callback := func() { callbackCalled = true }
	manager = NewSubscriptionManager(&SubscriptionManagerConfig{
		Interval:          5 * time.Minute,
		Db:                db,
		Downloader:        mockDownloader,
		AutoUpdate:        true,
		OnUpdatedCallback: callback,
	})
	assert.NotNil(t, manager.OnUpdatedCallback)
	manager.OnUpdatedCallback()
	assert.True(t, callbackCalled)
}

func TestSubscriptionManager_ChangeInterval(t *testing.T) {
	db := setupTestDB(t, "test4.db")
	defer removeTestDB(t, "test4.db")
	mockDownloader := new(MockDownloader)
	manager := NewSubscriptionManager(&SubscriptionManagerConfig{
		Interval:   5 * time.Minute,
		Db:         db,
		Downloader: mockDownloader,
		AutoUpdate: true,
	})

	// Test changing interval
	manager.SetInterval(10 * time.Minute)
	assert.Equal(t, 10*time.Minute, manager.Interval)
}

func TestSubscriptionManager_Close(t *testing.T) {
	db := setupTestDB(t, "test5.db")
	defer removeTestDB(t, "test5.db")
	mockDownloader := new(MockDownloader)
	manager := NewSubscriptionManager(&SubscriptionManagerConfig{
		Interval:   5 * time.Minute,
		Db:         db,
		Downloader: mockDownloader,
		AutoUpdate: true,
	})

	// Start the manager
	err := manager.Start()
	assert.NoError(t, err)

	// Close the manager
	err = manager.Close()
	assert.NoError(t, err)
	assert.False(t, manager.Running)
	assert.Nil(t, manager.Timer)
}

func TestSubscriptionManager_GetLastUpdate(t *testing.T) {
	db := setupTestDB(t, "test6.db")
	defer removeTestDB(t, "test6.db")
	mockDownloader := new(MockDownloader)
	manager := NewSubscriptionManager(&SubscriptionManagerConfig{
		Interval:   5 * time.Minute,
		Db:         db,
		Downloader: mockDownloader,
		AutoUpdate: true,
	})

	// Test with no subscriptions
	lastUpdate := manager.GetLastUpdate()
	assert.True(t, lastUpdate.IsZero())

	// Create a test subscription
	sub := &xsqlite.Subscription{
		Name:       "Test Sub",
		Link:       "http://test.com",
		LastUpdate: int(time.Now().UnixMilli()),
	}
	db.Create(sub)

	// Test with one subscription
	lastUpdate = manager.GetLastUpdate()
	assert.False(t, lastUpdate.IsZero())
}

func TestSubscriptionManager_StartDoesNotUpdateFreshSubscriptionImmediately(t *testing.T) {
	db := setupTestDB(t, "test7.db")
	defer removeTestDB(t, "test7.db")

	sub := &xsqlite.Subscription{
		Name:       "Fresh Sub",
		Link:       "https://example.com/subscription",
		LastUpdate: int(time.Now().UnixMilli()),
	}
	assert.NoError(t, db.Create(sub).Error)

	var callbackCount int32
	manager := NewSubscriptionManager(&SubscriptionManagerConfig{
		Interval: 200 * time.Millisecond,
		Db:       db,
		Downloader: funcDownloader{
			download: func(ctx context.Context, url string, headers map[string]string) ([]byte, http.Header, error) {
				return []byte(testSubscriptionContent), http.Header{}, nil
			},
		},
		AutoUpdate: true,
		OnUpdatedCallback: func() {
			atomic.AddInt32(&callbackCount, 1)
		},
	})
	defer manager.Close()

	assert.NoError(t, manager.Start())

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(&callbackCount))

	time.Sleep(250 * time.Millisecond)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&callbackCount), int32(1))
}

func TestSubscriptionManager_StartUpdatesStaleSubscriptionImmediately(t *testing.T) {
	db := setupTestDB(t, "test8.db")
	defer removeTestDB(t, "test8.db")

	sub := &xsqlite.Subscription{
		Name:       "Stale Sub",
		Link:       "https://example.com/subscription",
		LastUpdate: int(time.Now().Add(-time.Hour).UnixMilli()),
	}
	assert.NoError(t, db.Create(sub).Error)

	var callbackCount int32
	manager := NewSubscriptionManager(&SubscriptionManagerConfig{
		Interval: 200 * time.Millisecond,
		Db:       db,
		Downloader: funcDownloader{
			download: func(ctx context.Context, url string, headers map[string]string) ([]byte, http.Header, error) {
				return []byte(testSubscriptionContent), http.Header{}, nil
			},
		},
		AutoUpdate: true,
		OnUpdatedCallback: func() {
			atomic.AddInt32(&callbackCount, 1)
		},
	})
	defer manager.Close()

	assert.NoError(t, manager.Start())

	time.Sleep(80 * time.Millisecond)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&callbackCount), int32(1))
}

func TestSubscriptionManager_SetAutoUpdateStartsAndStopsManager(t *testing.T) {
	db := setupTestDB(t, "test9.db")
	defer removeTestDB(t, "test9.db")

	manager := NewSubscriptionManager(&SubscriptionManagerConfig{
		Interval:   200 * time.Millisecond,
		Db:         db,
		Downloader: funcDownloader{},
		AutoUpdate: false,
	})

	assert.False(t, manager.Running)

	manager.SetAutoUpdate(true)
	assert.True(t, manager.Running)
	assert.NotNil(t, manager.Timer)

	manager.SetAutoUpdate(false)
	assert.False(t, manager.Running)
	assert.Nil(t, manager.Timer)
}

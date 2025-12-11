package subscription_test

import (
	"context"
	"os"
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

func (m *MockDownloader) Download(ctx context.Context, url string) ([]byte, error) {
	args := m.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

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
	manager := NewSubscriptionManager(5*time.Minute, db, mockDownloader)
	assert.NotNil(t, manager)
	assert.Equal(t, 5*time.Minute, manager.Interval)
	assert.NotNil(t, manager.Db)
	assert.NotNil(t, manager.Downloader)

	// Test with callback option
	callbackCalled := false
	callback := func() { callbackCalled = true }
	manager = NewSubscriptionManager(5*time.Minute, db, mockDownloader, WithOnUpdatedCallback(callback))
	assert.NotNil(t, manager.OnUpdatedCallback)
	manager.OnUpdatedCallback()
	assert.True(t, callbackCalled)
}

func TestSubscriptionManager_ChangeInterval(t *testing.T) {
	db := setupTestDB(t, "test4.db")
	defer removeTestDB(t, "test4.db")
	mockDownloader := new(MockDownloader)
	manager := NewSubscriptionManager(5*time.Minute, db, mockDownloader)

	// Test changing interval
	manager.SetInterval(10 * time.Minute)
	assert.Equal(t, 10*time.Minute, manager.Interval)
}

func TestSubscriptionManager_Close(t *testing.T) {
	db := setupTestDB(t, "test5.db")
	defer removeTestDB(t, "test5.db")
	mockDownloader := new(MockDownloader)
	manager := NewSubscriptionManager(5*time.Minute, db, mockDownloader)

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
	manager := NewSubscriptionManager(5*time.Minute, db, mockDownloader)

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

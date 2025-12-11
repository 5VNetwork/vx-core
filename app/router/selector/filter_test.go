package selector

import (
	"fmt"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockDb is a mock implementation of the Db interface for testing
type MockDb struct {
	mock.Mock
}

func (m *MockDb) GetAllHandlers() ([]*xsqlite.OutboundHandler, error) {
	args := m.Called()
	return args.Get(0).([]*xsqlite.OutboundHandler), args.Error(1)
}

func (m *MockDb) GetHandlersByGroup(group string) ([]*xsqlite.OutboundHandler, error) {
	args := m.Called(group)
	return args.Get(0).([]*xsqlite.OutboundHandler), args.Error(1)
}

func (m *MockDb) GetBatchedHandlers(batchSize int, offset int) ([]*xsqlite.OutboundHandler, error) {
	args := m.Called(batchSize, offset)
	return args.Get(0).([]*xsqlite.OutboundHandler), args.Error(1)
}

func (m *MockDb) GetHandler(id int) *xsqlite.OutboundHandler {
	args := m.Called(id)
	return args.Get(0).(*xsqlite.OutboundHandler)
}

// createTestHandler creates a test handler with the given parameters
func createTestHandler(id int, tag string, subId *int, support6 int, selected bool) *xsqlite.OutboundHandler {
	// Create a simple config with the tag
	config := &configs.HandlerConfig{
		Type: &configs.HandlerConfig_Outbound{
			Outbound: &configs.OutboundHandlerConfig{
				Tag: tag,
			},
		},
	}

	// Serialize the config
	configBytes, err := proto.Marshal(config)
	if err != nil {
		panic("failed to marshal config")
	}

	return &xsqlite.OutboundHandler{
		ID:       id,
		Config:   configBytes,
		SubId:    subId,
		Support6: support6,
		Selected: selected,
		Ok:       1,
		Speed:    100.0,
		Ping:     50,
	}
}

func TestFilter_GetHandlers_All(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	// Create test handlers
	allHandlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "handler1", nil, 0, false),
		createTestHandler(2, "handler2", nil, 1, false),
		createTestHandler(3, "handler3", nil, 0, true),
	}

	mockDb.On("GetAllHandlers").Return(allHandlers, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		All: true,
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 3, "Should return all handlers when All is true")
}

func TestFilter_GetHandlers_PrefixFilter(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	// Create test handlers with different prefixes
	batchHandlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "proxy_handler1", nil, 0, false),
		createTestHandler(2, "proxy_handler2", nil, 1, false),
		createTestHandler(3, "direct_handler1", nil, 0, false),
		createTestHandler(4, "block_handler1", nil, 1, false),
	}

	// Mock the batched calls - first batch returns all handlers, second batch returns empty
	mockDb.On("GetBatchedHandlers", 100, 0).Return(batchHandlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 100).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		Prefixes: []string{"proxy_", "direct_"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 3, "Should return handlers matching the prefixes")

	// Verify only matching handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	assert.True(t, handlerIds[1], "Should include proxy_handler1")
	assert.True(t, handlerIds[2], "Should include proxy_handler2")
	assert.True(t, handlerIds[3], "Should include direct_handler1")
	assert.False(t, handlerIds[4], "Should not include block_handler1")
}

func TestFilter_GetHandlers_TagFilter(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	batchHandlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "handler1", nil, 0, false),
		createTestHandler(2, "handler2", nil, 1, false),
		createTestHandler(3, "handler3", nil, 0, false),
		createTestHandler(4, "handler4", nil, 1, false),
	}

	mockDb.On("GetBatchedHandlers", 100, 0).Return(batchHandlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 100).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		Tags: []string{"handler1", "handler3"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 2, "Should return handlers matching the exact tags")

	// Verify only matching handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	assert.True(t, handlerIds[1], "Should include handler1")
	assert.True(t, handlerIds[3], "Should include handler3")
	assert.False(t, handlerIds[2], "Should not include handler2")
	assert.False(t, handlerIds[4], "Should not include handler4")
}

func TestFilter_GetHandlers_GroupFilter(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	// Create test handlers for groups
	group1Handlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "handler1", nil, 0, false),
		createTestHandler(2, "handler2", nil, 1, false),
	}

	mockDb.On("GetHandlersByGroup", "group1").Return(group1Handlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 0).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		GroupTags: []string{"group1"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 2, "Should return handlers from the specified group")

	// Verify only group1 handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	assert.True(t, handlerIds[1], "Should include handler1 from group1")
	assert.True(t, handlerIds[2], "Should include handler2 from group1")
}

func TestFilter_GetHandlers_MultipleGroups(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	group1Handlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "handler1", nil, 0, false),
		createTestHandler(2, "handler2", nil, 1, false),
	}

	group2Handlers := []*xsqlite.OutboundHandler{
		createTestHandler(3, "handler3", nil, 0, false),
		createTestHandler(4, "handler4", nil, 1, false),
	}

	mockDb.On("GetHandlersByGroup", "group1").Return(group1Handlers, nil)
	mockDb.On("GetHandlersByGroup", "group2").Return(group2Handlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 0).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		GroupTags: []string{"group1", "group2"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 4, "Should return handlers from all specified groups")

	// Verify all handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	assert.True(t, handlerIds[1])
	assert.True(t, handlerIds[2])
	assert.True(t, handlerIds[3])
	assert.True(t, handlerIds[4])
}

func TestFilter_GetHandlers_HandlerIdsFilter(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	batchHandlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "handler1", nil, 0, false),
		createTestHandler(2, "handler2", nil, 1, false),
		createTestHandler(3, "handler3", nil, 0, false),
		createTestHandler(4, "handler4", nil, 1, false),
	}

	mockDb.On("GetBatchedHandlers", 100, 0).Return(batchHandlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 100).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		HandlerIds: []int64{1, 3},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 2, "Should return handlers matching the handler IDs")

	// Verify only matching handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	assert.True(t, handlerIds[1], "Should include handler with ID 1")
	assert.True(t, handlerIds[3], "Should include handler with ID 3")
	assert.False(t, handlerIds[2], "Should not include handler with ID 2")
	assert.False(t, handlerIds[4], "Should not include handler with ID 4")
}

func TestFilter_GetHandlers_SubIdFilter(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	subId1 := 100
	subId2 := 200

	batchHandlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "handler1", &subId1, 0, false),
		createTestHandler(2, "handler2", &subId2, 1, false),
		createTestHandler(3, "handler3", nil, 0, false),
		createTestHandler(4, "handler4", &subId1, 1, false),
	}

	mockDb.On("GetBatchedHandlers", 100, 0).Return(batchHandlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 100).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		SubIds: []int64{100, 200},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 3, "Should return handlers matching the sub IDs")

	// Verify only matching handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	assert.True(t, handlerIds[1], "Should include handler1 with subId 100")
	assert.True(t, handlerIds[2], "Should include handler2 with subId 200")
	assert.True(t, handlerIds[4], "Should include handler4 with subId 100")
	assert.False(t, handlerIds[3], "Should not include handler3 with nil subId")
}

func TestFilter_GetHandlers_SelectedFilter(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	batchHandlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "handler1", nil, 0, true),
		createTestHandler(2, "handler2", nil, 1, false),
		createTestHandler(3, "handler3", nil, 0, true),
		createTestHandler(4, "handler4", nil, 1, false),
	}

	mockDb.On("GetBatchedHandlers", 100, 0).Return(batchHandlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 100).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		Selected: true,
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 2, "Should return only selected handlers")

	// Verify only selected handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	assert.True(t, handlerIds[1], "Should include selected handler1")
	assert.True(t, handlerIds[3], "Should include selected handler3")
	assert.False(t, handlerIds[2], "Should not include unselected handler2")
	assert.False(t, handlerIds[4], "Should not include unselected handler4")
}

func TestFilter_GetHandlers_CombinedFilters(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	subId1 := 100
	subId2 := 200

	// Create handlers for groups
	group1Handlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "proxy_handler1", &subId1, 0, false),
		createTestHandler(3, "direct_handler1", &subId1, 0, false),
	}

	// Create handlers for batch processing
	batchHandlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "proxy_handler1", &subId1, 0, false),
		createTestHandler(2, "proxy_handler2", &subId2, 1, false),
		createTestHandler(3, "direct_handler1", &subId1, 0, false),
		createTestHandler(4, "block_handler1", nil, 1, false),
	}

	mockDb.On("GetHandlersByGroup", "group1").Return(group1Handlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 0).Return(batchHandlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 100).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		Prefixes:   []string{"proxy_"},
		Tags:       []string{"direct_handler1"},
		GroupTags:  []string{"group1"},
		HandlerIds: []int64{2},
		SubIds:     []int64{100},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 3, "Should return handlers matching any of the filter criteria")

	// Verify matching handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	assert.True(t, handlerIds[1], "Should include proxy_handler1 (matches prefix, group, and subId)")
	assert.True(t, handlerIds[2], "Should include proxy_handler2 (matches prefix and handlerId)")
	assert.True(t, handlerIds[3], "Should include direct_handler1 (matches tag, group, and subId)")
	assert.False(t, handlerIds[4], "Should not include block_handler1 (matches nothing)")
}

func TestFilter_GetHandlers_BatchProcessing(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	// Create handlers for first batch
	firstBatch := make([]*xsqlite.OutboundHandler, 100)
	for i := 0; i < 100; i++ {
		firstBatch[i] = createTestHandler(i+1, fmt.Sprintf("handler%d", i+1), nil, i%2, false)
	}

	// Create handlers for second batch
	secondBatch := make([]*xsqlite.OutboundHandler, 50)
	for i := 0; i < 50; i++ {
		secondBatch[i] = createTestHandler(i+101, fmt.Sprintf("handler%d", i+101), nil, i%2, false)
	}

	mockDb.On("GetBatchedHandlers", 100, 0).Return(firstBatch, nil)
	mockDb.On("GetBatchedHandlers", 100, 100).Return(secondBatch, nil)
	mockDb.On("GetBatchedHandlers", 100, 200).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		Prefixes: []string{"handler"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 150, "Should return all handlers when processing in batches")

	// Verify all handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	for i := 1; i <= 150; i++ {
		assert.True(t, handlerIds[i], "Should include handler%d", i)
	}
}

func TestFilter_GetHandlers_NoMatchingHandlers(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	batchHandlers := []*xsqlite.OutboundHandler{
		createTestHandler(1, "handler1", nil, 0, false),
		createTestHandler(2, "handler2", nil, 1, false),
	}

	mockDb.On("GetBatchedHandlers", 100, 0).Return(batchHandlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 100).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		Prefixes: []string{"nonexistent_"},
		Tags:     []string{"nonexistent_tag"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 0, "Should return empty slice when no handlers match")
}

func TestFilter_GetHandlers_EmptyBatches(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	mockDb.On("GetBatchedHandlers", 100, 0).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		Prefixes: []string{"handler"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 0, "Should return empty slice when no handlers exist")
}

func TestFilter_GetHandlers_DuplicateHandlers(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	// Create handlers that will appear in both group and batch results
	duplicateHandler := createTestHandler(1, "handler1", nil, 0, false)

	group1Handlers := []*xsqlite.OutboundHandler{
		duplicateHandler,
		createTestHandler(2, "handler2", nil, 1, false),
	}

	batchHandlers := []*xsqlite.OutboundHandler{
		duplicateHandler, // Same handler as in group
		createTestHandler(3, "handler3", nil, 0, false),
	}

	mockDb.On("GetHandlersByGroup", "group1").Return(group1Handlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 0).Return(batchHandlers, nil)
	mockDb.On("GetBatchedHandlers", 100, 100).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		GroupTags: []string{"group1"},
		Prefixes:  []string{"handler"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 3, "Should return unique handlers even when they appear in multiple sources")

	// Verify all unique handlers are returned
	handlerIds := make(map[int]bool)
	for _, h := range handlers {
		handlerIds[h.(*dbHandler).id] = true
	}
	assert.True(t, handlerIds[1])
	assert.True(t, handlerIds[2])
	assert.True(t, handlerIds[3])
}

func TestFilter_GetHandlers_ErrorRetry(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	// First two calls fail, third succeeds
	mockDb.On("GetBatchedHandlers", 100, 0).Return([]*xsqlite.OutboundHandler{}, fmt.Errorf("database error")).Times(2)
	mockDb.On("GetBatchedHandlers", 100, 0).Return([]*xsqlite.OutboundHandler{
		createTestHandler(1, "handler1", nil, 0, false),
	}, nil).Once()
	mockDb.On("GetBatchedHandlers", 100, 100).Return([]*xsqlite.OutboundHandler{}, nil)

	filterConfig := &configs.SelectorConfig_Filter{
		Prefixes: []string{"handler"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, handlers, 1, "Should succeed after retries")
}

func TestFilter_GetHandlers_ErrorExhaustion(t *testing.T) {
	// Setup
	mockDb := new(MockDb)

	// All calls fail
	mockDb.On("GetBatchedHandlers", 100, 0).Return([]*xsqlite.OutboundHandler{}, fmt.Errorf("database error")).Times(3)

	filterConfig := &configs.SelectorConfig_Filter{
		Prefixes: []string{"handler"},
	}
	filter := NewDbFilter(mockDb, filterConfig, nil, nil)

	// Act
	handlers, err := filter.GetHandlers()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, handlers)
	assert.Equal(t, "cannot get handlers", err.Error())
}

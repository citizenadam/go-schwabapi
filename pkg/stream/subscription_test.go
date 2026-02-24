package stream

import (
	"context"
	"testing"

	"github.com/citizenadam/go-schwabapi/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestNewManager(t *testing.T) {
	logger := slog.Default()
	manager := NewManager(logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.subscriptions)
	assert.Equal(t, logger, manager.logger)
}

func TestManager_RecordRequest_ADD(t *testing.T) {
	logger := slog.Default()
	manager := NewManager(logger)

	ctx := context.Background()
	req := &types.Subscription{
		Command: "ADD",
		Service: "LEVEL",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL,MSFT",
			Fields: "0,1,2",
		},
	}

	err := manager.RecordRequest(ctx, req)
	assert.NoError(t, err)

	subs := manager.GetSubscriptions()
	assert.Len(t, subs, 1)
	assert.Len(t, subs["LEVEL"], 2)
	assert.Contains(t, subs["LEVEL"], "AAPL")
	assert.Contains(t, subs["LEVEL"], "MSFT")
	assert.Equal(t, []string{"0", "1", "2"}, subs["LEVEL"]["AAPL"])
	assert.Equal(t, []string{"0", "1", "2"}, subs["LEVEL"]["MSFT"])
}

func TestManager_GetSubscriptions(t *testing.T) {
	logger := slog.Default()
	manager := NewManager(logger)

	ctx := context.Background()
	req := &types.Subscription{
		Command: "ADD",
		Service: "LEVEL",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL,MSFT",
			Fields: "0,1",
		},
	}
	err := manager.RecordRequest(ctx, req)
	require.NoError(t, err)

	subs := manager.GetSubscriptions()
	assert.NotNil(t, subs)
	assert.Len(t, subs, 1)
	assert.Len(t, subs["LEVEL"], 2)

	// Verify it's a deep copy (modifying returned map shouldn't affect internal state)
	subs["LEVEL"]["GOOGL"] = []string{"0", "1"}
	subs2 := manager.GetSubscriptions()
	assert.NotContains(t, subs2["LEVEL"], "GOOGL")
}

func TestManager_GetSubscriptions_Empty(t *testing.T) {
	logger := slog.Default()
	manager := NewManager(logger)

	subs := manager.GetSubscriptions()
	assert.NotNil(t, subs)
	assert.Empty(t, subs)
}

func TestManager_Clear(t *testing.T) {
	logger := slog.Default()
	manager := NewManager(logger)

	ctx := context.Background()
	req := &types.Subscription{
		Command: "ADD",
		Service: "LEVEL",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL,MSFT",
			Fields: "0,1",
		},
	}
	err := manager.RecordRequest(ctx, req)
	require.NoError(t, err)

	// Verify subscriptions exist
	subs := manager.GetSubscriptions()
	assert.Len(t, subs, 1)

	// Clear all subscriptions
	manager.Clear()

	// Verify all cleared
	subs = manager.GetSubscriptions()
	assert.Empty(t, subs)
}

func TestManager_MultipleServices(t *testing.T) {
	logger := slog.Default()
	manager := NewManager(logger)

	ctx := context.Background()

	// Add subscriptions for different services
	req1 := &types.Subscription{
		Command: "ADD",
		Service: "LEVEL",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL",
			Fields: "0,1",
		},
	}
	err := manager.RecordRequest(ctx, req1)
	require.NoError(t, err)

	req2 := &types.Subscription{
		Command: "ADD",
		Service: "QUOTE",
		Parameters: &types.SubscriptionParams{
			Keys:   "MSFT",
			Fields: "0,1,2",
		},
	}
	err = manager.RecordRequest(ctx, req2)
	require.NoError(t, err)

	subs := manager.GetSubscriptions()
	assert.Len(t, subs, 2)
	assert.Contains(t, subs, "LEVEL")
	assert.Contains(t, subs, "QUOTE")
	assert.Len(t, subs["LEVEL"], 1)
	assert.Len(t, subs["QUOTE"], 1)
}

func TestManager_ConcurrentAccess(t *testing.T) {
	logger := slog.Default()
	manager := NewManager(logger)

	ctx := context.Background()
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(i int) {
			req := &types.Subscription{
				Command: "ADD",
				Service: "LEVEL",
				Parameters: &types.SubscriptionParams{
					Keys:   "AAPL",
					Fields: "0,1",
				},
			}
			_ = manager.RecordRequest(ctx, req)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.GetSubscriptions()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestManager_EmptyKeys(t *testing.T) {
	logger := slog.Default()
	manager := NewManager(logger)

	ctx := context.Background()
	req := &types.Subscription{
		Command: "ADD",
		Service: "LEVEL",
		Parameters: &types.SubscriptionParams{
			Keys:   "",
			Fields: "0,1",
		},
	}

	err := manager.RecordRequest(ctx, req)
	assert.NoError(t, err)

	subs := manager.GetSubscriptions()
	// Service map is created but no keys are added when keys are empty
	assert.Len(t, subs["LEVEL"], 0)
}

func TestManager_EmptyFields(t *testing.T) {
	logger := slog.Default()
	manager := NewManager(logger)

	ctx := context.Background()
	req := &types.Subscription{
		Command: "ADD",
		Service: "LEVEL",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL",
			Fields: "",
		},
	}

	err := manager.RecordRequest(ctx, req)
	assert.NoError(t, err)

	subs := manager.GetSubscriptions()
	assert.Len(t, subs["LEVEL"], 1)
	assert.Empty(t, subs["LEVEL"]["AAPL"])
}

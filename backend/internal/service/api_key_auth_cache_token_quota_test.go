package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyAuthSnapshotTokenQuotaRoundtrip(t *testing.T) {
	groupID := int64(50)
	dailyLimit := int64(1_000)
	weeklyLimit := int64(10_000)
	apiKey := &APIKey{
		ID:      82,
		UserID:  40,
		GroupID: &groupID,
		Key:     "sk-token-quota-roundtrip",
		Status:  StatusActive,
		User:    &User{ID: 40, Status: StatusActive},
		Group: &Group{
			ID:               groupID,
			Name:             "token-quota-roundtrip",
			Platform:         PlatformAnthropic,
			Status:           StatusActive,
			SubscriptionType: SubscriptionTypeSubscription,
			DailyTokenLimit:  &dailyLimit,
			WeeklyTokenLimit: &weeklyLimit,
		},
	}
	service := &APIKeyService{}

	payload, err := json.Marshal(&APIKeyAuthCacheEntry{Snapshot: service.snapshotFromAPIKey(context.Background(), apiKey)})
	require.NoError(t, err)

	var cached APIKeyAuthCacheEntry
	require.NoError(t, json.Unmarshal(payload, &cached))

	materialized, used, err := service.applyAuthCacheEntry(apiKey.Key, &cached)
	require.NoError(t, err)
	require.True(t, used)
	require.NotNil(t, materialized.Group)
	require.NotNil(t, materialized.Group.DailyTokenLimit)
	require.Equal(t, dailyLimit, *materialized.Group.DailyTokenLimit)
	require.NotNil(t, materialized.Group.WeeklyTokenLimit)
	require.Equal(t, weeklyLimit, *materialized.Group.WeeklyTokenLimit)
	require.Nil(t, materialized.Group.MonthlyTokenLimit)
}

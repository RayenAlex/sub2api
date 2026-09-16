//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGetByKeyForAuthCarriesGroupTokenQuotaLimits(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewAPIKeyRepository(client)
	nonce := time.Now().UnixNano()
	dailyLimit := int64(1_000)
	weeklyLimit := int64(2_000)

	group := mustCreateGroup(t, client, &service.Group{
		Name:             fmt.Sprintf("token-quota-projection-%d", nonce),
		Platform:         service.PlatformOpenAI,
		RateMultiplier:   1,
		Status:           service.StatusActive,
		SubscriptionType: service.SubscriptionTypeSubscription,
		DailyTokenLimit:  &dailyLimit,
		WeeklyTokenLimit: &weeklyLimit,
	})
	user := mustCreateUser(t, client, &service.User{Email: fmt.Sprintf("token-quota-projection-%d@example.com", nonce)})
	groupID := group.ID
	keyValue := fmt.Sprintf("sk-token-quota-projection-%d", nonce)
	key := &service.APIKey{UserID: user.ID, Key: keyValue, Name: "token-quota-projection", GroupID: &groupID, Status: service.StatusActive}
	require.NoError(t, repo.Create(ctx, key))

	got, err := repo.GetByKeyForAuth(ctx, keyValue)
	require.NoError(t, err)
	require.NotNil(t, got.Group)
	require.NotNil(t, got.Group.DailyTokenLimit)
	require.Equal(t, dailyLimit, *got.Group.DailyTokenLimit)
	require.NotNil(t, got.Group.WeeklyTokenLimit)
	require.Equal(t, weeklyLimit, *got.Group.WeeklyTokenLimit)
	require.Nil(t, got.Group.MonthlyTokenLimit)
}

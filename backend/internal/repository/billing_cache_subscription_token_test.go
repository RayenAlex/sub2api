//go:build unit

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// Seam 4: 订阅缓存 V3 —— 原子累计 USD + Token，旧缓存强制回源

func TestSubscriptionCache_UpdateIncrementsUSDAndTokens(t *testing.T) {
	c, mr := newMiniRedisCache(t)
	ctx := context.Background()
	expires := time.Now().Add(time.Hour)

	// 初始化 V3 订阅缓存
	err := c.SetSubscriptionCache(ctx, 7, 10, &service.SubscriptionCacheData{
		Status:            service.SubscriptionStatusActive,
		ExpiresAt:         expires,
		DailyUsage:        1.0,
		WeeklyUsage:       1.0,
		MonthlyUsage:      1.0,
		DailyTokenUsage:   10,
		WeeklyTokenUsage:  10,
		MonthlyTokenUsage: 10,
		SchemaVersion:     service.SubscriptionCacheSchemaV3,
	})
	require.NoError(t, err)
	require.Equal(t, "3", mr.HGet(billingSubKey(7, 10), "schema_version"),
		"new subscription caches must use schema V3")

	// 更新：USD +2.5，Token +22
	require.NoError(t, c.UpdateSubscriptionUsage(ctx, 7, 10, 2.5, 22))

	got, err := c.GetSubscriptionCache(ctx, 7, 10)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.InDelta(t, 3.5, got.DailyUsage, 1e-9)
	require.InDelta(t, 3.5, got.WeeklyUsage, 1e-9)
	require.InDelta(t, 3.5, got.MonthlyUsage, 1e-9)
	require.Equal(t, int64(32), got.DailyTokenUsage)
	require.Equal(t, int64(32), got.WeeklyTokenUsage)
	require.Equal(t, int64(32), got.MonthlyTokenUsage)
}

func TestSubscriptionCache_LegacyV1MissingTokenFieldsReturnsMiss(t *testing.T) {
	c, mr := newMiniRedisCache(t)
	ctx := context.Background()
	expires := time.Now().Add(time.Hour)

	// 模拟旧版 V1 缓存：只有 USD 字段，无 schema_version / token 字段
	key := billingSubKey(7, 10)
	mr.HSet(key, "status", service.SubscriptionStatusActive)
	mr.HSet(key, "expires_at", fmt.Sprintf("%d", expires.Unix()))
	mr.HSet(key, "daily_usage", "1.0")
	mr.HSet(key, "weekly_usage", "1.0")
	mr.HSet(key, "monthly_usage", "1.0")
	mr.HSet(key, "version", "1")

	got, err := c.GetSubscriptionCache(ctx, 7, 10)
	require.NoError(t, err)
	require.Nil(t, got, "缺少 schema_version / token 字段的旧缓存必须视为 miss")
}

// V2 caches could contain zero token counters written before database hydration was fixed.
// They must not be accepted after the cache schema is upgraded, otherwise an already
// exhausted subscription can still pass the quota preflight until the TTL expires.
func TestSubscriptionCache_V2TokenCountersReturnMissAfterHydrationFix(t *testing.T) {
	c, mr := newMiniRedisCache(t)
	ctx := context.Background()
	key := billingSubKey(7, 10)

	mr.HSet(key,
		"status", service.SubscriptionStatusActive,
		"expires_at", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()),
		"daily_usage", "0",
		"weekly_usage", "0",
		"monthly_usage", "0",
		"version", "1",
		"schema_version", "2",
		"daily_token_usage", "0",
		"weekly_token_usage", "0",
		"monthly_token_usage", "0",
	)

	got, err := c.GetSubscriptionCache(ctx, 7, 10)
	require.NoError(t, err)
	require.Nil(t, got, "V2 cache must be a miss so token usage is reloaded from the database")
}

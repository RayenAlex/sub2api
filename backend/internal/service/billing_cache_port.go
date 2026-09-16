package service

import (
	"time"
)

// SubscriptionCacheData represents cached subscription data
type SubscriptionCacheData struct {
	Status       string
	ExpiresAt    time.Time
	DailyUsage   float64
	WeeklyUsage  float64
	MonthlyUsage float64
	// Token 用量（V2 新增）
	DailyTokenUsage   int64
	WeeklyTokenUsage  int64
	MonthlyTokenUsage int64
	// SchemaVersion 缓存结构版本；缺失/旧版本视为 miss
	SchemaVersion int64
	Version       int64
}

// SubscriptionCacheSchemaV3 forces a database reload of V2 Redis entries.
// V2 entries created before token-usage hydration was fixed can contain zero
// counters even when the database shows an exhausted subscription.
const SubscriptionCacheSchemaV3 int64 = 3

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"

	"github.com/stretchr/testify/require"
)

// ---- seam 1: 分组 Token 配额领域规则 ----

func TestGroup_TokenLimitPredicates_NilIsUnlimited(t *testing.T) {
	g := &Group{}
	require.False(t, g.HasDailyTokenLimit(), "nil daily_token_limit 应表示不限")
	require.False(t, g.HasWeeklyTokenLimit(), "nil weekly_token_limit 应表示不限")
	require.False(t, g.HasMonthlyTokenLimit(), "nil monthly_token_limit 应表示不限")
}

func TestGroup_TokenLimitPredicates_ZeroIsHardDeny(t *testing.T) {
	var zero int64
	g := &Group{
		DailyTokenLimit:   &zero,
		WeeklyTokenLimit:  &zero,
		MonthlyTokenLimit: &zero,
	}
	require.True(t, g.HasDailyTokenLimit(), "0 是已启用的硬禁止上限")
	require.True(t, g.HasWeeklyTokenLimit(), "0 是已启用的硬禁止上限")
	require.True(t, g.HasMonthlyTokenLimit(), "0 是已启用的硬禁止上限")
}

func TestGroup_TokenLimitPredicates_PositiveIsLimited(t *testing.T) {
	var d, w, m int64 = 100, 500, 2000
	g := &Group{
		DailyTokenLimit:   &d,
		WeeklyTokenLimit:  &w,
		MonthlyTokenLimit: &m,
	}
	require.True(t, g.HasDailyTokenLimit())
	require.True(t, g.HasWeeklyTokenLimit())
	require.True(t, g.HasMonthlyTokenLimit())
}

// ---- seam 2: 实际 Token 计量 ----

func TestUsageBillingCommand_BillableTokens_SumsRealUsage(t *testing.T) {
	cmd := &UsageBillingCommand{
		InputTokens:         11,
		OutputTokens:        5,
		CacheCreationTokens: 4,
		CacheReadTokens:     2,
	}
	require.Equal(t, int64(22), cmd.BillableTokens(),
		"应累计 input+output+cache_creation+cache_read，不重复计入 5m/1h 明细")
}

func TestUsageBillingCommand_BillableTokens_NoTokenUsageIsZero(t *testing.T) {
	// 图像/按次计费等非 Token 请求不产生 Token 增量
	cmd := &UsageBillingCommand{
		ImageCount: 2,
	}
	require.Equal(t, int64(0), cmd.BillableTokens())
}

func TestUsageBillingCommand_BillableTokens_NegativeDefensiveToZero(t *testing.T) {
	cmd := &UsageBillingCommand{
		InputTokens:         -3,
		OutputTokens:        5,
		CacheCreationTokens: 0,
		CacheReadTokens:     0,
	}
	require.Equal(t, int64(5), cmd.BillableTokens(),
		"负的防御性 Token 字段不得使总量倒退")
}

// ---- seam 1 续: 订阅资格前置校验（错误码与拦截） ----

type tokenQuotaSubRepoStub struct {
	UserSubscriptionRepository
	sub *UserSubscription
	err error
}

func (s *tokenQuotaSubRepoStub) GetActiveByUserIDAndGroupID(_ context.Context, _, _ int64) (*UserSubscription, error) {
	return s.sub, s.err
}

func newTokenQuotaBillingSvc(t *testing.T, cache BillingCache, subRepo UserSubscriptionRepository) *BillingCacheService {
	t.Helper()
	svc := NewBillingCacheService(cache, nil, subRepo, nil, nil, nil, &config.Config{}, nil)
	t.Cleanup(svc.Stop)
	return svc
}

func TestCheckSubscriptionEligibility_DailyTokenLimitExceeded(t *testing.T) {
	var limit int64 = 100
	now := time.Now()
	svc := newTokenQuotaBillingSvc(t, nil, &tokenQuotaSubRepoStub{sub: &UserSubscription{
		ID:        1,
		GroupID:   10,
		Status:    SubscriptionStatusActive,
		ExpiresAt: now.Add(time.Hour),
		// DailyTokenUsage >= DailyTokenLimit → 拒绝
		DailyTokenUsage: 100,
	}})
	group := &Group{ID: 10, DailyTokenLimit: &limit}
	err := svc.checkSubscriptionEligibility(context.Background(), 7, group, &UserSubscription{ID: 1})
	require.ErrorIs(t, err, ErrDailyTokenLimitExceeded)
}

func TestCheckSubscriptionEligibility_ZeroTokenLimitBlocksFirstRequest(t *testing.T) {
	var zero int64
	now := time.Now()
	svc := newTokenQuotaBillingSvc(t, nil, &tokenQuotaSubRepoStub{sub: &UserSubscription{
		ID:              1,
		GroupID:         10,
		Status:          SubscriptionStatusActive,
		ExpiresAt:       now.Add(time.Hour),
		DailyTokenUsage: 0,
	}})
	group := &Group{ID: 10, DailyTokenLimit: &zero}
	err := svc.checkSubscriptionEligibility(context.Background(), 7, group, &UserSubscription{ID: 1})
	require.ErrorIs(t, err, ErrDailyTokenLimitExceeded, "limit=0 必须拒绝首个有 Token 用量的请求前置校验")
}

func TestCheckSubscriptionEligibility_TokenBelowLimitPasses(t *testing.T) {
	var limit int64 = 100
	now := time.Now()
	svc := newTokenQuotaBillingSvc(t, nil, &tokenQuotaSubRepoStub{sub: &UserSubscription{
		ID:              1,
		GroupID:         10,
		Status:          SubscriptionStatusActive,
		ExpiresAt:       now.Add(time.Hour),
		DailyTokenUsage: 99,
	}})
	group := &Group{ID: 10, DailyTokenLimit: &limit}
	err := svc.checkSubscriptionEligibility(context.Background(), 7, group, &UserSubscription{ID: 1})
	require.NoError(t, err)
}

func TestCheckSubscriptionEligibility_NilTokenLimitIsUnlimited(t *testing.T) {
	now := time.Now()
	svc := newTokenQuotaBillingSvc(t, nil, &tokenQuotaSubRepoStub{sub: &UserSubscription{
		ID:              1,
		GroupID:         10,
		Status:          SubscriptionStatusActive,
		ExpiresAt:       now.Add(time.Hour),
		DailyTokenUsage: 1 << 40,
	}})
	group := &Group{ID: 10} // 未配置 Token limit
	err := svc.checkSubscriptionEligibility(context.Background(), 7, group, &UserSubscription{ID: 1})
	require.NoError(t, err)
}

func TestCheckSubscriptionEligibility_WeeklyAndMonthlyTokenLimits(t *testing.T) {
	var wl, ml int64 = 10, 10
	now := time.Now()
	cases := []struct {
		name    string
		sub     *UserSubscription
		wantErr error
	}{
		{
			name: "weekly exceeded",
			sub: &UserSubscription{
				Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour),
				WeeklyTokenUsage: 10,
			},
			wantErr: ErrWeeklyTokenLimitExceeded,
		},
		{
			name: "monthly exceeded",
			sub: &UserSubscription{
				Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour),
				MonthlyTokenUsage: 10,
			},
			wantErr: ErrMonthlyTokenLimitExceeded,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTokenQuotaBillingSvc(t, nil, &tokenQuotaSubRepoStub{sub: tc.sub})
			group := &Group{ID: 10, WeeklyTokenLimit: &wl, MonthlyTokenLimit: &ml}
			err := svc.checkSubscriptionEligibility(context.Background(), 7, group, &UserSubscription{ID: 1})
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// ---- seam 3 续: buildUsageBillingCommand 把 UsageLog token 计入订阅命令 ----

func TestBuildUsageBillingCommand_SubscriptionCarriesPeakAdjustedTokenCount(t *testing.T) {
	groupID := int64(7)
	subID := int64(42)
	p := &postUsageBillingParams{
		Cost:               &CostBreakdown{TotalCost: 1.0, ActualCost: 3.0},
		User:               &User{ID: 1},
		APIKey:             &APIKey{ID: 2, GroupID: &groupID},
		Account:            &Account{ID: 3},
		Subscription:       &UserSubscription{ID: subID},
		IsSubscriptionBill: true,
		// 高峰 3 倍已在请求时刻计算完成，账单命令必须使用该预处理值，
		// 不能重新从 usageLog 取原始 22 tokens。
		BillableTokens: 66,
	}
	usageLog := &UsageLog{
		InputTokens:         11,
		OutputTokens:        5,
		CacheCreationTokens: 4,
		CacheReadTokens:     2,
	}

	cmd := buildUsageBillingCommand("req-1", usageLog, p)
	require.NotNil(t, cmd)
	require.Equal(t, int64(66), cmd.SubscriptionTokens, "高峰 3 倍时 22 个实际 tokens 应扣除 66 个 token 配额")
}

func TestBuildUsageBillingCommand_NonSubscriptionLeavesTokenZero(t *testing.T) {
	groupID := int64(7)
	subID := int64(42)
	p := &postUsageBillingParams{
		Cost:               &CostBreakdown{TotalCost: 1.0, ActualCost: 1.0},
		User:               &User{ID: 1},
		APIKey:             &APIKey{ID: 2, GroupID: &groupID},
		Account:            &Account{ID: 3},
		Subscription:       &UserSubscription{ID: subID},
		IsSubscriptionBill: false,
	}
	usageLog := &UsageLog{
		InputTokens:  11,
		OutputTokens: 5,
	}
	cmd := buildUsageBillingCommand("req-1", usageLog, p)
	require.NotNil(t, cmd)
	require.Equal(t, int64(0), cmd.SubscriptionTokens,
		"非订阅请求不应设置订阅 Token 增量")
}

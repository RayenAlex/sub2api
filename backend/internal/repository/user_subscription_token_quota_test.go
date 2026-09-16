package repository

import (
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

func TestUserSubscriptionEntityToServiceMapsTokenUsage(t *testing.T) {
	subscription := userSubscriptionEntityToService(&dbent.UserSubscription{
		DailyTokenUsage:   1_000,
		WeeklyTokenUsage:  2_000,
		MonthlyTokenUsage: 3_000,
	})

	require.EqualValues(t, 1_000, subscription.DailyTokenUsage)
	require.EqualValues(t, 2_000, subscription.WeeklyTokenUsage)
	require.EqualValues(t, 3_000, subscription.MonthlyTokenUsage)
}

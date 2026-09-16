package repository

import (
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

func TestGroupEntityToServiceMapsTokenQuota(t *testing.T) {
	dailyLimit := int64(50_000_000)
	weeklyLimit := int64(75_000_000)

	group := groupEntityToService(&dbent.Group{
		DailyTokenLimit:  &dailyLimit,
		WeeklyTokenLimit: &weeklyLimit,
	})

	require.NotNil(t, group.DailyTokenLimit)
	require.Equal(t, dailyLimit, *group.DailyTokenLimit)
	require.NotNil(t, group.WeeklyTokenLimit)
	require.Equal(t, weeklyLimit, *group.WeeklyTokenLimit)
	require.Nil(t, group.MonthlyTokenLimit)
}

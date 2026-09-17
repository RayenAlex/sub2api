package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionTokenQuotaConstraintRepairMigrationGuardsReplays(t *testing.T) {
	content, err := FS.ReadFile("239_ensure_subscription_token_quota_constraints.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	for _, name := range []string{
		"chk_groups_daily_token_limit_nonneg",
		"chk_groups_weekly_token_limit_nonneg",
		"chk_groups_monthly_token_limit_nonneg",
		"chk_user_subscriptions_daily_token_usage_nonneg",
		"chk_user_subscriptions_weekly_token_usage_nonneg",
		"chk_user_subscriptions_monthly_token_usage_nonneg",
	} {
		require.Contains(t, sql, "ADD CONSTRAINT "+name)
	}
	require.Contains(t, sql, "EXCEPTION WHEN duplicate_object THEN NULL")
}

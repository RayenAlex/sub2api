package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUsageLogAppliedPeakMultiplierMigrationKeepsHistoricalRowsUnmarked(t *testing.T) {
	content, err := FS.ReadFile("240_add_usage_log_applied_peak_multiplier.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS applied_peak_multiplier DECIMAL(10,4)")
	require.NotContains(t, strings.ToUpper(sql), "NOT NULL")
	require.NotContains(t, strings.ToUpper(sql), "DEFAULT 1")
}

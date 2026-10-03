package migrations

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestUsageWebSearchEventsMigration(t *testing.T) {
	content, err := FS.ReadFile("241_add_usage_web_search_events.sql")
	require.NoError(t, err)
	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "REFERENCES usage_logs(id) ON DELETE CASCADE")
	require.Contains(t, sql, "UNIQUE (usage_log_id, sequence)")
	require.Contains(t, sql, "CHECK (sequence > 0)")
	require.Contains(t, sql, "CHECK (source_count >= 0)")
	require.Contains(t, sql, "sources JSONB NOT NULL DEFAULT '[]'::jsonb")
	require.Contains(t, sql, "ON usage_web_search_events (call_id)")
	require.NotContains(t, sql, "CREATE INDEX IF NOT EXISTS idx_usage_web_search_events_usage_log_sequence")
	for _, field := range []string{"input_tokens", "output_tokens", "total_cost", "actual_cost"} {
		require.NotContains(t, sql, field)
	}
	require.NotContains(t, strings.ToUpper(sql), "ALTER TABLE usage_logs")
}

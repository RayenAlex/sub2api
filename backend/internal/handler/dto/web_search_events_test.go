package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestWebSearchEventsAdminJSONContract(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	log := &service.UsageLog{ID: 10, InputTokens: 5, TotalCost: 1, WebSearchEvents: []service.WebSearchEvent{{ID: 11, UsageLogID: 10, Sequence: 1, CallID: "call_1", Query: "private search", Status: "completed", SourceCount: 99, CreatedAt: now}, {ID: 12, Sequence: 2, Sources: []service.WebSearchSource{{URL: "https://example.com", Title: "Example"}}, CreatedAt: now}}}
	admin := UsageLogFromServiceAdmin(log)
	payload, err := json.Marshal(admin)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(payload, &decoded))
	events := decoded["web_search_events"].([]any)
	require.Len(t, events, 2)
	first := events[0].(map[string]any)
	require.Equal(t, map[string]any{"id": float64(11), "sequence": float64(1), "call_id": "call_1", "query": "private search", "status": "completed", "source_count": float64(0), "sources": []any{}, "created_at": "2026-01-02T03:04:05Z"}, first)
	second := events[1].(map[string]any)
	require.Equal(t, float64(1), second["source_count"])
	require.Equal(t, []any{map[string]any{"url": "https://example.com", "title": "Example"}}, second["sources"])
	// Explicit child DTOs must not inherit parent billing/account fields.
	for _, event := range events {
		for _, field := range []string{"input_tokens", "output_tokens", "total_cost", "actual_cost", "usage_log_id", "account_id"} {
			require.NotContains(t, event, field)
		}
	}
	require.Equal(t, 5, admin.InputTokens)
	require.Equal(t, float64(1), admin.TotalCost)
	regular, err := json.Marshal(UsageLogFromService(log))
	require.NoError(t, err)
	require.NotContains(t, string(regular), "web_search_events")
	require.NotContains(t, string(regular), "private search")
	require.NotContains(t, string(regular), "example.com")
	admin.WebSearchEvents[1].Sources[0].Title = "changed"
	require.Equal(t, "Example", log.WebSearchEvents[1].Sources[0].Title)
}

func TestWebSearchEventsAbsentPreservesOldDTO(t *testing.T) {
	for _, events := range [][]service.WebSearchEvent{nil, {}} {
		payload, err := json.Marshal(UsageLogFromServiceAdmin(&service.UsageLog{WebSearchEvents: events}))
		require.NoError(t, err)
		require.NotContains(t, string(payload), "web_search_events")
	}
	require.Nil(t, UsageLogFromServiceAdmin(nil))
}

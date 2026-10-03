package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type adminWebSearchUsageRepo struct {
	service.UsageLogRepository
}

func (*adminWebSearchUsageRepo) ListWithFilters(_ context.Context, params pagination.PaginationParams, _ usagestats.UsageLogFilters) ([]service.UsageLog, *pagination.PaginationResult, error) {
	return []service.UsageLog{{
		ID: 9, Model: "gpt-test", InputTokens: 30, OutputTokens: 10, TotalCost: 0.5, ActualCost: 0.5,
		WebSearchEvents: []service.WebSearchEvent{
			{ID: 101, Sequence: 1, CallID: "ws_1", Query: "query", Status: "completed", CreatedAt: time.Now(), Sources: []service.WebSearchSource{{URL: "https://example.com", Title: "Example"}}},
			{ID: 102, Sequence: 2, CallID: "ws_2", Query: "next", Status: "completed", CreatedAt: time.Now()},
		},
	}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func TestAdminUsageWebSearchEventsAPIContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUsageHandler(service.NewUsageService(&adminWebSearchUsageRepo{}, nil, nil, nil), nil, nil, nil)
	router := gin.New()
	router.GET("/api/v1/admin/usage", h.List)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/admin/usage?page=1&page_size=10", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	payload := recorder.Body.Bytes()
	require.Equal(t, int64(1), gjson.GetBytes(payload, "data.total").Int())
	require.Len(t, gjson.GetBytes(payload, "data.items").Array(), 1)
	parent := gjson.GetBytes(payload, "data.items.0")
	require.Equal(t, int64(30), parent.Get("input_tokens").Int())
	require.Equal(t, 0.5, parent.Get("actual_cost").Float())
	require.Len(t, parent.Get("web_search_events").Array(), 2)
	require.Equal(t, "query", parent.Get("web_search_events.0.query").String())
	require.Equal(t, int64(1), parent.Get("web_search_events.0.source_count").Int())
	require.Equal(t, "https://example.com", parent.Get("web_search_events.0.sources.0.url").String())
	require.Equal(t, "Example", parent.Get("web_search_events.0.sources.0.title").String())
	require.Equal(t, "[]", parent.Get("web_search_events.1.sources").Raw)
	require.False(t, parent.Get("web_search_events.0.input_tokens").Exists())
	require.False(t, parent.Get("web_search_events.0.actual_cost").Exists())
}

//go:build unit

package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const usageWebSearchResponseFixture = `{"id":"resp_search_usage","object":"response","model":"gpt-5.6-sol","status":"completed","output":[{"type":"web_search_call","id":"ws_one","status":"completed","action":{"type":"search","query":"example query","sources":[null,{"url":"https://example.com/docs","title":"Example docs"}]}},{"type":"web_search_call","id":"ws_two","status":"completed","action":{"query":"follow-up"}},{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Answer"}]}],"usage":{"input_tokens":30,"output_tokens":10,"input_tokens_details":{"cached_tokens":4}}}`

func usageWebSearchSSEFixture() string {
	return `data: {"type":"response.output_item.done","output_index":0,"item":{"type":"web_search_call","id":"ws_one","status":"completed","action":{"query":"example query"}}}` + "\n\n" +
		`data: {"type":"response.completed","response":` + usageWebSearchResponseFixture + "}\n\n"
}

func TestForwardWebSearchEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, passthrough := range []bool{false, true} {
		for _, wire := range []string{"json", "sse", "buffered-sse"} {
			t.Run(fmt.Sprintf("passthrough=%v/%s", passthrough, wire), func(t *testing.T) {
				payload, contentType := usageWebSearchResponseFixture, "application/json"
				if wire != "json" {
					payload, contentType = usageWebSearchSSEFixture(), "text/event-stream"
				}
				upstream := &httpUpstreamRecorder{resp: &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{contentType}},
					Body:       io.NopCloser(strings.NewReader(payload)),
				}}
				svc := newOpenAIImageGenerationControlTestService(upstream)
				c, recorder := newOpenAIImageGenerationControlTestContext(true, "codex_cli_rs/0.144.1")
				account := newOpenAIImageGenerationControlTestAccount()
				account.Extra = map[string]any{"openai_passthrough": passthrough, "openai_responses_mode": "force_responses"}
				body := []byte(fmt.Sprintf(`{"model":"gpt-5.6-sol","stream":%t,"input":"search example","tools":[{"type":"web_search"}]}`, wire == "sse"))

				result, err := svc.Forward(context.Background(), c, account, body)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Contains(t, recorder.Body.String(), "Answer")
				require.NotNil(t, result)
				require.Len(t, result.WebSearchEvents, 2)
				require.Equal(t, "example query", result.WebSearchEvents[0].Query)
				require.Equal(t, "ws_one", result.WebSearchEvents[0].CallID)
				require.Equal(t, []WebSearchSource{{URL: "https://example.com/docs", Title: "Example docs"}}, result.WebSearchEvents[0].Sources)
				require.Equal(t, 2, result.WebSearchEvents[1].Sequence)
				require.Equal(t, 30, result.Usage.InputTokens)
				require.Equal(t, 10, result.Usage.OutputTokens)
				require.Equal(t, 4, result.Usage.CacheReadInputTokens)
				require.Zero(t, result.WebSearchCalls, "display events must not activate standalone alpha/search billing")
				require.Zero(t, result.SearchCount, "display events must not activate Grok search surcharge")
			})
		}
	}
}

func TestForwardWebSearchEventsFailureHasNoUsage(t *testing.T) {
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusBadRequest,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"invalid request"},"output":[{"type":"web_search_call","id":"failed_attempt"}]}`)),
	}}
	svc := newOpenAIImageGenerationControlTestService(upstream)
	c, _ := newOpenAIImageGenerationControlTestContext(true, "codex_cli_rs/0.144.1")
	account := newOpenAIImageGenerationControlTestAccount()
	account.Extra = map[string]any{"openai_responses_mode": "force_responses"}
	result, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-5.6-sol","input":"query","stream":false}`))
	require.Error(t, err)
	require.Nil(t, result)
}

func TestRecordUsageWebSearchEventsDoesNotChangeBilling(t *testing.T) {
	var baseline *UsageLog
	for _, withEvents := range []bool{false, true} {
		repo := &openAIRecordUsageLogRepoStub{inserted: true}
		billing := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
		svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(repo, billing, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
		result := &OpenAIForwardResult{
			RequestID: "resp-search-usage",
			Model:     "gpt-5.1",
			Usage:     OpenAIUsage{InputTokens: 1000, OutputTokens: 200, CacheReadInputTokens: 100},
			Duration:  time.Second,
		}
		if withEvents {
			result.WebSearchEvents = parseWebSearchEventsFromJSONBytes([]byte(usageWebSearchResponseFixture))
		}
		err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
			Result:          result,
			APIKey:          &APIKey{ID: 1000, Group: &Group{RateMultiplier: 1}},
			User:            &User{ID: 2000},
			Account:         &Account{ID: 3000, Platform: PlatformOpenAI, Type: AccountTypeAPIKey},
			InboundEndpoint: "/v1/responses", UpstreamEndpoint: "/v1/responses",
		})
		require.NoError(t, err)
		require.Equal(t, 1, repo.calls, "only the parent is a UsageLog")
		require.Equal(t, 1, billing.calls)
		if !withEvents {
			baseline = repo.lastLog
			continue
		}
		require.Len(t, repo.lastLog.WebSearchEvents, 2)
		require.Equal(t, baseline.TotalTokens(), repo.lastLog.TotalTokens())
		require.Equal(t, baseline.BillableTokens(), repo.lastLog.BillableTokens())
		require.Equal(t, baseline.TotalCost, repo.lastLog.TotalCost)
		require.Equal(t, baseline.ActualCost, repo.lastLog.ActualCost)
		result.WebSearchEvents[0].Sources[0].Title = "modified after capture"
		require.Equal(t, "Example docs", repo.lastLog.WebSearchEvents[0].Sources[0].Title)
	}
}

func TestWebSearchEventsWSHTTPBridgeTurnIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := &httpUpstreamRecorder{}
	svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}}, httpUpstream: upstream}
	account := &Account{ID: 8001, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Concurrency: 1}
	payload := []byte(`{"type":"response.create","model":"gpt-5.6-sol","stream":true,"input":"search"}`)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)
	var first *OpenAIForwardResult
	for turn := 1; turn <= 2; turn++ {
		sse := usageWebSearchSSEFixture()
		if turn == 2 {
			sse = `data: {"type":"response.completed","response":{"id":"resp_without_search","model":"gpt-5.6-sol","usage":{"input_tokens":1,"output_tokens":2}}}` + "\n\n"
		}
		upstream.resp = &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(sse))}
		result, err := svc.proxyOpenAIWSHTTPBridgeTurn(context.Background(), c, account, "test-token", payload, len(payload), "gpt-5.6-sol", "", "", "", "", turn, func([]byte) error { return nil })
		require.NoError(t, err)
		if turn == 1 {
			require.Len(t, result.WebSearchEvents, 2)
			first = result
		} else {
			require.Empty(t, result.WebSearchEvents)
			require.Len(t, first.WebSearchEvents, 2)
			require.Equal(t, "Example docs", first.WebSearchEvents[0].Sources[0].Title)
		}
	}
}

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// A Chat Completions client did not ask for encrypted reasoning or a summary.
// Both fields are synthesized by the gateway, so an opt-in NInfer account must
// not send them to a Responses upstream that cannot represent either field.
func TestForwardAsChatCompletions_NInferResponsesCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name                string
		endpoint            string
		compat              any
		request             string
		wantInclude         bool
		wantSummary         bool
		wantReasoningEffort string
	}{
		{
			name:                "opted in chat request",
			compat:              true,
			request:             `{"model":"qwen3.8-27b","messages":[{"role":"user","content":"hello"}],"stream":true,"reasoning_effort":"medium"}`,
			wantReasoningEffort: "medium",
		},
		{
			name:                "ordinary provider unchanged",
			request:             `{"model":"qwen3.8-27b","messages":[{"role":"user","content":"hello"}],"stream":true,"reasoning_effort":"medium"}`,
			wantInclude:         true,
			wantSummary:         true,
			wantReasoningEffort: "medium",
		},
		{
			name:        "invalid opt in ignored",
			compat:      "true",
			request:     `{"model":"qwen3.8-27b","messages":[{"role":"user","content":"hello"}],"stream":true}`,
			wantInclude: true,
		},
		{
			name:        "explicit Responses include preserved",
			endpoint:    "/v1/responses",
			compat:      true,
			request:     `{"model":"qwen3.8-27b","input":"hello","include":["reasoning.encrypted_content"],"stream":true}`,
			wantInclude: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(tc.request)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			endpoint := tc.endpoint
			if endpoint == "" {
				endpoint = "/v1/chat/completions"
			}
			c.Request = httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"stop after recording request"}}`)),
			}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			account := &Account{
				ID: 89, Name: "test-ninfer-account", Platform: PlatformOpenAI,
				Type: AccountTypeAPIKey, Concurrency: 1,
				Credentials: map[string]any{"api_key": "test-only-key"},
				Extra: map[string]any{
					"openai_responses_mode":   "force_responses",
					"ninfer_responses_compat": tc.compat,
				},
			}
			var err error
			if endpoint == "/v1/responses" {
				_, err = svc.Forward(context.Background(), c, account, body)
			} else {
				_, err = svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "qwen3.8-27b")
			}
			require.Error(t, err)
			require.NotEmpty(t, upstream.lastBody)
			require.Equal(t, "https://api.openai.com/v1/responses", upstream.lastReq.URL.String())
			require.Equal(t, tc.wantInclude, gjson.GetBytes(upstream.lastBody, "include").Exists())
			require.Equal(t, tc.wantSummary, gjson.GetBytes(upstream.lastBody, "reasoning.summary").Exists())
			require.Equal(t, tc.wantReasoningEffort, gjson.GetBytes(upstream.lastBody, "reasoning.effort").String())
		})
	}
}

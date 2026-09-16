//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func newDevinConversationIdentityContext(t *testing.T, apiKeyID int64, headers map[string]string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(nil))
	for key, value := range headers {
		c.Request.Header.Set(key, value)
	}
	c.Set("api_key", &APIKey{ID: apiKeyID})
	return c
}

func TestNormalizeDevinConversationIdentityUsesPromptCacheKeyForBodyAndHeaders(t *testing.T) {
	c := newDevinConversationIdentityContext(t, 41, map[string]string{
		"session_id": "header-session-must-converge",
	})
	body := []byte(`{"model":"devin/swe-2","prompt_cache_key":"conversation-123","input":"hello"}`)

	updated, identity, err := normalizeDevinConversationIdentity(c, body)
	require.NoError(t, err)

	want := generateSessionUUID(isolateOpenAISessionID(41, "conversation-123"))
	require.Equal(t, want, identity)
	require.Equal(t, want, gjson.GetBytes(updated, "prompt_cache_key").String())

	headers := http.Header{}
	applyDevinConversationIdentityHeaders(headers, identity)
	require.Equal(t, want, headers.Get("session_id"))
	require.Equal(t, want, headers.Get("X-Session-Id"))
}

func TestNormalizeDevinConversationIdentityUsesHeaderWhenBodyKeyMissing(t *testing.T) {
	c := newDevinConversationIdentityContext(t, 42, map[string]string{
		"X-Session-Id": "conversation-from-header",
	})
	body := []byte(`{"model":"devin/swe-2-medium","messages":[{"role":"user","content":"hello"}]}`)

	updated, identity, err := normalizeDevinConversationIdentity(c, body)
	require.NoError(t, err)

	want := generateSessionUUID(isolateOpenAISessionID(42, "conversation-from-header"))
	require.Equal(t, want, identity)
	require.Equal(t, want, gjson.GetBytes(updated, "prompt_cache_key").String())
}

func TestNormalizeDevinConversationIdentityFallbackIsStableAcrossTurns(t *testing.T) {
	firstContext := newDevinConversationIdentityContext(t, 43, nil)
	laterContext := newDevinConversationIdentityContext(t, 43, nil)
	first := []byte(`{"model":"devin/swe-2","messages":[{"role":"system","content":"system"},{"role":"user","content":"first"}]}`)
	later := []byte(`{"model":"devin/swe-2","messages":[{"role":"system","content":"system"},{"role":"user","content":"first"},{"role":"assistant","content":"answer"},{"role":"user","content":"next"}]}`)

	firstBody, firstIdentity, err := normalizeDevinConversationIdentity(firstContext, first)
	require.NoError(t, err)
	laterBody, laterIdentity, err := normalizeDevinConversationIdentity(laterContext, later)
	require.NoError(t, err)

	require.NotEmpty(t, firstIdentity)
	require.Equal(t, firstIdentity, laterIdentity)
	require.Equal(t, firstIdentity, gjson.GetBytes(firstBody, "prompt_cache_key").String())
	require.Equal(t, laterIdentity, gjson.GetBytes(laterBody, "prompt_cache_key").String())
}

func TestNormalizeDevinConversationIdentityIsTenantIsolated(t *testing.T) {
	body := []byte(`{"model":"devin/swe-2","prompt_cache_key":"same-client-id","input":"hello"}`)

	_, first, err := normalizeDevinConversationIdentity(newDevinConversationIdentityContext(t, 100, nil), body)
	require.NoError(t, err)
	_, second, err := normalizeDevinConversationIdentity(newDevinConversationIdentityContext(t, 101, nil), body)
	require.NoError(t, err)

	require.NotEqual(t, first, second)
}

func TestNormalizeDevinConversationIdentityIgnoresOtherModels(t *testing.T) {
	c := newDevinConversationIdentityContext(t, 44, map[string]string{"session_id": "conversation"})
	body := []byte(`{"model":"gpt-5.6-sol","prompt_cache_key":"client-key","input":"hello"}`)

	updated, identity, err := normalizeDevinConversationIdentity(c, body)
	require.NoError(t, err)
	require.Empty(t, identity)
	require.Equal(t, body, updated)
}

func TestBuildUpstreamRequestOpenAIPassthroughInjectsDevinIdentity(t *testing.T) {
	c := newDevinConversationIdentityContext(t, 45, map[string]string{
		"X-Session-Id": "responses-conversation",
	})
	body := []byte(`{"model":"devin/swe-2","input":"hello","stream":true}`)
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}

	req, err := (&OpenAIGatewayService{}).buildUpstreamRequestOpenAIPassthrough(
		c.Request.Context(), c, account, body, "upstream-token",
	)
	require.NoError(t, err)

	want := generateSessionUUID(isolateOpenAISessionID(45, "responses-conversation"))
	require.Equal(t, want, req.Header.Get("session_id"))
	require.Equal(t, want, req.Header.Get("X-Session-Id"))
	requestBody, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.Equal(t, want, gjson.GetBytes(requestBody, "prompt_cache_key").String())
}

func TestSendCCUpstreamRequestInjectsDevinIdentity(t *testing.T) {
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"chatcmpl_test","choices":[]}`))),
	}}
	c := newDevinConversationIdentityContext(t, 46, map[string]string{
		"session_id": "chat-conversation",
	})
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	body := []byte(`{"model":"devin/swe-2-medium","messages":[{"role":"user","content":"hello"}]}`)

	resp, err := (&OpenAIGatewayService{httpUpstream: upstream}).sendCCUpstreamRequest(
		c.Request.Context(), c, account, "https://cpa.example/v1/chat/completions", body, false, "upstream-token", "", "",
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	want := generateSessionUUID(isolateOpenAISessionID(46, "chat-conversation"))
	require.Equal(t, want, upstream.lastReq.Header.Get("session_id"))
	require.Equal(t, want, upstream.lastReq.Header.Get("X-Session-Id"))
	require.Equal(t, want, gjson.GetBytes(upstream.lastBody, "prompt_cache_key").String())
}

func TestBuildUpstreamRequestOpenAIPassthroughCapturesDevinPromptProfile(t *testing.T) {
	c := newDevinConversationIdentityContext(t, 47, map[string]string{
		"X-Session-Id": "responses-profile",
	})
	body := []byte(`{
		"model":"devin/swe-2",
		"instructions":"stable system",
		"input":[{"role":"user","content":"hello"}],
		"tools":[
			{"type":"function","name":"mcp_list_servers","parameters":{"type":"object"}},
			{"type":"function","name":"mcp_list_tools","parameters":{"type":"object"}},
			{"type":"function","name":"mcp_call_tool","parameters":{"type":"object"}},
			{"type":"function","name":"mcp_read_resource","parameters":{"type":"object"}}
		]
	}`)
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			openCodeCacheDiagnosticsCredentialKey: true,
		},
	}
	cfg := &config.Config{}
	cfg.JWT.Secret = "diagnostic-secret"

	req, err := (&OpenAIGatewayService{cfg: cfg}).buildUpstreamRequestOpenAIPassthrough(
		context.Background(),
		c,
		account,
		body,
		"upstream-token",
	)
	require.NoError(t, err)
	require.NotNil(t, req)

	diagnostic := openCodeCacheDiagnosticFromContext(c)
	require.NotNil(t, diagnostic)
	require.True(t, diagnostic.LazyMCP)
	require.Equal(t, 4, diagnostic.ToolCount)
	require.NotEmpty(t, diagnostic.CanonicalSessionHMAC)
	require.Equal(t, diagnostic.CanonicalSessionHMAC, diagnostic.PromptCacheKeyHMAC)
}

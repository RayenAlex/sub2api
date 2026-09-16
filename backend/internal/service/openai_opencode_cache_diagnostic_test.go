package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBuildOpenCodeCacheDiagnosticRequiresAccountOptIn(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "stable-test-secret"}}
	account := &Account{Credentials: map[string]any{}}
	body := []byte(`{"model":"gpt-5.6-luna","messages":[{"role":"system","content":"private system instruction"},{"role":"user","content":"private user prompt"}]}`)
	headers := http.Header{openCodeSessionAffinityHeader: []string{"forwarded-session"}}

	require.Nil(t, buildOpenCodeCacheDiagnostic(cfg, account, "inbound-session", headers, body))

	account.Credentials[openCodeCacheDiagnosticsCredentialKey] = true
	diagnostic := buildOpenCodeCacheDiagnostic(cfg, account, "inbound-session", headers, body)
	require.NotNil(t, diagnostic)
	require.Equal(t, 1, diagnostic.Version)
	require.Equal(t, []string{"system", "user"}, diagnostic.MessageRoles)
	require.Equal(t, 2, diagnostic.MessageCount)
	require.NotEmpty(t, diagnostic.InboundSessionHMAC)
	require.NotEmpty(t, diagnostic.ForwardedSessionAffinityHMAC)
	require.NotEmpty(t, diagnostic.RequestBodyHMAC)
	require.NotEmpty(t, diagnostic.RequestShapeHMAC)
	require.NotEmpty(t, diagnostic.SystemPrefixHMAC)

	serialized := diagnostic.String()
	for _, sensitive := range []string{"private system instruction", "private user prompt", "inbound-session", "forwarded-session", "stable-test-secret"} {
		require.NotContains(t, serialized, sensitive)
	}
}

func TestBuildOpenCodeCacheDiagnosticSeparatesContentStructureAndToolSchema(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "stable-test-secret"}}
	account := &Account{Credentials: map[string]any{openCodeCacheDiagnosticsCredentialKey: true}}
	headers := http.Header{openCodeSessionAffinityHeader: []string{"forwarded-session"}}
	base := []byte(`{"model":"gpt-5.6-luna","messages":[{"role":"system","content":"system v1"},{"role":"user","content":"first prompt"}],"tools":[{"type":"function","function":{"name":"lookup","parameters":{"type":"object","properties":{"query":{"type":"string"}}}}}]}`)

	first := buildOpenCodeCacheDiagnostic(cfg, account, "inbound-session", headers, base)
	second := buildOpenCodeCacheDiagnostic(cfg, account, "inbound-session", headers, base)
	require.Equal(t, first, second)
	require.NotEmpty(t, first.ToolSchemaHMAC)

	changedPrompt := buildOpenCodeCacheDiagnostic(cfg, account, "inbound-session", headers, []byte(strings.Replace(string(base), "first prompt", "second prompt", 1)))
	require.NotEqual(t, first.RequestBodyHMAC, changedPrompt.RequestBodyHMAC)
	require.Equal(t, first.RequestShapeHMAC, changedPrompt.RequestShapeHMAC)
	require.Equal(t, first.SystemPrefixHMAC, changedPrompt.SystemPrefixHMAC)
	require.Equal(t, first.ToolSchemaHMAC, changedPrompt.ToolSchemaHMAC)

	changedSystem := buildOpenCodeCacheDiagnostic(cfg, account, "inbound-session", headers, []byte(strings.Replace(string(base), "system v1", "system v2", 1)))
	require.NotEqual(t, first.SystemPrefixHMAC, changedSystem.SystemPrefixHMAC)

	changedTool := buildOpenCodeCacheDiagnostic(cfg, account, "inbound-session", headers, []byte(strings.Replace(string(base), "lookup", "search", 1)))
	require.NotEqual(t, first.ToolSchemaHMAC, changedTool.ToolSchemaHMAC)
}

func TestBuildOpenCodeCacheDiagnosticProfilesDevinLazyMCPPayload(t *testing.T) {
	cfg := &config.Config{}
	cfg.JWT.Secret = "diagnostic-secret"
	account := &Account{Credentials: map[string]any{openCodeCacheDiagnosticsCredentialKey: true}}
	headers := http.Header{}
	headers.Set("session_id", "canonical-session")
	body := []byte(`{
		"model":"devin/swe-2",
		"prompt_cache_key":"canonical-session",
		"instructions":"stable system instructions",
		"input":[
			{"role":"developer","content":"stable developer prefix"},
			{"role":"user","content":"hello"}
		],
		"tools":[
			{"type":"function","name":"mcp_list_servers","parameters":{"type":"object"}},
			{"type":"function","name":"mcp_list_tools","parameters":{"type":"object"}},
			{"type":"function","name":"mcp_call_tool","parameters":{"type":"object"}},
			{"type":"function","name":"mcp_read_resource","parameters":{"type":"object"}},
			{"type":"function","name":"read","parameters":{"type":"object"}}
		]
	}`)

	diagnostic := buildOpenCodeCacheDiagnostic(cfg, account, "client-session", headers, body)
	require.NotNil(t, diagnostic)
	require.Equal(t, len(body), diagnostic.RequestBytes)
	require.Positive(t, diagnostic.SystemPrefixBytes)
	require.Positive(t, diagnostic.ToolSchemaBytes)
	require.Equal(t, 5, diagnostic.ToolCount)
	require.Equal(t, 4, diagnostic.MetaToolCount)
	require.True(t, diagnostic.LazyMCP)
	require.Equal(t, []string{"developer", "user"}, diagnostic.MessageRoles)
	require.Equal(t, 2, diagnostic.MessageCount)
	require.Equal(
		t,
		openCodeCacheDiagnosticHMAC(openCodeCacheDiagnosticKey(cfg), []byte("canonical-session")),
		diagnostic.CanonicalSessionHMAC,
	)
	require.Equal(t, diagnostic.CanonicalSessionHMAC, diagnostic.PromptCacheKeyHMAC)
}

func TestAttachOpenCodeCacheDiagnosticCopiesContextValue(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	diagnostic := &OpenCodeCacheDiagnostic{Version: 1, RequestBytes: 123}
	setOpenCodeCacheDiagnostic(c, diagnostic)

	result := attachOpenCodeCacheDiagnostic(c, &OpenAIForwardResult{RequestID: "req-1"})

	require.Same(t, diagnostic, result.CacheDiagnostic)
	require.Nil(t, attachOpenCodeCacheDiagnostic(c, nil))
}

func TestBuildOpenCodeCacheDiagnosticRecognizesChatCompletionsMetaTools(t *testing.T) {
	cfg := &config.Config{}
	cfg.JWT.Secret = "diagnostic-secret"
	account := &Account{Credentials: map[string]any{openCodeCacheDiagnosticsCredentialKey: true}}
	body := []byte(`{
		"model":"devin/swe-2",
		"messages":[{"role":"system","content":"stable system"},{"role":"user","content":"hello"}],
		"tools":[
			{"type":"function","function":{"name":"mcp_list_servers","parameters":{"type":"object"}}},
			{"type":"function","function":{"name":"mcp_list_tools","parameters":{"type":"object"}}},
			{"type":"function","function":{"name":"mcp_call_tool","parameters":{"type":"object"}}},
			{"type":"function","function":{"name":"mcp_read_resource","parameters":{"type":"object"}}}
		]
	}`)

	diagnostic := buildOpenCodeCacheDiagnostic(cfg, account, "", nil, body)

	require.NotNil(t, diagnostic)
	require.Equal(t, 4, diagnostic.ToolCount)
	require.Equal(t, 4, diagnostic.MetaToolCount)
	require.True(t, diagnostic.LazyMCP)
	require.Positive(t, diagnostic.SystemPrefixBytes)
	require.Positive(t, diagnostic.ToolSchemaBytes)
}

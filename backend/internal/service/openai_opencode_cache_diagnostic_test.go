package service

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
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

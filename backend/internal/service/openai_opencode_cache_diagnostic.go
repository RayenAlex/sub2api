package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
)

const (
	// openCodeCacheDiagnosticsCredentialKey explicitly enables bounded cache
	// diagnostics for a trusted OpenAI-compatible account. It must never be
	// inferred from the account name, URL, or numeric ID.
	openCodeCacheDiagnosticsCredentialKey = "opencode_cache_diagnostics"
	openCodeCacheDiagnosticVersion        = 1
	openCodeCacheDiagnosticDomain         = "sub2api/opencode-cache-diagnostics/v1"
	openCodeCacheDiagnosticContextKey     = "opencode_cache_diagnostic"
)

// OpenCodeCacheDiagnostic contains only non-reversible request correlation
// values. It intentionally excludes raw prompts, tool arguments, session IDs,
// credentials, cookies, and model output.
type OpenCodeCacheDiagnostic struct {
	Version                      int      `json:"version"`
	InboundSessionHMAC           string   `json:"inbound_session_hmac,omitempty"`
	ForwardedSessionAffinityHMAC string   `json:"forwarded_session_affinity_hmac,omitempty"`
	RequestBodyHMAC              string   `json:"request_body_hmac,omitempty"`
	RequestShapeHMAC             string   `json:"request_shape_hmac,omitempty"`
	SystemPrefixHMAC             string   `json:"system_prefix_hmac,omitempty"`
	ToolSchemaHMAC               string   `json:"tool_schema_hmac,omitempty"`
	MessageRoles                 []string `json:"message_roles,omitempty"`
	MessageCount                 int      `json:"message_count"`
}

// String returns the persisted JSON representation. All fields are either
// fixed metadata, role labels, counts, or HMACs; raw request data is excluded.
func (d *OpenCodeCacheDiagnostic) String() string {
	if d == nil {
		return ""
	}
	encoded, err := json.Marshal(d)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func openCodeCacheDiagnosticsEnabled(account *Account) bool {
	if account == nil || account.Credentials == nil {
		return false
	}
	enabled, _ := account.Credentials[openCodeCacheDiagnosticsCredentialKey].(bool)
	return enabled
}

func openCodeCacheDiagnosticKey(cfg *config.Config) []byte {
	if cfg == nil || strings.TrimSpace(cfg.JWT.Secret) == "" {
		return nil
	}
	mac := hmac.New(sha256.New, []byte(cfg.JWT.Secret))
	_, _ = mac.Write([]byte(openCodeCacheDiagnosticDomain))
	return mac.Sum(nil)
}

func openCodeCacheDiagnosticHMAC(key []byte, value []byte) string {
	if len(key) == 0 || len(value) == 0 {
		return ""
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(value)
	return hex.EncodeToString(mac.Sum(nil))
}

// buildOpenCodeCacheDiagnostic runs immediately before the raw Chat
// Completions request is sent to the selected upstream. body and headers are
// therefore the final outbound representations after all request rewrites and
// account header overrides.
func buildOpenCodeCacheDiagnostic(cfg *config.Config, account *Account, inboundSessionID string, headers http.Header, body []byte) *OpenCodeCacheDiagnostic {
	if !openCodeCacheDiagnosticsEnabled(account) {
		return nil
	}
	key := openCodeCacheDiagnosticKey(cfg)
	if len(key) == 0 {
		return nil
	}

	diagnostic := &OpenCodeCacheDiagnostic{
		Version:         openCodeCacheDiagnosticVersion,
		RequestBodyHMAC: openCodeCacheDiagnosticHMAC(key, body),
	}
	if value := strings.TrimSpace(inboundSessionID); value != "" {
		diagnostic.InboundSessionHMAC = openCodeCacheDiagnosticHMAC(key, []byte(value))
	}
	if headers != nil {
		if value := strings.TrimSpace(headers.Get(openCodeSessionAffinityHeader)); value != "" {
			diagnostic.ForwardedSessionAffinityHMAC = openCodeCacheDiagnosticHMAC(key, []byte(value))
		}
	}

	root, ok := decodeOpenCodeCacheDiagnosticJSON(body)
	if !ok {
		return diagnostic
	}
	messages := openCodeCacheDiagnosticMessages(root)
	diagnostic.MessageCount = len(messages)
	diagnostic.MessageRoles = openCodeCacheDiagnosticRoles(messages)
	if shape, err := json.Marshal(openCodeCacheDiagnosticShape(root)); err == nil {
		diagnostic.RequestShapeHMAC = openCodeCacheDiagnosticHMAC(key, shape)
	}
	if prefix, ok := openCodeCacheDiagnosticSystemPrefix(messages); ok {
		if encoded, err := json.Marshal(prefix); err == nil {
			diagnostic.SystemPrefixHMAC = openCodeCacheDiagnosticHMAC(key, encoded)
		}
	}
	if tools, ok := openCodeCacheDiagnosticToolSchema(root); ok {
		if encoded, err := json.Marshal(tools); err == nil {
			diagnostic.ToolSchemaHMAC = openCodeCacheDiagnosticHMAC(key, encoded)
		}
	}
	return diagnostic
}

func decodeOpenCodeCacheDiagnosticJSON(body []byte) (map[string]any, bool) {
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil || root == nil {
		return nil, false
	}
	return root, true
}

func openCodeCacheDiagnosticMessages(root map[string]any) []map[string]any {
	items, _ := root["messages"].([]any)
	messages := make([]map[string]any, 0, len(items))
	for _, item := range items {
		message, ok := item.(map[string]any)
		if !ok {
			continue
		}
		messages = append(messages, message)
	}
	return messages
}

func openCodeCacheDiagnosticRoles(messages []map[string]any) []string {
	roles := make([]string, 0, len(messages))
	for _, message := range messages {
		role, _ := message["role"].(string)
		roles = append(roles, strings.TrimSpace(role))
	}
	return roles
}

func openCodeCacheDiagnosticSystemPrefix(messages []map[string]any) ([]map[string]any, bool) {
	prefix := make([]map[string]any, 0)
	for _, message := range messages {
		role, _ := message["role"].(string)
		role = strings.ToLower(strings.TrimSpace(role))
		if role != "system" && role != "developer" {
			break
		}
		prefix = append(prefix, message)
	}
	return prefix, len(prefix) > 0
}

func openCodeCacheDiagnosticToolSchema(root map[string]any) (map[string]any, bool) {
	tools := make(map[string]any, 2)
	for _, field := range []string{"tools", "functions"} {
		if value, ok := root[field]; ok {
			tools[field] = value
		}
	}
	return tools, len(tools) > 0
}

func openCodeCacheDiagnosticShape(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			switch key {
			case "content", "arguments", "input", "output", "text":
				result[key] = "<redacted>"
			default:
				result[key] = openCodeCacheDiagnosticShape(child)
			}
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, child := range typed {
			result[i] = openCodeCacheDiagnosticShape(child)
		}
		return result
	default:
		return typed
	}
}

func setOpenCodeCacheDiagnostic(c *gin.Context, diagnostic *OpenCodeCacheDiagnostic) {
	if c == nil {
		return
	}
	if diagnostic == nil {
		c.Set(openCodeCacheDiagnosticContextKey, nil)
		return
	}
	c.Set(openCodeCacheDiagnosticContextKey, diagnostic)
}

func openCodeCacheDiagnosticFromContext(c *gin.Context) *OpenCodeCacheDiagnostic {
	if c == nil {
		return nil
	}
	value, ok := c.Get(openCodeCacheDiagnosticContextKey)
	if !ok {
		return nil
	}
	diagnostic, _ := value.(*OpenCodeCacheDiagnostic)
	return diagnostic
}

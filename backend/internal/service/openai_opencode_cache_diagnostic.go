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
	// inferred from account name, URL, or numeric ID.
	openCodeCacheDiagnosticsCredentialKey = "opencode_cache_diagnostics"
	openCodeCacheDiagnosticVersion        = 1
	openCodeCacheDiagnosticDomain         = "sub2api/opencode-cache-diagnostics/v1"
	openCodeCacheDiagnosticContextKey     = "opencode_cache_diagnostic"
)

var openCodeLazyMCPMetaTools = map[string]struct{}{
	"mcp_list_servers":  {},
	"mcp_list_tools":    {},
	"mcp_call_tool":     {},
	"mcp_read_resource": {},
}

// OpenCodeCacheDiagnostic contains only non-reversible request correlation
// values and bounded numeric prompt-profile measurements. It intentionally
// excludes raw prompts, tool arguments, session IDs, credentials, cookies,
// and model output.
type OpenCodeCacheDiagnostic struct {
	Version                      int      `json:"version"`
	InboundSessionHMAC           string   `json:"inbound_session_hmac,omitempty"`
	CanonicalSessionHMAC         string   `json:"canonical_session_hmac,omitempty"`
	PromptCacheKeyHMAC           string   `json:"prompt_cache_key_hmac,omitempty"`
	ForwardedSessionAffinityHMAC string   `json:"forwarded_session_affinity_hmac,omitempty"`
	RequestBodyHMAC              string   `json:"request_body_hmac,omitempty"`
	RequestShapeHMAC             string   `json:"request_shape_hmac,omitempty"`
	SystemPrefixHMAC             string   `json:"system_prefix_hmac,omitempty"`
	ToolSchemaHMAC               string   `json:"tool_schema_hmac,omitempty"`
	RequestBytes                 int      `json:"request_bytes,omitempty"`
	SystemPrefixBytes            int      `json:"system_prefix_bytes,omitempty"`
	ToolSchemaBytes              int      `json:"tool_schema_bytes,omitempty"`
	ToolCount                    int      `json:"tool_count,omitempty"`
	MetaToolCount                int      `json:"meta_tool_count,omitempty"`
	LazyMCP                      bool     `json:"lazy_mcp,omitempty"`
	MessageRoles                 []string `json:"message_roles,omitempty"`
	MessageCount                 int      `json:"message_count"`
}

// String returns the persisted JSON representation. All fields are fixed
// metadata, role labels, counts, byte sizes, booleans, or HMACs; raw request
// data is excluded.
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

// buildOpenCodeCacheDiagnostic runs immediately before an OpenAI-compatible
// request is sent to the selected upstream. body and headers are therefore the
// final outbound representations after request rewrites and account overrides.
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
		RequestBytes:    len(body),
	}
	if value := strings.TrimSpace(inboundSessionID); value != "" {
		diagnostic.InboundSessionHMAC = openCodeCacheDiagnosticHMAC(key, []byte(value))
	}
	if headers != nil {
		if value := strings.TrimSpace(headers.Get("session_id")); value != "" {
			diagnostic.CanonicalSessionHMAC = openCodeCacheDiagnosticHMAC(key, []byte(value))
		}
		if value := strings.TrimSpace(headers.Get(openCodeSessionAffinityHeader)); value != "" {
			diagnostic.ForwardedSessionAffinityHMAC = openCodeCacheDiagnosticHMAC(key, []byte(value))
		}
	}

	root, ok := decodeOpenCodeCacheDiagnosticJSON(body)
	if !ok {
		return diagnostic
	}
	if value, _ := root["prompt_cache_key"].(string); strings.TrimSpace(value) != "" {
		diagnostic.PromptCacheKeyHMAC = openCodeCacheDiagnosticHMAC(key, []byte(strings.TrimSpace(value)))
	}

	messages := openCodeCacheDiagnosticMessages(root)
	diagnostic.MessageCount = len(messages)
	diagnostic.MessageRoles = openCodeCacheDiagnosticRoles(messages)
	if shape, err := json.Marshal(openCodeCacheDiagnosticShape(root)); err == nil {
		diagnostic.RequestShapeHMAC = openCodeCacheDiagnosticHMAC(key, shape)
	}
	if prefix, ok := openCodeCacheDiagnosticSystemPrefix(root, messages); ok {
		if encoded, err := json.Marshal(prefix); err == nil {
			diagnostic.SystemPrefixHMAC = openCodeCacheDiagnosticHMAC(key, encoded)
			diagnostic.SystemPrefixBytes = len(encoded)
		}
	}
	if tools, ok := openCodeCacheDiagnosticToolSchema(root); ok {
		if encoded, err := json.Marshal(tools); err == nil {
			diagnostic.ToolSchemaHMAC = openCodeCacheDiagnosticHMAC(key, encoded)
			diagnostic.ToolSchemaBytes = len(encoded)
		}
	}
	diagnostic.ToolCount, diagnostic.MetaToolCount = openCodeCacheDiagnosticToolCounts(root)
	diagnostic.LazyMCP = diagnostic.MetaToolCount == len(openCodeLazyMCPMetaTools)
	return diagnostic
}

func decodeOpenCodeCacheDiagnosticJSON(body []byte) (map[string]any, bool) {
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil || root == nil {
		return nil, false
	}
	if response, ok := root["response"].(map[string]any); ok {
		return response, true
	}
	return root, true
}

func openCodeCacheDiagnosticMessages(root map[string]any) []map[string]any {
	items, _ := root["messages"].([]any)
	if len(items) == 0 {
		items, _ = root["input"].([]any)
	}
	messages := make([]map[string]any, 0, len(items))
	for _, item := range items {
		message, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if _, ok := message["role"]; !ok {
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

func openCodeCacheDiagnosticSystemPrefix(root map[string]any, messages []map[string]any) (map[string]any, bool) {
	prefix := make([]map[string]any, 0)
	for _, message := range messages {
		role, _ := message["role"].(string)
		role = strings.ToLower(strings.TrimSpace(role))
		if role != "system" && role != "developer" {
			break
		}
		prefix = append(prefix, message)
	}

	result := make(map[string]any, 2)
	if instructions, ok := root["instructions"]; ok {
		result["instructions"] = instructions
	}
	if len(prefix) > 0 {
		result["messages"] = prefix
	}
	return result, len(result) > 0
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

func openCodeCacheDiagnosticToolCounts(root map[string]any) (toolCount int, metaToolCount int) {
	seenMeta := make(map[string]struct{}, len(openCodeLazyMCPMetaTools))
	for _, field := range []string{"tools", "functions"} {
		items, _ := root[field].([]any)
		toolCount += len(items)
		for _, item := range items {
			name := openCodeCacheDiagnosticToolName(item)
			if _, ok := openCodeLazyMCPMetaTools[name]; ok {
				seenMeta[name] = struct{}{}
			}
		}
	}
	return toolCount, len(seenMeta)
}

func openCodeCacheDiagnosticToolName(value any) string {
	tool, _ := value.(map[string]any)
	if name, _ := tool["name"].(string); strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}
	function, _ := tool["function"].(map[string]any)
	name, _ := function["name"].(string)
	return strings.TrimSpace(name)
}

func openCodeCacheDiagnosticShape(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			result[key] = openCodeCacheDiagnosticShape(child)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, child := range typed {
			result[i] = openCodeCacheDiagnosticShape(child)
		}
		return result
	case nil:
		return nil
	default:
		return ""
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
	if !ok || value == nil {
		return nil
	}
	diagnostic, _ := value.(*OpenCodeCacheDiagnostic)
	return diagnostic
}

func attachOpenCodeCacheDiagnostic(c *gin.Context, result *OpenAIForwardResult) *OpenAIForwardResult {
	if result == nil {
		return nil
	}
	result.CacheDiagnostic = openCodeCacheDiagnosticFromContext(c)
	return result
}

package service

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	openCodeSessionHeader                       = "X-OpenCode-Session"
	forwardOpenCodeSessionAffinityCredentialKey = "forward_opencode_session_affinity"
)

// applyOpenCodeSessionHeader forwards the caller-owned conversation identifier
// only to OpenCode's official API origin. The caller applies this after account
// header overrides so a per-conversation value cannot be replaced by a fixed
// account-wide override.
func applyOpenCodeSessionHeader(c *gin.Context, account *Account, targetURL string, headers http.Header) {
	if c == nil || c.Request == nil || account == nil || account.Type != AccountTypeAPIKey || headers == nil {
		return
	}

	parsed, err := url.Parse(targetURL)
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || !strings.EqualFold(parsed.Hostname(), "opencode.ai") {
		return
	}

	sessionID := strings.TrimSpace(c.GetHeader(openCodeSessionHeader))
	if sessionID == "" {
		return
	}
	for key := range headers {
		if strings.EqualFold(key, openCodeSessionHeader) {
			delete(headers, key)
		}
	}
	headers.Set(openCodeSessionHeader, sessionID)
}

// applyOpenCodeSessionAffinityHeader forwards a caller-owned session identifier
// to an explicitly trusted API-key upstream such as a locally operated CPA.
//
// This is deliberately opt-in per account. Forwarding a conversation identifier
// to every OpenAI-compatible target would violate the trust boundary enforced by
// applyOpenCodeSessionHeader. The target receives the canonical CPA header even
// when the client used another supported OpenCode session-header spelling.
func applyOpenCodeSessionAffinityHeader(c *gin.Context, account *Account, headers http.Header) {
	if c == nil || c.Request == nil || account == nil || account.Type != AccountTypeAPIKey || headers == nil {
		return
	}
	enabled, _ := account.Credentials[forwardOpenCodeSessionAffinityCredentialKey].(bool)
	if !enabled {
		return
	}

	sessionID := strings.TrimSpace(explicitOpenAIHeaderSessionID(c))
	if sessionID == "" {
		return
	}
	for key := range headers {
		if strings.EqualFold(key, openCodeSessionAffinityHeader) {
			delete(headers, key)
		}
	}
	headers.Set(openCodeSessionAffinityHeader, sessionID)
}

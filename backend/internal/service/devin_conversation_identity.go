package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const devinConversationIdentityContextKey = "devin_conversation_identity"

// normalizeDevinConversationIdentity converges every Devin request onto one
// tenant-isolated conversation identity. The same value is written into the
// request body as prompt_cache_key and later projected to session_id and
// X-Session-Id headers by applyDevinConversationIdentityHeaders.
//
// Explicit prompt_cache_key wins over headers because it survives every proxy
// and protocol bridge. When clients provide no explicit identity, the existing
// stable-prefix derivation (model + instructions/tools + first user input) is
// used as a compatibility fallback.
func normalizeDevinConversationIdentity(c *gin.Context, body []byte) ([]byte, string, error) {
	if len(body) == 0 {
		return body, "", nil
	}
	view := openAIRequestPayloadView(body)
	if !isDevinModel(view.Get("model").String()) {
		return body, "", nil
	}

	if c != nil {
		if staged, ok := c.Get(devinConversationIdentityContextKey); ok {
			if identity, okIdentity := staged.(string); okIdentity && identity != "" {
				updated, err := setOpenAIRequestPromptCacheKey(body, identity)
				return updated, identity, err
			}
		}
	}

	rawIdentity := sanitizeSessionID(view.Get("prompt_cache_key").String())
	if rawIdentity == "" {
		rawIdentity = sanitizeSessionID(explicitOpenAIHeaderSessionID(c))
	}
	if rawIdentity == "" {
		rawIdentity = deriveOpenAIContentSessionSeed(body)
	}
	if strings.TrimSpace(rawIdentity) == "" {
		return body, "", nil
	}

	canonical := generateSessionUUID(isolateOpenAISessionID(getAPIKeyIDFromContext(c), rawIdentity))
	updated, err := setOpenAIRequestPromptCacheKey(body, canonical)
	if err != nil {
		return body, "", err
	}
	if c != nil {
		c.Set(devinConversationIdentityContextKey, canonical)
	}
	return updated, canonical, nil
}

func isDevinModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "devin/")
}

func setOpenAIRequestPromptCacheKey(body []byte, identity string) ([]byte, error) {
	root := gjson.ParseBytes(body)
	eventType := strings.ToLower(strings.TrimSpace(root.Get("type").String()))
	path := "prompt_cache_key"
	if strings.HasPrefix(eventType, "response.") && root.Get("response").IsObject() {
		path = "response.prompt_cache_key"
	}
	updated, err := sjson.SetBytes(body, path, identity)
	if err != nil {
		return body, fmt.Errorf("set Devin prompt_cache_key: %w", err)
	}
	return updated, nil
}

func applyDevinConversationIdentityHeaders(headers http.Header, identity string) {
	identity = strings.TrimSpace(identity)
	if headers == nil || identity == "" {
		return
	}
	deleteHeaderAllForms(headers, "session_id")
	deleteHeaderAllForms(headers, "X-Session-Id")
	headers.Set("session_id", identity)
	headers.Set("X-Session-Id", identity)
}

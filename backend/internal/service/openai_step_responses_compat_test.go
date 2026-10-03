package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNormalizeOpenAIStepResponsesAssistantTextContentConvertsTypedAndImplicitMessages(t *testing.T) {
	for _, tc := range []struct {
		name   string
		prefix string
	}{
		{
			name:   "typed message",
			prefix: `{"type":"message",`,
		},
		{
			name:   "implicit easy input message",
			prefix: `{`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"step-5-preview","input":[` + tc.prefix +
				`"role":"assistant","status":"completed","id":"msg_1","phase":"final","content":[` +
				`{"type":"output_text","text":"keep"},` +
				`{"type":"output_text","text":" whitespace","annotations":[]},` +
				`{"type":"output_text","text":"\n and unicode 世界"}]}]}`)

			got, convertedMessages, err := normalizeOpenAIStepResponsesAssistantTextContent(body)

			require.NoError(t, err)
			require.Equal(t, 1, convertedMessages)
			var input []map[string]json.RawMessage
			require.NoError(t, json.Unmarshal([]byte(gjson.GetBytes(got, "input").Raw), &input))
			require.NotContains(t, input[0], "type", "string content must use the untagged easy-input-message discriminator")
			require.JSONEq(t, `"keep whitespace\n and unicode 世界"`, string(input[0]["content"]))
			require.JSONEq(t, `"completed"`, string(input[0]["status"]))
			require.JSONEq(t, `"msg_1"`, string(input[0]["id"]))
			require.JSONEq(t, `"final"`, string(input[0]["phase"]))
		})
	}
}

func TestNormalizeOpenAIStepResponsesAssistantTextContentIsIdempotent(t *testing.T) {
	body := []byte(`{"model":"step-5-preview","input":[` +
		`{"type":"message","role":"assistant","status":"completed","id":"msg_1","content":[` +
		`{"type":"output_text","text":"keep"},` +
		`{"type":"output_text","text":" whitespace","annotations":[]},` +
		`{"type":"output_text","text":"\\n and unicode 世界"}]},` +
		`{"type":"message","role":"user","content":[{"type":"input_text","text":"unchanged"}]}` +
		`]}`)

	got, convertedMessages, err := normalizeOpenAIStepResponsesAssistantTextContent(body)

	require.NoError(t, err)
	require.Equal(t, 1, convertedMessages)
	require.NotEqual(t, body, got)
	require.Equal(t, "keep whitespace\\n and unicode 世界", gjsonString(t, got, "input.0.content"))
	require.Equal(t, "unchanged", gjsonString(t, got, "input.1.content.0.text"))

	idempotent, convertedMessages, err := normalizeOpenAIStepResponsesAssistantTextContent(got)
	require.NoError(t, err)
	require.Zero(t, convertedMessages)
	require.True(t, bytes.Equal(got, idempotent), "a converted string message must remain untouched")
}

func TestNormalizeOpenAIStepResponsesAssistantTextContentPreservesConservativeCases(t *testing.T) {
	body := []byte(`{"model":"step-5-preview","input":[` +
		`{"type":"message","role":"assistant","content":[{"type":"output_text","text":"annotated","annotations":[{"type":"url_citation"}]}]},` +
		`{"type":"message","role":"assistant","content":[{"type":"output_text","text":"unknown","extra":true}]},` +
		`{"type":"message","role":"assistant","content":[{"type":"output_text","text":"mixed"},{"type":"input_text","text":"not pure"}]},` +
		`{"type":"message","role":"assistant","content":"already a string"},` +
		`{"type":"message","role":"assistant","content":null},` +
		`{"type":"message","role":"assistant","content":[]},` +
		`{"type":"message","role":"user","content":[{"type":"output_text","text":"user text"}]}` +
		`]}`)

	got, convertedMessages, err := normalizeOpenAIStepResponsesAssistantTextContent(body)

	require.NoError(t, err)
	require.Zero(t, convertedMessages)
	require.True(t, bytes.Equal(body, got), "non-pure assistant content must preserve the original body")
}

func TestNormalizeOpenAIStepResponsesAssistantTextContentRejectsNullTextAndAnnotations(t *testing.T) {
	for _, body := range []string{
		`{"model":"step-5-preview","input":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":null}]}]}`,
		`{"model":"step-5-preview","input":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"text","annotations":null}]}]}`,
	} {
		got, convertedMessages, err := normalizeOpenAIStepResponsesAssistantTextContent([]byte(body))
		require.NoError(t, err)
		require.Zero(t, convertedMessages)
		require.Equal(t, []byte(body), got)
	}
}

func TestNormalizeOpenAIStepResponsesAssistantTextContentLeavesInvalidJSONUntouched(t *testing.T) {
	body := []byte(`{"model":"step-5-preview","input":[`)

	got, convertedMessages, err := normalizeOpenAIStepResponsesAssistantTextContent(body)

	require.NoError(t, err)
	require.Zero(t, convertedMessages)
	require.Equal(t, body, got)
}

func TestShouldNormalizeOpenAIStepResponsesAssistantTextContent(t *testing.T) {
	newAccount := func(baseURL string, accountType string, mode string) *Account {
		return &Account{
			Platform: PlatformOpenAI,
			Type:     accountType,
			Credentials: map[string]any{
				"base_url": baseURL,
			},
			Extra: map[string]any{
				openai_compat.ExtraKeyResponsesMode: mode,
			},
		}
	}

	for _, tc := range []struct {
		name       string
		account    *Account
		transport  OpenAIUpstreamTransport
		compact    bool
		wantEnable bool
	}{
		{
			name:       "step native HTTP responses",
			account:    newAccount("https://api.stepfun.com/v1", AccountTypeAPIKey, string(openai_compat.ResponsesSupportModeForceResponses)),
			transport:  OpenAIUpstreamTransportHTTPSSE,
			wantEnable: true,
		},
		{
			name:       "step port is still exact hostname",
			account:    newAccount("https://api.stepfun.com:443/v1", AccountTypeAPIKey, string(openai_compat.ResponsesSupportModeForceResponses)),
			transport:  OpenAIUpstreamTransportHTTPSSE,
			wantEnable: true,
		},
		{
			name:      "other host",
			account:   newAccount("https://api.stepfun.com.evil.example/v1", AccountTypeAPIKey, string(openai_compat.ResponsesSupportModeForceResponses)),
			transport: OpenAIUpstreamTransportHTTPSSE,
		},
		{
			name:      "OAuth is excluded",
			account:   newAccount("https://api.stepfun.com/v1", AccountTypeOAuth, string(openai_compat.ResponsesSupportModeForceResponses)),
			transport: OpenAIUpstreamTransportHTTPSSE,
		},
		{
			name:      "websocket is excluded",
			account:   newAccount("https://api.stepfun.com/v1", AccountTypeAPIKey, string(openai_compat.ResponsesSupportModeForceResponses)),
			transport: OpenAIUpstreamTransportResponsesWebsocketV2,
		},
		{
			name:      "compact is excluded",
			account:   newAccount("https://api.stepfun.com/v1", AccountTypeAPIKey, string(openai_compat.ResponsesSupportModeForceResponses)),
			transport: OpenAIUpstreamTransportHTTPSSE,
			compact:   true,
		},
		{
			name:      "chat completions route is excluded",
			account:   newAccount("https://api.stepfun.com/v1", AccountTypeAPIKey, string(openai_compat.ResponsesSupportModeForceChatCompletions)),
			transport: OpenAIUpstreamTransportHTTPSSE,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.wantEnable, shouldNormalizeOpenAIStepResponsesAssistantTextContent(tc.account, tc.transport, tc.compact))
		})
	}
}

func TestForwardOpenAIStepResponsesAssistantTextContentNormalizesHTTPAndPassthrough(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(map[bool]string{false: "native HTTP", true: "passthrough"}[passthrough], func(t *testing.T) {
			account := &Account{
				ID:       101,
				Name:     "step-apikey",
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"api_key":  "sk-test",
					"base_url": "https://api.stepfun.com/v1",
				},
				Extra: map[string]any{
					openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceResponses),
					"openai_passthrough":                passthrough,
				},
				Concurrency: 1,
			}
			body := []byte(`{"model":"step-5-preview","input":[{"type":"message","role":"assistant","id":"msg_1","status":"completed","content":[{"type":"output_text","text":"first"},{"type":"output_text","text":" second","annotations":[]}]}],"stream":false}`)
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":"resp_step","object":"response","model":"step-5-preview","status":"completed","output":[],"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}}`)),
			}}
			svc := &OpenAIGatewayService{cfg: &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}, httpUpstream: upstream}

			result, err := svc.Forward(context.Background(), c, account, body)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, "first second", gjson.GetBytes(upstream.lastBody, "input.0.content").String())
			require.False(t, gjson.GetBytes(upstream.lastBody, "input.0.type").Exists(), "upstream must receive an implicit easy input message")
			require.Equal(t, "msg_1", gjson.GetBytes(upstream.lastBody, "input.0.id").String())
			require.Equal(t, "completed", gjson.GetBytes(upstream.lastBody, "input.0.status").String())
		})
	}
}

func gjsonString(t *testing.T, body []byte, path string) string {
	t.Helper()
	var decoded any
	require.NoError(t, json.Unmarshal(body, &decoded))
	return rawJSONPathString(decoded, path)
}

func rawJSONPathString(value any, path string) string {
	var current = value
	for _, part := range strings.Split(path, ".") {
		switch typed := current.(type) {
		case map[string]any:
			current = typed[part]
		case []any:
			var index int
			if _, err := fmt.Sscanf(part, "%d", &index); err != nil || index < 0 || index >= len(typed) {
				return ""
			}
			current = typed[index]
		default:
			return ""
		}
	}
	if stringValue, ok := current.(string); ok {
		return stringValue
	}
	return ""
}

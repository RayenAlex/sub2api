package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type codexModelReasoningHandlerRepo struct{ *settingHandlerRepoStub }

func (r *codexModelReasoningHandlerRepo) Set(_ context.Context, key, value string) error {
	if r.values == nil {
		r.values = make(map[string]string)
	}
	r.values[key] = value
	return nil
}

func TestCodexModelReasoningAdminEndpointIsolatedPersistence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &codexModelReasoningHandlerRepo{settingHandlerRepoStub: &settingHandlerRepoStub{values: map[string]string{service.SettingKeySiteName: "Example Gateway"}}}
	handler := NewSettingHandler(service.NewSettingService(repo, &config.Config{}), nil, nil, nil, nil, nil, nil)
	invoke := func(method, body string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = httptest.NewRequest(method, "/api/v1/admin/settings/codex-model-reasoning", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		if method == http.MethodGet {
			handler.GetCodexModelReasoningSettings(ctx)
		} else {
			handler.UpdateCodexModelReasoningSettings(ctx)
		}
		return rec
	}
	require.Equal(t, http.StatusOK, invoke(http.MethodGet, "").Code)
	payload := `{"rules":[{"model":"gpt-6-sol","supported_reasoning_levels":["low","medium","high"],"default_reasoning_level":"medium"}]}`
	rec := invoke(http.MethodPut, payload)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "Example Gateway", repo.values[service.SettingKeySiteName])
	got := invoke(http.MethodGet, "")
	require.Equal(t, http.StatusOK, got.Code)
	require.Equal(t, "gpt-6-sol", gjson.Get(got.Body.String(), "data.rules.0.model").String())
	require.Equal(t, "medium", gjson.Get(got.Body.String(), "data.rules.0.default_reasoning_level").String())
	for _, invalid := range []string{`{}`, `{"rules":null}`, `{"rules":[{"model":"gpt-6-*","supported_reasoning_levels":["high"],"default_reasoning_level":"high"}]}`} {
		require.Equal(t, http.StatusBadRequest, invoke(http.MethodPut, invalid).Code)
	}
	require.Equal(t, http.StatusBadRequest, invoke(http.MethodPut, `{"rules":[],"ignored":"`+strings.Repeat("x", 33<<10)+`"}`).Code)
	normalized := invoke(http.MethodPut, `{"rules":[{"model":" gpt-6-sol ","supported_reasoning_levels":["medium"],"default_reasoning_level":"medium"}]}`)
	require.Equal(t, http.StatusOK, normalized.Code)
	require.Equal(t, "gpt-6-sol", gjson.Get(normalized.Body.String(), "data.rules.0.model").String())
	var stored service.CodexModelReasoningSettings
	require.NoError(t, json.Unmarshal([]byte(repo.values[service.SettingKeyCodexModelReasoningRules]), &stored))
	require.Len(t, stored.Rules, 1)
	require.Equal(t, http.StatusOK, invoke(http.MethodPut, `{"rules":[]}`).Code)
	require.Empty(t, gjson.Get(invoke(http.MethodGet, "").Body.String(), "data.rules").Array())
}

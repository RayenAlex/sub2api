package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodexModelReasoningConvertedUpstreamListAndETag(t *testing.T) {
	upstream := &codexModelsHTTPUpstreamStub{do: func(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Etag": []string{`"upstream-list"`}}, Body: io.NopCloser(bytes.NewBufferString(`{"object":"list","data":[{"id":"gpt-6-sol"},{"id":"plain-model"}]}`))}, nil
	}}
	svc := newCodexModelsAPIKeyTestService(upstream)
	svc.accountRepo = codexModelsVisibilityAccountRepo{}
	repo := &codexReasoningSettingRepo{}
	svc.settingService = NewSettingService(repo, nil)
	account := newCodexModelsAPIKeyTestAccount("https://compatible.example/v1")
	group := &Group{ID: 82, Platform: PlatformOpenAI}
	load := func(ifNoneMatch string) *OpenAIModelsResponse {
		t.Helper()
		manifest, err := svc.FetchCodexModelsManifest(context.Background(), account, "0.150.0", "")
		require.NoError(t, err)
		require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))
		require.NoError(t, svc.MergeGroupConfiguredCodexModels(context.Background(), group, manifest, ifNoneMatch))
		return manifest
	}
	before := load("")
	beforeModels := decodeCodexManifestModels(t, before.Body)
	rules := CodexModelReasoningSettings{Rules: []CodexModelReasoningRule{{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"none", "low", "medium"}, DefaultReasoningLevel: "medium"}}}
	require.NoError(t, svc.settingService.SetCodexModelReasoningSettings(context.Background(), rules))
	after := load(before.ETag)
	require.False(t, after.NotModified)
	require.NotEqual(t, before.ETag, after.ETag)
	require.Equal(t, codexModelsManifestBodyETag(after.Body), after.ETag)
	models := decodeCodexManifestModels(t, after.Body)
	require.Equal(t, "medium", models[0]["default_reasoning_level"])
	require.Equal(t, rules.Rules[0].SupportedReasoningLevels, effortsFromManifestModel(t, models[0]))
	require.Equal(t, beforeModels[1]["default_reasoning_level"], models[1]["default_reasoning_level"])
	unchanged := load(after.ETag)
	require.True(t, unchanged.NotModified)
	require.Nil(t, unchanged.Body)
}

func TestCodexModelReasoningConvertedUpstreamListCapabilityRestrictions(t *testing.T) {
	svc := newCodexModelsAPIKeyTestService(nil)
	repo := &codexReasoningSettingRepo{}
	svc.settingService = NewSettingService(repo, nil)
	require.NoError(t, svc.settingService.SetCodexModelReasoningSettings(context.Background(), CodexModelReasoningSettings{Rules: []CodexModelReasoningRule{
		{Model: "disabled", SupportedReasoningLevels: []string{"low", "high"}, DefaultReasoningLevel: "high"},
		{Model: "narrow", SupportedReasoningLevels: []string{"low", "high"}, DefaultReasoningLevel: "high"},
	}}))
	account := newCodexModelsAPIKeyTestAccount("https://compatible.example/v1")
	no := false
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"disabled": {Reasoning: &no},
		"narrow":   {SupportedReasoningLevels: []string{"low"}},
	}})
	source := []byte(`{"object":"list","data":[{"id":"disabled","reasoning":false},{"id":"narrow","supported_reasoning_levels":["low"]}]}`)
	manifest := &OpenAIModelsResponse{Body: convertOpenAIModelListToCodexManifestForAccount(source, account), upstreamSourceBody: source, convertedFromOpenAIModelList: true}
	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Equal(t, []string{"none"}, effortsFromManifestModel(t, models[0]))
	require.Equal(t, []string{"low"}, effortsFromManifestModel(t, models[1]))
	require.Equal(t, "low", models[1]["default_reasoning_level"])
}

func TestCodexModelReasoningConvertedListUsesRuleWhenUpstreamLevelsUnknown(t *testing.T) {
	svc := newCodexModelsAPIKeyTestService(nil)
	svc.settingService = NewSettingService(&codexReasoningSettingRepo{}, nil)
	require.NoError(t, svc.settingService.SetCodexModelReasoningSettings(context.Background(), CodexModelReasoningSettings{Rules: []CodexModelReasoningRule{{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"none", "medium", "max"}, DefaultReasoningLevel: "medium"}}}))
	account := newCodexModelsAPIKeyTestAccount("https://compatible.example/v1")
	yes := true
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{"gpt-6-sol": {Reasoning: &yes}}})
	source := []byte(`{"object":"list","data":[{"id":"gpt-6-sol","reasoning":true}]}`)
	manifest := &OpenAIModelsResponse{Body: convertOpenAIModelListToCodexManifestForAccount(source, account), upstreamSourceBody: source, convertedFromOpenAIModelList: true}
	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))
	model := decodeCodexManifestModels(t, manifest.Body)[0]
	require.Equal(t, "medium", model["default_reasoning_level"])
	require.Equal(t, []string{"none", "medium", "max"}, effortsFromManifestModel(t, model))
}

func TestCodexModelReasoningConvertedListHonorsRawReasoningDenialWithoutSnapshot(t *testing.T) {
	source := []byte(`{"object":"list","data":[{"id":"gpt-6-sol","reasoning":false}]}`)
	account := newCodexModelsAPIKeyTestAccount("https://compatible.example/v1")
	converted := convertOpenAIModelListToCodexManifestForAccount(source, account)
	models := decodeCodexManifestModels(t, converted)
	require.Equal(t, []string{"none"}, effortsFromManifestModel(t, models[0]))
}

func TestCodexModelReasoningConvertedListConflictingSourcesDoNotAdvertiseExtraLevels(t *testing.T) {
	svc := newCodexModelsAPIKeyTestService(nil)
	svc.settingService = NewSettingService(&codexReasoningSettingRepo{}, nil)
	require.NoError(t, svc.settingService.SetCodexModelReasoningSettings(context.Background(), CodexModelReasoningSettings{Rules: []CodexModelReasoningRule{{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"low", "high"}, DefaultReasoningLevel: "high"}}}))
	yes := true
	account := newCodexModelsAPIKeyTestAccount("https://compatible.example/v1")
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{"gpt-6-sol": {Reasoning: &yes, SupportedReasoningLevels: []string{"high"}}}})
	for _, rawEntry := range []string{
		`{"id":"gpt-6-sol","reasoning":false}`,
		`{"id":"gpt-6-sol","reasoning":true,"supported_reasoning_levels":["low"]}`,
	} {
		source := []byte(`{"object":"list","data":[` + rawEntry + `]}`)
		manifest := &OpenAIModelsResponse{Body: convertOpenAIModelListToCodexManifestForAccount(source, account), upstreamSourceBody: source, convertedFromOpenAIModelList: true}
		require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))
		models := decodeCodexManifestModels(t, manifest.Body)
		require.Empty(t, effortsFromManifestModel(t, models[0]))
	}
}

func TestCodexModelReasoningOpenAIGroupAndOtherPlatformCatalogs(t *testing.T) {
	repo := &codexReasoningSettingRepo{}
	settings := NewSettingService(repo, nil)
	const groupID int64 = 98
	openai := &OpenAIGatewayService{
		settingService: settings,
		accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
			groupID: {{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
				"model_mapping": map[string]any{"gpt-6-sol": "gpt-6-sol"},
			}}},
		}},
	}
	group := &Group{ID: groupID, Platform: PlatformOpenAI}
	before, configured, err := openai.BuildGroupConfiguredCodexModelsManifest(context.Background(), group, "")
	require.NoError(t, err)
	require.True(t, configured)
	require.NoError(t, settings.SetCodexModelReasoningSettings(context.Background(), CodexModelReasoningSettings{Rules: []CodexModelReasoningRule{{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"none", "low"}, DefaultReasoningLevel: "low"}}}))
	after, configured, err := openai.BuildGroupConfiguredCodexModelsManifest(context.Background(), group, before.ETag)
	require.NoError(t, err)
	require.True(t, configured)
	require.False(t, after.NotModified)
	require.NotEqual(t, before.ETag, after.ETag)
	models := decodeCodexManifestModels(t, after.Body)
	require.Equal(t, "low", models[0]["default_reasoning_level"])
	require.Equal(t, []string{"none", "low"}, effortsFromManifestModel(t, models[0]))
	matching, _, err := openai.BuildGroupConfiguredCodexModelsManifest(context.Background(), group, after.ETag)
	require.NoError(t, err)
	require.True(t, matching.NotModified)

	other := &GatewayService{settingService: settings}
	otherBody, err := other.BuildCodexModelsManifestForGroup(context.Background(), &Group{Platform: PlatformGrok}, "", []string{"gpt-6-sol"})
	require.NoError(t, err)
	require.Equal(t, []string{"none", "low"}, effortsFromManifestModel(t, decodeCodexManifestModels(t, otherBody)[0]))
}

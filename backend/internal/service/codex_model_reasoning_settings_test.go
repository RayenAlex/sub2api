package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type codexReasoningSettingRepo struct {
	SettingRepository
	value  string
	writes int
	err    error
}

type blockedCodexReasoningSettingRepo struct {
	SettingRepository
	mu      sync.Mutex
	value   string
	once    sync.Once
	started chan struct{}
	release chan struct{}
}

func (r *blockedCodexReasoningSettingRepo) GetValue(context.Context, string) (string, error) {
	r.mu.Lock()
	value := r.value
	r.mu.Unlock()
	r.once.Do(func() { close(r.started); <-r.release })
	return value, nil
}

func (r *blockedCodexReasoningSettingRepo) Set(_ context.Context, _ string, value string) error {
	r.mu.Lock()
	r.value = value
	r.mu.Unlock()
	return nil
}

func TestCodexModelReasoningConcurrentReadCannotOverwriteSave(t *testing.T) {
	repo := &blockedCodexReasoningSettingRepo{
		value:   `{"rules":[{"model":"gpt-6-sol","supported_reasoning_levels":["low"],"default_reasoning_level":"low"}]}`,
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	svc := NewSettingService(repo, nil)
	readDone := make(chan error, 1)
	go func() { _, err := svc.GetCodexModelReasoningSettings(context.Background()); readDone <- err }()
	<-repo.started
	saveDone := make(chan error, 1)
	go func() {
		saveDone <- svc.SetCodexModelReasoningSettings(context.Background(), CodexModelReasoningSettings{Rules: []CodexModelReasoningRule{{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"high"}, DefaultReasoningLevel: "high"}}})
	}()
	// Before serialization, the save could complete while the stale read was
	// blocked, after which that read would replace the new cache entry.
	savedBeforeRead := false
	select {
	case err := <-saveDone:
		require.NoError(t, err)
		savedBeforeRead = true
	case <-time.After(50 * time.Millisecond):
	}
	close(repo.release)
	require.NoError(t, <-readDone)
	if !savedBeforeRead {
		select {
		case err := <-saveDone:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("save did not finish")
		}
	}
	got, err := svc.GetCodexModelReasoningSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "high", got.Rules[0].DefaultReasoningLevel)
}

func (r *codexReasoningSettingRepo) GetValue(context.Context, string) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	if r.writes == 0 && r.value == "" {
		return "", ErrSettingNotFound
	}
	return r.value, nil
}
func (r *codexReasoningSettingRepo) Set(_ context.Context, key, value string) error {
	if key != SettingKeyCodexModelReasoningRules {
		panic("wrong settings key")
	}
	r.value = value
	r.writes++
	return nil
}

func TestCodexModelReasoningSettingsValidationAndPersistence(t *testing.T) {
	ctx := context.Background()
	repo := &codexReasoningSettingRepo{}
	svc := NewSettingService(repo, nil)
	before, err := svc.GetCodexModelReasoningSettings(ctx)
	require.NoError(t, err)
	require.Empty(t, before.Rules)
	rules := CodexModelReasoningSettings{Rules: []CodexModelReasoningRule{{
		Model: " gpt-6-sol ", SupportedReasoningLevels: []string{"none", "low", "medium", "high", "xhigh", "max"}, DefaultReasoningLevel: "medium",
	}}}
	require.NoError(t, svc.SetCodexModelReasoningSettings(ctx, rules))
	require.Equal(t, 1, repo.writes)
	got, err := svc.GetCodexModelReasoningSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, "gpt-6-sol", got.Rules[0].Model)
	var persisted CodexModelReasoningSettings
	require.NoError(t, json.Unmarshal([]byte(repo.value), &persisted))
	require.Equal(t, got, persisted)
	got.Rules[0].SupportedReasoningLevels[0] = "ultra"
	again, err := svc.GetCodexModelReasoningSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, "none", again.Rules[0].SupportedReasoningLevels[0])
	require.NoError(t, svc.SetCodexModelReasoningSettings(ctx, CodexModelReasoningSettings{Rules: []CodexModelReasoningRule{}}))
	cleared, err := svc.GetCodexModelReasoningSettings(ctx)
	require.NoError(t, err)
	require.Empty(t, cleared.Rules)
}

func TestCodexModelReasoningSettingsRejectInvalidWithoutWrite(t *testing.T) {
	valid := CodexModelReasoningRule{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"low", "max"}, DefaultReasoningLevel: "low"}
	cases := map[string][]CodexModelReasoningRule{
		"empty model":           {{Model: "  ", SupportedReasoningLevels: []string{"low"}, DefaultReasoningLevel: "low"}},
		"wildcard":              {{Model: "gpt-6-*", SupportedReasoningLevels: []string{"low"}, DefaultReasoningLevel: "low"}},
		"bracket wildcard":      {{Model: "gpt-6-[ab]", SupportedReasoningLevels: []string{"low"}, DefaultReasoningLevel: "low"}},
		"duplicate models":      {valid, valid},
		"no levels":             {{Model: "gpt-6-sol", DefaultReasoningLevel: "medium"}},
		"duplicate levels":      {{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"low", "low"}, DefaultReasoningLevel: "low"}},
		"unsupported effort":    {{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"low", "impossible"}, DefaultReasoningLevel: "low"}},
		"default not in levels": {{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"low"}, DefaultReasoningLevel: "medium"}},
		"long model":            {{Model: string(make([]byte, 201)), SupportedReasoningLevels: []string{"low"}, DefaultReasoningLevel: "low"}},
	}
	for label, rules := range cases {
		t.Run(label, func(t *testing.T) {
			repo := &codexReasoningSettingRepo{}
			require.Error(t, NewSettingService(repo, nil).SetCodexModelReasoningSettings(context.Background(), CodexModelReasoningSettings{Rules: rules}))
			require.Zero(t, repo.writes)
		})
	}
	oversized := make([]CodexModelReasoningRule, 65)
	for i := range oversized {
		oversized[i] = valid
	}
	repo := &codexReasoningSettingRepo{}
	require.Error(t, NewSettingService(repo, nil).SetCodexModelReasoningSettings(context.Background(), CodexModelReasoningSettings{Rules: oversized}))
	require.Zero(t, repo.writes)
}

func TestCodexModelReasoningRuntimeReadFailureLeavesCatalogUntouched(t *testing.T) {
	svc := NewSettingService(&codexReasoningSettingRepo{err: errors.New("database unavailable")}, nil)
	_, err := svc.GetCodexModelReasoningSettings(context.Background())
	require.Error(t, err)
	require.Empty(t, svc.CodexModelReasoningRulesForManifest(context.Background()))
}

func TestCodexModelReasoningAppliesOnlyToLocallyGeneratedModels(t *testing.T) {
	rules := []CodexModelReasoningRule{{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"low", "high"}, DefaultReasoningLevel: "high"}}
	body, err := buildCodexModelsManifestForAccounts(PlatformOpenAI, []string{"gpt-6-sol", "gpt-6-astra"}, nil, nil, nil, true, rules)
	require.NoError(t, err)
	models := decodeCodexManifestModels(t, body)
	require.Equal(t, "high", models[0]["default_reasoning_level"])
	require.Equal(t, []string{"low", "high"}, effortsFromManifestModel(t, models[0]))
	require.Contains(t, effortsFromManifestModel(t, models[1]), "ultra")
	native := []byte(`{"models":[{"slug":"gpt-6-sol","default_reasoning_level":"medium","supported_reasoning_levels":[{"effort":"low"},{"effort":"medium"}],"custom":{"keep":true}}]}`)
	merged, changed, err := mergeConfiguredCodexModelsManifest(native, []string{"gpt-6-sol", "gpt-6-astra"}, nil, false, rules)
	require.NoError(t, err)
	require.True(t, changed)
	mergedModels := decodeCodexManifestModels(t, merged)
	require.Equal(t, "medium", mergedModels[0]["default_reasoning_level"])
	require.Equal(t, []string{"low", "medium"}, effortsFromManifestModel(t, mergedModels[0]))
	require.Equal(t, map[string]any{"keep": true}, mergedModels[0]["custom"])
}

func TestCodexModelReasoningRespectsExplicitUpstreamCapabilities(t *testing.T) {
	rule := CodexModelReasoningRule{Model: "gpt-6-sol", SupportedReasoningLevels: []string{"low", "medium", "high"}, DefaultReasoningLevel: "high"}
	descriptor := newConfiguredCodexModelDescriptor("gpt-6-sol")
	supported := codexModelMetadataOverride{UpstreamModelMetadata: UpstreamModelMetadata{SupportedReasoningLevels: []string{"low", "medium"}}}
	applyCodexModelReasoningRule(&descriptor, rule, &supported)
	require.Equal(t, []string{"low", "medium"}, effortsFromConfiguredCodexLevels(descriptor.SupportedReasoningLevels))
	require.Equal(t, "low", *descriptor.DefaultReasoningLevel)
	noReasoning := false
	denied := codexModelMetadataOverride{UpstreamModelMetadata: UpstreamModelMetadata{Reasoning: &noReasoning}}
	descriptor = newConfiguredCodexModelDescriptor("gpt-6-sol")
	applyUpstreamModelMetadataToCodexDescriptor(&descriptor, denied)
	applyCodexModelReasoningRule(&descriptor, rule, &denied)
	require.Equal(t, []string{"none"}, effortsFromConfiguredCodexLevels(descriptor.SupportedReasoningLevels))
	// Reasoning=true does not, by itself, assert a narrower effort list.
	yes := true
	unknownLevels := codexModelMetadataOverride{UpstreamModelMetadata: UpstreamModelMetadata{Reasoning: &yes}}
	descriptor = newConfiguredCodexModelDescriptor("gpt-6-sol")
	applyUpstreamModelMetadataToCodexDescriptor(&descriptor, unknownLevels)
	applyCodexModelReasoningRule(&descriptor, rule, &unknownLevels)
	require.Equal(t, []string{"low", "medium", "high"}, effortsFromConfiguredCodexLevels(descriptor.SupportedReasoningLevels))
	require.Equal(t, "high", *descriptor.DefaultReasoningLevel)
}

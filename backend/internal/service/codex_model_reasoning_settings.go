package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"
)

// CodexModelReasoningRule changes only locally generated Codex model entries.
// Model matches the public manifest slug exactly (after trimming whitespace).
type CodexModelReasoningRule struct {
	Model                    string   `json:"model"`
	SupportedReasoningLevels []string `json:"supported_reasoning_levels"`
	DefaultReasoningLevel    string   `json:"default_reasoning_level"`
}
type CodexModelReasoningSettings struct {
	Rules []CodexModelReasoningRule `json:"rules"`
}
type cachedCodexModelReasoningSettings struct {
	settings  CodexModelReasoningSettings
	err       error
	expiresAt int64
}

const (
	codexModelReasoningCacheTTL     = 60 * time.Second
	codexModelReasoningErrorTTL     = 5 * time.Second
	codexModelReasoningDBTimeout    = 5 * time.Second
	codexModelReasoningMaxRules     = 64
	codexModelReasoningMaxSlugBytes = 200
)

var ErrInvalidCodexModelReasoningSettings = errors.New("invalid Codex model reasoning settings")

var codexReasoningEfforts = map[string]struct{}{
	"none": {}, "minimal": {}, "low": {}, "medium": {}, "high": {}, "xhigh": {}, "max": {}, "ultra": {},
}

func cloneCodexModelReasoningSettings(value CodexModelReasoningSettings) CodexModelReasoningSettings {
	out := CodexModelReasoningSettings{Rules: make([]CodexModelReasoningRule, 0, len(value.Rules))}
	for _, rule := range value.Rules {
		rule.SupportedReasoningLevels = append([]string(nil), rule.SupportedReasoningLevels...)
		out.Rules = append(out.Rules, rule)
	}
	return out
}
func normalizeCodexModelReasoningSettings(settings CodexModelReasoningSettings) (CodexModelReasoningSettings, error) {
	if len(settings.Rules) > codexModelReasoningMaxRules {
		return CodexModelReasoningSettings{}, fmt.Errorf("too many Codex model rules: max %d", codexModelReasoningMaxRules)
	}
	out := CodexModelReasoningSettings{Rules: make([]CodexModelReasoningRule, 0, len(settings.Rules))}
	seen := map[string]struct{}{}
	for _, rule := range settings.Rules {
		slug := strings.TrimSpace(rule.Model)
		if slug == "" || len(slug) > codexModelReasoningMaxSlugBytes || strings.ContainsAny(slug, "*?[]") || strings.IndexFunc(slug, unicode.IsSpace) >= 0 || strings.IndexFunc(slug, unicode.IsControl) >= 0 {
			return CodexModelReasoningSettings{}, fmt.Errorf("invalid Codex model slug %q", slug)
		}
		if _, ok := seen[slug]; ok {
			return CodexModelReasoningSettings{}, fmt.Errorf("duplicate Codex model slug %q", slug)
		}
		seen[slug] = struct{}{}
		if len(rule.SupportedReasoningLevels) == 0 || len(rule.SupportedReasoningLevels) > len(codexReasoningEfforts) {
			return CodexModelReasoningSettings{}, fmt.Errorf("model %q must have 1-%d reasoning levels", slug, len(codexReasoningEfforts))
		}
		efforts := map[string]struct{}{}
		levels := make([]string, 0, len(rule.SupportedReasoningLevels))
		for _, raw := range rule.SupportedReasoningLevels {
			effort := strings.TrimSpace(raw)
			if _, ok := codexReasoningEfforts[effort]; !ok {
				return CodexModelReasoningSettings{}, fmt.Errorf("unsupported Codex reasoning level %q", effort)
			}
			if _, ok := efforts[effort]; ok {
				return CodexModelReasoningSettings{}, fmt.Errorf("duplicate Codex reasoning level %q", effort)
			}
			efforts[effort] = struct{}{}
			levels = append(levels, effort)
		}
		defaultEffort := strings.TrimSpace(rule.DefaultReasoningLevel)
		if _, ok := efforts[defaultEffort]; !ok {
			return CodexModelReasoningSettings{}, fmt.Errorf("default reasoning level %q is not supported by model %q", defaultEffort, slug)
		}
		out.Rules = append(out.Rules, CodexModelReasoningRule{Model: slug, SupportedReasoningLevels: levels, DefaultReasoningLevel: defaultEffort})
	}
	return out, nil
}
func (s *SettingService) GetCodexModelReasoningSettings(ctx context.Context) (CodexModelReasoningSettings, error) {
	empty := CodexModelReasoningSettings{Rules: []CodexModelReasoningRule{}}
	if s == nil || s.settingRepo == nil {
		return empty, nil
	}
	if cached, ok := s.codexModelReasoningCache.Load().(*cachedCodexModelReasoningSettings); ok && cached != nil && time.Now().UnixNano() < cached.expiresAt {
		return cloneCodexModelReasoningSettings(cached.settings), cached.err
	}
	s.codexModelReasoningMu.Lock()
	defer s.codexModelReasoningMu.Unlock()
	if cached, ok := s.codexModelReasoningCache.Load().(*cachedCodexModelReasoningSettings); ok && cached != nil && time.Now().UnixNano() < cached.expiresAt {
		return cloneCodexModelReasoningSettings(cached.settings), cached.err
	}
	dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), codexModelReasoningDBTimeout)
	defer cancel()
	raw, err := s.settingRepo.GetValue(dbCtx, SettingKeyCodexModelReasoningRules)
	if errors.Is(err, ErrSettingNotFound) {
		err = nil
		raw = ""
	}
	settings := empty
	if err == nil && strings.TrimSpace(raw) != "" {
		if err = json.Unmarshal([]byte(raw), &settings); err == nil {
			settings, err = normalizeCodexModelReasoningSettings(settings)
		}
	}
	ttl := codexModelReasoningCacheTTL
	if err != nil {
		ttl = codexModelReasoningErrorTTL
		settings = empty
	}
	s.codexModelReasoningCache.Store(&cachedCodexModelReasoningSettings{settings: settings, err: err, expiresAt: time.Now().Add(ttl).UnixNano()})
	return cloneCodexModelReasoningSettings(settings), err
}
func (s *SettingService) SetCodexModelReasoningSettings(ctx context.Context, settings CodexModelReasoningSettings) error {
	if s == nil || s.settingRepo == nil {
		return errors.New("settings repository unavailable")
	}
	normalized, err := normalizeCodexModelReasoningSettings(settings)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCodexModelReasoningSettings, err)
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	s.codexModelReasoningMu.Lock()
	defer s.codexModelReasoningMu.Unlock()
	if err = s.settingRepo.Set(ctx, SettingKeyCodexModelReasoningRules, string(encoded)); err != nil {
		return err
	}
	s.codexModelReasoningCache.Store(&cachedCodexModelReasoningSettings{settings: normalized, expiresAt: time.Now().Add(codexModelReasoningCacheTTL).UnixNano()})
	return nil
}

// Runtime lookup fails safely: never advertise unverified capabilities on a DB outage.
func (s *SettingService) CodexModelReasoningRulesForManifest(ctx context.Context) []CodexModelReasoningRule {
	settings, err := s.GetCodexModelReasoningSettings(ctx)
	if err != nil {
		slog.Warn("Codex model reasoning settings unavailable; using catalog defaults", "error", err)
		return nil
	}
	return settings.Rules
}
func findCodexModelReasoningRule(rules []CodexModelReasoningRule, slug string) (CodexModelReasoningRule, bool) {
	for _, rule := range rules {
		if rule.Model == slug {
			return rule, true
		}
	}
	return CodexModelReasoningRule{}, false
}
func applyCodexModelReasoningRule(descriptor *configuredCodexModelDescriptor, rule CodexModelReasoningRule, upstream *codexModelMetadataOverride) {
	if descriptor == nil {
		return
	}
	levels := rule.SupportedReasoningLevels
	if upstream != nil {
		if upstream.reasoningConflict || (upstream.Reasoning != nil && !*upstream.Reasoning) {
			return
		}
		if len(upstream.SupportedReasoningLevels) > 0 {
			levels = intersectOrderedStrings(levels, normalizeReasoningLevels(upstream.SupportedReasoningLevels))
			if len(levels) == 0 {
				return
			}
		}
	}
	defaultLevel := rule.DefaultReasoningLevel
	if !stringSliceContains(levels, defaultLevel) {
		defaultLevel = levels[0]
	}
	descriptor.DefaultReasoningLevel = &defaultLevel
	descriptor.SupportedReasoningLevels = make([]configuredCodexReasoningLevel, 0, len(levels))
	for _, level := range levels {
		descriptor.SupportedReasoningLevels = append(descriptor.SupportedReasoningLevels, configuredCodexReasoningLevel{Effort: level, Description: configuredCodexReasoningLevelDescription(level)})
	}
}

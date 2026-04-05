package config

// normalize.go — normalization helpers for scheduling, cloudflared, and thinking settings.
// Extracted from manager.go to keep that file focused on state management.

import "strings"

func normalizeSchedulingMode(mode SchedulingMode) SchedulingMode {
	switch SchedulingMode(strings.TrimSpace(string(mode))) {
	case SchedulingModeCacheFirst:
		return SchedulingModeCacheFirst
	case SchedulingModePerformance:
		return SchedulingModePerformance
	default:
		return SchedulingModeBalance
	}
}

func normalizeCloudflaredMode(mode CloudflaredMode) CloudflaredMode {
	switch CloudflaredMode(strings.TrimSpace(string(mode))) {
	case CloudflaredModeAuth:
		return CloudflaredModeAuth
	default:
		return CloudflaredModeQuick
	}
}

func defaultCloudflaredSettings() CloudflaredSettings {
	return CloudflaredSettings{
		Enabled:  false,
		Mode:     CloudflaredModeQuick,
		Token:    "",
		UseHTTP2: true,
	}
}

func normalizeCloudflaredSettings(settings CloudflaredSettings) CloudflaredSettings {
	normalized := defaultCloudflaredSettings()
	normalized.Enabled = settings.Enabled
	normalized.Mode = normalizeCloudflaredMode(settings.Mode)
	normalized.Token = strings.TrimSpace(settings.Token)
	normalized.UseHTTP2 = settings.UseHTTP2
	return normalized
}

func normalizeThinkingMode(mode ThinkingMode) ThinkingMode {
	switch ThinkingMode(strings.TrimSpace(string(mode))) {
	case ThinkingModeOff:
		return ThinkingModeOff
	case ThinkingModeForce:
		return ThinkingModeForce
	default:
		return ThinkingModeAuto
	}
}

func defaultThinkingSettings() ThinkingSettings {
	return ThinkingSettings{
		Suffix:                    defaultThinkingSuffix,
		Mode:                      ThinkingModeAuto,
		FallbackTags:              []string{"<thinking>", "<think>"},
		RequireAnthropicSignature: true,
		ForceForAnthropic:         false,
		MaxForcedThinkingTokens:   defaultMaxForcedThinkingTokens,
	}
}

func normalizeThinkingSettings(settings ThinkingSettings) ThinkingSettings {
	normalized := defaultThinkingSettings()
	if suffix := strings.TrimSpace(settings.Suffix); suffix != "" {
		normalized.Suffix = suffix
	}
	normalized.Mode = normalizeThinkingMode(settings.Mode)
	normalized.FallbackTags = normalizeThinkingFallbackTags(settings.FallbackTags)
	normalized.RequireAnthropicSignature = settings.RequireAnthropicSignature
	normalized.ForceForAnthropic = settings.ForceForAnthropic
	if settings.MaxForcedThinkingTokens > 0 {
		normalized.MaxForcedThinkingTokens = settings.MaxForcedThinkingTokens
	}
	return normalized
}

func normalizeThinkingFallbackTags(tags []string) []string {
	if len(tags) == 0 {
		return append([]string(nil), defaultThinkingSettings().FallbackTags...)
	}

	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return append([]string(nil), defaultThinkingSettings().FallbackTags...)
	}
	return normalized
}

func cloneThinkingSettings(settings ThinkingSettings) ThinkingSettings {
	clone := settings
	clone.FallbackTags = append([]string(nil), settings.FallbackTags...)
	return clone
}

package kiro

import (
	"encoding/json"
	"io"
	"strings"

	adapterencode "cliro-go/internal/adapter/encode"
	"cliro-go/internal/config"
	provider "cliro-go/internal/provider"

	"github.com/google/uuid"
)

func collectCompletion(body io.Reader, req provider.ChatRequest) (provider.CompletionOutcome, error) {
	return collectCompletionWithCallback(body, req, nil)
}

func collectCompletionWithCallback(body io.Reader, req provider.ChatRequest, callback func(StreamEvent)) (provider.CompletionOutcome, error) {
	outcome := provider.CompletionOutcome{
		ID:    "chatcmpl-" + uuid.NewString(),
		Model: req.Model,
	}

	parser := NewStreamParser(body)
	var textBuilder strings.Builder
	var thinkingBuilder strings.Builder

	for {
		event, err := parser.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return outcome, err
		}
		if event.Text != "" {
			textBuilder.WriteString(event.Text)
		}
		if event.Thinking != "" {
			thinkingBuilder.WriteString(event.Thinking)
		}
		mergeUsage(&outcome.Usage, event.Usage)

		// Forward event to callback for live streaming
		if callback != nil && (event.Text != "" || event.Thinking != "" || event.Usage.TotalTokens > 0) {
			callback(event)
		}
	}

	toolUses := deduplicateToolUses(parser.ToolUses())
	text := sanitizeModelOutputText(textBuilder.String())
	if extracted, ok := extractBracketToolUses(text); ok {
		toolUses = deduplicateToolUses(append(toolUses, extracted...))
		text = ""
	}

	outcome.Text = text
	outcome.Thinking = sanitizeModelOutputText(thinkingBuilder.String())
	outcome.ThinkingSignature = adapterencode.StableThinkingSignature(outcome.Thinking)
	outcome.ToolUses = toolUses
	estimateUsageIfMissing(&outcome.Usage, req, &outcome)
	return outcome, nil
}

func mergeUsage(stats *config.ProxyStats, usage UsageSnapshot) {
	if stats == nil {
		return
	}
	if usage.PromptTokens > 0 {
		stats.PromptTokens = usage.PromptTokens
	}
	if usage.CompletionTokens > 0 {
		stats.CompletionTokens = usage.CompletionTokens
	}
	if usage.TotalTokens > 0 {
		stats.TotalTokens = usage.TotalTokens
	}
}

func deduplicateToolUses(toolUses []provider.ToolUse) []provider.ToolUse {
	if len(toolUses) == 0 {
		return nil
	}

	byID := make(map[string]provider.ToolUse)
	orderedIDs := make([]string, 0, len(toolUses))
	withoutID := make([]provider.ToolUse, 0)
	for _, toolUse := range toolUses {
		toolID := strings.TrimSpace(toolUse.ID)
		if toolID == "" {
			withoutID = append(withoutID, toolUse)
			continue
		}
		existing, ok := byID[toolID]
		if !ok {
			orderedIDs = append(orderedIDs, toolID)
			byID[toolID] = toolUse
			continue
		}
		if toolUsePayloadSize(toolUse) > toolUsePayloadSize(existing) {
			byID[toolID] = toolUse
		}
	}

	candidates := make([]provider.ToolUse, 0, len(toolUses))
	for _, toolID := range orderedIDs {
		candidates = append(candidates, byID[toolID])
	}
	candidates = append(candidates, withoutID...)

	seen := make(map[string]struct{}, len(candidates))
	unique := make([]provider.ToolUse, 0, len(candidates))
	for _, toolUse := range candidates {
		key := strings.TrimSpace(toolUse.Name) + "|" + marshalToolInput(toolUse.Input)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, toolUse)
	}
	return unique
}

func toolUsePayloadSize(toolUse provider.ToolUse) int {
	return len(marshalToolInput(toolUse.Input))
}

func marshalToolInput(input map[string]any) string {
	encoded, err := json.Marshal(defaultIfNilMap(input))
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func extractBracketToolUses(text string) ([]provider.ToolUse, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" || (!strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "{")) {
		return nil, false
	}

	var rawItems []any
	if strings.HasPrefix(trimmed, "{") {
		rawItems = []any{map[string]any{}}
		if err := json.Unmarshal([]byte(trimmed), &rawItems[0]); err != nil {
			return nil, false
		}
	} else if err := json.Unmarshal([]byte(trimmed), &rawItems); err != nil {
		return nil, false
	}

	toolUses := make([]provider.ToolUse, 0, len(rawItems))
	for _, item := range rawItems {
		toolUse, ok := bracketToolUse(item)
		if !ok {
			return nil, false
		}
		toolUses = append(toolUses, toolUse)
	}
	if len(toolUses) == 0 {
		return nil, false
	}
	return toolUses, true
}

func bracketToolUse(item any) (provider.ToolUse, bool) {
	object, ok := item.(map[string]any)
	if !ok {
		return provider.ToolUse{}, false
	}

	name := strings.TrimSpace(asString(object["name"]))
	arguments := anyToMap(object["input"])
	if function, ok := object["function"].(map[string]any); ok {
		if name == "" {
			name = strings.TrimSpace(asString(function["name"]))
		}
		if len(arguments) == 0 {
			arguments = anyToMap(function["arguments"])
			if len(arguments) == 0 {
				arguments = parseToolArguments(asString(function["arguments"]))
			}
		}
	}
	if len(arguments) == 0 {
		arguments = anyToMap(object["arguments"])
		if len(arguments) == 0 {
			arguments = parseToolArguments(asString(object["arguments"]))
		}
	}
	if name == "" {
		return provider.ToolUse{}, false
	}
	return provider.ToolUse{
		ID:    strings.TrimSpace(firstNonEmpty(asString(object["id"]), asString(object["toolUseId"]), asString(object["call_id"]))),
		Name:  name,
		Input: defaultIfNilMap(arguments),
	}, true
}

package kiro

import (
	"encoding/json"
	"strings"

	"cliro-go/internal/config"
	provider "cliro-go/internal/provider"
)

func estimateUsageIfMissing(stats *config.ProxyStats, req provider.ChatRequest, outcome *provider.CompletionOutcome) {
	if stats == nil || outcome == nil {
		return
	}
	if stats.PromptTokens <= 0 {
		stats.PromptTokens = estimatePromptTokens(req)
	}
	if stats.CompletionTokens <= 0 {
		stats.CompletionTokens = estimateCompletionTokens(*outcome)
	}
	if stats.TotalTokens <= 0 {
		stats.TotalTokens = stats.PromptTokens + stats.CompletionTokens
	}
	outcome.Usage = *stats
}

func estimatePromptTokens(req provider.ChatRequest) int {
	parts := make([]string, 0, len(req.Messages)+(len(req.Tools)*2)+4)
	if model := strings.TrimSpace(req.Model); model != "" {
		parts = append(parts, model)
	}
	for _, message := range req.Messages {
		if role := strings.TrimSpace(message.Role); role != "" {
			parts = append(parts, role)
		}
		if name := strings.TrimSpace(message.Name); name != "" {
			parts = append(parts, name)
		}
		if text := sanitizePromptText(messageTextContent(message.Content)); text != "" {
			parts = append(parts, text)
		}
		for _, toolCall := range message.ToolCalls {
			parts = append(parts, strings.TrimSpace(toolCall.Function.Name), strings.TrimSpace(toolCall.Function.Arguments))
		}
		if toolCallID := strings.TrimSpace(message.ToolCallID); toolCallID != "" {
			parts = append(parts, toolCallID)
		}
	}
	for _, tool := range req.Tools {
		parts = append(parts, strings.TrimSpace(tool.Function.Name), strings.TrimSpace(tool.Function.Description), marshalAny(tool.Function.Parameters))
	}
	if user := strings.TrimSpace(req.User); user != "" {
		parts = append(parts, user)
	}
	if req.ToolChoice != nil {
		parts = append(parts, marshalAny(req.ToolChoice))
	}
	return estimateTokenText(strings.Join(nonEmptyStrings(parts), "\n"))
}

func estimateCompletionTokens(outcome provider.CompletionOutcome) int {
	parts := []string{
		sanitizeModelOutputText(outcome.Thinking),
		sanitizeModelOutputText(outcome.Text),
	}
	for _, toolUse := range outcome.ToolUses {
		parts = append(parts, strings.TrimSpace(toolUse.Name), marshalAny(toolUse.Input))
	}
	return estimateTokenText(strings.Join(nonEmptyStrings(parts), "\n"))
}

func estimateTokenText(text string) int {
	runeCount := len([]rune(strings.TrimSpace(text)))
	if runeCount <= 0 {
		return 0
	}
	estimated := runeCount / 4
	if estimated <= 0 {
		estimated = 1
	}
	return estimated
}

func marshalAny(value any) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func nonEmptyStrings(parts []string) []string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			filtered = append(filtered, strings.TrimSpace(part))
		}
	}
	return filtered
}

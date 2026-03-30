package decode

import (
	"encoding/json"
	"fmt"
	"strings"

	"cliro-go/internal/adapter/ir"
	"cliro-go/internal/protocol/openai"
)

func roleFromString(value string) ir.Role {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "system":
		return ir.RoleSystem
	case "developer":
		return ir.RoleDeveloper
	case "assistant":
		return ir.RoleAssistant
	case "tool":
		return ir.RoleTool
	default:
		return ir.RoleUser
	}
}

func validateModel(model string) error {
	if strings.TrimSpace(model) == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

func mergeConsecutiveOpenAIMessages(messages []openai.Message) []openai.Message {
	if len(messages) <= 1 {
		return messages
	}

	merged := make([]openai.Message, 0, len(messages))
	current := cloneOpenAIMessage(messages[0])

	for idx := 1; idx < len(messages); idx++ {
		next := messages[idx]
		if !strings.EqualFold(strings.TrimSpace(current.Role), strings.TrimSpace(next.Role)) {
			merged = append(merged, current)
			current = cloneOpenAIMessage(next)
			continue
		}
		if strings.EqualFold(strings.TrimSpace(current.Role), "tool") && strings.TrimSpace(current.ToolCallID) != strings.TrimSpace(next.ToolCallID) {
			merged = append(merged, current)
			current = cloneOpenAIMessage(next)
			continue
		}

		current.Content = mergeOpenAIContent(current.Content, next.Content)
		current.ToolCalls = append(current.ToolCalls, next.ToolCalls...)
		current.Name = firstNonEmptyString(current.Name, next.Name)
		current.ToolCallID = firstNonEmptyString(current.ToolCallID, next.ToolCallID)
		current.AdditionalKwargs = mergeAdditionalKwargs(current.AdditionalKwargs, next.AdditionalKwargs)
	}

	merged = append(merged, current)
	return merged
}

func cloneOpenAIMessage(message openai.Message) openai.Message {
	cloned := message
	if len(message.ToolCalls) > 0 {
		cloned.ToolCalls = append([]openai.ToolCall(nil), message.ToolCalls...)
	}
	if len(message.AdditionalKwargs) > 0 {
		cloned.AdditionalKwargs = mergeAdditionalKwargs(nil, message.AdditionalKwargs)
	}
	return cloned
}

func mergeOpenAIContent(current any, next any) any {
	left := strings.TrimSpace(openAIContentToText(current))
	right := strings.TrimSpace(openAIContentToText(next))
	switch {
	case left == "" && right == "":
		return nil
	case left == "":
		return right
	case right == "":
		return left
	default:
		return left + "\n\n" + right
	}
}

func openAIContentToText(content any) string {
	switch typed := content.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(openAIContentToText(item))
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	case map[string]any:
		if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}
		if thinking, ok := typed["thinking"].(string); ok && strings.TrimSpace(thinking) != "" {
			return strings.TrimSpace(thinking)
		}
		encoded, _ := json.Marshal(typed)
		return strings.TrimSpace(string(encoded))
	default:
		encoded, _ := json.Marshal(typed)
		return strings.TrimSpace(string(encoded))
	}
}

func mergeAdditionalKwargs(current map[string]any, next map[string]any) map[string]any {
	if len(current) == 0 && len(next) == 0 {
		return nil
	}

	merged := make(map[string]any, len(current)+len(next))
	for key, value := range current {
		merged[key] = value
	}
	for key, value := range next {
		if key == "thinking_blocks" {
			merged[key] = appendThinkingBlockMetadata(merged[key], value)
			continue
		}
		if _, exists := merged[key]; !exists {
			merged[key] = value
		}
	}
	return merged
}

func appendThinkingBlockMetadata(current any, next any) any {
	blocks := make([]any, 0)
	blocks = append(blocks, thinkingBlockMetadataSlice(current)...)
	blocks = append(blocks, thinkingBlockMetadataSlice(next)...)
	if len(blocks) == 0 {
		return nil
	}
	return blocks
}

func thinkingBlockMetadataSlice(value any) []any {
	switch typed := value.(type) {
	case nil:
		return nil
	case []any:
		return append([]any(nil), typed...)
	case []ir.ThinkingBlock:
		result := make([]any, 0, len(typed))
		for _, block := range typed {
			result = append(result, map[string]any{"thinking": block.Thinking, "signature": block.Signature})
		}
		return result
	default:
		return []any{typed}
	}
}

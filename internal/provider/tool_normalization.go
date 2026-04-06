package provider

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

const DefaultToolNameLimit = 64

var unsupportedKiroToolNames = map[string]struct{}{
	"web_search":     {},
	"websearch":      {},
	"code_execution": {},
	"text_editor":    {},
	"computer":       {},
}

type ToolNameMapping struct {
	forward map[string]string
	reverse map[string]string
}

func RemapChatRequestToolNames(req ChatRequest, mapping ToolNameMapping) ChatRequest {
	cloned := req
	if len(req.Tools) > 0 {
		cloned.Tools = append([]Tool(nil), req.Tools...)
		for index := range cloned.Tools {
			cloned.Tools[index].Function.Name = mapping.Remap(cloned.Tools[index].Function.Name)
			cloned.Tools[index].Function.Parameters = NormalizeToolSchema(cloned.Tools[index].Function.Parameters)
		}
	}
	if len(req.Messages) > 0 {
		cloned.Messages = append([]Message(nil), req.Messages...)
		for index := range cloned.Messages {
			if len(req.Messages[index].ToolCalls) == 0 {
				continue
			}
			cloned.Messages[index].ToolCalls = append([]ToolCall(nil), req.Messages[index].ToolCalls...)
			for toolIndex := range cloned.Messages[index].ToolCalls {
				cloned.Messages[index].ToolCalls[toolIndex].Function.Name = mapping.Remap(cloned.Messages[index].ToolCalls[toolIndex].Function.Name)
			}
		}
	}
	return cloned
}

func RestoreToolUseNames(toolUses []ToolUse, mapping ToolNameMapping) []ToolUse {
	if len(toolUses) == 0 {
		return nil
	}
	cloned := append([]ToolUse(nil), toolUses...)
	for index := range cloned {
		cloned[index].Name = mapping.Restore(cloned[index].Name)
	}
	return cloned
}

func NormalizeToolsForProvider(providerName string, tools []Tool) []Tool {
	if len(tools) == 0 {
		return nil
	}
	normalized := make([]Tool, 0, len(tools))
	for _, tool := range tools {
		if !ToolSupportedByProvider(providerName, tool) {
			continue
		}
		normalized = append(normalized, tool)
	}
	return normalized
}

func ToolSupportedByProvider(providerName string, tool Tool) bool {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	toolType := strings.ToLower(strings.TrimSpace(tool.Type))
	toolName := strings.ToLower(strings.TrimSpace(tool.Function.Name))

	switch providerName {
	case "kiro":
		if toolType != "" && toolType != "function" {
			return false
		}
		if _, ok := unsupportedKiroToolNames[toolName]; ok {
			return false
		}
	}
	return true
}

func BuildToolNameMapping(tools []Tool, messages []Message, maxLen int) ToolNameMapping {
	if maxLen <= 0 {
		maxLen = DefaultToolNameLimit
	}
	mapping := ToolNameMapping{forward: map[string]string{}, reverse: map[string]string{}}
	seenShort := make(map[string]struct{})
	register := func(name string) {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return
		}
		if _, ok := mapping.forward[trimmed]; ok {
			return
		}
		short := shortenToolName(trimmed, maxLen)
		for short != trimmed {
			if _, exists := seenShort[short]; !exists {
				break
			}
			short = shortenToolName(trimmed+"_", maxLen)
		}
		mapping.forward[trimmed] = short
		mapping.reverse[short] = trimmed
		seenShort[short] = struct{}{}
	}
	for _, tool := range tools {
		register(tool.Function.Name)
	}
	for _, message := range messages {
		for _, call := range message.ToolCalls {
			register(call.Function.Name)
		}
	}
	return mapping
}

func (m ToolNameMapping) Remap(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	if mapped, ok := m.forward[trimmed]; ok {
		return mapped
	}
	return trimmed
}

func (m ToolNameMapping) Restore(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	if restored, ok := m.reverse[trimmed]; ok {
		return restored
	}
	return trimmed
}

func NormalizeToolSchema(schema any) any {
	mapSchema, ok := schema.(map[string]any)
	if !ok || mapSchema == nil {
		return map[string]any{"type": "object", "properties": map[string]any{}, "required": []any{}}
	}
	cloned := make(map[string]any, len(mapSchema)+2)
	for key, value := range mapSchema {
		cloned[key] = value
	}
	if _, ok := cloned["type"]; !ok {
		cloned["type"] = "object"
	}
	if _, ok := cloned["properties"]; !ok {
		cloned["properties"] = map[string]any{}
	}
	if _, ok := cloned["required"]; !ok {
		cloned["required"] = []any{}
	}
	return cloned
}

func shortenToolName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	sum := sha1.Sum([]byte(name))
	suffix := "_" + hex.EncodeToString(sum[:])[:10]
	limit := maxLen - len(suffix)
	if limit <= 0 {
		return name[:maxLen]
	}
	return name[:limit] + suffix
}

package decode

import (
	"encoding/json"
	"fmt"
	"strings"

	"cliro-go/internal/adapter/ir"
	"cliro-go/internal/protocol/anthropic"
	"cliro-go/internal/protocol/openai"

	"github.com/google/uuid"
)

const anthropicFallbackUserContent = "."

func AnthropicMessagesToIR(req anthropic.MessagesRequest) (ir.Request, error) {
	chatReq, err := convertAnthropicToOpenAI(req)
	if err != nil {
		return ir.Request{}, err
	}

	out, err := OpenAIChatToIR(chatReq)
	if err != nil {
		return ir.Request{}, err
	}
	out.Protocol = ir.ProtocolAnthropic
	out.Endpoint = ir.EndpointAnthropicMessages
	return out, nil
}

func convertAnthropicToOpenAI(req anthropic.MessagesRequest) (openai.ChatRequest, error) {
	out := openai.ChatRequest{
		Model:       strings.TrimSpace(req.Model),
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		User:        strings.TrimSpace(req.User),
		ToolChoice:  req.ToolChoice,
		Tools:       convertAnthropicTools(req.Tools),
	}
	if req.MaxTokens > 0 {
		maxTokens := req.MaxTokens
		out.MaxTokens = &maxTokens
	}

	if systemText := strings.TrimSpace(anthropicSystemToText(req.System)); systemText != "" {
		out.Messages = append(out.Messages, openai.Message{Role: "system", Content: systemText})
	}

	for _, message := range req.Messages {
		messageContent := stripAnthropicCacheControl(message.Content)
		role := strings.ToLower(strings.TrimSpace(message.Role))
		if role == "" {
			role = "user"
		}
		if role != "assistant" {
			role = "user"
		}

		text, toolCalls, toolResults, thinkingBlocks := convertAnthropicMessageContent(role, messageContent)
		if role == "assistant" {
			if strings.TrimSpace(text) == "" && len(toolCalls) == 0 && len(thinkingBlocks) == 0 {
				continue
			}
			assistantMessage := openai.Message{Role: "assistant", ToolCalls: toolCalls}
			if len(thinkingBlocks) > 0 {
				assistantMessage.AdditionalKwargs = thinkingBlocksToAdditionalKwargs(thinkingBlocks)
			}
			if strings.TrimSpace(text) != "" {
				assistantMessage.Content = text
			} else {
				assistantMessage.Content = nil
			}
			out.Messages = append(out.Messages, assistantMessage)
			continue
		}

		if len(toolResults) > 0 {
			out.Messages = append(out.Messages, toolResults...)
		}
		if strings.TrimSpace(text) != "" {
			userMessage := openai.Message{Role: "user", Content: text}
			if len(thinkingBlocks) > 0 {
				userMessage.AdditionalKwargs = thinkingBlocksToAdditionalKwargs(thinkingBlocks)
			}
			out.Messages = append(out.Messages, userMessage)
		}
	}

	out.Messages = mergeConsecutiveOpenAIMessages(out.Messages)

	if len(out.Messages) == 0 {
		return out, fmt.Errorf("messages are empty")
	}
	if strings.TrimSpace(out.Model) == "" {
		return out, fmt.Errorf("model is required")
	}

	return out, nil
}

func convertAnthropicTools(tools []anthropic.Tool) []openai.Tool {
	if len(tools) == 0 {
		return nil
	}

	result := make([]openai.Tool, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			continue
		}
		result = append(result, openai.Tool{
			Type: "function",
			Function: openai.ToolFunction{
				Name:        name,
				Description: strings.TrimSpace(tool.Description),
				Parameters:  tool.InputSchema,
			},
		})
	}
	return result
}

func convertAnthropicMessageContent(role string, content any) (string, []openai.ToolCall, []openai.Message, []ir.ThinkingBlock) {
	switch typed := content.(type) {
	case string:
		return strings.TrimSpace(typed), nil, nil, nil
	case []any:
		textParts := make([]string, 0, len(typed))
		toolCalls := make([]openai.ToolCall, 0)
		toolResults := make([]openai.Message, 0)
		thinkingBlocks := make([]ir.ThinkingBlock, 0)

		for _, item := range typed {
			sanitized := stripAnthropicCacheControl(item)
			block, ok := sanitized.(map[string]any)
			if !ok {
				fallback := strings.TrimSpace(anthropicContentToText(sanitized))
				if fallback != "" {
					textParts = append(textParts, fallback)
				}
				continue
			}

			blockType, _ := block["type"].(string)
			switch strings.ToLower(strings.TrimSpace(blockType)) {
			case "text":
				if text, ok := block["text"].(string); ok && strings.TrimSpace(text) != "" {
					textParts = append(textParts, text)
				}
			case "thinking":
				if text, ok := block["thinking"].(string); ok && strings.TrimSpace(text) != "" {
					signature, _ := block["signature"].(string)
					thinkingBlocks = append(thinkingBlocks, ir.ThinkingBlock{Thinking: text, Signature: strings.TrimSpace(signature)})
				}
			case "redacted_thinking":
				continue
			case "tool_use":
				if role != "assistant" {
					continue
				}
				name, _ := block["name"].(string)
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				id, _ := block["id"].(string)
				id = strings.TrimSpace(id)
				if id == "" {
					id = "toolu_" + uuid.NewString()[:8]
				}

				input := map[string]any{}
				if parsed, ok := block["input"].(map[string]any); ok && parsed != nil {
					input = parsed
				}
				encodedInput, _ := json.Marshal(input)

				toolCalls = append(toolCalls, openai.ToolCall{
					ID:   id,
					Type: "function",
					Function: openai.ToolCallFunction{
						Name:      name,
						Arguments: string(encodedInput),
					},
				})
			case "tool_result":
				if role == "assistant" {
					continue
				}
				toolUseID, _ := block["tool_use_id"].(string)
				toolUseID = strings.TrimSpace(toolUseID)
				if toolUseID == "" {
					continue
				}
				toolContent := strings.TrimSpace(anthropicContentToText(stripAnthropicCacheControl(block["content"])))
				if toolContent == "" {
					toolContent = anthropicFallbackUserContent
				}
				toolResults = append(toolResults, openai.Message{
					Role:       "tool",
					ToolCallID: toolUseID,
					Content:    toolContent,
				})
			default:
				fallback := strings.TrimSpace(anthropicContentToText(block))
				if fallback != "" {
					textParts = append(textParts, fallback)
				}
			}
		}

		return strings.TrimSpace(strings.Join(textParts, "\n")), toolCalls, toolResults, thinkingBlocks
	default:
		return strings.TrimSpace(anthropicContentToText(stripAnthropicCacheControl(content))), nil, nil, nil
	}
}

func anthropicSystemToText(system any) string {
	switch typed := system.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(anthropicContentToText(item))
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	default:
		return strings.TrimSpace(anthropicContentToText(system))
	}
}

func anthropicContentToText(content any) string {
	content = stripAnthropicCacheControl(content)
	switch typed := content.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if block, ok := item.(map[string]any); ok {
				blockType, _ := block["type"].(string)
				switch strings.ToLower(strings.TrimSpace(blockType)) {
				case "text":
					if text, ok := block["text"].(string); ok && strings.TrimSpace(text) != "" {
						parts = append(parts, text)
					}
					continue
				case "thinking":
					if thinking, ok := block["thinking"].(string); ok && strings.TrimSpace(thinking) != "" {
						parts = append(parts, thinking)
					}
					continue
				case "redacted_thinking":
					if data := strings.TrimSpace(anthropicContentToText(block["data"])); data != "" {
						parts = append(parts, "[Redacted Thinking: "+data+"]")
					}
					continue
				case "tool_result":
					contentText := strings.TrimSpace(anthropicContentToText(block["content"]))
					if contentText != "" {
						parts = append(parts, contentText)
					}
					continue
				case "tool_use":
					if name, ok := block["name"].(string); ok && strings.TrimSpace(name) != "" {
						parts = append(parts, name)
					}
					if input := strings.TrimSpace(anthropicContentToText(block["input"])); input != "" {
						parts = append(parts, input)
					}
					continue
				}
			}
			fallback := strings.TrimSpace(anthropicContentToText(item))
			if fallback != "" {
				parts = append(parts, fallback)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	case map[string]any:
		if blockType, _ := typed["type"].(string); strings.EqualFold(strings.TrimSpace(blockType), "redacted_thinking") {
			if data := strings.TrimSpace(anthropicContentToText(typed["data"])); data != "" {
				return "[Redacted Thinking: " + data + "]"
			}
		}
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

func thinkingBlocksToAdditionalKwargs(blocks []ir.ThinkingBlock) map[string]any {
	if len(blocks) == 0 {
		return nil
	}

	metadata := make([]any, 0, len(blocks))
	for _, block := range blocks {
		metadata = append(metadata, map[string]any{
			"thinking":  block.Thinking,
			"signature": block.Signature,
		})
	}
	return map[string]any{"thinking_blocks": metadata}
}

func stripAnthropicCacheControl(value any) any {
	switch typed := value.(type) {
	case []any:
		cleaned := make([]any, 0, len(typed))
		for _, item := range typed {
			cleaned = append(cleaned, stripAnthropicCacheControl(item))
		}
		return cleaned
	case map[string]any:
		cleaned := make(map[string]any, len(typed))
		for key, item := range typed {
			if key == "cache_control" {
				continue
			}
			cleaned[key] = stripAnthropicCacheControl(item)
		}
		return cleaned
	default:
		return value
	}
}

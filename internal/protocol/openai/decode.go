package openai

import (
	"encoding/json"
	"fmt"
	"strings"

	contract "cliro/internal/contract"
	"cliro/internal/util"

	"github.com/google/uuid"
)

func ResponsesToIR(req ResponsesRequest) (contract.Request, error) {
	if err := validateModel(req.Model); err != nil {
		return contract.Request{}, err
	}

	metadata := make(map[string]any)
	mergeMetadata(metadata, req.Metadata)
	if strings.TrimSpace(req.PreviousResponseID) != "" {
		metadata["previousResponseID"] = strings.TrimSpace(req.PreviousResponseID)
	}
	if req.ParallelToolCalls != nil {
		metadata["parallelToolCalls"] = *req.ParallelToolCalls
	}
	if req.Store != nil {
		metadata["store"] = *req.Store
	}

	messages := responseInputToIRMessages(req.Input, metadata)
	if instructions := strings.TrimSpace(req.Instructions); instructions != "" {
		messages = append([]contract.Message{{Role: contract.RoleSystem, Content: instructions}}, messages...)
	}
	if len(messages) == 0 {
		messages = []contract.Message{{Role: contract.RoleUser, Content: req.Input}}
	}
	messages = sanitizeToolHistory(messages)

	tools := make([]contract.Tool, 0, len(req.Tools))
	for _, tool := range req.Tools {
		toolType := strings.TrimSpace(tool.Type)
		if toolType == "" {
			toolType = "function"
		}
		tools = append(tools, contract.Tool{
			Type:        toolType,
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			Schema:      tool.Function.Parameters,
		})
	}

	return contract.Request{
		Protocol:    contract.ProtocolOpenAI,
		Endpoint:    contract.EndpointOpenAIResponses,
		Model:       req.Model,
		Thinking:    parseThinkingConfig(req.Reasoning),
		Messages:    messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxOutputTokens,
		Tools:       tools,
		ToolChoice:  req.ToolChoice,
		User:        req.User,
		Metadata:    metadata,
	}, nil
}

func ChatToIR(req ChatRequest) (contract.Request, error) {
	if err := validateModel(req.Model); err != nil {
		return contract.Request{}, err
	}

	messages := make([]contract.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		toolCalls := make([]contract.ToolCall, 0, len(msg.ToolCalls))
		for _, toolCall := range msg.ToolCalls {
			id := strings.TrimSpace(toolCall.ID)
			if id == "" {
				id = "toolu_" + uuid.NewString()[:8]
			}
			toolCalls = append(toolCalls, contract.ToolCall{
				ID:        id,
				Name:      toolCall.Function.Name,
				Arguments: toolCall.Function.Arguments,
			})
		}
		messages = append(messages, contract.Message{
			Role:           roleFromString(msg.Role),
			Content:        msg.Content,
			Name:           msg.Name,
			ToolCalls:      toolCalls,
			ToolCallID:     msg.ToolCallID,
			ThinkingBlocks: thinkingBlocksFromAdditionalKwargs(msg.AdditionalKwargs),
		})
	}
	messages = sanitizeToolHistory(messages)

	tools := make([]contract.Tool, 0, len(req.Tools))
	for _, tool := range req.Tools {
		toolType := strings.TrimSpace(tool.Type)
		if toolType == "" {
			toolType = "function"
		}
		tools = append(tools, contract.Tool{
			Type:        toolType,
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			Schema:      tool.Function.Parameters,
		})
	}

	return contract.Request{
		Protocol:    contract.ProtocolOpenAI,
		Endpoint:    contract.EndpointOpenAIChat,
		Model:       req.Model,
		Thinking:    parseThinkingConfig(req.Reasoning),
		Messages:    messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Tools:       tools,
		ToolChoice:  req.ToolChoice,
		User:        req.User,
		Metadata:    metadataFromChatRequest(req),
	}, nil
}

func CompletionsToIR(req CompletionsRequest) (contract.Request, error) {
	if err := validateModel(req.Model); err != nil {
		return contract.Request{}, err
	}

	messages := []contract.Message{{
		Role:    contract.RoleUser,
		Content: req.Prompt,
	}}

	return contract.Request{
		Protocol:    contract.ProtocolOpenAI,
		Endpoint:    contract.EndpointOpenAICompletions,
		Model:       req.Model,
		Messages:    messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		User:        req.User,
		Metadata:    map[string]any{},
	}, nil
}

func responseInputToIRMessages(input any, metadata map[string]any) []contract.Message {
	switch typed := input.(type) {
	case nil:
		return nil
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		return []contract.Message{{Role: contract.RoleUser, Content: typed}}
	case []any:
		messages := make([]contract.Message, 0, len(typed))
		for _, item := range typed {
			messages = append(messages, responsesInputItemToIRMessages(item, metadata)...)
		}
		return messages
	case map[string]any:
		return responsesInputItemToIRMessages(typed, metadata)
	default:
		return []contract.Message{{Role: contract.RoleUser, Content: typed}}
	}
}

func responsesInputItemToIRMessages(item any, metadata map[string]any) []contract.Message {
	object, ok := item.(map[string]any)
	if !ok {
		return responseInputToIRMessages(item, metadata)
	}

	if additional, ok := object["additional_kwargs"].(map[string]any); ok {
		captureAdditionalKwargsMetadata(additional, metadata)
	}

	itemType := strings.ToLower(strings.TrimSpace(asString(object["type"])))
	if itemType == "" && strings.TrimSpace(asString(object["role"])) != "" {
		itemType = "message"
	}

	switch itemType {
	case "message":
		role := roleFromString(strings.TrimSpace(asString(object["role"])))
		if role == "" {
			role = contract.RoleUser
		}
		content := normalizeResponsesMessageContent(object["content"], role)
		if strings.TrimSpace(content) == "" {
			return nil
		}
		return []contract.Message{{Role: role, Content: content}}
	case "function_call", "custom_tool_call":
		callID := strings.TrimSpace(asString(object["call_id"]))
		if callID == "" {
			callID = "toolu_" + uuid.NewString()[:8]
		}
		name := strings.TrimSpace(asString(object["name"]))
		arguments := strings.TrimSpace(asString(object["arguments"]))
		if itemType == "custom_tool_call" {
			arguments = strings.TrimSpace(asString(object["input"]))
		}
		if arguments == "" {
			if encoded, err := json.Marshal(object["input"]); err == nil && string(encoded) != "null" {
				arguments = string(encoded)
			}
		}
		return []contract.Message{{
			Role: contract.RoleAssistant,
			ToolCalls: []contract.ToolCall{{
				ID:        callID,
				Name:      name,
				Arguments: util.FirstNonEmpty(arguments, "{}"),
			}},
		}}
	case "function_call_output", "custom_tool_call_output":
		toolCallID := strings.TrimSpace(asString(object["call_id"]))
		if toolCallID == "" {
			return nil
		}
		return []contract.Message{{Role: contract.RoleTool, ToolCallID: toolCallID, Content: normalizeResponsesToolOutput(object["output"])}}
	case "input_text":
		text := strings.TrimSpace(asString(object["text"]))
		if text == "" {
			return nil
		}
		return []contract.Message{{Role: contract.RoleUser, Content: text}}
	case "output_text", "text", "refusal":
		text := strings.TrimSpace(asString(object["text"]))
		if text == "" {
			text = strings.TrimSpace(asString(object["refusal"]))
		}
		if text == "" {
			return nil
		}
		return []contract.Message{{Role: contract.RoleAssistant, Content: text}}
	default:
		content := normalizeResponsesMessageContent(object["content"], roleFromString(strings.TrimSpace(asString(object["role"]))))
		if strings.TrimSpace(content) == "" {
			content = strings.TrimSpace(asString(object["text"]))
		}
		if content == "" {
			encoded, _ := json.Marshal(object)
			content = strings.TrimSpace(string(encoded))
		}
		return []contract.Message{{Role: contract.RoleUser, Content: content}}
	}
}

func normalizeResponsesMessageContent(content any, role contract.Role) string {
	switch typed := content.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			partObject, ok := item.(map[string]any)
			if !ok {
				if text := strings.TrimSpace(asString(item)); text != "" {
					parts = append(parts, text)
				}
				continue
			}
			partType := strings.ToLower(strings.TrimSpace(asString(partObject["type"])))
			switch partType {
			case "input_text", "output_text", "text":
				if text := strings.TrimSpace(asString(partObject["text"])); text != "" {
					parts = append(parts, text)
				}
			case "refusal":
				if text := strings.TrimSpace(asString(partObject["refusal"])); text != "" {
					parts = append(parts, text)
				}
			default:
				if text := strings.TrimSpace(asString(partObject["text"])); text != "" {
					parts = append(parts, text)
				}
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	default:
		encoded, _ := json.Marshal(typed)
		return strings.TrimSpace(string(encoded))
	}
}

func normalizeResponsesToolOutput(output any) string {
	switch typed := output.(type) {
	case nil:
		return ""
	case string:
		return typed
	default:
		encoded, _ := json.Marshal(typed)
		return strings.TrimSpace(string(encoded))
	}
}

func metadataFromChatRequest(req ChatRequest) map[string]any {
	metadata := make(map[string]any)
	mergeMetadata(metadata, req.Metadata)
	for _, message := range req.Messages {
		captureAdditionalKwargsMetadata(message.AdditionalKwargs, metadata)
	}
	return metadata
}

func captureAdditionalKwargsMetadata(additional map[string]any, metadata map[string]any) {
	if len(additional) == 0 || metadata == nil {
		return
	}
	if conversationID, ok := additional["conversationId"].(string); ok && strings.TrimSpace(conversationID) != "" {
		metadata["conversationId"] = strings.TrimSpace(conversationID)
	}
	if continuationID, ok := additional["continuationId"].(string); ok && strings.TrimSpace(continuationID) != "" {
		metadata["continuationId"] = strings.TrimSpace(continuationID)
	}
	if profileARN, ok := additional["profileArn"].(string); ok && strings.TrimSpace(profileARN) != "" {
		metadata["profileArn"] = strings.TrimSpace(profileARN)
	}
}

func thinkingBlocksFromAdditionalKwargs(additional map[string]any) []contract.ThinkingBlock {
	if len(additional) == 0 {
		return nil
	}
	raw, ok := additional["thinking_blocks"]
	if !ok {
		return nil
	}

	switch typed := raw.(type) {
	case []contract.ThinkingBlock:
		return append([]contract.ThinkingBlock(nil), typed...)
	case []any:
		blocks := make([]contract.ThinkingBlock, 0, len(typed))
		for _, item := range typed {
			block, ok := item.(map[string]any)
			if !ok {
				continue
			}
			thinking, _ := block["thinking"].(string)
			signature, _ := block["signature"].(string)
			if strings.TrimSpace(thinking) == "" && strings.TrimSpace(signature) == "" {
				continue
			}
			blocks = append(blocks, contract.ThinkingBlock{Thinking: thinking, Signature: signature})
		}
		if len(blocks) == 0 {
			return nil
		}
		return blocks
	default:
		return nil
	}
}

func mergeMetadata(dst map[string]any, src map[string]any) {
	if dst == nil || len(src) == 0 {
		return
	}
	for key, value := range src {
		dst[key] = value
	}
}

func roleFromString(value string) contract.Role {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "system":
		return contract.RoleSystem
	case "developer":
		return contract.RoleDeveloper
	case "assistant":
		return contract.RoleAssistant
	case "tool":
		return contract.RoleTool
	default:
		return contract.RoleUser
	}
}

func validateModel(model string) error {
	if strings.TrimSpace(model) == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

func sanitizeToolHistory(messages []contract.Message) []contract.Message {
	if len(messages) == 0 {
		return nil
	}

	declared := make(map[string]struct{})
	seenResults := make(map[string]struct{})
	sanitized := make([]contract.Message, 0, len(messages))
	for _, message := range messages {
		current := message
		if current.Role == contract.RoleAssistant && len(current.ToolCalls) > 0 {
			filteredCalls := make([]contract.ToolCall, 0, len(current.ToolCalls))
			seenCalls := make(map[string]struct{}, len(current.ToolCalls))
			for _, call := range current.ToolCalls {
				name := strings.TrimSpace(call.Name)
				if name == "" {
					continue
				}
				id := strings.TrimSpace(call.ID)
				if id == "" {
					id = "toolu_" + uuid.NewString()[:8]
				}
				if _, ok := seenCalls[id]; ok {
					continue
				}
				seenCalls[id] = struct{}{}
				declared[id] = struct{}{}
				call.ID = id
				call.Name = name
				if strings.TrimSpace(call.Arguments) == "" {
					call.Arguments = "{}"
				}
				filteredCalls = append(filteredCalls, call)
			}
			current.ToolCalls = filteredCalls
			if len(current.ToolCalls) == 0 && strings.TrimSpace(openAIContentToText(current.Content)) == "" {
				continue
			}
			sanitized = append(sanitized, current)
			continue
		}

		if current.Role == contract.RoleTool {
			id := strings.TrimSpace(current.ToolCallID)
			if id == "" {
				current.Role = contract.RoleUser
				current.ToolCallID = ""
				if strings.TrimSpace(openAIContentToText(current.Content)) == "" {
					continue
				}
				sanitized = append(sanitized, current)
				continue
			}
			if _, ok := declared[id]; !ok {
				current.Role = contract.RoleUser
				current.ToolCallID = ""
				if strings.TrimSpace(openAIContentToText(current.Content)) == "" {
					continue
				}
				sanitized = append(sanitized, current)
				continue
			}
			if _, ok := seenResults[id]; ok {
				continue
			}
			seenResults[id] = struct{}{}
		}

		sanitized = append(sanitized, current)
	}
	return sanitized
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
		for _, key := range []string{"text", "output", "content", "refusal"} {
			if value, ok := typed[key]; ok {
				text := strings.TrimSpace(openAIContentToText(value))
				if text != "" {
					return text
				}
			}
		}
		encoded, _ := json.Marshal(typed)
		return strings.TrimSpace(string(encoded))
	default:
		encoded, _ := json.Marshal(typed)
		return strings.TrimSpace(string(encoded))
	}
}

func parseThinkingConfig(reasoning map[string]any) contract.ThinkingConfig {
	if len(reasoning) == 0 {
		return contract.ThinkingConfig{}
	}
	return contract.ThinkingConfig{
		Requested: true,
		Mode:      contract.ThinkingModeAuto,
		RawParams: reasoning,
	}
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		encoded, _ := json.Marshal(typed)
		return strings.TrimSpace(string(encoded))
	}
}

package kiro

import (
	"fmt"
	"strings"

	"cliro-go/internal/config"
	provider "cliro-go/internal/provider"

	"github.com/google/uuid"
)

var errMessagesEmpty = fmt.Errorf("messages are empty")

func buildRequestPayload(req provider.ChatRequest, account config.Account) (requestPayload, error) {
	messages, systemPrompt, err := normalizeRequest(req)
	if err != nil {
		return requestPayload{}, err
	}
	if len(messages) == 0 {
		return requestPayload{}, errMessagesEmpty
	}

	historyMessages := append([]normalizedMessage(nil), messages...)
	current := historyMessages[len(historyMessages)-1]
	historyMessages = historyMessages[:len(historyMessages)-1]

	if current.Role == "assistant" {
		historyMessages = append(historyMessages, current)
		current = normalizedMessage{Role: "user", Content: "Continue"}
	}

	injectSystemPrompt(systemPrompt, &historyMessages, &current)
	current.Content = sanitizePromptText(current.Content)
	if strings.TrimSpace(current.Content) == "" {
		if len(current.ToolResults) > 0 {
			current.Content = "Tool results provided."
		} else {
			current.Content = "Continue"
		}
	}

	payload := requestPayload{
		ConversationState: conversationState{
			AgentTaskType:   kiroAgentMode,
			ChatTriggerType: "MANUAL",
			ConversationID:  resolveConversationID(req.Metadata),
			CurrentMessage: currentMessage{
				UserInputMessage: buildCurrentUserInput(req, current),
			},
		},
	}

	if history := buildHistory(req.Model, historyMessages); len(history) > 0 {
		payload.ConversationState.History = history
	}
	if continuationID := resolveContinuationID(req.Metadata); continuationID != "" {
		payload.ConversationState.AgentContinuationID = continuationID
	}
	if profileARN := resolveProfileARN(req.Metadata, account); profileARN != "" {
		payload.ProfileARN = profileARN
	}
	if config := buildInferenceConfig(req); config != nil {
		payload.InferenceConfig = config
	}

	return payload, nil
}

func buildHistory(model string, messages []normalizedMessage) []historyMessage {
	history := make([]historyMessage, 0, len(messages))
	for _, message := range messages {
		switch message.Role {
		case "assistant":
			assistantContent := sanitizePromptText(message.Content)
			if assistantContent == "" {
				if len(message.ToolUses) > 0 {
					assistantContent = "."
				} else {
					assistantContent = "(empty)"
				}
			}
			entry := historyMessage{AssistantResponseMessage: &assistantResponseMessage{Content: assistantContent}}
			if len(message.ToolUses) > 0 {
				entry.AssistantResponseMessage.ToolUses = append([]toolUsePayload(nil), message.ToolUses...)
			}
			history = append(history, entry)
		default:
			userContent := sanitizePromptText(message.Content)
			if userContent == "" {
				if len(message.ToolResults) > 0 {
					userContent = "Tool results provided."
				} else {
					userContent = "Continue"
				}
			}
			entry := historyMessage{UserInputMessage: &userInputMessage{Content: userContent, ModelID: model, Origin: kiroDefaultOrigin}}
			if len(message.ToolResults) > 0 {
				entry.UserInputMessage.UserInputMessageContext = &userInputMessageContext{ToolResults: append([]toolResult(nil), message.ToolResults...)}
			}
			history = append(history, entry)
		}
	}
	return history
}

func buildCurrentUserInput(req provider.ChatRequest, current normalizedMessage) userInputMessage {
	content := sanitizePromptText(current.Content)
	if content == "" {
		if len(current.ToolResults) > 0 {
			content = "Tool results provided."
		} else {
			content = "Continue"
		}
	}

	message := userInputMessage{
		Content: content,
		ModelID: req.Model,
		Origin:  kiroDefaultOrigin,
	}

	context := &userInputMessageContext{}
	if tools := buildToolSpecifications(req.Tools); len(tools) > 0 {
		context.Tools = tools
	}
	if len(current.ToolResults) > 0 {
		context.ToolResults = append([]toolResult(nil), current.ToolResults...)
	}
	if len(context.Tools) > 0 || len(context.ToolResults) > 0 {
		message.UserInputMessageContext = context
	}

	return message
}

func sanitizeHistoryToolResults(history []historyMessage) []historyMessage {
	if len(history) == 0 {
		return history
	}

	validToolUseIDs := collectHistoryToolUseIDs(history)
	for index := range history {
		if history[index].UserInputMessage == nil || history[index].UserInputMessage.UserInputMessageContext == nil {
			continue
		}
		ctx := history[index].UserInputMessage.UserInputMessageContext
		if len(ctx.ToolResults) == 0 {
			continue
		}
		filtered := filterToolResultsByKnownIDs(ctx.ToolResults, validToolUseIDs)
		if len(filtered) == 0 {
			ctx.ToolResults = nil
			if len(ctx.Tools) == 0 {
				history[index].UserInputMessage.UserInputMessageContext = nil
			}
			continue
		}
		ctx.ToolResults = filtered
	}

	return history
}

func sanitizeCurrentToolResults(current *currentMessage, history []historyMessage) {
	if current == nil || current.UserInputMessage.UserInputMessageContext == nil {
		return
	}
	ctx := current.UserInputMessage.UserInputMessageContext
	if len(ctx.ToolResults) == 0 {
		return
	}

	filtered := filterToolResultsByKnownIDs(ctx.ToolResults, collectHistoryToolUseIDs(history))
	if len(filtered) == 0 {
		ctx.ToolResults = nil
		if len(ctx.Tools) == 0 {
			current.UserInputMessage.UserInputMessageContext = nil
		}
		return
	}
	ctx.ToolResults = filtered
}

func collectHistoryToolUseIDs(history []historyMessage) map[string]struct{} {
	ids := make(map[string]struct{})
	for _, item := range history {
		if item.AssistantResponseMessage == nil {
			continue
		}
		for _, toolUse := range item.AssistantResponseMessage.ToolUses {
			id := strings.TrimSpace(toolUse.ToolUseID)
			if id != "" {
				ids[id] = struct{}{}
			}
		}
	}
	return ids
}

func filterToolResultsByKnownIDs(results []toolResult, validIDs map[string]struct{}) []toolResult {
	if len(results) == 0 {
		return results
	}
	filtered := make([]toolResult, 0, len(results))
	seen := make(map[string]struct{}, len(results))
	for _, result := range results {
		id := strings.TrimSpace(result.ToolUseID)
		if id == "" {
			continue
		}
		if len(validIDs) > 0 {
			if _, ok := validIDs[id]; !ok {
				continue
			}
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		filtered = append(filtered, result)
	}
	return filtered
}

func buildToolSpecifications(tools []provider.Tool) []toolSpecification {
	result := make([]toolSpecification, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Function.Name)
		if name == "" {
			continue
		}
		result = append(result, toolSpecification{
			ToolSpecification: toolSpecificationDetails{
				Name:        name,
				Description: defaultIfEmpty(strings.TrimSpace(tool.Function.Description), "Tool: "+name),
				InputSchema: toolInputSchema{JSON: normalizeToolSchema(tool.Function.Parameters)},
			},
		})
	}
	return result
}

func normalizeToolSchema(schema any) any {
	mapSchema, ok := schema.(map[string]any)
	if !ok || mapSchema == nil {
		return map[string]any{"type": "object", "properties": map[string]any{}, "required": []any{}}
	}
	cloned := make(map[string]any, len(mapSchema)+1)
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

func injectSystemPrompt(systemPrompt string, history *[]normalizedMessage, current *normalizedMessage) {
	trimmed := sanitizePromptText(systemPrompt)
	if trimmed == "" {
		return
	}
	if history != nil && len(*history) > 0 && (*history)[0].Role == "user" {
		(*history)[0].Content = joinNonEmpty(trimmed, (*history)[0].Content)
		return
	}
	current.Content = joinNonEmpty(trimmed, current.Content)
}

func resolveConversationID(metadata map[string]any) string {
	if metadata != nil {
		if value, ok := metadata["conversationId"].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return uuid.NewString()
}

func resolveContinuationID(metadata map[string]any) string {
	if metadata != nil {
		if value, ok := metadata["continuationId"].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func resolveProfileARN(metadata map[string]any, account config.Account) string {
	if strings.TrimSpace(account.ClientID) != "" && strings.TrimSpace(account.ClientSecret) != "" {
		return ""
	}
	if metadata != nil {
		if value, ok := metadata["profileArn"].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return strings.TrimSpace(account.AccountID)
}

func buildInferenceConfig(req provider.ChatRequest) *inferenceConfig {
	config := &inferenceConfig{}
	hasConfig := false

	if req.MaxTokens != nil {
		value := *req.MaxTokens
		if value <= 0 {
			value = kiroDefaultMaxTokens
		}
		config.MaxTokens = &value
		hasConfig = true
	}
	if req.Temperature != nil {
		config.Temperature = req.Temperature
		hasConfig = true
	}
	if req.TopP != nil {
		config.TopP = req.TopP
		hasConfig = true
	}
	if !hasConfig {
		return nil
	}
	return config
}

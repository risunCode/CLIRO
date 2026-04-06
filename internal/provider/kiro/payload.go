package kiro

import (
	"fmt"
	"strings"

	"cliro/internal/config"
	contract "cliro/internal/contract"
	provider "cliro/internal/provider"
	thinkingctrl "cliro/internal/provider/thinking"

	"github.com/google/uuid"
)

var errMessagesEmpty = fmt.Errorf("messages are empty")

const (
	forcedThinkingModeTag         = "<thinking_mode>enabled</thinking_mode>"
	adaptiveThinkingModeTag     = "<thinking_mode>adaptive</thinking_mode>"
	thinkingEffortLowTag      = "<thinking_effort>low</thinking_effort>"
	thinkingEffortMediumTag       = "<thinking_effort>medium</thinking_effort>"
	thinkingEffortHighTag         = "<thinking_effort>high</thinking_effort>"
)

type requestPayload struct {
	ConversationState conversationState `json:"conversationState"`
	ProfileARN        string            `json:"profileArn,omitempty"`
	InferenceConfig   *inferenceConfig  `json:"inferenceConfig,omitempty"`
	ThinkingBudget    *int              `json:"thinking_budget,omitempty"`
}

type conversationState struct {
	AgentContinuationID string           `json:"agentContinuationId,omitempty"`
	AgentTaskType       string           `json:"agentTaskType,omitempty"`
	ChatTriggerType     string           `json:"chatTriggerType"`
	ConversationID      string           `json:"conversationId"`
	CurrentMessage      currentMessage   `json:"currentMessage"`
	History             []historyMessage `json:"history,omitempty"`
}

type inferenceConfig struct {
	MaxTokens   *int     `json:"maxTokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"topP,omitempty"`
}

type currentMessage struct {
	UserInputMessage userInputMessage `json:"userInputMessage"`
}

type historyMessage struct {
	UserInputMessage         *userInputMessage         `json:"userInputMessage,omitempty"`
	AssistantResponseMessage *assistantResponseMessage `json:"assistantResponseMessage,omitempty"`
}

type userInputMessage struct {
	Content                 string                   `json:"content"`
	ModelID                 string                   `json:"modelId"`
	Origin                  string                   `json:"origin"`
	UserInputMessageContext *userInputMessageContext `json:"userInputMessageContext,omitempty"`
}

type assistantResponseMessage struct {
	Content  string           `json:"content"`
	ToolUses []toolUsePayload `json:"toolUses,omitempty"`
}

type userInputMessageContext struct {
	Tools       []toolSpecification `json:"tools,omitempty"`
	ToolResults []toolResult        `json:"toolResults,omitempty"`
}

type toolSpecification struct {
	ToolSpecification toolSpecificationDetails `json:"toolSpecification"`
}

type toolSpecificationDetails struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema toolInputSchema `json:"inputSchema"`
}

type toolInputSchema struct {
	JSON any `json:"json"`
}

type toolResult struct {
	ToolUseID string              `json:"toolUseId"`
	Content   []toolResultContent `json:"content"`
	Status    string              `json:"status"`
}

type toolResultContent struct {
	Text string `json:"text,omitempty"`
}

type toolUsePayload struct {
	ToolUseID string         `json:"toolUseId"`
	Name      string         `json:"name"`
	Input     map[string]any `json:"input"`
}

func buildRequestPayload(req provider.ChatRequest, account config.Account, thinkingSettings config.ThinkingSettings) (requestPayload, error) {
	payload, _, err := buildRequestPayloadWithToolNames(req, account, thinkingSettings)
	return payload, err
}

func buildRequestPayloadWithToolNames(req provider.ChatRequest, account config.Account, thinkingSettings config.ThinkingSettings) (requestPayload, provider.ToolNameMapping, error) {
	req.Tools = provider.NormalizeToolsForProvider("kiro", req.Tools)
	mapping := provider.BuildToolNameMapping(req.Tools, req.Messages, provider.DefaultToolNameLimit)
	req = provider.RemapChatRequestToolNames(req, mapping)
	messages, systemPrompt, err := normalizeRequest(req)
	if err != nil {
		return requestPayload{}, mapping, err
	}
	if len(messages) == 0 {
		return requestPayload{}, mapping, errMessagesEmpty
	}

	historyMessages := append([]normalizedMessage(nil), messages...)
	current := historyMessages[len(historyMessages)-1]
	historyMessages = historyMessages[:len(historyMessages)-1]

	if current.Role == "assistant" {
		historyMessages = append(historyMessages, current)
		current = normalizedMessage{Role: "user", Content: "Continue"}
	}

	injectSystemPrompt(systemPrompt, &historyMessages, &current)
	injectForcedThinkingFallback(req, thinkingSettings, &current)
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
		payload.ConversationState.History = sanitizeHistoryToolResults(history)
	}
	sanitizeCurrentToolResults(&payload.ConversationState.CurrentMessage, payload.ConversationState.History)
	if continuationID := resolveContinuationID(req.Metadata); continuationID != "" {
		payload.ConversationState.AgentContinuationID = continuationID
	}
	if profileARN := resolveProfileARN(req.Metadata, account); profileARN != "" {
		payload.ProfileARN = profileARN
	}
	if config := buildInferenceConfig(req); config != nil {
		payload.InferenceConfig = config
	}
	if thinkingBudget := extractThinkingBudget(req); thinkingBudget > 0 {
		payload.ThinkingBudget = &thinkingBudget
	}

	return payload, mapping, nil
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
	for _, tool := range provider.NormalizeToolsForProvider("kiro", tools) {
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
	return provider.NormalizeToolSchema(schema)
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

func injectForcedThinkingFallback(req provider.ChatRequest, settings config.ThinkingSettings, current *normalizedMessage) {
	if current == nil || !shouldInjectForcedThinking(req, settings) {
		return
	}

	// Check if adaptive mode is requested
	if req.Thinking.Requested && len(req.Thinking.RawParams) > 0 {
		if thinkingType, ok := req.Thinking.RawParams["type"].(string); ok {
			if strings.ToLower(strings.TrimSpace(thinkingType)) == "adaptive" {
				// Use adaptive mode with effort level
				effort := "high"
				if effortVal, ok := req.Thinking.RawParams["effort"].(string); ok {
					effort = effortVal
				}
				current.Content = joinNonEmpty(buildAdaptiveThinkingPrompt(effort), current.Content)
				return
			}
		}
	}

	// Default to enabled mode with budget tokens
	current.Content = joinNonEmpty(buildForcedThinkingPrompt(settings.MaxForcedThinkingTokens), current.Content)
}

func shouldInjectForcedThinking(req provider.ChatRequest, settings config.ThinkingSettings) bool {
	if req.RouteFamily != string(contract.EndpointAnthropicMessages) {
		return false
	}
	effective := contract.ThinkingConfig{
		Requested: req.Thinking.Requested,
		Mode:      thinkingModeFromSettings(settings.Mode),
	}
	return thinkingctrl.ForceEligible(effective, settings.ForceForAnthropic)
}

func thinkingModeFromSettings(mode config.ThinkingMode) contract.ThinkingMode {
	switch mode {
	case config.ThinkingModeOff:
		return contract.ThinkingModeOff
	case config.ThinkingModeForce:
		return contract.ThinkingModeForce
	default:
		return contract.ThinkingModeAuto
	}
}

func buildForcedThinkingPrompt(maxTokens int) string {
	if maxTokens <= 0 {
		maxTokens = 4000
	}
	return forcedThinkingModeTag + "\n<max_thinking_length>" + fmt.Sprintf("%d", maxTokens) + "</max_thinking_length>"
}

func buildAdaptiveThinkingPrompt(effort string) string {
	effortTag := thinkingEffortHighTag
	switch strings.ToLower(strings.TrimSpace(effort)) {
	case "low":
		effortTag = thinkingEffortLowTag
	case "medium":
		effortTag = thinkingEffortMediumTag
	case "high":
		effortTag = thinkingEffortHighTag
	default:
		effortTag = thinkingEffortHighTag
	}
	return adaptiveThinkingModeTag + "\n" + effortTag
}

func extractThinkingBudget(req provider.ChatRequest) int {
	if !req.Thinking.Requested || len(req.Thinking.RawParams) == 0 {
		return 0
	}
	// Try Anthropic format first (thinking.budget_tokens)
	if budgetTokens, ok := req.Thinking.RawParams["budget_tokens"].(float64); ok {
		return int(budgetTokens)
	}
	if budgetTokens, ok := req.Thinking.RawParams["budget_tokens"].(int); ok {
		return budgetTokens
	}
	// Try OpenAI format (reasoning.effort) - not supported by Kiro, return 0
	// Kiro API doesn't support reasoning effort levels
	return 0
}

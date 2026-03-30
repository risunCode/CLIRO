package kiro

type requestPayload struct {
	ConversationState conversationState `json:"conversationState"`
	ProfileARN        string            `json:"profileArn,omitempty"`
	InferenceConfig   *inferenceConfig  `json:"inferenceConfig,omitempty"`
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

type normalizedMessage struct {
	Role        string
	Content     string
	ToolUses    []toolUsePayload
	ToolResults []toolResult
}

type StreamEvent struct {
	Text     string
	Thinking string
	Usage    UsageSnapshot
}

type UsageSnapshot struct {
	PromptTokens   int
	CompletionTokens int
	TotalTokens      int
}

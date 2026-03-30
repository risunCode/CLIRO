package decode

import (
	"testing"

	"cliro-go/internal/adapter/ir"
	"cliro-go/internal/protocol/anthropic"
)

func TestAnthropicMessagesToIR_MergesMessagesAndPreservesThinkingBlocks(t *testing.T) {
	request, err := AnthropicMessagesToIR(anthropic.MessagesRequest{
		Model:     "claude-sonnet-4.5",
		MaxTokens: 256,
		Messages: []anthropic.Message{
			{Role: "assistant", Content: []any{
				map[string]any{"type": "thinking", "thinking": "plan", "signature": "sig_plan", "cache_control": map[string]any{"type": "ephemeral"}},
				map[string]any{"type": "tool_use", "id": "call_1", "name": "Grep", "input": map[string]any{"query": "needle"}},
			}},
			{Role: "assistant", Content: []any{map[string]any{"type": "text", "text": "searching"}}},
			{Role: "user", Content: []any{map[string]any{"type": "text", "text": "continue"}}},
			{Role: "user", Content: []any{map[string]any{"type": "redacted_thinking", "data": "secret", "cache_control": map[string]any{"type": "ephemeral"}}}},
		},
	})
	if err != nil {
		t.Fatalf("AnthropicMessagesToIR: %v", err)
	}
	if len(request.Messages) != 2 {
		t.Fatalf("message count = %d, want 2 messages=%#v", len(request.Messages), request.Messages)
	}

	assistant := request.Messages[0]
	if assistant.Role != ir.RoleAssistant {
		t.Fatalf("assistant role = %q", assistant.Role)
	}
	if assistant.Content != "searching" {
		t.Fatalf("assistant content = %#v", assistant.Content)
	}
	if len(assistant.ToolCalls) != 1 || assistant.ToolCalls[0].Name != "Grep" {
		t.Fatalf("assistant tool calls = %#v", assistant.ToolCalls)
	}
	if len(assistant.ThinkingBlocks) != 1 {
		t.Fatalf("assistant thinking blocks = %#v", assistant.ThinkingBlocks)
	}
	if assistant.ThinkingBlocks[0] != (ir.ThinkingBlock{Thinking: "plan", Signature: "sig_plan"}) {
		t.Fatalf("unexpected thinking block: %#v", assistant.ThinkingBlocks[0])
	}

	user := request.Messages[1]
	if user.Role != ir.RoleUser {
		t.Fatalf("user role = %q", user.Role)
	}
	if user.Content != "continue" {
		t.Fatalf("user content = %#v", user.Content)
	}
}

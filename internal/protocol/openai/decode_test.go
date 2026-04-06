package openai

import (
	"reflect"
	"testing"

	contract "cliro/internal/contract"
)

func TestResponsesToIR_PreservesConversationMetadata(t *testing.T) {
	request, err := ResponsesToIR(ResponsesRequest{
		Model: "gpt-5.4",
		Input: []any{map[string]any{
			"type":              "message",
			"role":              "user",
			"content":           []any{map[string]any{"type": "input_text", "text": "hello"}},
			"additional_kwargs": map[string]any{"conversationId": "conv-1", "continuationId": "cont-1"},
		}},
	})
	if err != nil {
		t.Fatalf("ResponsesToIR: %v", err)
	}
	if request.Metadata["conversationId"] != "conv-1" {
		t.Fatalf("conversationId = %#v", request.Metadata["conversationId"])
	}
	if request.Metadata["continuationId"] != "cont-1" {
		t.Fatalf("continuationId = %#v", request.Metadata["continuationId"])
	}
}

func TestResponsesToIR_ParsesAssistantOutputAndToolResult(t *testing.T) {
	request, err := ResponsesToIR(ResponsesRequest{
		Model: "gpt-5.4",
		Input: []any{
			map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": "hello"}}},
			map[string]any{"type": "function_call", "call_id": "call_1", "name": "Read", "arguments": `{"path":"README.md"}`},
			map[string]any{"type": "function_call_output", "call_id": "call_1", "output": "done"},
		},
	})
	if err != nil {
		t.Fatalf("ResponsesToIR: %v", err)
	}
	if len(request.Messages) != 3 {
		t.Fatalf("message count = %d", len(request.Messages))
	}
	if request.Messages[0].Role != contract.RoleAssistant || request.Messages[0].Content != "hello" {
		t.Fatalf("unexpected assistant message: %+v", request.Messages[0])
	}
	if request.Messages[1].Role != contract.RoleAssistant || len(request.Messages[1].ToolCalls) != 1 || request.Messages[1].ToolCalls[0].ID != "call_1" {
		t.Fatalf("unexpected tool call message: %+v", request.Messages[1])
	}
	if request.Messages[2].Role != contract.RoleTool || request.Messages[2].ToolCallID != "call_1" || request.Messages[2].Content != "done" {
		t.Fatalf("unexpected tool result message: %+v", request.Messages[2])
	}
}

func TestChatToIR_PreservesBuiltinToolType(t *testing.T) {
	request, err := ChatToIR(ChatRequest{
		Model:      "gpt-5.4",
		Messages:   []Message{{Role: "user", Content: "search this"}},
		Tools:      []Tool{{Type: "web_search"}},
		ToolChoice: map[string]any{"type": "web_search"},
	})
	if err != nil {
		t.Fatalf("ChatToIR: %v", err)
	}
	if len(request.Tools) != 1 {
		t.Fatalf("tool count = %d", len(request.Tools))
	}
	if request.Tools[0].Type != "web_search" {
		t.Fatalf("tool type = %q", request.Tools[0].Type)
	}
	if request.Tools[0].Name != "" {
		t.Fatalf("tool name = %q", request.Tools[0].Name)
	}
	if !reflect.DeepEqual(request.ToolChoice, map[string]any{"type": "web_search"}) {
		t.Fatalf("tool choice = %#v", request.ToolChoice)
	}
}

func TestChatToIR_ConvertsOrphanToolResultsIntoUserMessages(t *testing.T) {
	request, err := ChatToIR(ChatRequest{
		Model: "gpt-5.4",
		Messages: []Message{
			{Role: "user", Content: "hello"},
			{Role: "tool", ToolCallID: "missing_call", Content: "orphan result"},
		},
	})
	if err != nil {
		t.Fatalf("ChatToIR: %v", err)
	}
	if len(request.Messages) != 2 {
		t.Fatalf("message count = %d", len(request.Messages))
	}
	if request.Messages[1].Role != contract.RoleUser {
		t.Fatalf("orphan tool role = %q", request.Messages[1].Role)
	}
	if request.Messages[1].ToolCallID != "" {
		t.Fatalf("unexpected tool_call_id = %q", request.Messages[1].ToolCallID)
	}
	if request.Messages[1].Content != "orphan result" {
		t.Fatalf("content = %#v", request.Messages[1].Content)
	}
}

func TestResponsesToIR_ConvertsOrphanFunctionCallOutputIntoUserMessage(t *testing.T) {
	request, err := ResponsesToIR(ResponsesRequest{
		Model: "gpt-5.4",
		Input: []any{
			map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": "hello"}}},
			map[string]any{"type": "function_call_output", "call_id": "missing_call", "output": "done"},
		},
	})
	if err != nil {
		t.Fatalf("ResponsesToIR: %v", err)
	}
	if len(request.Messages) != 2 {
		t.Fatalf("message count = %d", len(request.Messages))
	}
	if request.Messages[1].Role != contract.RoleUser {
		t.Fatalf("orphan tool result role = %q", request.Messages[1].Role)
	}
	if request.Messages[1].Content != "done" {
		t.Fatalf("orphan tool result content = %#v", request.Messages[1].Content)
	}
}

package ir

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestIRStructuresRoundTripJSON(t *testing.T) {
	message := Message{
		Role:    RoleAssistant,
		Content: "answer",
		ToolCalls: []ToolCall{{
			ID:        "toolu_1",
			Name:      "Read",
			Arguments: `{"path":"main.go"}`,
		}},
		ThinkingBlocks: []ThinkingBlock{{Thinking: "plan", Signature: "sig_plan"}},
	}
	encodedMessage, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	var decodedMessage Message
	if err := json.Unmarshal(encodedMessage, &decodedMessage); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if !reflect.DeepEqual(decodedMessage, message) {
		t.Fatalf("message mismatch: %#v", decodedMessage)
	}

	response := Response{
		ID:                "msg_test",
		Model:             "claude-sonnet-4.5",
		Text:              "done",
		Thinking:          "reasoning",
		ThinkingSignature: "sig_reasoning",
		ToolCalls:         message.ToolCalls,
		Usage:             Usage{PromptTokens: 3, CompletionTokens: 5, TotalTokens: 8, InputTokens: 3, OutputTokens: 5},
		StopReason:        "tool_calls",
	}
	encodedResponse, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var decodedResponse Response
	if err := json.Unmarshal(encodedResponse, &decodedResponse); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !reflect.DeepEqual(decodedResponse, response) {
		t.Fatalf("response mismatch: %#v", decodedResponse)
	}
}

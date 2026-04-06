package openai

import (
	"testing"

	contract "cliro/internal/contract"
)

func TestStreamState_StopsAfterCompletion(t *testing.T) {
	var state StreamState
	if !state.Apply(contract.Event{ThinkDelta: "plan"}) {
		t.Fatalf("expected first event to apply")
	}
	if !state.Apply(contract.Event{Done: true, Type: "stop"}) {
		t.Fatalf("expected completion event to apply")
	}
	if state.Apply(contract.Event{TextDelta: "late"}) {
		t.Fatalf("expected post-completion event to be ignored")
	}
	if !state.HasThinking || !state.Completed {
		t.Fatalf("state = %#v", state)
	}
}

func TestResponsesStreamState_EmitsReasoningTextAndToolLifecycle(t *testing.T) {
	var names []string
	state := NewResponsesStreamState("resp_123", "gpt-5.4", 123, func(name string, payload map[string]any) {
		names = append(names, name)
		if name == "response.function_call_arguments.delta" {
			if payload["delta"] != `{"path":"README.md"}` {
				t.Fatalf("delta = %#v", payload["delta"])
			}
		}
	})

	state.Start()
	state.EmitReasoningDelta("plan first")
	state.EmitTextDelta("hello")
	state.CloseMessageItem("hello", "plan first")
	state.EmitFunctionCall(contract.ToolCall{ID: "call_1", Name: "Read", Arguments: `{"path":"README.md"}`})
	state.Complete(ResponsesResponse{ID: "resp_123", Model: "gpt-5.4", Status: "completed"})

	want := []string{
		"response.created",
		"response.in_progress",
		"response.reasoning_summary_text.delta",
		"response.output_item.added",
		"response.content_part.added",
		"response.output_text.delta",
		"response.output_text.done",
		"response.content_part.done",
		"response.output_item.done",
		"response.output_item.added",
		"response.function_call_arguments.delta",
		"response.function_call_arguments.done",
		"response.output_item.done",
		"response.completed",
	}
	if len(names) != len(want) {
		t.Fatalf("event count = %d, want %d events=%#v", len(names), len(want), names)
	}
	for index := range want {
		if names[index] != want[index] {
			t.Fatalf("event %d = %q, want %q full=%#v", index, names[index], want[index], names)
		}
	}
}

package gateway

import (
	"net/http/httptest"
	"strings"
	"testing"

	contract "cliro/internal/contract"
)

func TestStreamOpenAICompletions_EmitsReasoningContentBeforeText(t *testing.T) {
	server := &Server{}
	recorder := httptest.NewRecorder()

	server.streamOpenAICompletions(recorder, "gpt-5.4", contract.Response{
		ID:       "cmpl_123",
		Model:    "gpt-5.4",
		Thinking: "plan first",
		Text:     "final answer",
	})

	body := recorder.Body.String()
	if !strings.Contains(body, `"reasoning_content":"plan first"`) {
		t.Fatalf("expected reasoning_content in stream: %s", body)
	}
	if !strings.Contains(body, `"text":"final answer"`) {
		t.Fatalf("expected text in stream: %s", body)
	}
	if strings.Index(body, `"reasoning_content":"plan first"`) > strings.Index(body, `"text":"final answer"`) {
		t.Fatalf("expected reasoning_content before text: %s", body)
	}
}

func TestStreamOpenAIResponses_EmitsReasoningContentBeforeText(t *testing.T) {
	server := &Server{}
	recorder := httptest.NewRecorder()

	server.streamOpenAIResponses(recorder, "gpt-5.4", contract.Response{
		ID:       "resp_123",
		Model:    "gpt-5.4",
		Thinking: "plan first",
		Text:     "final answer",
	})

	body := recorder.Body.String()
	if !strings.Contains(body, `event: response.reasoning_summary_text.delta`) {
		t.Fatalf("expected reasoning delta event in stream: %s", body)
	}
	if !strings.Contains(body, `"delta":"final answer"`) {
		t.Fatalf("expected text delta in stream: %s", body)
	}
	if strings.Index(body, `event: response.reasoning_summary_text.delta`) > strings.Index(body, `"delta":"final answer"`) {
		t.Fatalf("expected reasoning event before text delta: %s", body)
	}
}

func TestStreamOpenAIResponses_EmitsFunctionCallLifecycle(t *testing.T) {
	server := &Server{}
	recorder := httptest.NewRecorder()

	server.streamOpenAIResponses(recorder, "gpt-5.4", contract.Response{
		ID:    "resp_456",
		Model: "gpt-5.4",
		ToolCalls: []contract.ToolCall{{
			ID:        "call_1",
			Name:      "Read",
			Arguments: `{"path":"README.md"}`,
		}},
	})

	body := recorder.Body.String()
	if !strings.Contains(body, `event: response.function_call_arguments.delta`) {
		t.Fatalf("expected function call argument delta: %s", body)
	}
	if !strings.Contains(body, `event: response.function_call_arguments.done`) {
		t.Fatalf("expected function call argument done: %s", body)
	}
	if !strings.Contains(body, `"type":"function_call"`) {
		t.Fatalf("expected function_call output item: %s", body)
	}
	if !strings.Contains(body, `event: response.output_item.done`) {
		t.Fatalf("expected function_call output item done: %s", body)
	}
}

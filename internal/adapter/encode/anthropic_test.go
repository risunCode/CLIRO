package encode

import (
	"testing"

	"cliro-go/internal/adapter/ir"
)

func TestIRToAnthropicMessages_ThinkingFirstStableSignatureAndToolRemap(t *testing.T) {
	response := ir.Response{
		Model:      "claude-sonnet-4.5",
		Thinking:   "plan",
		Text:       "done",
		ToolCalls:  []ir.ToolCall{{ID: "toolu_1", Name: "Glob", Arguments: `{"query":"*.go","paths":["internal"]}`}},
		StopReason: "tool_calls",
		Usage:      ir.Usage{InputTokens: 4, OutputTokens: 6},
	}

	first := IRToAnthropicMessages(response)
	second := IRToAnthropicMessages(response)

	content, ok := first.Content.([]map[string]any)
	if !ok {
		t.Fatalf("content type = %T", first.Content)
	}
	if len(content) != 3 {
		t.Fatalf("content count = %d", len(content))
	}
	if content[0]["type"] != "thinking" || content[1]["type"] != "text" || content[2]["type"] != "tool_use" {
		t.Fatalf("unexpected block order: %#v", content)
	}
	if content[0]["signature"] != StableThinkingSignature("plan") {
		t.Fatalf("thinking signature = %#v", content[0]["signature"])
	}
	secondContent := second.Content.([]map[string]any)
	if content[0]["signature"] != secondContent[0]["signature"] {
		t.Fatalf("signature not stable: %#v vs %#v", content[0]["signature"], secondContent[0]["signature"])
	}

	input, ok := content[2]["input"].(map[string]any)
	if !ok {
		t.Fatalf("tool input = %#v", content[2]["input"])
	}
	if input["pattern"] != "*.go" || input["path"] != "internal" {
		t.Fatalf("remapped input = %#v", input)
	}
	if _, exists := input["query"]; exists {
		t.Fatalf("unexpected query key in %#v", input)
	}
	if first.StopReason != "tool_use" {
		t.Fatalf("stop reason = %q", first.StopReason)
	}
}

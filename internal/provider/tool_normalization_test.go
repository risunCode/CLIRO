package provider

import "testing"

func TestBuildToolNameMapping_RoundTripsLongNames(t *testing.T) {
	longName := "ThisIsAnExtremelyLongToolNameThatShouldBeShortenedToStayWithinProviderLimits_ReadVeryLongName"
	mapping := BuildToolNameMapping([]Tool{{Type: "function", Function: ToolFunction{Name: longName}}}, nil, 32)
	short := mapping.Remap(longName)
	if short == longName {
		t.Fatalf("expected remapped name to be shortened")
	}
	if len(short) > 32 {
		t.Fatalf("short name length = %d, want <= 32", len(short))
	}
	if restored := mapping.Restore(short); restored != longName {
		t.Fatalf("restored = %q, want %q", restored, longName)
	}
}

func TestNormalizeToolSchema_RepairsMissingFields(t *testing.T) {
	normalized := NormalizeToolSchema(map[string]any{"properties": map[string]any{"path": map[string]any{"type": "string"}}})
	object, ok := normalized.(map[string]any)
	if !ok {
		t.Fatalf("schema type = %T", normalized)
	}
	if object["type"] != "object" {
		t.Fatalf("type = %#v", object["type"])
	}
	if _, ok := object["required"]; !ok {
		t.Fatalf("required missing: %#v", object)
	}
	if _, ok := object["properties"]; !ok {
		t.Fatalf("properties missing: %#v", object)
	}
}

func TestRemapChatRequestToolNames_RenamesDefinitionsAndCalls(t *testing.T) {
	longName := "ThisIsAnExtremelyLongToolNameThatShouldBeShortenedToStayWithinProviderLimits_ReadVeryLongName"
	mapping := BuildToolNameMapping([]Tool{{Type: "function", Function: ToolFunction{Name: longName, Parameters: map[string]any{}}}}, []Message{{ToolCalls: []ToolCall{{Function: ToolCallTarget{Name: longName, Arguments: `{}`}}}}}, 32)
	req := RemapChatRequestToolNames(ChatRequest{
		Tools:    []Tool{{Type: "function", Function: ToolFunction{Name: longName, Parameters: map[string]any{}}}},
		Messages: []Message{{ToolCalls: []ToolCall{{Function: ToolCallTarget{Name: longName, Arguments: `{}`}}}}},
	}, mapping)
	if req.Tools[0].Function.Name == longName {
		t.Fatalf("expected tool definition name remap")
	}
	if req.Messages[0].ToolCalls[0].Function.Name != req.Tools[0].Function.Name {
		t.Fatalf("tool call name = %q, tool definition name = %q", req.Messages[0].ToolCalls[0].Function.Name, req.Tools[0].Function.Name)
	}
}

func TestNormalizeToolsForProvider_FiltersUnsupportedKiroTools(t *testing.T) {
	tools := []Tool{
		{Type: "function", Function: ToolFunction{Name: "Read", Parameters: map[string]any{}}},
		{Type: "function", Function: ToolFunction{Name: "web_search", Parameters: map[string]any{}}},
		{Type: "computer", Function: ToolFunction{Name: "computer", Parameters: map[string]any{}}},
	}

	normalized := NormalizeToolsForProvider("kiro", tools)
	if len(normalized) != 1 {
		t.Fatalf("normalized tools = %#v", normalized)
	}
	if normalized[0].Function.Name != "Read" {
		t.Fatalf("kept tool = %#v", normalized[0])
	}
}

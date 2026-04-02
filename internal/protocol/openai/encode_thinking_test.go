package openai

import (
	"encoding/json"
	"testing"

	contract "cliro-go/internal/contract"
)

func TestIRToChat_PreservesThinkingBlocksInAdditionalKwargs(t *testing.T) {
	resp := contract.Response{
		ID:                "test-id",
		Model:             "claude-3-5-sonnet",
		Text:              "Hello",
		Thinking:          "Let me think about this...",
		ThinkingSignature: "sig_abc123",
		Usage: contract.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	result := IRToChat(resp)

	if result.Choices[0].Message.ReasoningContent != "Let me think about this..." {
		t.Errorf("Expected reasoning_content to be set")
	}

	if result.Choices[0].Message.AdditionalKwargs == nil {
		t.Fatalf("Expected additional_kwargs to be set")
	}

	thinkingBlocks, ok := result.Choices[0].Message.AdditionalKwargs["thinking_blocks"]
	if !ok {
		t.Fatalf("Expected thinking_blocks in additional_kwargs")
	}

	blocks, ok := thinkingBlocks.([]map[string]any)
	if !ok || len(blocks) != 1 {
		t.Fatalf("Expected thinking_blocks to be array with 1 element")
	}

	if blocks[0]["thinking"] != "Let me think about this..." {
		t.Errorf("Expected thinking content to match")
	}

	if blocks[0]["signature"] != "sig_abc123" {
		t.Errorf("Expected signature to match")
	}
}

func TestIRToChat_RoundTripWithThinkingBlocks(t *testing.T) {
	// Simulate Claude response with thinking
	claudeResp := contract.Response{
		ID:                "msg_123",
		Model:             "claude-3-5-sonnet",
		Text:              "The answer is 42",
		Thinking:          "First, I need to calculate...",
		ThinkingSignature: "sig_xyz",
	}

	// Convert to OpenAI format
	openaiResp := IRToChat(claudeResp)

	// Serialize to JSON (simulating API response)
	jsonData, err := json.Marshal(openaiResp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Deserialize back
	var parsed ChatResponse
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify thinking blocks are preserved
	if parsed.Choices[0].Message.AdditionalKwargs == nil {
		t.Fatalf("additional_kwargs lost after round-trip")
	}

	thinkingBlocks := parsed.Choices[0].Message.AdditionalKwargs["thinking_blocks"]
	if thinkingBlocks == nil {
		t.Fatalf("thinking_blocks lost after round-trip")
	}

	// Now convert back to IR (simulating client sending this back)
	chatReq := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{
				Role:             parsed.Choices[0].Message.Role,
				Content:          parsed.Choices[0].Message.Content,
				AdditionalKwargs: parsed.Choices[0].Message.AdditionalKwargs,
			},
		},
	}

	ir, err := ChatToIR(chatReq)
	if err != nil {
		t.Fatalf("Failed to convert back to IR: %v", err)
	}

	// Verify thinking blocks are restored
	if len(ir.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(ir.Messages))
	}

	if len(ir.Messages[0].ThinkingBlocks) != 1 {
		t.Fatalf("Expected 1 thinking block, got %d", len(ir.Messages[0].ThinkingBlocks))
	}

	block := ir.Messages[0].ThinkingBlocks[0]
	if block.Thinking != "First, I need to calculate..." {
		t.Errorf("Thinking content not preserved: got %q", block.Thinking)
	}

	if block.Signature != "sig_xyz" {
		t.Errorf("Signature not preserved: got %q", block.Signature)
	}
}

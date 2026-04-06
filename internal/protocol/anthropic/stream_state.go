package anthropic

type MessageStreamState struct {
	emit           StreamEventEmitter
	messageID      string
	model          string
	messageStarted bool
	textOpen       bool
	textIndex      int
	nextIndex      int
	completed      bool
}

func NewMessageStreamState(messageID string, model string, emit StreamEventEmitter) *MessageStreamState {
	if emit == nil {
		emit = func(string, map[string]any) {}
	}
	return &MessageStreamState{emit: emit, messageID: messageID, model: model}
}

func (s *MessageStreamState) Started() bool {
	return s != nil && s.messageStarted
}

func (s *MessageStreamState) StartMessage(inputTokens int) {
	if s == nil || s.messageStarted {
		return
	}
	s.emit("message_start", map[string]any{
		"type": "message_start",
		"message": map[string]any{
			"id":            s.messageID,
			"type":          "message",
			"role":          "assistant",
			"model":         s.model,
			"content":       []any{},
			"stop_reason":   nil,
			"stop_sequence": nil,
			"usage": map[string]int{
				"input_tokens":  inputTokens,
				"output_tokens": 0,
			},
		},
	})
	s.messageStarted = true
	if s.nextIndex < 0 {
		s.nextIndex = 0
	}
}

func (s *MessageStreamState) EnsureTextBlock() int {
	if s == nil {
		return 0
	}
	if s.textOpen {
		return s.textIndex
	}
	s.textIndex = s.nextIndex
	s.nextIndex++
	s.textOpen = true
	s.emit("content_block_start", map[string]any{
		"type":  "content_block_start",
		"index": s.textIndex,
		"content_block": map[string]any{
			"type": "text",
			"text": "",
		},
	})
	return s.textIndex
}

func (s *MessageStreamState) EmitTextDelta(delta string) {
	if s == nil || delta == "" {
		return
	}
	index := s.EnsureTextBlock()
	s.emit("content_block_delta", map[string]any{
		"type":  "content_block_delta",
		"index": index,
		"delta": map[string]any{
			"type": "text_delta",
			"text": delta,
		},
	})
	if s.nextIndex <= index {
		s.nextIndex = index + 1
	}
}

func (s *MessageStreamState) CloseTextBlock() {
	if s == nil || !s.textOpen {
		return
	}
	s.emit("content_block_stop", map[string]any{"type": "content_block_stop", "index": s.textIndex})
	s.textOpen = false
}

func (s *MessageStreamState) NextIndex() int {
	if s == nil {
		return 0
	}
	if s.textOpen && s.nextIndex <= s.textIndex {
		return s.textIndex + 1
	}
	return s.nextIndex
}

func (s *MessageStreamState) MarkIndex(index int) {
	if s == nil {
		return
	}
	if index > s.nextIndex {
		s.nextIndex = index
	}
	if s.textOpen && s.nextIndex <= s.textIndex {
		s.nextIndex = s.textIndex + 1
	}
}

func (s *MessageStreamState) Complete(stopReason string, outputTokens int) {
	if s == nil || s.completed {
		return
	}
	s.completed = true
	s.emit("message_delta", map[string]any{
		"type": "message_delta",
		"delta": map[string]any{
			"stop_reason":   stopReason,
			"stop_sequence": nil,
		},
		"usage": map[string]int{
			"output_tokens": outputTokens,
		},
	})
	s.emit("message_stop", map[string]any{"type": "message_stop"})
}

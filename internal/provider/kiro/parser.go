package kiro

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"strings"

	provider "cliro-go/internal/provider"
)

type StreamParser struct {
	reader           io.Reader
	assistantContent string
	thinkingContent  string
	currentTool      *toolAccumulator
	toolUses         []provider.ToolUse
}

type eventFrame struct {
	EventType   string
	MessageType string
	Payload     []byte
}

type toolAccumulator struct {
	ID         string
	Name       string
	InputParts strings.Builder
	HasInput   bool
}

func NewStreamParser(reader io.Reader) *StreamParser {
	return &StreamParser{reader: reader}
}

func (p *StreamParser) Next() (StreamEvent, error) {
	for {
		frame, err := readEventFrame(p.reader)
		if err != nil {
			if err == io.EOF {
				p.finalizeCurrentTool()
			}
			return StreamEvent{}, err
		}
		event, err := p.parseFrame(frame)
		if err != nil {
			return StreamEvent{}, err
		}
		if event.Text != "" || event.Thinking != "" || event.Usage.TotalTokens > 0 || event.Usage.PromptTokens > 0 || event.Usage.CompletionTokens > 0 {
			return event, nil
		}
	}
}

func (p *StreamParser) ToolUses() []provider.ToolUse {
	p.finalizeCurrentTool()
	return append([]provider.ToolUse(nil), p.toolUses...)
}

func (p *StreamParser) parseFrame(frame eventFrame) (StreamEvent, error) {
	if strings.EqualFold(frame.MessageType, "error") || strings.EqualFold(frame.MessageType, "exception") {
		return StreamEvent{}, fmt.Errorf(errorMessageFromPayload(frame.Payload))
	}
	if len(frame.Payload) == 0 {
		return StreamEvent{}, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(frame.Payload, &payload); err != nil {
		return StreamEvent{}, nil
	}

	switch resolveEventType(frame.EventType, payload) {
	case "assistantResponseEvent":
		return StreamEvent{Text: deltaFromCumulative(&p.assistantContent, resolveTextField(payload, "content", "text")), Usage: extractUsage(payload)}, nil
	case "reasoningContentEvent":
		return StreamEvent{Thinking: deltaFromCumulative(&p.thinkingContent, resolveTextField(payload, "text", "content")), Usage: extractUsage(payload)}, nil
	case "toolUseEvent":
		p.handleToolUseEvent(payload)
		return StreamEvent{Usage: extractUsage(payload)}, nil
	default:
		return StreamEvent{Usage: extractUsage(payload)}, nil
	}
}

func (p *StreamParser) handleToolUseEvent(payload map[string]any) {
	toolID := strings.TrimSpace(resolveTextField(payload, "toolUseId", "id"))
	toolName := strings.TrimSpace(resolveTextField(payload, "name"))
	stop, _ := payload["stop"].(bool)

	if toolID != "" || toolName != "" {
		if p.currentTool != nil && p.currentTool.ID != "" && toolID != "" && p.currentTool.ID != toolID {
			p.finalizeCurrentTool()
		}
		if p.currentTool == nil {
			p.currentTool = &toolAccumulator{}
		}
		if toolID != "" {
			p.currentTool.ID = toolID
		}
		if toolName != "" {
			p.currentTool.Name = toolName
		}
	}

	if p.currentTool != nil {
		switch input := payload["input"].(type) {
		case string:
			if strings.TrimSpace(input) != "" {
				p.currentTool.InputParts.WriteString(input)
				p.currentTool.HasInput = true
			}
		case map[string]any:
			encoded, _ := json.Marshal(input)
			p.currentTool.InputParts.Reset()
			p.currentTool.InputParts.Write(encoded)
			p.currentTool.HasInput = true
		}
	}

	if stop {
		p.finalizeCurrentTool()
	}
}

func (p *StreamParser) finalizeCurrentTool() {
	if p.currentTool == nil {
		return
	}

	toolName := strings.TrimSpace(p.currentTool.Name)
	if toolName != "" {
		toolUse := provider.ToolUse{
			ID:    strings.TrimSpace(p.currentTool.ID),
			Name:  toolName,
			Input: map[string]any{},
		}
		if p.currentTool.HasInput {
			toolUse.Input = parseToolArguments(p.currentTool.InputParts.String())
		}
		p.toolUses = append(p.toolUses, toolUse)
	}
	p.currentTool = nil
}

func readEventFrame(reader io.Reader) (eventFrame, error) {
	prelude := make([]byte, 12)
	if _, err := io.ReadFull(reader, prelude); err != nil {
		return eventFrame{}, err
	}

	totalLen := int(binary.BigEndian.Uint32(prelude[0:4]))
	headersLen := int(binary.BigEndian.Uint32(prelude[4:8]))
	if totalLen < 16 {
		return eventFrame{}, fmt.Errorf("invalid AWS event-stream frame length: %d", totalLen)
	}
	if binary.BigEndian.Uint32(prelude[8:12]) != crc32.ChecksumIEEE(prelude[:8]) {
		return eventFrame{}, fmt.Errorf("invalid AWS event-stream prelude CRC")
	}

	remaining := make([]byte, totalLen-12)
	if _, err := io.ReadFull(reader, remaining); err != nil {
		return eventFrame{}, err
	}
	if crc32.ChecksumIEEE(append(prelude[:], remaining[:len(remaining)-4]...)) != binary.BigEndian.Uint32(remaining[len(remaining)-4:]) {
		return eventFrame{}, fmt.Errorf("invalid AWS event-stream message CRC")
	}
	if headersLen > len(remaining)-4 {
		return eventFrame{}, fmt.Errorf("invalid AWS event-stream headers length")
	}

	headers := remaining[:headersLen]
	payload := remaining[headersLen : len(remaining)-4]
	frame := eventFrame{Payload: payload}
	frame.EventType, frame.MessageType = parseEventHeaders(headers)
	return frame, nil
}

func parseEventHeaders(headers []byte) (string, string) {
	offset := 0
	eventType := ""
	messageType := ""
	for offset < len(headers) {
		nameLen := int(headers[offset])
		offset++
		if offset+nameLen > len(headers) || offset >= len(headers) {
			break
		}
		name := string(headers[offset : offset+nameLen])
		offset += nameLen
		valueType := headers[offset]
		offset++
		if valueType != 7 || offset+2 > len(headers) {
			break
		}
		valueLen := int(binary.BigEndian.Uint16(headers[offset : offset+2]))
		offset += 2
		if offset+valueLen > len(headers) {
			break
		}
		value := string(headers[offset : offset+valueLen])
		offset += valueLen
		switch name {
		case ":event-type":
			eventType = value
		case ":message-type":
			messageType = value
		}
	}
	return eventType, messageType
}

func resolveEventType(headerType string, payload map[string]any) string {
	if strings.TrimSpace(headerType) != "" {
		return strings.TrimSpace(headerType)
	}
	if _, ok := payload["toolUseId"]; ok {
		return "toolUseEvent"
	}
	if _, ok := payload["name"]; ok {
		if _, hasInput := payload["input"]; hasInput {
			return "toolUseEvent"
		}
	}
	if _, ok := payload["text"]; ok {
		return "reasoningContentEvent"
	}
	if _, ok := payload["content"]; ok {
		return "assistantResponseEvent"
	}
	return ""
}

func resolveTextField(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if text, ok := payload[key].(string); ok && strings.TrimSpace(text) != "" {
			return text
		}
	}
	return ""
}

func deltaFromCumulative(previous *string, current string) string {
	if strings.TrimSpace(current) == "" {
		return ""
	}
	if previous == nil || *previous == "" {
		if previous != nil {
			*previous = current
		}
		return current
	}
	if current == *previous {
		return ""
	}
	if strings.HasPrefix(current, *previous) {
		delta := current[len(*previous):]
		*previous = current
		return delta
	}
	if strings.HasPrefix(*previous, current) {
		return ""
	}
	maxOverlap := 0
	maxLength := len(*previous)
	if len(current) < maxLength {
		maxLength = len(current)
	}
	for size := maxLength; size > 0; size-- {
		if strings.HasSuffix(*previous, current[:size]) {
			maxOverlap = size
			break
		}
	}
	*previous = current
	if maxOverlap > 0 {
		return current[maxOverlap:]
	}
	return current
}

func extractUsage(payload map[string]any) UsageSnapshot {
	usageMaps := make([]map[string]any, 0, 4)
	collectUsageMaps(payload, &usageMaps)
	usage := UsageSnapshot{}
	for _, item := range usageMaps {
		if item == nil {
			continue
		}
		if value, ok := readTokenNumber(item, "inputTokens", "promptTokens", "input_tokens", "prompt_tokens", "totalInputTokens", "total_input_tokens"); ok {
			usage.PromptTokens = value
		}
		if value, ok := readTokenNumber(item, "outputTokens", "completionTokens", "output_tokens", "completion_tokens", "totalOutputTokens", "total_output_tokens"); ok {
			usage.CompletionTokens = value
		}
		if value, ok := readTokenNumber(item, "totalTokens", "total_tokens"); ok {
			usage.TotalTokens = value
		}
	}
	if usage.TotalTokens == 0 && (usage.PromptTokens > 0 || usage.CompletionTokens > 0) {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	return usage
}

func collectUsageMaps(value any, usageMaps *[]map[string]any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			lowerKey := strings.ToLower(strings.TrimSpace(key))
			if lowerKey == "usage" || lowerKey == "tokenusage" || lowerKey == "token_usage" {
				if usage, ok := child.(map[string]any); ok {
					*usageMaps = append(*usageMaps, usage)
				}
			}
			collectUsageMaps(child, usageMaps)
		}
	case []any:
		for _, child := range typed {
			collectUsageMaps(child, usageMaps)
		}
	}
}

func readTokenNumber(values map[string]any, keys ...string) (int, bool) {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int(typed), true
		case int:
			return typed, true
		case int64:
			return int(typed), true
		case json.Number:
			parsed, err := typed.Int64()
			if err == nil {
				return int(parsed), true
			}
		case string:
			var parsed int
			_, err := fmt.Sscanf(strings.TrimSpace(typed), "%d", &parsed)
			if err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func errorMessageFromPayload(payload []byte) string {
	trimmed := strings.TrimSpace(string(payload))
	if trimmed == "" {
		return "upstream stream error"
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return trimmed
	}
	for _, key := range []string{"message", "Message", "errorMessage"} {
		if value, ok := decoded[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return trimmed
}

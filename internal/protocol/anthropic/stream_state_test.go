package anthropic

import "testing"

func TestMessageStreamState_EmitsOrderedTextLifecycle(t *testing.T) {
	var events []recordedStreamEvent
	state := NewMessageStreamState("msg_1", "claude-sonnet-4.5", func(name string, payload map[string]any) {
		events = append(events, recordedStreamEvent{name: name, payload: payload})
	})
	state.StartMessage(3)
	state.EmitTextDelta("hello")
	state.CloseTextBlock()
	state.Complete("end_turn", 5)

	if len(events) != 6 {
		t.Fatalf("event count = %d, want 6", len(events))
	}
	assertRecordedEventName(t, events[0], "message_start")
	assertRecordedEventName(t, events[1], "content_block_start")
	assertRecordedEventName(t, events[2], "content_block_delta")
	assertRecordedEventName(t, events[3], "content_block_stop")
	assertRecordedEventName(t, events[4], "message_delta")
	assertRecordedEventName(t, events[5], "message_stop")
	if state.NextIndex() != 1 {
		t.Fatalf("next index = %d, want 1", state.NextIndex())
	}
}

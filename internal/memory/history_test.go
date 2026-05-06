package memory

import "testing"

func TestTranscriptIsSessionScopedAndOrdered(t *testing.T) {
	h := NewHistory(t.TempDir())
	h.Append(HistoryEntry{SessionID: "a", RunID: "1", Role: "user", Content: "hello"})
	h.Append(HistoryEntry{SessionID: "b", RunID: "2", Role: "user", Content: "other"})
	h.Append(HistoryEntry{SessionID: "a", RunID: "3", Role: "assistant", Content: "hi"})
	h.Append(HistoryEntry{SessionID: "a", RunID: "4", Role: "tool", Content: "tool output"})

	got := h.Transcript("a", 10)
	if len(got) != 2 {
		t.Fatalf("expected 2 transcript messages, got %d: %#v", len(got), got)
	}
	if got[0].Role != "user" || got[0].Content != "hello" {
		t.Fatalf("unexpected first message: %#v", got[0])
	}
	if got[1].Role != "assistant" || got[1].Content != "hi" {
		t.Fatalf("unexpected second message: %#v", got[1])
	}
}

func TestTranscriptLimitKeepsMostRecent(t *testing.T) {
	h := NewHistory(t.TempDir())
	h.Append(HistoryEntry{SessionID: "a", RunID: "1", Role: "user", Content: "one"})
	h.Append(HistoryEntry{SessionID: "a", RunID: "2", Role: "assistant", Content: "two"})
	h.Append(HistoryEntry{SessionID: "a", RunID: "3", Role: "user", Content: "three"})

	got := h.Transcript("a", 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 transcript messages, got %d", len(got))
	}
	if got[0].Content != "two" || got[1].Content != "three" {
		t.Fatalf("unexpected limited transcript: %#v", got)
	}
}

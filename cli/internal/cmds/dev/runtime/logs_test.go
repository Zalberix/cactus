package runtime

import (
	"testing"
	"time"
)

func TestLogBufferKeepsLastEntries(t *testing.T) {
	buf := NewLogBuffer(3)

	for _, line := range []string{"one", "two", "three", "four"} {
		buf.Append(LogEntry{
			Time:     time.Unix(1, 0),
			TargetID: "manager",
			Stream:   StreamStdout,
			Line:     line,
		})
	}

	got := buf.Entries()
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if got[0].Line != "two" || got[1].Line != "three" || got[2].Line != "four" {
		t.Fatalf("unexpected entries: %#v", got)
	}
}

func TestLogStoreCreatesBuffersByTarget(t *testing.T) {
	store := NewLogStore(2)
	store.Append(LogEntry{TargetID: "a", Stream: StreamStdout, Line: "a1"})
	store.Append(LogEntry{TargetID: "b", Stream: StreamStderr, Line: "b1"})

	if got := store.Entries("a"); len(got) != 1 || got[0].Line != "a1" {
		t.Fatalf("unexpected entries for a: %#v", got)
	}
	if got := store.Entries("b"); len(got) != 1 || got[0].Line != "b1" {
		t.Fatalf("unexpected entries for b: %#v", got)
	}
}

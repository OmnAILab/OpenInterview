package textdeal

import (
	"testing"
	"unicode/utf8"
)

func TestBufferSubmitStopSplitsBetweenMarkers(t *testing.T) {
	var buffer Buffer

	snapshot := buffer.AppendStable("What did you do")
	snapshot = buffer.AppendStable("in that project?")

	if got, want := snapshot.StableText, "What did you do in that project?"; got != want {
		t.Fatalf("StableText = %q, want %q", got, want)
	}

	firstStop := utf8.RuneCountInString("What did you do")
	segment, snapshot, err := buffer.SubmitStop(firstStop)
	if err != nil {
		t.Fatalf("SubmitStop(first): %v", err)
	}
	if got, want := segment.Text, "What did you do"; got != want {
		t.Fatalf("first segment = %q, want %q", got, want)
	}
	if got, want := snapshot.PendingText, "in that project?"; got != want {
		t.Fatalf("PendingText after first stop = %q, want %q", got, want)
	}

	secondStop := utf8.RuneCountInString(snapshot.StableText)
	segment, snapshot, err = buffer.SubmitStop(secondStop)
	if err != nil {
		t.Fatalf("SubmitStop(second): %v", err)
	}
	if got, want := segment.Text, "in that project?"; got != want {
		t.Fatalf("second segment = %q, want %q", got, want)
	}
	if got := snapshot.PendingText; got != "" {
		t.Fatalf("PendingText after second stop = %q, want empty", got)
	}
}

func TestBufferRejectsNonForwardStop(t *testing.T) {
	var buffer Buffer
	buffer.AppendStable("abc")

	if _, _, err := buffer.SubmitStop(2); err != nil {
		t.Fatalf("SubmitStop(2): %v", err)
	}
	if _, _, err := buffer.SubmitStop(2); err == nil {
		t.Fatal("SubmitStop with same stop index succeeded, want error")
	}
}

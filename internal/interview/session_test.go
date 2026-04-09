package interview

import (
	"context"
	"io"
	"log"
	"testing"
	"time"
	"unicode/utf8"

	"openinterview/internal/llm"
	"openinterview/internal/stt"
)

type recordingLLM struct {
	questions chan string
}

func (r *recordingLLM) StreamAnswer(ctx context.Context, request llm.Request, sink llm.TokenSink) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r.questions <- request.Question:
	}

	answer := "answer:" + request.Question
	if sink != nil {
		sink(answer)
	}
	return answer, nil
}

func TestSessionSubmitTextSegmentStartsAnswer(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	llmClient := &recordingLLM{questions: make(chan string, 1)}
	session := newSession("session-1", Config{
		MaxTurns:         5,
		ExpectedRate:     16000,
		ExpectedChannels: 1,
		ExpectedEncoding: "pcm16",
		MaxChunkBytes:    262144,
	}, stt.NewFactory(stt.Config{Provider: "mock"}, logger), llmClient, logger)

	if _, err := session.InjectTranscript(context.Background(), stt.TranscriptEvent{
		Kind: stt.EventFinal,
		Text: "第一段",
	}); err != nil {
		t.Fatalf("InjectTranscript(first): %v", err)
	}
	if _, err := session.InjectTranscript(context.Background(), stt.TranscriptEvent{
		Kind: stt.EventFinal,
		Text: "第二段",
	}); err != nil {
		t.Fatalf("InjectTranscript(second): %v", err)
	}

	select {
	case question := <-llmClient.questions:
		t.Fatalf("llm triggered before textdeal stop submission: %q", question)
	default:
	}

	stop := utf8.RuneCountInString("第一段")
	snapshot, err := session.SubmitTextSegment(stop)
	if err != nil {
		t.Fatalf("SubmitTextSegment: %v", err)
	}
	if got, want := snapshot.CurrentQuestion, "第一段"; got != want {
		t.Fatalf("CurrentQuestion = %q, want %q", got, want)
	}
	if got, want := snapshot.TextDeal.PendingText, "第二段"; got != want {
		t.Fatalf("PendingText = %q, want %q", got, want)
	}

	select {
	case question := <-llmClient.questions:
		if got, want := question, "第一段"; got != want {
			t.Fatalf("llm question = %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for llm question")
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		current := session.Snapshot()
		if current.CurrentAnswer == "answer:第一段" && len(current.History) == 1 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("snapshot not updated in time: %#v", current)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

package interview

import (
	"context"
	"io"
	"log"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"openinterview/internal/knowledge"
	"openinterview/internal/llm"
	"openinterview/internal/stt"
)

type recordingLLM struct {
	requests chan llm.Request
}

func (r *recordingLLM) StreamAnswer(ctx context.Context, request llm.Request, sink llm.TokenSink) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r.requests <- request:
	}

	answer := "answer:" + request.Question
	if sink != nil {
		sink(answer)
	}
	return answer, nil
}

type stubKnowledgeClient struct {
	docs []knowledge.Document
	err  error
}

func (s stubKnowledgeClient) Retrieve(_ context.Context, _ string) ([]knowledge.Document, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]knowledge.Document(nil), s.docs...), nil
}

func TestSessionSubmitTextSegmentStartsAnswer(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	llmClient := &recordingLLM{requests: make(chan llm.Request, 1)}
	session := newSession("session-1", Config{
		MaxTurns:         5,
		ExpectedRate:     16000,
		ExpectedChannels: 1,
		ExpectedEncoding: "pcm16",
		MaxChunkBytes:    262144,
	}, stt.NewFactory(stt.Config{Provider: "mock"}, logger), llmClient, stubKnowledgeClient{}, logger)

	if _, err := session.InjectTranscript(context.Background(), stt.TranscriptEvent{
		Kind: stt.EventFinal,
		Text: "first segment",
	}); err != nil {
		t.Fatalf("InjectTranscript(first): %v", err)
	}
	if _, err := session.InjectTranscript(context.Background(), stt.TranscriptEvent{
		Kind: stt.EventFinal,
		Text: "second segment",
	}); err != nil {
		t.Fatalf("InjectTranscript(second): %v", err)
	}

	select {
	case request := <-llmClient.requests:
		t.Fatalf("llm triggered before textdeal stop submission: %q", request.Question)
	default:
	}

	stop := utf8.RuneCountInString("first segment")
	snapshot, err := session.SubmitTextSegment(stop)
	if err != nil {
		t.Fatalf("SubmitTextSegment: %v", err)
	}
	if got, want := snapshot.CurrentQuestion, "first segment"; got != want {
		t.Fatalf("CurrentQuestion = %q, want %q", got, want)
	}
	if got, want := snapshot.TextDeal.PendingText, "second segment"; got != want {
		t.Fatalf("PendingText = %q, want %q", got, want)
	}

	select {
	case request := <-llmClient.requests:
		if got, want := request.Question, "first segment"; got != want {
			t.Fatalf("llm question = %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for llm question")
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		current := session.Snapshot()
		if current.CurrentAnswer == "answer:first segment" && len(current.History) == 1 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("snapshot not updated in time: %#v", current)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestSessionAskQuestionInjectsKnowledgeContext(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	llmClient := &recordingLLM{requests: make(chan llm.Request, 1)}
	knowledgeClient := stubKnowledgeClient{
		docs: []knowledge.Document{
			{
				Title:   "rag-design",
				Content: "Use sentence-transformers embeddings and cosine similarity for top-k retrieval.",
				Path:    "knowledge/rag.md",
				Score:   0.91,
			},
		},
	}

	session := newSession("session-2", Config{
		MaxTurns:         5,
		ExpectedRate:     16000,
		ExpectedChannels: 1,
		ExpectedEncoding: "pcm16",
		MaxChunkBytes:    262144,
	}, stt.NewFactory(stt.Config{Provider: "mock"}, logger), llmClient, knowledgeClient, logger)

	if _, err := session.AskQuestion("How would you add RAG to this project?"); err != nil {
		t.Fatalf("AskQuestion: %v", err)
	}

	var request llm.Request
	select {
	case request = <-llmClient.requests:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for llm request")
	}

	if len(request.Messages) == 0 {
		t.Fatal("request.Messages is empty")
	}
	if got := request.Messages[0].Content; !strings.Contains(got, "Retrieved knowledge:") {
		t.Fatalf("system message missing retrieved knowledge section: %q", got)
	}
	if got := request.Messages[0].Content; !strings.Contains(got, "sentence-transformers embeddings") {
		t.Fatalf("system message missing knowledge content: %q", got)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		current := session.Snapshot()
		if len(current.Knowledge) == 1 && current.Knowledge[0].Title == "rag-design" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("knowledge snapshot not updated in time: %#v", current)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

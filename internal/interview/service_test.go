package interview

import (
	"io"
	"log"
	"testing"

	"openinterview/internal/llm"
	"openinterview/internal/stt"
)

func TestServiceDeleteSessionRemovesSession(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	service := NewService(Config{
		MaxTurns:         5,
		ExpectedRate:     16000,
		ExpectedChannels: 1,
		ExpectedEncoding: "pcm16",
		MaxChunkBytes:    262144,
	}, stt.NewFactory(stt.Config{Provider: "mock"}, logger), llm.NewClient(llm.Config{Provider: "mock"}, logger), logger)

	snapshot := service.CreateSession()
	if err := service.DeleteSession(snapshot.ID); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	if _, err := service.GetSnapshot(snapshot.ID); err != ErrNotFound {
		t.Fatalf("GetSnapshot after delete err = %v, want %v", err, ErrNotFound)
	}
}

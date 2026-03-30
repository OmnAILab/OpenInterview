package stt

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

type Config struct {
	Provider string
	WSURL    string
}

type EventKind string

const (
	EventPartial EventKind = "partial"
	EventFinal   EventKind = "final"
)

type AudioChunk struct {
	Data       []byte
	SampleRate int
	Channels   int
	Encoding   string
	At         time.Time
}

type TranscriptEvent struct {
	Kind     EventKind `json:"kind"`
	Text     string    `json:"text"`
	Endpoint bool      `json:"endpoint"`
}

type Sink interface {
	HandleTranscriptEvent(ctx context.Context, event TranscriptEvent)
	HandleSTTError(ctx context.Context, err error)
}

type Stream interface {
	Push(ctx context.Context, chunk AudioChunk) error
	Close() error
}

type Factory interface {
	NewStream(sessionID string, sink Sink) (Stream, error)
}

func NewFactory(cfg Config, logger *log.Logger) Factory {
	switch strings.ToLower(cfg.Provider) {
	case "", "mock":
		return &mockFactory{logger: logger}
	case "sherpa", "sherpa-websocket", "sherpa_onnx", "sherpa-onnx":
		return &sherpaWebSocketFactory{
			wsURL:  cfg.WSURL,
			logger: logger,
		}
	default:
		return &unsupportedFactory{
			err: fmt.Errorf("unknown stt provider %q", cfg.Provider),
		}
	}
}

type unsupportedFactory struct {
	err error
}

func (f *unsupportedFactory) NewStream(string, Sink) (Stream, error) {
	return nil, f.err
}

type mockFactory struct {
	logger *log.Logger
}

func (f *mockFactory) NewStream(sessionID string, sink Sink) (Stream, error) {
	return &mockStream{
		sessionID: sessionID,
		sink:      sink,
		logger:    f.logger,
	}, nil
}

type mockStream struct {
	sessionID string
	sink      Sink
	logger    *log.Logger
	closed    bool
}

func (m *mockStream) Push(ctx context.Context, chunk AudioChunk) error {
	if m.closed {
		return fmt.Errorf("stream closed")
	}
	if m.logger != nil && len(chunk.Data) > 0 {
		m.logger.Printf("[stt] mock stream received %d bytes for session %s", len(chunk.Data), m.sessionID)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (m *mockStream) Close() error {
	m.closed = true
	return nil
}

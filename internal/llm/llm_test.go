package llm

import (
	"io"
	"log"
	"testing"
)

func TestParseOpenAIChunk(t *testing.T) {
	token, done, err := parseOpenAIChunk([]byte(`{"choices":[{"delta":{"content":"hello"},"finish_reason":""}]}`))
	if err != nil {
		t.Fatalf("parseOpenAIChunk returned error: %v", err)
	}
	if token != "hello" {
		t.Fatalf("unexpected token: %q", token)
	}
	if done {
		t.Fatal("expected done=false")
	}

	token, done, err = parseOpenAIChunk([]byte(`{"choices":[{"delta":{},"finish_reason":"stop"}]}`))
	if err != nil {
		t.Fatalf("parseOpenAIChunk returned error: %v", err)
	}
	if token != "" || !done {
		t.Fatalf("unexpected final state token=%q done=%v", token, done)
	}
}

func TestNewClient_SelectsProviderImplementation(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	if _, ok := NewClient(Config{Provider: "groq"}, logger).(*groqClient); !ok {
		t.Fatal("groq provider should return *groqClient")
	}

	if _, ok := NewClient(Config{Provider: "openai-compatible"}, logger).(*openAICompatibleClient); !ok {
		t.Fatal("openai-compatible provider should return *openAICompatibleClient")
	}

	if _, ok := NewClient(Config{Provider: "mock"}, logger).(*mockClient); !ok {
		t.Fatal("mock provider should return *mockClient")
	}
}

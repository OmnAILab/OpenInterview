package llm

import (
	"io"
	"log"
	"testing"
)

func TestParseAnthropicChunk(t *testing.T) {
	// content_block_delta with text_delta
	token, done, err := parseAnthropicChunk([]byte(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hello"}}`))
	if err != nil {
		t.Fatalf("parseAnthropicChunk returned error: %v", err)
	}
	if token != "hello" {
		t.Fatalf("unexpected token: %q", token)
	}
	if done {
		t.Fatal("expected done=false")
	}

	// message_stop signals end
	token, done, err = parseAnthropicChunk([]byte(`{"type":"message_stop"}`))
	if err != nil {
		t.Fatalf("parseAnthropicChunk returned error: %v", err)
	}
	if token != "" || !done {
		t.Fatalf("unexpected final state token=%q done=%v", token, done)
	}
}

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

	if _, ok := NewClient(Config{Provider: "openai"}, logger).(*openAIClient); !ok {
		t.Fatal("openai provider should return *openAIClient")
	}

	if _, ok := NewClient(Config{Provider: "anthropic"}, logger).(*anthropicClient); !ok {
		t.Fatal("anthropic provider should return *anthropicClient")
	}

	if _, ok := NewClient(Config{Provider: "claude"}, logger).(*anthropicClient); !ok {
		t.Fatal("claude provider should return *anthropicClient")
	}

	if _, ok := NewClient(Config{Provider: "deepseek"}, logger).(*deepSeekClient); !ok {
		t.Fatal("deepseek provider should return *deepSeekClient")
	}

	if _, ok := NewClient(Config{Provider: "zhipu"}, logger).(*zhipuClient); !ok {
		t.Fatal("zhipu provider should return *zhipuClient")
	}

	if _, ok := NewClient(Config{Provider: "glm"}, logger).(*zhipuClient); !ok {
		t.Fatal("glm provider should return *zhipuClient")
	}

	if _, ok := NewClient(Config{Provider: "qianwen"}, logger).(*qianwenClient); !ok {
		t.Fatal("qianwen provider should return *qianwenClient")
	}

	if _, ok := NewClient(Config{Provider: "moonshot"}, logger).(*moonshotClient); !ok {
		t.Fatal("moonshot provider should return *moonshotClient")
	}

	if _, ok := NewClient(Config{Provider: "kimi"}, logger).(*moonshotClient); !ok {
		t.Fatal("kimi provider should return *moonshotClient")
	}

	if _, ok := NewClient(Config{Provider: "gemini"}, logger).(*geminiClient); !ok {
		t.Fatal("gemini provider should return *geminiClient")
	}

	if _, ok := NewClient(Config{Provider: "ollama"}, logger).(*ollamaClient); !ok {
		t.Fatal("ollama provider should return *ollamaClient")
	}

	if _, ok := NewClient(Config{Provider: "vllm"}, logger).(*openAICompatibleClient); !ok {
		t.Fatal("vllm provider should return *openAICompatibleClient")
	}
}

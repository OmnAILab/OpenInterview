package llm

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOpenAICompatibleClient_StreamAnswer_Integration(t *testing.T) {
	if os.Getenv("RUN_LLM_INTEGRATION") != "1" {
		t.Skip("set RUN_LLM_INTEGRATION=1 to run the external LLM integration test")
	}

	baseURL := strings.TrimSpace(os.Getenv("INTERVIEW_LLM_BASE_URL"))
	endpoint := strings.TrimSpace(os.Getenv("INTERVIEW_LLM_ENDPOINT"))
	apiKey := strings.TrimSpace(os.Getenv("INTERVIEW_LLM_API_KEY"))
	model := strings.TrimSpace(os.Getenv("INTERVIEW_LLM_MODEL"))

	if baseURL == "" || model == "" {
		t.Fatal("INTERVIEW_LLM_BASE_URL and INTERVIEW_LLM_MODEL are required for the integration test")
	}

	cfg := Config{
		Provider:     "openai-compatible",
		BaseURL:      baseURL,
		Endpoint:     endpoint,
		APIKey:       apiKey,
		Model:        model,
		SystemPrompt: "You are a concise assistant. Answer in Chinese.",
		Temperature:  0.2,
		Timeout:      60 * time.Second,
	}

	logger := log.New(log.Writer(), "[llm-test] ", log.LstdFlags|log.Lshortfile)
	client := NewClient(cfg, logger)

	req := Request{
		Question: "1 + 1 equals what?",
		Messages: []Message{{
			Role:    RoleUser,
			Content: "1 + 1 equals what? Reply with the result only.",
		}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var (
		mu         sync.Mutex
		chunks     []string
		chunkCount int
	)

	answer, err := client.StreamAnswer(ctx, req, func(token string) {
		mu.Lock()
		defer mu.Unlock()

		chunkCount++
		chunks = append(chunks, token)
		t.Logf("chunk #%d: %q", chunkCount, token)
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	streamText := strings.Join(chunks, "")
	if strings.TrimSpace(answer) == "" {
		t.Fatal("answer is empty")
	}
	if streamText != answer {
		t.Fatalf("stream text and final answer differ\nstream=%q\nanswer=%q", streamText, answer)
	}
	if chunkCount == 0 {
		t.Fatal("no streamed chunks received")
	}
}

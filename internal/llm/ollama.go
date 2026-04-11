package llm

import (
	"context"
	"log"
	"net/http"
)

// ollamaClient wraps Ollama's OpenAI-compatible local API.
type ollamaClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newOllamaClient(cfg Config, logger *log.Logger) Client {
	return &ollamaClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *ollamaClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamOpenAICompatibleAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

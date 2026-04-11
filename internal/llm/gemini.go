package llm

import (
	"context"
	"log"
	"net/http"
)

// geminiClient wraps the Google Gemini OpenAI-compatible API
// (https://generativelanguage.googleapis.com/v1beta/openai/).
type geminiClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newGeminiClient(cfg Config, logger *log.Logger) Client {
	return &geminiClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *geminiClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamOpenAICompatibleAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

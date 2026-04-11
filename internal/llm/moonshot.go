package llm

import (
	"context"
	"log"
	"net/http"
)

// moonshotClient wraps the Moonshot AI (月之暗面 / Kimi) OpenAI-compatible API.
type moonshotClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newMoonshotClient(cfg Config, logger *log.Logger) Client {
	return &moonshotClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *moonshotClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamOpenAICompatibleAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

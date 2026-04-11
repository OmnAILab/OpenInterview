package llm

import (
	"context"
	"log"
	"net/http"
)

type groqClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newGroqClient(cfg Config, logger *log.Logger) Client {
	return &groqClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *groqClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamOpenAICompatibleAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

package llm

import (
	"context"
	"log"
	"net/http"
)

type openAIClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newOpenAIClient(cfg Config, logger *log.Logger) Client {
	return &openAIClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *openAIClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamOpenAICompatibleAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

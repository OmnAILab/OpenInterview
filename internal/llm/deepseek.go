package llm

import (
	"context"
	"log"
	"net/http"
)

type deepSeekClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newDeepSeekClient(cfg Config, logger *log.Logger) Client {
	return &deepSeekClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *deepSeekClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamOpenAICompatibleAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

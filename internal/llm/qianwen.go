package llm

import (
	"context"
	"log"
	"net/http"
)

// qianwenClient wraps Alibaba Cloud Qianwen (通义千问 / DashScope) OpenAI-compatible API.
type qianwenClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newQianwenClient(cfg Config, logger *log.Logger) Client {
	return &qianwenClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *qianwenClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamOpenAICompatibleAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

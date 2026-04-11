package llm

import (
	"context"
	"log"
	"net/http"
)

// zhipuClient wraps the Zhipu AI (智谱 GLM) OpenAI-compatible chat API.
type zhipuClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newZhipuClient(cfg Config, logger *log.Logger) Client {
	return &zhipuClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *zhipuClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamOpenAICompatibleAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

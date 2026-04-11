package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	anthropicAPIVersion  = "2023-06-01"
	anthropicDefaultURL  = "https://api.anthropic.com"
	anthropicEndpoint    = "/v1/messages"
	anthropicMaxTokens   = 4096
)

// anthropicClient calls Anthropic's Messages API (/v1/messages) with SSE streaming.
type anthropicClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newAnthropicClient(cfg Config, logger *log.Logger) Client {
	return &anthropicClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *anthropicClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamAnthropicAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

func streamAnthropicAnswer(ctx context.Context, cfg Config, httpClient *http.Client, request Request, sink TokenSink) (string, error) {
	// Separate system prompt from conversation messages.
	systemText := strings.TrimSpace(cfg.SystemPrompt)
	msgs := request.Messages

	// Build Anthropic messages (only user / assistant roles).
	type anthropicMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var anthropicMsgs []anthropicMsg
	for _, m := range msgs {
		if m.Role == RoleSystem {
			if systemText == "" {
				systemText = m.Content
			} else {
				systemText = systemText + "\n\n" + m.Content
			}
			continue
		}
		anthropicMsgs = append(anthropicMsgs, anthropicMsg{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	payload := map[string]any{
		"model":      cfg.Model,
		"messages":   anthropicMsgs,
		"max_tokens": anthropicMaxTokens,
		"stream":     true,
	}
	if systemText != "" {
		payload["system"] = systemText
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = anthropicEndpoint
	}
	url := joinURL(cfg.BaseURL, endpoint)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)
	if cfg.APIKey != "" {
		httpReq.Header.Set("x-api-key", cfg.APIKey)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("anthropic request failed: %s: %s", resp.Status, strings.TrimSpace(string(msg)))
	}

	var answer strings.Builder
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		if strings.HasPrefix(line, "event:") {
			continue
		}

		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}

		token, done, err := parseAnthropicChunk([]byte(data))
		if err != nil {
			// Non-fatal: skip unknown event types.
			continue
		}
		if token != "" {
			answer.WriteString(token)
			if sink != nil {
				sink(token)
			}
		}
		if done {
			break
		}
	}

	return answer.String(), nil
}

// parseAnthropicChunk extracts a text delta from an Anthropic SSE data payload.
func parseAnthropicChunk(payload []byte) (token string, done bool, err error) {
	var event struct {
		Type  string `json:"type"`
		Index int    `json:"index"`
		Delta struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"delta"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return "", false, err
	}

	switch event.Type {
	case "content_block_delta":
		if event.Delta.Type == "text_delta" {
			return event.Delta.Text, false, nil
		}
	case "message_stop":
		return "", true, nil
	case "message_delta":
		// Contains usage info; not a terminal signal on its own.
	}
	return "", false, nil
}

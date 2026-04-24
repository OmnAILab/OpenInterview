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
	"time"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Config struct {
	Provider     string
	BaseURL      string
	Endpoint     string
	APIKey       string
	Model        string
	SystemPrompt string
	Temperature  float64
	Timeout      time.Duration
}

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Question string
	Messages []Message
}

type TokenSink func(token string)

type Client interface {
	StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error)
}

func NewClient(cfg Config, logger *log.Logger) Client {
	switch strings.ToLower(cfg.Provider) {
	case "", "mock":
		return &mockClient{}
	case "openai-compatible", "openai_compatible":
		return newOpenAICompatibleClient(cfg, logger)
	case "openai":
		return newOpenAIClient(cfg, logger)
	case "groq":
		return newGroqClient(cfg, logger)
	case "anthropic", "claude":
		return newAnthropicClient(cfg, logger)
	case "deepseek":
		return newDeepSeekClient(cfg, logger)
	case "zhipu", "glm":
		return newZhipuClient(cfg, logger)
	case "qianwen", "dashscope", "tongyi":
		return newQianwenClient(cfg, logger)
	case "moonshot", "kimi":
		return newMoonshotClient(cfg, logger)
	case "gemini", "google":
		return newGeminiClient(cfg, logger)
	case "ollama":
		return newOllamaClient(cfg, logger)
	case "vllm":
		return newOpenAICompatibleClient(cfg, logger)
	default:
		return &mockClient{}
	}
}

type mockClient struct{}

func (m *mockClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	answer := buildMockAnswer(request.Question)
	chunks := splitIntoChunks(answer, 18)

	var builder strings.Builder
	for _, chunk := range chunks {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		builder.WriteString(chunk)
		if sink != nil {
			sink(chunk)
		}
	}

	return builder.String(), nil
}

type openAICompatibleClient struct {
	cfg        Config
	httpClient *http.Client
	logger     *log.Logger
}

func newOpenAICompatibleClient(cfg Config, logger *log.Logger) Client {
	return &openAICompatibleClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *openAICompatibleClient) StreamAnswer(ctx context.Context, request Request, sink TokenSink) (string, error) {
	return streamOpenAICompatibleAnswer(ctx, c.cfg, c.httpClient, request, sink)
}

func streamOpenAICompatibleAnswer(ctx context.Context, cfg Config, httpClient *http.Client, request Request, sink TokenSink) (string, error) {
	payload := map[string]any{
		"model":       cfg.Model,
		"messages":    prependSystemPrompt(cfg.SystemPrompt, request.Messages),
		"stream":      true,
		"temperature": cfg.Temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(cfg.BaseURL, cfg.Endpoint), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("llm request failed: %s: %s", resp.Status, strings.TrimSpace(string(msg)))
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
		if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}

		token, done, err := parseOpenAIChunk([]byte(data))
		if err != nil {
			return "", err
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

func prependSystemPrompt(systemPrompt string, messages []Message) []Message {
	if strings.TrimSpace(systemPrompt) == "" {
		return messages
	}

	if len(messages) > 0 && messages[0].Role == RoleSystem {
		result := append([]Message(nil), messages...)
		result[0].Content = strings.TrimSpace(systemPrompt + "\n\n" + result[0].Content)
		return result
	}

	result := make([]Message, 0, len(messages)+1)
	result = append(result, Message{
		Role:    RoleSystem,
		Content: systemPrompt,
	})
	result = append(result, messages...)
	return result
}

func parseOpenAIChunk(payload []byte) (token string, done bool, err error) {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(payload, &chunk); err != nil {
		return "", false, err
	}
	if len(chunk.Choices) == 0 {
		return "", false, nil
	}

	if chunk.Choices[0].Delta.Content != "" {
		return chunk.Choices[0].Delta.Content, false, nil
	}
	if chunk.Choices[0].Message.Content != "" {
		return chunk.Choices[0].Message.Content, false, nil
	}
	return "", chunk.Choices[0].FinishReason != "", nil
}

func splitIntoChunks(text string, chunkSize int) []string {
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}
	if chunkSize <= 0 {
		chunkSize = len(runes)
	}

	var chunks []string
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}
	return chunks
}

func buildMockAnswer(question string) string {
	return fmt.Sprintf("I would answer this by stating the conclusion first, then covering the business context, my ownership, the key trade-offs, and the final outcome. For the question %q, that structure will sound much more like a real interview answer.", strings.TrimSpace(question))
}

func joinURL(base, endpoint string) string {
	if endpoint == "" {
		return base
	}
	if strings.HasSuffix(base, "/") && strings.HasPrefix(endpoint, "/") {
		return base + endpoint[1:]
	}
	if strings.HasSuffix(base, "/") || strings.HasPrefix(endpoint, "/") {
		return base + endpoint
	}
	return base + "/" + endpoint
}

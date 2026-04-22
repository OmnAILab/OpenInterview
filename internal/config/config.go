package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	Session   SessionConfig
	Audio     AudioConfig
	STT       STTConfig
	LLM       LLMConfig
	Knowledge KnowledgeConfig
}

type ServerConfig struct {
	Addr string
}

type SessionConfig struct {
	MaxTurns int
}

type AudioConfig struct {
	SampleRate    int
	Channels      int
	Encoding      string
	MaxChunkBytes int
}

type STTConfig struct {
	Provider string
	Sherpa   *SherpaConfig
	Tencent  *TencentConfig
}

type SherpaConfig struct {
	WSURL string
}

type TencentConfig struct {
	WSURL         string
	AppID         string
	SecretID      string
	SecretKey     string
	EngineType    string
	NeedVAD       int
	NoEmptyResult int
}

type LLMConfig struct {
	Provider     string
	BaseURL      string
	Endpoint     string
	APIKey       string
	Model        string
	SystemPrompt string
	Temperature  float64
	Timeout      time.Duration
}

type KnowledgeConfig struct {
	SearchEndpoint    string
	Path              string
	APIKey            string
	EmbeddingEndpoint string
	EmbeddingModel    string
	MaxResults        int
	Timeout           time.Duration
	LocalAddr         string
}

func Load() Config {
	loadDotEnv(filepath.Join(".", ".env"))

	sttProvider := strings.ToLower(envString("INTERVIEW_STT_PROVIDER", "mock"))
	llmProvider := strings.ToLower(envString("INTERVIEW_LLM_PROVIDER", "mock"))

	return Config{
		Server: ServerConfig{
			Addr: envString("INTERVIEW_ADDR", ":8080"),
		},
		Session: SessionConfig{
			MaxTurns: envInt("INTERVIEW_MAX_TURNS", 5),
		},
		Audio: AudioConfig{
			SampleRate:    envInt("INTERVIEW_AUDIO_SAMPLE_RATE", 16000),
			Channels:      envInt("INTERVIEW_AUDIO_CHANNELS", 1),
			Encoding:      strings.ToLower(envString("INTERVIEW_AUDIO_ENCODING", "pcm16")),
			MaxChunkBytes: envInt("INTERVIEW_AUDIO_MAX_CHUNK_BYTES", 262144),
		},
		STT: STTConfig{
			Provider: sttProvider,
			Sherpa: &SherpaConfig{
				WSURL: envString("INTERVIEW_SHERPA_WS_URL", ""),
			},
			Tencent: &TencentConfig{
				WSURL:         envString("INTERVIEW_TENCENT_WS_URL", ""),
				AppID:         envString("INTERVIEW_TENCENT_APP_ID", ""),
				SecretID:      envString("INTERVIEW_TENCENT_SECRET_ID", ""),
				SecretKey:     envString("INTERVIEW_TENCENT_SECRET_KEY", ""),
				EngineType:    envString("INTERVIEW_TENCENT_ENGINE_TYPE", "16k_zh"),
				NeedVAD:       envInt("INTERVIEW_TENCENT_NEED_VAD", 0),
				NoEmptyResult: envInt("INTERVIEW_TENCENT_NO_EMPTY_RESULT", 1),
			},
		},
		LLM: LLMConfig{
			Provider:     llmProvider,
			BaseURL:      strings.TrimRight(defaultLLMResolvedBaseURL(llmProvider), "/"),
			Endpoint:     envString("INTERVIEW_LLM_ENDPOINT", defaultLLMEndpoint(llmProvider)),
			APIKey:       defaultLLMAPIKey(llmProvider),
			Model:        defaultLLMModel(llmProvider),
			SystemPrompt: envString("INTERVIEW_LLM_SYSTEM_PROMPT", defaultSystemPrompt),
			Temperature:  envFloat("INTERVIEW_LLM_TEMPERATURE", 0.2),
			Timeout:      envDuration("INTERVIEW_LLM_TIMEOUT", 90*time.Second),
		},
		Knowledge: KnowledgeConfig{
			SearchEndpoint:    envString("INTERVIEW_KNOWLEDGE_ENDPOINT", ""),
			Path:              firstNonEmptyEnv("INTERVIEW_KNOWLEDGE_LOCAL_PATH", "INTERVIEW_KNOWLEDGE_PATH"),
			APIKey:            firstNonEmptyEnv("INTERVIEW_KNOWLEDGE_API_KEY", "INTERVIEW_EMBEDDING_API_KEY"),
			EmbeddingEndpoint: firstNonEmptyEnv("INTERVIEW_KNOWLEDGE_EMBEDDING_ENDPOINT", "INTERVIEW_EMBEDDING_ENDPOINT"),
			EmbeddingModel:    firstNonEmptyEnv("INTERVIEW_KNOWLEDGE_EMBEDDING_MODEL", "INTERVIEW_EMBEDDING_MODEL"),
			MaxResults:        envInt("INTERVIEW_KNOWLEDGE_MAX_RESULTS", 5),
			Timeout:           envDuration("INTERVIEW_KNOWLEDGE_TIMEOUT", 10*time.Second),
			LocalAddr:         envString("INTERVIEW_KNOWLEDGE_LOCAL_ADDR", ":7007"),
		},
	}
}

const defaultSystemPrompt = "You are a local interview copilot. Help the candidate answer interview questions in a concise, credible, first-person style. Use the candidate profile and recent context, do not invent experience, and reply in the same language as the question."

func defaultLLMBaseURL(provider string) string {
	switch provider {
	case "groq":
		return "https://api.groq.com/openai/v1"
	case "openai":
		return "https://api.openai.com/v1"
	case "anthropic", "claude":
		return "https://api.anthropic.com"
	case "deepseek":
		return "https://api.deepseek.com/v1"
	case "zhipu", "glm":
		return "https://open.bigmodel.cn/api/paas/v4"
	case "qianwen", "dashscope", "tongyi":
		return "https://dashscope.aliyuncs.com/compatible-mode/v1"
	case "moonshot", "kimi":
		return "https://api.moonshot.cn/v1"
	case "gemini", "google":
		return "https://generativelanguage.googleapis.com/v1beta/openai"
	case "ollama":
		return "http://127.0.0.1:11434/v1"
	default:
		return "http://127.0.0.1:1234/v1"
	}
}

// defaultLLMResolvedBaseURL resolves the base URL for the given provider.
// Lookup order: INTERVIEW_LLM_BASE_URL → {PREFIX}_BASE_URL → hardcoded default.
func defaultLLMResolvedBaseURL(provider string) string {
	if v := os.Getenv("INTERVIEW_LLM_BASE_URL"); strings.TrimSpace(v) != "" {
		return v
	}
	if prefix := providerEnvPrefix(provider); prefix != "" {
		if v := os.Getenv(prefix + "_BASE_URL"); strings.TrimSpace(v) != "" {
			return v
		}
	}
	return defaultLLMBaseURL(provider)
}

// providerEnvPrefix returns the uppercase environment variable prefix for a provider.
// For example, "groq" → "GROQ", "anthropic" → "ANTHROPIC".
// Returns an empty string for providers without a standard prefix.
func providerEnvPrefix(provider string) string {
	switch provider {
	case "groq":
		return "GROQ"
	case "openai":
		return "OPENAI"
	case "anthropic", "claude":
		return "ANTHROPIC"
	case "deepseek":
		return "DEEPSEEK"
	case "zhipu", "glm":
		return "ZHIPU"
	case "qianwen", "dashscope", "tongyi":
		return "DASHSCOPE"
	case "moonshot", "kimi":
		return "MOONSHOT"
	case "gemini", "google":
		return "GEMINI"
	case "ollama":
		return "OLLAMA"
	default:
		return ""
	}
}

// defaultLLMModel resolves the model name for the given provider.
// Lookup order: {PREFIX}_MODEL → INTERVIEW_LLM_MODEL → hardcoded default.
func defaultLLMModel(provider string) string {
	if prefix := providerEnvPrefix(provider); prefix != "" {
		if v := os.Getenv(prefix + "_MODEL"); strings.TrimSpace(v) != "" {
			return v
		}
	}
	if v := os.Getenv("INTERVIEW_LLM_MODEL"); strings.TrimSpace(v) != "" {
		return v
	}
	switch provider {
	case "groq":
		return "llama-3.3-70b-versatile"
	case "openai":
		return "gpt-4o"
	case "anthropic", "claude":
		return "claude-opus-4-5"
	case "deepseek":
		return "deepseek-chat"
	case "zhipu", "glm":
		return "glm-4-flash"
	case "qianwen", "dashscope", "tongyi":
		return "qwen-max"
	case "moonshot", "kimi":
		return "moonshot-v1-8k"
	case "gemini", "google":
		return "gemini-2.0-flash"
	case "ollama":
		return "qwen2.5:7b"
	default:
		return "local-model"
	}
}

// defaultSTTWSURL returns the default WebSocket URL for a locally-running STT server
// listening on the given port.
func defaultSTTWSURL(port int) string {
	return fmt.Sprintf("ws://127.0.0.1:%d/", port)
}

// defaultLLMEndpoint returns the default chat-completion endpoint path for a provider.
// Anthropic uses a different path; all others follow the OpenAI convention.
func defaultLLMEndpoint(provider string) string {
	if provider == "anthropic" || provider == "claude" {
		return "/v1/messages"
	}
	return "/chat/completions"
}

// defaultLLMAPIKey resolves the API key for the given provider.
// Lookup order: {PREFIX}_API_KEY → INTERVIEW_LLM_API_KEY.
func defaultLLMAPIKey(provider string) string {
	var candidates []string
	if prefix := providerEnvPrefix(provider); prefix != "" {
		candidates = append(candidates, prefix+"_API_KEY")
	}
	candidates = append(candidates, "INTERVIEW_LLM_API_KEY")
	for _, key := range candidates {
		if v := os.Getenv(key); strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func loadDotEnv(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		key, value, ok := parseDotEnvLine(scanner.Text())
		if !ok {
			continue
		}
		// ← 移除这个条件判断，总是设置 .env 中的值
		// if _, exists := os.LookupEnv(key); exists {
		// 	continue
		// }
		_ = os.Setenv(key, value)
	}
}

func parseDotEnvLine(line string) (key string, value string, ok bool) {
	line = strings.TrimSpace(strings.TrimPrefix(line, "\ufeff"))
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}

	if strings.HasPrefix(line, "export ") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	key = strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", false
	}

	rawValue := strings.TrimSpace(parts[1])
	if len(rawValue) >= 2 {
		if (rawValue[0] == '"' && rawValue[len(rawValue)-1] == '"') || (rawValue[0] == '\'' && rawValue[len(rawValue)-1] == '\'') {
			return key, rawValue[1 : len(rawValue)-1], true
		}
	}

	return key, rawValue, true
}

func envString(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := envString(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envFloat(key string, fallback float64) float64 {
	value := envString(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := envString(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := envString(key, ""); strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

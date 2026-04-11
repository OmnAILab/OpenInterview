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
	Server  ServerConfig
	Session SessionConfig
	Audio   AudioConfig
	STT     STTConfig
	LLM     LLMConfig
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
			BaseURL:      strings.TrimRight(envString("INTERVIEW_LLM_BASE_URL", defaultLLMBaseURL(llmProvider)), "/"),
			Endpoint:     envString("INTERVIEW_LLM_ENDPOINT", defaultLLMEndpoint(llmProvider)),
			APIKey:       envString("INTERVIEW_LLM_API_KEY", defaultLLMAPIKey(llmProvider)),
			Model:        envString("INTERVIEW_LLM_MODEL", "local-model"),
			SystemPrompt: envString("INTERVIEW_LLM_SYSTEM_PROMPT", defaultSystemPrompt),
			Temperature:  envFloat("INTERVIEW_LLM_TEMPERATURE", 0.2),
			Timeout:      envDuration("INTERVIEW_LLM_TIMEOUT", 90*time.Second),
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

// defaultLLMAPIKey looks for well-known provider-specific environment variables as
// a convenience so users do not have to set INTERVIEW_LLM_API_KEY when they already
// have a standard key variable exported in their shell.
func defaultLLMAPIKey(provider string) string {
	candidates := []string{"INTERVIEW_LLM_API_KEY"}
	switch provider {
	case "groq":
		candidates = append(candidates, "GROQ_API_KEY")
	case "openai":
		candidates = append(candidates, "OPENAI_API_KEY")
	case "anthropic", "claude":
		candidates = append(candidates, "ANTHROPIC_API_KEY")
	case "deepseek":
		candidates = append(candidates, "DEEPSEEK_API_KEY")
	case "zhipu", "glm":
		candidates = append(candidates, "ZHIPU_API_KEY")
	case "qianwen", "dashscope", "tongyi":
		candidates = append(candidates, "DASHSCOPE_API_KEY")
	case "moonshot", "kimi":
		candidates = append(candidates, "MOONSHOT_API_KEY")
	case "gemini", "google":
		candidates = append(candidates, "GEMINI_API_KEY")
	}
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

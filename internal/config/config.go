package config

import (
	"bufio"
	"bytes"
	"errors"
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
	Port     int
	WSURL    string
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

	sttPort := envInt("INTERVIEW_STT_PORT", 6006)
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
			Port:     sttPort,
			WSURL:    envString("INTERVIEW_STT_WS_URL", defaultSTTWSURL(sttPort)),
		},
		LLM: LLMConfig{
			Provider:     llmProvider,
			BaseURL:      strings.TrimRight(envString("INTERVIEW_LLM_BASE_URL", defaultLLMBaseURL(llmProvider)), "/"),
			Endpoint:     envString("INTERVIEW_LLM_ENDPOINT", "/chat/completions"),
			APIKey:       envString("INTERVIEW_LLM_API_KEY", envString("GROQ_API_KEY", "")),
			Model:        envString("INTERVIEW_LLM_MODEL", "local-model"),
			SystemPrompt: envString("INTERVIEW_LLM_SYSTEM_PROMPT", defaultSystemPrompt),
			Temperature:  envFloat("INTERVIEW_LLM_TEMPERATURE", 0.2),
			Timeout:      envDuration("INTERVIEW_LLM_TIMEOUT", 90*time.Second),
		},
	}
}

const defaultSystemPrompt = "You are a local interview copilot. Help the candidate answer interview questions in a concise, credible, first-person style. Use the candidate profile and recent context, do not invent experience, and reply in the same language as the question."

func defaultSTTWSURL(port int) string {
	return "ws://localhost:" + strconv.Itoa(port) + "/"
}

func defaultLLMBaseURL(provider string) string {
	if provider == "groq" {
		return "https://api.groq.com/openai/v1"
	}
	return "http://127.0.0.1:1234/v1"
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
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
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

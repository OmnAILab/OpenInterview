package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvDoesNotOverrideExistingEnv(t *testing.T) {
	t.Setenv("INTERVIEW_LLM_MODEL", "from-env")

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := []byte("INTERVIEW_LLM_MODEL=from-dotenv\nINTERVIEW_STT_PORT=7001\nINTERVIEW_STT_WS_URL=ws://127.0.0.1:7001/\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loadDotEnv(path)

	if got := os.Getenv("INTERVIEW_LLM_MODEL"); got != "from-env" {
		t.Fatalf("INTERVIEW_LLM_MODEL=%q, want from-env", got)
	}
	if got := os.Getenv("INTERVIEW_STT_PORT"); got != "7001" {
		t.Fatalf("INTERVIEW_STT_PORT=%q, want 7001", got)
	}
	if got := os.Getenv("INTERVIEW_STT_WS_URL"); got != "ws://127.0.0.1:7001/" {
		t.Fatalf("INTERVIEW_STT_WS_URL=%q, want ws://127.0.0.1:7001/", got)
	}
}

func TestParseDotEnvLine(t *testing.T) {
	cases := []struct {
		line     string
		wantKey  string
		wantVal  string
		wantOkay bool
	}{
		{line: "", wantOkay: false},
		{line: "# comment", wantOkay: false},
		{line: "INTERVIEW_LLM_BASE_URL=http://127.0.0.1:1234/v1", wantKey: "INTERVIEW_LLM_BASE_URL", wantVal: "http://127.0.0.1:1234/v1", wantOkay: true},
		{line: " export INTERVIEW_LLM_MODEL = qwen2.5 ", wantKey: "INTERVIEW_LLM_MODEL", wantVal: "qwen2.5", wantOkay: true},
		{line: "INTERVIEW_LLM_API_KEY=\"secret value\"", wantKey: "INTERVIEW_LLM_API_KEY", wantVal: "secret value", wantOkay: true},
		{line: "INTERVIEW_NOTE='candidate profile'", wantKey: "INTERVIEW_NOTE", wantVal: "candidate profile", wantOkay: true},
	}

	for _, tc := range cases {
		gotKey, gotVal, gotOK := parseDotEnvLine(tc.line)
		if gotKey != tc.wantKey || gotVal != tc.wantVal || gotOK != tc.wantOkay {
			t.Fatalf("parseDotEnvLine(%q) = (%q, %q, %v), want (%q, %q, %v)", tc.line, gotKey, gotVal, gotOK, tc.wantKey, tc.wantVal, tc.wantOkay)
		}
	}
}

func TestDefaultURLs(t *testing.T) {
	if got := defaultSTTWSURL(6006); got != "ws://127.0.0.1:6006/" {
		t.Fatalf("defaultSTTWSURL(6006) = %q", got)
	}
	if got := defaultLLMBaseURL("groq"); got != "https://api.groq.com/openai/v1" {
		t.Fatalf("defaultLLMBaseURL(groq) = %q", got)
	}
	if got := defaultLLMBaseURL("mock"); got != "http://127.0.0.1:1234/v1" {
		t.Fatalf("defaultLLMBaseURL(mock) = %q", got)
	}
}

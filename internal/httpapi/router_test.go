package httpapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"openinterview/internal/interview"
	"openinterview/internal/llm"
	"openinterview/internal/stt"
)

func TestHandleHealthIncludesRuntime(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	service := interview.NewService(interview.Config{
		MaxTurns:         5,
		ExpectedRate:     16000,
		ExpectedChannels: 1,
		ExpectedEncoding: "pcm16",
		MaxChunkBytes:    262144,
	}, stt.NewFactory(stt.Config{Provider: "mock"}, logger), llm.NewClient(llm.Config{Provider: "mock"}, logger), logger)

	handler := NewRouter(Config{
		Addr: ":8080",
		Runtime: RuntimeConfig{
			STT: RuntimeSTTConfig{
				Provider:                 "sherpa-websocket",
				DirectWebSocketAvailable: true,
				DirectWebSocketURL:       "ws://127.0.0.1:6006/",
			},
			LLM: RuntimeLLMConfig{
				Provider: "groq",
				Model:    "llama-3.3-70b-versatile",
			},
		},
	}, service, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload struct {
		Status  string         `json:"status"`
		Audio   map[string]any `json:"audio"`
		Runtime RuntimeConfig  `json:"runtime"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal health response: %v", err)
	}

	if payload.Status != "ok" {
		t.Fatalf("status payload = %q, want ok", payload.Status)
	}
	if got := payload.Runtime.STT.Provider; got != "sherpa-websocket" {
		t.Fatalf("runtime.stt.provider = %q, want sherpa-websocket", got)
	}
	if !payload.Runtime.STT.DirectWebSocketAvailable {
		t.Fatalf("runtime.stt.directWebSocketAvailable = false, want true")
	}
	if got := payload.Runtime.STT.DirectWebSocketURL; got != "ws://127.0.0.1:6006/" {
		t.Fatalf("runtime.stt.directWebSocketUrl = %q, want ws://127.0.0.1:6006/", got)
	}
	if got := payload.Runtime.LLM.Provider; got != "groq" {
		t.Fatalf("runtime.llm.provider = %q, want groq", got)
	}
	if got := payload.Runtime.LLM.Model; got != "llama-3.3-70b-versatile" {
		t.Fatalf("runtime.llm.model = %q, want llama-3.3-70b-versatile", got)
	}
}

func TestDeleteSessionRemovesSession(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	service := interview.NewService(interview.Config{
		MaxTurns:         5,
		ExpectedRate:     16000,
		ExpectedChannels: 1,
		ExpectedEncoding: "pcm16",
		MaxChunkBytes:    262144,
	}, stt.NewFactory(stt.Config{Provider: "mock"}, logger), llm.NewClient(llm.Config{Provider: "mock"}, logger), logger)

	handler := NewRouter(Config{
		Addr:    ":8080",
		Runtime: RuntimeConfig{},
	}, service, logger)

	created := service.CreateSession()

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/sessions/"+created.ID, nil)
	deleteRec := httptest.NewRecorder()
	handler.ServeHTTP(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleteRec.Code, http.StatusNoContent)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/sessions/"+created.ID, nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, want %d", getRec.Code, http.StatusNotFound)
	}
}

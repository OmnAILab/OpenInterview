package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"openinterview/internal/config"
	"openinterview/internal/httpapi"
	"openinterview/internal/interview"
	"openinterview/internal/llm"
	"openinterview/internal/stt"
)

func main() {
	logger := log.New(os.Stdout, "interviewd ", log.LstdFlags|log.Lmicroseconds)

	cfg := config.Load()
	sttFactory := stt.NewFactory(stt.Config{
		Provider: cfg.STT.Provider,
		Sherpa: &stt.SherpaConfig{
			WSURL: cfg.STT.Sherpa.WSURL,
		},
		Tencent: &stt.TencentConfig{
			WSURL:         cfg.STT.Tencent.WSURL,
			AppID:         cfg.STT.Tencent.AppID,
			SecretID:      cfg.STT.Tencent.SecretID,
			SecretKey:     cfg.STT.Tencent.SecretKey,
			EngineType:    cfg.STT.Tencent.EngineType,
			NeedVAD:       cfg.STT.Tencent.NeedVAD,
			NoEmptyResult: cfg.STT.Tencent.NoEmptyResult,
		},
	}, logger)
	llmClient := llm.NewClient(llm.Config{
		Provider:     cfg.LLM.Provider,
		BaseURL:      cfg.LLM.BaseURL,
		Endpoint:     cfg.LLM.Endpoint,
		APIKey:       cfg.LLM.APIKey,
		Model:        cfg.LLM.Model,
		SystemPrompt: cfg.LLM.SystemPrompt,
		Temperature:  cfg.LLM.Temperature,
		Timeout:      cfg.LLM.Timeout,
	}, logger)

	service := interview.NewService(interview.Config{
		MaxTurns:         cfg.Session.MaxTurns,
		ExpectedEncoding: cfg.Audio.Encoding,
		ExpectedChannels: cfg.Audio.Channels,
		ExpectedRate:     cfg.Audio.SampleRate,
		MaxChunkBytes:    cfg.Audio.MaxChunkBytes,
	}, sttFactory, llmClient, logger)
	directSTTWSURL := directSTTWebSocketURL(cfg.STT.Provider)

	handler := httpapi.NewRouter(httpapi.Config{
		Addr: cfg.Server.Addr,
		Runtime: httpapi.RuntimeConfig{
			STT: httpapi.RuntimeSTTConfig{
				Provider:                 cfg.STT.Provider,
				DirectWebSocketAvailable: directSTTWSURL != "",
				DirectWebSocketURL:       directSTTWSURL,
			},
			LLM: httpapi.RuntimeLLMConfig{
				Provider: cfg.LLM.Provider,
				Model:    cfg.LLM.Model,
			},
		},
	}, service, logger)

	server := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Printf("listening on %s", cfg.Server.Addr)
		logger.Printf("health check: http://localhost%s/api/healthz", cfg.Server.Addr)
		logger.Printf("create a session with POST /api/sessions, then connect SSE on /api/sessions/{id}/events")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server failed: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("shutdown failed: %v", err)
	}
}

func directSTTWebSocketURL(provider string) string {
	cfg := config.Load()
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "sherpa", "sherpa-websocket", "sherpa_onnx", "sherpa-onnx":
		return strings.TrimSpace(cfg.STT.Sherpa.WSURL)
	case "tencent", "tencent-asr":
		return strings.TrimSpace(cfg.STT.Tencent.WSURL)
	default:
		return ""
	}
}

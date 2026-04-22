package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"openinterview/internal/config"
	"openinterview/internal/knowledge"
)

func main() {
	logger := log.New(os.Stdout, "kbd ", log.LstdFlags|log.Lmicroseconds)
	cfg := config.Load()

	if strings.TrimSpace(cfg.Knowledge.Path) == "" {
		logger.Fatal("knowledge path is empty: set INTERVIEW_KNOWLEDGE_LOCAL_PATH")
	}
	if strings.TrimSpace(cfg.Knowledge.EmbeddingEndpoint) == "" {
		logger.Fatal("embedding endpoint is empty: set INTERVIEW_KNOWLEDGE_EMBEDDING_ENDPOINT")
	}

	client := knowledge.NewClient(knowledge.Config{
		Path:              cfg.Knowledge.Path,
		MaxResults:        cfg.Knowledge.MaxResults,
		EmbeddingEndpoint: cfg.Knowledge.EmbeddingEndpoint,
		EmbeddingAPIKey:   cfg.Knowledge.APIKey,
		EmbeddingModel:    cfg.Knowledge.EmbeddingModel,
		Timeout:           cfg.Knowledge.Timeout,
	}, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"path":   cfg.Knowledge.Path,
			"mode":   "local-vector",
		})
	})
	mux.HandleFunc("POST /search", func(w http.ResponseWriter, req *http.Request) {
		query, limit, err := readSearchRequest(req, cfg.Knowledge.MaxResults)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}

		results, err := client.Retrieve(req.Context(), query)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if limit > 0 && len(results) > limit {
			results = results[:limit]
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": results})
	})

	server := &http.Server{
		Addr:              cfg.Knowledge.LocalAddr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Printf("listening on %s", cfg.Knowledge.LocalAddr)
		logger.Printf("health check: http://localhost%s/api/healthz", cfg.Knowledge.LocalAddr)
		logger.Printf("search endpoint: http://localhost%s/search", cfg.Knowledge.LocalAddr)
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

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func readSearchRequest(req *http.Request, fallbackLimit int) (query string, limit int, err error) {
	defer req.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(io.LimitReader(req.Body, 1<<20)).Decode(&payload); err != nil {
		return "", 0, err
	}

	query = firstNonEmptyString(payload, "query", "question", "text")
	if strings.TrimSpace(query) == "" {
		return "", 0, errors.New("query is empty")
	}

	limit = firstPositiveInt(payload, "top_k", "topK", "limit")
	if limit <= 0 {
		limit = fallbackLimit
	}
	return query, limit, nil
}

func firstNonEmptyString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok {
			continue
		}
		value, ok := raw.(string)
		if ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstPositiveInt(payload map[string]any, keys ...string) int {
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok {
			continue
		}
		switch value := raw.(type) {
		case float64:
			if value > 0 {
				return int(value)
			}
		case int:
			if value > 0 {
				return value
			}
		}
	}
	return 0
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

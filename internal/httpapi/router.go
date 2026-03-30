package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"openinterview/internal/interview"
	"openinterview/internal/stt"
)

type Config struct {
	Addr string
}

type Router struct {
	cfg     Config
	service *interview.Service
	logger  *log.Logger
}

func NewRouter(cfg Config, service *interview.Service, logger *log.Logger) http.Handler {
	router := &Router{
		cfg:     cfg,
		service: service,
		logger:  logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", router.handleHealth)
	mux.HandleFunc("POST /api/sessions", router.handleCreateSession)
	mux.HandleFunc("GET /api/sessions/{sessionID}", router.handleGetSession)
	mux.HandleFunc("GET /api/sessions/{sessionID}/events", router.handleEvents)
	mux.HandleFunc("GET /api/sessions/{sessionID}/profile", router.handleGetProfile)
	mux.HandleFunc("PUT /api/sessions/{sessionID}/profile", router.handlePutProfile)
	mux.HandleFunc("POST /api/sessions/{sessionID}/listen/start", router.handleStartListening)
	mux.HandleFunc("POST /api/sessions/{sessionID}/listen/stop", router.handleStopListening)
	mux.HandleFunc("POST /api/sessions/{sessionID}/audio", router.handleAudioChunk)
	mux.HandleFunc("POST /api/sessions/{sessionID}/reset", router.handleReset)
	mux.HandleFunc("POST /api/sessions/{sessionID}/debug/transcript", router.handleDebugTranscript)
	mux.HandleFunc("POST /api/sessions/{sessionID}/ask", router.handleAskQuestion)

	return router.withMiddleware(mux)
}

func (r *Router) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, req)
		r.logger.Printf("%s %s %s", req.Method, req.URL.Path, time.Since(start))
	})
}

func (r *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"audio": map[string]any{
			"sampleRate": r.service.ExpectedSampleRate(),
			"channels":   r.service.ExpectedChannels(),
			"encoding":   r.service.ExpectedEncoding(),
		},
	})
}

func (r *Router) handleCreateSession(w http.ResponseWriter, _ *http.Request) {
	snapshot := r.service.CreateSession()
	writeJSON(w, http.StatusCreated, snapshot)
}

func (r *Router) handleGetSession(w http.ResponseWriter, req *http.Request) {
	snapshot, err := r.service.GetSnapshot(req.PathValue("sessionID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (r *Router) handleEvents(w http.ResponseWriter, req *http.Request) {
	sessionID := req.PathValue("sessionID")
	snapshot, events, unsubscribe, err := r.service.Subscribe(sessionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	if err := writeSSE(w, "snapshot", snapshot); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-req.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := io.WriteString(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeSSE(w, event.Type, event); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (r *Router) handleGetProfile(w http.ResponseWriter, req *http.Request) {
	profile, err := r.service.GetProfile(req.PathValue("sessionID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (r *Router) handlePutProfile(w http.ResponseWriter, req *http.Request) {
	var profile interview.CandidateProfile
	if err := decodeJSON(req, &profile); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	snapshot, err := r.service.UpdateProfile(req.PathValue("sessionID"), profile)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (r *Router) handleStartListening(w http.ResponseWriter, req *http.Request) {
	snapshot, err := r.service.StartListening(req.Context(), req.PathValue("sessionID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (r *Router) handleStopListening(w http.ResponseWriter, req *http.Request) {
	snapshot, err := r.service.StopListening(req.PathValue("sessionID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (r *Router) handleAudioChunk(w http.ResponseWriter, req *http.Request) {
	chunk, err := readAudioChunk(req, r.service.MaxChunkBytes())
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	snapshot, err := r.service.IngestAudio(req.Context(), req.PathValue("sessionID"), chunk)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, snapshot)
}

func (r *Router) handleReset(w http.ResponseWriter, req *http.Request) {
	snapshot, err := r.service.Reset(req.Context(), req.PathValue("sessionID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (r *Router) handleDebugTranscript(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Kind     string `json:"kind"`
		Text     string `json:"text"`
		Endpoint bool   `json:"endpoint"`
	}
	if err := decodeJSON(req, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	kind := strings.ToLower(strings.TrimSpace(body.Kind))
	if kind == "" {
		kind = string(stt.EventFinal)
	}

	event := stt.TranscriptEvent{
		Kind:     stt.EventKind(kind),
		Text:     body.Text,
		Endpoint: body.Endpoint,
	}

	snapshot, err := r.service.InjectTranscript(req.Context(), req.PathValue("sessionID"), event)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, snapshot)
}

func (r *Router) handleAskQuestion(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Question string `json:"question"`
	}
	if err := decodeJSON(req, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	snapshot, err := r.service.AskQuestion(req.PathValue("sessionID"), body.Question)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, snapshot)
}

func decodeJSON(req *http.Request, target any) error {
	defer req.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(req.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func readAudioChunk(req *http.Request, maxBytes int) (stt.AudioChunk, error) {
	sampleRate, err := readIntQuery(req, "sampleRate", 16000)
	if err != nil {
		return stt.AudioChunk{}, err
	}
	channels, err := readIntQuery(req, "channels", 1)
	if err != nil {
		return stt.AudioChunk{}, err
	}

	encoding := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("encoding")))
	if encoding == "" {
		encoding = "pcm16"
	}

	body, err := io.ReadAll(io.LimitReader(req.Body, int64(maxBytes)+1))
	if err != nil {
		return stt.AudioChunk{}, fmt.Errorf("read audio chunk: %w", err)
	}
	if len(body) == 0 {
		return stt.AudioChunk{}, errors.New("audio body is empty")
	}
	if len(body) > maxBytes {
		return stt.AudioChunk{}, fmt.Errorf("audio chunk exceeds %d bytes", maxBytes)
	}

	return stt.AudioChunk{
		Data:       body,
		SampleRate: sampleRate,
		Channels:   channels,
		Encoding:   encoding,
		At:         time.Now(),
	}, nil
}

func readIntQuery(req *http.Request, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(req.URL.Query().Get(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return value, nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, interview.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, interview.ErrInvalidAudioFormat), errors.Is(err, interview.ErrBadRequest):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	case errors.Is(err, interview.ErrSessionNotListening):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSSE(w io.Writer, event string, payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, encoded)
	return err
}

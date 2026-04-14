package interview

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"

	"openinterview/internal/llm"
	"openinterview/internal/stt"
)

type Config struct {
	MaxTurns         int
	ExpectedRate     int
	ExpectedChannels int
	ExpectedEncoding string
	MaxChunkBytes    int
}

type Service struct {
	cfg        Config
	sttFactory stt.Factory
	llm        llm.Client
	logger     *log.Logger

	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewService(cfg Config, sttFactory stt.Factory, llmClient llm.Client, logger *log.Logger) *Service {
	if cfg.MaxTurns <= 0 {
		cfg.MaxTurns = 5
	}
	return &Service{
		cfg:        cfg,
		sttFactory: sttFactory,
		llm:        llmClient,
		logger:     logger,
		sessions:   make(map[string]*Session),
	}
}

func (s *Service) CreateSession() Snapshot {
	id := randomID()
	session := newSession(id, s.cfg, s.sttFactory, s.llm, s.logger)

	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()

	return session.Snapshot()
}

func (s *Service) DeleteSession(sessionID string) error {
	s.mu.Lock()
	session, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return ErrNotFound
	}
	delete(s.sessions, sessionID)
	s.mu.Unlock()

	session.Close()
	return nil
}

func (s *Service) GetSnapshot(sessionID string) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.Snapshot(), nil
}

func (s *Service) GetProfile(sessionID string) (CandidateProfile, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return CandidateProfile{}, err
	}
	return session.Profile(), nil
}

func (s *Service) UpdateProfile(sessionID string, profile CandidateProfile) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.UpdateProfile(profile), nil
}

func (s *Service) Subscribe(sessionID string) (Snapshot, <-chan Event, func(), error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, nil, nil, err
	}
	snapshot := session.Snapshot()
	events, unsubscribe := session.Subscribe()
	return snapshot, events, unsubscribe, nil
}

func (s *Service) StartListening(ctx context.Context, sessionID string) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.StartListening(ctx)
}

func (s *Service) StopListening(sessionID string) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.StopListening()
}

func (s *Service) IngestAudio(ctx context.Context, sessionID string, chunk stt.AudioChunk) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.IngestAudio(ctx, chunk)
}

func (s *Service) InjectTranscript(ctx context.Context, sessionID string, event stt.TranscriptEvent) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.InjectTranscript(ctx, event)
}

func (s *Service) AskQuestion(sessionID string, question string) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.AskQuestion(question)
}

func (s *Service) SubmitTextSegment(sessionID string, stop int) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.SubmitTextSegment(stop)
}

func (s *Service) AddTextStopMarker(sessionID string, stop int) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.AddTextStopMarker(stop)
}

func (s *Service) Reset(ctx context.Context, sessionID string) (Snapshot, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	return session.Reset(ctx)
}

func (s *Service) ExpectedSampleRate() int {
	return s.cfg.ExpectedRate
}

func (s *Service) ExpectedChannels() int {
	return s.cfg.ExpectedChannels
}

func (s *Service) ExpectedEncoding() string {
	return s.cfg.ExpectedEncoding
}

func (s *Service) MaxChunkBytes() int {
	return s.cfg.MaxChunkBytes
}

func (s *Service) session(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return session, nil
}

func randomID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("random id: %v", err))
	}
	return hex.EncodeToString(buf)
}

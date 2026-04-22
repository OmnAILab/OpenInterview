package interview

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"openinterview/internal/knowledge"
	"openinterview/internal/llm"
	"openinterview/internal/stt"
	"openinterview/internal/textdeal"
)

type Session struct {
	id         string
	cfg        Config
	sttFactory stt.Factory
	llm        llm.Client
	knowledge  knowledge.Client
	logger     *log.Logger
	broker     *broker

	mu                sync.RWMutex
	listening         bool
	answerInProgress  bool
	partialTranscript string
	finalTranscripts  []string
	textDeal          textdeal.Buffer
	currentQuestion   string
	currentAnswer     string
	profile           CandidateProfile
	history           []Turn
	knowledgeHits     []KnowledgeHit
	audio             AudioStats
	lastError         string
	stream            stt.Stream
	sttGeneration     int64
	answerCancel      context.CancelFunc
	sequence          int64
	activeAnswerID    int64
}

type sessionSTTSink struct {
	session    *Session
	generation int64
}

func (s sessionSTTSink) HandleTranscriptEvent(ctx context.Context, event stt.TranscriptEvent) {
	s.session.handleTranscriptForGeneration(ctx, s.generation, event)
}

func (s sessionSTTSink) HandleSTTError(ctx context.Context, err error) {
	s.session.handleSTTErrorForGeneration(ctx, s.generation, err)
}

func newSession(id string, cfg Config, sttFactory stt.Factory, llmClient llm.Client, knowledgeClient knowledge.Client, logger *log.Logger) *Session {
	return &Session{
		id:         id,
		cfg:        cfg,
		sttFactory: sttFactory,
		llm:        llmClient,
		knowledge:  knowledgeClient,
		logger:     logger,
		broker:     newBroker(),
	}
}

func (s *Session) Subscribe() (<-chan Event, func()) {
	return s.broker.subscribe()
}

func (s *Session) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshotLocked()
}

func (s *Session) Profile() CandidateProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profile
}

func (s *Session) UpdateProfile(profile CandidateProfile) Snapshot {
	s.mu.Lock()
	s.profile = normalizeProfile(profile)
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	s.publish("profile.updated", snapshot.Profile)
	s.log("profile", "candidate profile updated", map[string]any{
		"targetRole": snapshot.Profile.TargetRole,
		"techStack":  snapshot.Profile.TechStack,
	})
	return snapshot
}

func (s *Session) StartListening(ctx context.Context) (Snapshot, error) {
	s.mu.Lock()
	if s.listening {
		snapshot := s.snapshotLocked()
		s.mu.Unlock()
		return snapshot, nil
	}
	if err := s.ensureStreamLocked(); err != nil {
		s.lastError = err.Error()
		s.mu.Unlock()
		s.publish("error", map[string]any{"message": err.Error()})
		return Snapshot{}, err
	}
	s.listening = true
	s.lastError = ""
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	s.publish("state", map[string]any{"listening": true, "answerInProgress": snapshot.AnswerInProgress})
	s.log("audio", "listening started", map[string]any{
		"sampleRate": s.cfg.ExpectedRate,
		"channels":   s.cfg.ExpectedChannels,
		"encoding":   s.cfg.ExpectedEncoding,
	})

	select {
	case <-ctx.Done():
	default:
	}

	return snapshot, nil
}

func (s *Session) StopListening() (Snapshot, error) {
	s.mu.Lock()
	if !s.listening {
		snapshot := s.snapshotLocked()
		s.mu.Unlock()
		return snapshot, nil
	}
	s.listening = false
	stream := s.detachStreamLocked()
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	if stream != nil {
		_ = stream.Close()
	}

	s.publish("state", map[string]any{"listening": false, "answerInProgress": snapshot.AnswerInProgress})
	s.log("audio", "listening stopped", nil)
	return snapshot, nil
}

func (s *Session) IngestAudio(ctx context.Context, chunk stt.AudioChunk) (Snapshot, error) {
	if err := s.validateAudio(chunk); err != nil {
		return Snapshot{}, err
	}

	s.mu.Lock()
	if !s.listening {
		s.mu.Unlock()
		return Snapshot{}, ErrSessionNotListening
	}
	if err := s.ensureStreamLocked(); err != nil {
		s.lastError = err.Error()
		snapshot := s.snapshotLocked()
		s.mu.Unlock()
		s.publish("error", map[string]any{"message": err.Error()})
		return snapshot, err
	}

	stream := s.stream
	now := chunk.At
	s.audio.Chunks++
	s.audio.Bytes += int64(len(chunk.Data))
	s.audio.LastChunkAt = &now
	shouldLog := s.audio.Chunks == 1 || s.audio.Chunks%25 == 0
	snapshot := s.snapshotLocked()
	audioStats := s.audio
	s.mu.Unlock()

	if shouldLog {
		s.log("audio", "audio chunk accepted", map[string]any{
			"chunks": audioStats.Chunks,
			"bytes":  audioStats.Bytes,
		})
	}

	if err := stream.Push(ctx, chunk); err != nil {
		s.setError(err)
		return s.Snapshot(), err
	}
	return snapshot, nil
}

func (s *Session) InjectTranscript(_ context.Context, event stt.TranscriptEvent) (Snapshot, error) {
	if strings.TrimSpace(event.Text) == "" {
		return Snapshot{}, fmt.Errorf("%w: transcript text is empty", ErrBadRequest)
	}
	s.handleTranscript(event)
	return s.Snapshot(), nil
}

func (s *Session) AskQuestion(question string) (Snapshot, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return Snapshot{}, fmt.Errorf("%w: question is empty", ErrBadRequest)
	}

	s.mu.Lock()
	s.currentQuestion = question
	s.mu.Unlock()

	s.publish("question.manual", map[string]any{"text": question})
	s.log("question", "manual question submitted", map[string]any{"text": question})
	s.startAnswer(question)
	return s.Snapshot(), nil
}

func (s *Session) SubmitTextSegment(stop int) (Snapshot, error) {
	s.mu.Lock()
	segment, textDealSnapshot, err := s.textDeal.SubmitStop(stop)
	if err != nil {
		s.mu.Unlock()
		return Snapshot{}, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	s.currentQuestion = segment.Text
	s.lastError = ""
	s.mu.Unlock()

	s.publish("textdeal.updated", textDealSnapshot)
	s.publish("question.detected", map[string]any{
		"text":   segment.Text,
		"source": "textdeal",
		"start":  segment.Start,
		"end":    segment.End,
	})
	s.log("textdeal", "segment submitted to llm", map[string]any{
		"text":  segment.Text,
		"start": segment.Start,
		"end":   segment.End,
	})

	s.startAnswer(segment.Text)
	return s.Snapshot(), nil
}

func (s *Session) AddTextStopMarker(stop int) (Snapshot, error) {
	s.mu.Lock()
	textDealSnapshot, err := s.textDeal.AddStopMarker(stop)
	if err != nil {
		s.mu.Unlock()
		return Snapshot{}, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	s.lastError = ""
	s.mu.Unlock()

	s.publish("textdeal.updated", textDealSnapshot)
	s.log("textdeal", "stop marker added", map[string]any{"stop": stop})
	return s.Snapshot(), nil
}

func (s *Session) Reset(ctx context.Context) (Snapshot, error) {
	s.mu.Lock()
	cancel := s.answerCancel
	s.answerCancel = nil
	s.activeAnswerID++
	stream := s.detachStreamLocked()
	keepListening := s.listening
	s.answerInProgress = false
	s.partialTranscript = ""
	s.finalTranscripts = nil
	s.textDeal.Reset()
	s.currentQuestion = ""
	s.currentAnswer = ""
	s.history = nil
	s.knowledgeHits = nil
	s.audio = AudioStats{}
	s.lastError = ""
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if stream != nil {
		_ = stream.Close()
	}
	if keepListening {
		s.mu.Lock()
		if err := s.ensureStreamLocked(); err != nil {
			s.lastError = err.Error()
		}
		s.mu.Unlock()
	}

	select {
	case <-ctx.Done():
	default:
	}

	snapshot := s.Snapshot()
	s.publish("session.reset", map[string]any{"sessionId": s.id})
	s.log("session", "session context reset", nil)
	return snapshot, nil
}

func (s *Session) Close() {
	s.mu.Lock()
	cancel := s.answerCancel
	s.answerCancel = nil
	s.activeAnswerID++
	s.answerInProgress = false
	s.listening = false
	stream := s.detachStreamLocked()
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if stream != nil {
		_ = stream.Close()
	}

	s.broker.close()
}

func (s *Session) HandleTranscriptEvent(ctx context.Context, event stt.TranscriptEvent) {
	_ = ctx
	s.handleTranscript(event)
}

func (s *Session) HandleSTTError(ctx context.Context, err error) {
	s.handleSTTError(ctx, err)
}

func (s *Session) handleTranscriptForGeneration(_ context.Context, generation int64, event stt.TranscriptEvent) {
	s.mu.RLock()
	currentGeneration := s.sttGeneration
	s.mu.RUnlock()

	if generation != currentGeneration {
		return
	}
	s.handleTranscript(event)
}

func (s *Session) handleSTTErrorForGeneration(ctx context.Context, generation int64, err error) {
	if err == nil {
		return
	}

	s.mu.Lock()
	if generation != s.sttGeneration {
		s.mu.Unlock()
		return
	}
	s.stream = nil
	s.sttGeneration++
	s.listening = false
	s.lastError = err.Error()
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	s.publish("error", map[string]any{"message": err.Error()})
	s.publish("state", map[string]any{"listening": snapshot.Listening, "answerInProgress": snapshot.AnswerInProgress})
	s.log("stt", "stt stream failed", map[string]any{"error": err.Error()})

	select {
	case <-ctx.Done():
	default:
	}
}

func (s *Session) handleSTTError(_ context.Context, err error) {
	s.setError(err)
}

func (s *Session) handleTranscript(event stt.TranscriptEvent) {
	normalized := strings.TrimSpace(event.Text)
	if normalized == "" {
		return
	}

	switch event.Kind {
	case stt.EventPartial:
		s.mu.Lock()
		s.partialTranscript = normalized
		s.lastError = ""
		s.mu.Unlock()
		s.publish("stt.partial", map[string]any{"text": normalized})
	case stt.EventFinal:
		var textDealSnapshot textdeal.Snapshot

		s.mu.Lock()
		s.partialTranscript = ""
		s.finalTranscripts = appendAndTrim(s.finalTranscripts, normalized, 20)
		textDealSnapshot = s.textDeal.AppendStable(normalized)
		s.lastError = ""
		s.mu.Unlock()

		s.publish("stt.final", map[string]any{
			"text":     normalized,
			"endpoint": event.Endpoint,
		})
		s.publish("textdeal.updated", textDealSnapshot)
		s.log("stt", "final transcript received", map[string]any{"text": normalized})
		s.log("textdeal", "stable transcript appended", map[string]any{
			"text":             normalized,
			"stableTextLength": len([]rune(textDealSnapshot.StableText)),
		})
	default:
		s.log("stt", "unsupported transcript event kind", map[string]any{"kind": event.Kind})
	}
}

func (s *Session) startAnswer(question string) {
	profile := s.profileSnapshot()
	history := s.historySnapshot()

	ctx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	oldCancel := s.answerCancel
	s.activeAnswerID++
	answerID := s.activeAnswerID
	s.answerCancel = cancel
	s.answerInProgress = true
	s.currentAnswer = ""
	s.knowledgeHits = nil
	s.lastError = ""
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	if oldCancel != nil {
		oldCancel()
		s.publish("llm.interrupted", map[string]any{"question": question})
		s.log("llm", "previous answer interrupted by a new question", nil)
	}

	s.publish("state", map[string]any{"listening": snapshot.Listening, "answerInProgress": true})
	s.publish("llm.started", map[string]any{"question": question})
	s.log("llm", "starting answer generation", map[string]any{"question": question})

	go func() {
		retrieved := s.retrieveKnowledge(ctx, question)
		s.setKnowledgeHits(answerID, retrieved)

		request := llm.Request{
			Question: question,
			Messages: buildMessages(profile, history, question, retrieved),
		}

		answer, err := s.llm.StreamAnswer(ctx, request, func(token string) {
			s.appendAnswerToken(answerID, token)
		})
		s.finishAnswer(answerID, question, answer, err)
	}()
}

func (s *Session) finishAnswer(answerID int64, question, answer string, err error) {
	s.mu.Lock()
	if answerID != s.activeAnswerID {
		s.mu.Unlock()
		return
	}

	s.answerInProgress = false
	s.answerCancel = nil
	if err == nil {
		s.currentAnswer = answer
		s.history = append(s.history, Turn{
			Question:  question,
			Answer:    answer,
			CreatedAt: time.Now(),
		})
		if len(s.history) > s.cfg.MaxTurns {
			s.history = append([]Turn(nil), s.history[len(s.history)-s.cfg.MaxTurns:]...)
		}
		s.lastError = ""
	} else if !errors.Is(err, context.Canceled) {
		s.lastError = err.Error()
	}
	snapshot := s.snapshotLocked()
	lastError := s.lastError
	s.mu.Unlock()

	if err == nil {
		s.publish("llm.done", map[string]any{"question": question, "answer": answer})
		s.log("llm", "answer generation finished", map[string]any{"question": question})
	} else if errors.Is(err, context.Canceled) {
		s.publish("llm.cancelled", map[string]any{"question": question})
		return
	} else {
		s.publish("error", map[string]any{"message": lastError})
		s.log("llm", "answer generation failed", map[string]any{"error": lastError})
	}

	s.publish("state", map[string]any{"listening": snapshot.Listening, "answerInProgress": false})
}

func (s *Session) appendAnswerToken(answerID int64, token string) {
	if token == "" {
		return
	}

	s.mu.Lock()
	if answerID != s.activeAnswerID {
		s.mu.Unlock()
		return
	}
	s.currentAnswer += token
	s.mu.Unlock()

	s.publish("llm.token", map[string]any{"token": token})
}

func (s *Session) retrieveKnowledge(ctx context.Context, question string) []knowledge.Document {
	if s.knowledge == nil {
		return nil
	}

	docs, err := s.knowledge.Retrieve(ctx, question)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		s.log("knowledge", "knowledge retrieval failed", map[string]any{"error": err.Error()})
		return nil
	}
	if len(docs) == 0 {
		s.log("knowledge", "knowledge retrieval returned no results", map[string]any{"question": question})
		return nil
	}

	s.log("knowledge", "knowledge retrieved", map[string]any{
		"question": question,
		"results":  len(docs),
	})
	return docs
}

func (s *Session) setKnowledgeHits(answerID int64, docs []knowledge.Document) {
	hits := make([]KnowledgeHit, 0, len(docs))
	for _, doc := range docs {
		hits = append(hits, KnowledgeHit{
			Title:   doc.Title,
			Content: doc.Content,
			Path:    doc.Path,
			Score:   doc.Score,
		})
	}

	s.mu.Lock()
	if answerID != s.activeAnswerID {
		s.mu.Unlock()
		return
	}
	s.knowledgeHits = hits
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	s.publish("knowledge.retrieved", map[string]any{
		"question": snapshot.CurrentQuestion,
		"results":  snapshot.Knowledge,
	})
}

func (s *Session) profileSnapshot() CandidateProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profile
}

func (s *Session) historySnapshot() []Turn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Turn(nil), s.history...)
}

func (s *Session) ensureStreamLocked() error {
	if s.stream != nil {
		return nil
	}

	generation := s.sttGeneration + 1
	stream, err := s.sttFactory.NewStream(s.id, sessionSTTSink{
		session:    s,
		generation: generation,
	})
	if err != nil {
		return err
	}

	s.stream = stream
	s.sttGeneration = generation
	return nil
}

func (s *Session) detachStreamLocked() stt.Stream {
	stream := s.stream
	s.stream = nil
	s.sttGeneration++
	return stream
}

func (s *Session) validateAudio(chunk stt.AudioChunk) error {
	if chunk.SampleRate != s.cfg.ExpectedRate || chunk.Channels != s.cfg.ExpectedChannels || strings.ToLower(chunk.Encoding) != s.cfg.ExpectedEncoding {
		return fmt.Errorf("%w: expected %dHz/%dch/%s, got %dHz/%dch/%s",
			ErrInvalidAudioFormat,
			s.cfg.ExpectedRate,
			s.cfg.ExpectedChannels,
			s.cfg.ExpectedEncoding,
			chunk.SampleRate,
			chunk.Channels,
			chunk.Encoding,
		)
	}
	return nil
}

func (s *Session) publish(kind string, payload any) {
	s.mu.Lock()
	s.sequence++
	event := Event{
		Type:      kind,
		Sequence:  s.sequence,
		SessionID: s.id,
		Time:      time.Now(),
		Data:      payload,
	}
	s.mu.Unlock()
	s.broker.publish(event)
}

func (s *Session) log(scope, message string, fields map[string]any) {
	data := map[string]any{
		"scope":   scope,
		"message": message,
	}
	for key, value := range fields {
		data[key] = value
	}
	s.publish("log", data)
	if s.logger != nil {
		s.logger.Printf("[%s] %s", scope, message)
	}
}

func (s *Session) setError(err error) {
	s.mu.Lock()
	s.lastError = err.Error()
	s.mu.Unlock()
	s.publish("error", map[string]any{"message": err.Error()})
}

func (s *Session) snapshotLocked() Snapshot {
	history := append([]Turn(nil), s.history...)
	finals := append([]string(nil), s.finalTranscripts...)
	knowledgeHits := append([]KnowledgeHit(nil), s.knowledgeHits...)
	return Snapshot{
		ID:                s.id,
		Listening:         s.listening,
		AnswerInProgress:  s.answerInProgress,
		PartialTranscript: s.partialTranscript,
		FinalTranscripts:  finals,
		TextDeal:          s.textDeal.Snapshot(),
		CurrentQuestion:   s.currentQuestion,
		CurrentAnswer:     s.currentAnswer,
		Profile:           s.profile,
		History:           history,
		Knowledge:         knowledgeHits,
		Audio:             s.audio,
		LastError:         s.lastError,
	}
}

func appendAndTrim(values []string, value string, limit int) []string {
	values = append(values, value)
	if len(values) > limit {
		values = append([]string(nil), values[len(values)-limit:]...)
	}
	return values
}

func normalizeProfile(profile CandidateProfile) CandidateProfile {
	profile.TargetRole = strings.TrimSpace(profile.TargetRole)
	profile.ProjectSummary = strings.TrimSpace(profile.ProjectSummary)
	profile.Strengths = strings.TrimSpace(profile.Strengths)
	profile.AnswerStyle = strings.TrimSpace(profile.AnswerStyle)
	profile.TechStack = compactStrings(profile.TechStack)
	return profile
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func buildMessages(profile CandidateProfile, history []Turn, question string, docs []knowledge.Document) []llm.Message {
	var builder strings.Builder
	builder.WriteString("Candidate profile:\n")
	if profile.TargetRole != "" {
		builder.WriteString("Target role: ")
		builder.WriteString(profile.TargetRole)
		builder.WriteString("\n")
	}
	if len(profile.TechStack) > 0 {
		builder.WriteString("Tech stack: ")
		builder.WriteString(strings.Join(profile.TechStack, ", "))
		builder.WriteString("\n")
	}
	if profile.ProjectSummary != "" {
		builder.WriteString("Project summary: ")
		builder.WriteString(profile.ProjectSummary)
		builder.WriteString("\n")
	}
	if profile.Strengths != "" {
		builder.WriteString("Strengths: ")
		builder.WriteString(profile.Strengths)
		builder.WriteString("\n")
	}
	if profile.AnswerStyle != "" {
		builder.WriteString("Answer style: ")
		builder.WriteString(profile.AnswerStyle)
		builder.WriteString("\n")
	}
	if len(docs) > 0 {
		builder.WriteString("\nRetrieved knowledge:\n")
		for i, doc := range docs {
			builder.WriteString(fmt.Sprintf("[%d] %s", i+1, doc.Title))
			if doc.Path != "" {
				builder.WriteString(" (")
				builder.WriteString(doc.Path)
				builder.WriteString(")")
			}
			if doc.Score > 0 {
				builder.WriteString(fmt.Sprintf(" score=%.3f", doc.Score))
			}
			builder.WriteString("\n")
			builder.WriteString(doc.Content)
			builder.WriteString("\n\n")
		}
		builder.WriteString("Use retrieved knowledge only when it is directly relevant. Do not invent first-person experience that is not supported by the candidate profile.\n")
	}

	messages := []llm.Message{
		{
			Role:    llm.RoleSystem,
			Content: strings.TrimSpace(builder.String()),
		},
	}

	for _, turn := range history {
		messages = append(messages,
			llm.Message{Role: llm.RoleUser, Content: turn.Question},
			llm.Message{Role: llm.RoleAssistant, Content: turn.Answer},
		)
	}

	messages = append(messages, llm.Message{
		Role:    llm.RoleUser,
		Content: question,
	})
	return messages
}

package interview

import (
	"time"

	"openinterview/internal/textdeal"
)

type CandidateProfile struct {
	TargetRole     string   `json:"targetRole"`
	TechStack      []string `json:"techStack"`
	ProjectSummary string   `json:"projectSummary"`
	Strengths      string   `json:"strengths"`
	AnswerStyle    string   `json:"answerStyle"`
}

type Turn struct {
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"createdAt"`
}

type AudioStats struct {
	Chunks      int64      `json:"chunks"`
	Bytes       int64      `json:"bytes"`
	LastChunkAt *time.Time `json:"lastChunkAt,omitempty"`
}

type KnowledgeHit struct {
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Path    string  `json:"path,omitempty"`
	Score   float64 `json:"score,omitempty"`
}

type Snapshot struct {
	ID                string            `json:"id"`
	Listening         bool              `json:"listening"`
	AnswerInProgress  bool              `json:"answerInProgress"`
	PartialTranscript string            `json:"partialTranscript"`
	FinalTranscripts  []string          `json:"finalTranscripts"`
	TextDeal          textdeal.Snapshot `json:"textDeal"`
	CurrentQuestion   string            `json:"currentQuestion"`
	CurrentAnswer     string            `json:"currentAnswer"`
	Profile           CandidateProfile  `json:"profile"`
	History           []Turn            `json:"history"`
	Knowledge         []KnowledgeHit    `json:"knowledge"`
	Audio             AudioStats        `json:"audio"`
	LastError         string            `json:"lastError,omitempty"`
}

type Event struct {
	Type      string    `json:"type"`
	Sequence  int64     `json:"sequence"`
	SessionID string    `json:"sessionId"`
	Time      time.Time `json:"time"`
	Data      any       `json:"data,omitempty"`
}

package interview

import "errors"

var (
	ErrNotFound            = errors.New("session not found")
	ErrBadRequest          = errors.New("bad request")
	ErrInvalidAudioFormat  = errors.New("invalid audio format")
	ErrSessionNotListening = errors.New("session is not listening")
)

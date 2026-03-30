package stt

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type sherpaWebSocketFactory struct {
	wsURL  string
	logger *log.Logger
}

func (f *sherpaWebSocketFactory) NewStream(sessionID string, sink Sink) (Stream, error) {
	if strings.TrimSpace(f.wsURL) == "" {
		return nil, fmt.Errorf("stt websocket url is empty")
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, resp, err := dialer.DialContext(ctx, f.wsURL, http.Header{})
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("dial stt websocket: %w (status %s)", err, resp.Status)
		}
		return nil, fmt.Errorf("dial stt websocket: %w", err)
	}

	stream := &sherpaWebSocketStream{
		sessionID:     sessionID,
		sink:          sink,
		logger:        f.logger,
		conn:          conn,
		activeSegment: -1,
		done:          make(chan struct{}),
	}
	go stream.readLoop()

	return stream, nil
}

type sherpaWebSocketStream struct {
	sessionID string
	sink      Sink
	logger    *log.Logger
	conn      *websocket.Conn

	writeMu sync.Mutex
	stateMu sync.Mutex

	activeSegment int
	activeText    string
	expectClose   bool
	closed        bool
	done          chan struct{}
}

type sherpaServerMessage struct {
	Text    string `json:"text"`
	Segment int    `json:"segment"`
}

func (s *sherpaWebSocketStream) Push(ctx context.Context, chunk AudioChunk) error {
	payload, err := pcm16ToFloat32Bytes(chunk.Data)
	if err != nil {
		return err
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	if s.isClosed() {
		return fmt.Errorf("stt stream closed")
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = s.conn.SetWriteDeadline(deadline)
	} else {
		_ = s.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	}

	if err := s.conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
		return fmt.Errorf("write audio to stt websocket: %w", err)
	}

	return nil
}

func (s *sherpaWebSocketStream) Close() error {
	s.stateMu.Lock()
	if s.closed {
		s.stateMu.Unlock()
		return nil
	}
	s.expectClose = true
	s.closed = true
	s.stateMu.Unlock()

	s.writeMu.Lock()
	_ = s.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	err := s.conn.WriteMessage(websocket.TextMessage, []byte("Done"))
	s.writeMu.Unlock()

	closeErr := s.conn.Close()
	<-s.done

	if err != nil && !isNormalWebSocketClose(err) {
		return fmt.Errorf("send Done to stt websocket: %w", err)
	}
	if closeErr != nil && !isNormalWebSocketClose(closeErr) {
		return fmt.Errorf("close stt websocket: %w", closeErr)
	}
	return nil
}

func (s *sherpaWebSocketStream) readLoop() {
	defer close(s.done)
	defer func() { _ = s.conn.Close() }()

	for {
		msgType, payload, err := s.conn.ReadMessage()
		if err != nil {
			s.emitFinalIfNeeded(true)
			if !s.isExpectedClose() && !isNormalWebSocketClose(err) {
				s.reportError(fmt.Errorf("read stt websocket: %w", err))
			}
			return
		}

		if msgType != websocket.TextMessage {
			continue
		}

		var message sherpaServerMessage
		if err := json.Unmarshal(payload, &message); err != nil {
			s.reportError(fmt.Errorf("decode stt message: %w", err))
			continue
		}

		s.stateMu.Lock()
		events := mapSherpaMessage(&s.activeSegment, &s.activeText, message)
		s.stateMu.Unlock()

		for _, event := range events {
			s.sink.HandleTranscriptEvent(context.Background(), event)
		}
	}
}

func (s *sherpaWebSocketStream) emitFinalIfNeeded(endpoint bool) {
	s.stateMu.Lock()
	text := strings.TrimSpace(s.activeText)
	s.activeText = ""
	s.stateMu.Unlock()

	if text == "" {
		return
	}

	s.sink.HandleTranscriptEvent(context.Background(), TranscriptEvent{
		Kind:     EventFinal,
		Text:     text,
		Endpoint: endpoint,
	})
}

func (s *sherpaWebSocketStream) reportError(err error) {
	if err == nil {
		return
	}
	if s.logger != nil {
		s.logger.Printf("[stt] %s: %v", s.sessionID, err)
	}
	s.sink.HandleSTTError(context.Background(), err)
}

func (s *sherpaWebSocketStream) isExpectedClose() bool {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	return s.expectClose
}

func (s *sherpaWebSocketStream) isClosed() bool {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	return s.closed
}

func mapSherpaMessage(activeSegment *int, activeText *string, message sherpaServerMessage) []TranscriptEvent {
	text := strings.TrimSpace(message.Text)
	if *activeSegment == -1 {
		*activeSegment = message.Segment
		*activeText = text
		if text == "" {
			return nil
		}
		return []TranscriptEvent{{
			Kind: EventPartial,
			Text: text,
		}}
	}

	switch {
	case message.Segment < *activeSegment:
		return nil
	case message.Segment == *activeSegment:
		if text == "" || text == *activeText {
			return nil
		}
		*activeText = text
		return []TranscriptEvent{{
			Kind: EventPartial,
			Text: text,
		}}
	default:
		var events []TranscriptEvent
		previous := strings.TrimSpace(*activeText)
		if previous != "" {
			events = append(events, TranscriptEvent{
				Kind:     EventFinal,
				Text:     previous,
				Endpoint: true,
			})
		}

		*activeSegment = message.Segment
		*activeText = text

		if text != "" {
			events = append(events, TranscriptEvent{
				Kind: EventPartial,
				Text: text,
			})
		}
		return events
	}
}

func pcm16ToFloat32Bytes(data []byte) ([]byte, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("pcm16 payload must be aligned to 2 bytes")
	}

	result := make([]byte, len(data)*2)
	for i := 0; i < len(data); i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		value := float32(sample) / 32768.0
		binary.LittleEndian.PutUint32(result[i*2:], math.Float32bits(value))
	}

	return result, nil
}

func isNormalWebSocketClose(err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, websocket.ErrCloseSent) {
		return true
	}
	var closeError *websocket.CloseError
	return errors.As(err, &closeError)
}

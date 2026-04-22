package stt

import (
	"context"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultTencentASRWSURL = "wss://asr.cloud.tencent.com/asr/v2"
	tencentVoiceFormatPCM  = "1"

	tencentSliceSentenceBegin  = 0
	tencentSliceResultChange   = 1
	tencentSliceSentenceFinish = 2

	tencentVoiceIDLength = 16
)

var tencentVoiceIDAlphabet = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// TencentAsrConfig holds Tencent ASR API credentials and parameters.
type TencentAsrConfig struct {
	WSURL         string
	AppID         string
	SecretID      string
	SecretKey     string
	EngineType    string
	NeedVAD       int
	NoEmptyResult int
	Logger        *log.Logger
}

type tencentAsrFactory struct {
	config TencentAsrConfig
}

func (f *tencentAsrFactory) NewStream(sessionID string, sink Sink) (Stream, error) {
	cfg := normalizeTencentAsrConfig(f.config)
	if err := validateTencentAsrConfig(cfg); err != nil {
		return nil, err
	}

	voiceID, err := newTencentVoiceID()
	if err != nil {
		return nil, fmt.Errorf("generate voice_id: %w", err)
	}

	stream := &tencentAsrWebSocketStream{
		sessionID: sessionID,
		voiceID:   voiceID,
		config:    cfg,
		connDone:  make(chan struct{}),
		listener:  newTencentTranscriptListener(sessionID, sink, cfg.Logger),
	}

	wsURL, err := buildTencentSignedURL(cfg, voiceID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("build signed url: %w", err)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, resp, err := dialer.DialContext(ctx, wsURL, http.Header{})
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("dial tencent asr websocket: %w (status %s)", err, resp.Status)
		}
		return nil, fmt.Errorf("dial tencent asr websocket: %w", err)
	}

	stream.conn = conn
	go stream.readLoop()

	return stream, nil
}

type tencentAsrWebSocketStream struct {
	sessionID   string
	voiceID     string
	config      TencentAsrConfig
	conn        *websocket.Conn
	listener    *tencentTranscriptListener
	writeMu     sync.Mutex
	stateMu     sync.Mutex
	started     bool
	expectClose bool
	closed      bool
	connDone    chan struct{}
}

func (s *tencentAsrWebSocketStream) Push(ctx context.Context, chunk AudioChunk) error {
	if len(chunk.Data)%2 != 0 {
		return fmt.Errorf("pcm16 payload must be aligned to 2 bytes")
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	s.stateMu.Lock()
	if s.closed {
		s.stateMu.Unlock()
		return fmt.Errorf("tencent asr stream closed")
	}
	s.stateMu.Unlock()

	deadline := time.Now().Add(5 * time.Second)
	if ctxDeadline, ok := ctx.Deadline(); ok {
		deadline = ctxDeadline
	}
	_ = s.conn.SetWriteDeadline(deadline)

	return s.conn.WriteMessage(websocket.BinaryMessage, chunk.Data)
}

func (s *tencentAsrWebSocketStream) Close() error {
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
	endPayload, _ := json.Marshal(map[string]string{"type": "end"})
	writeErr := s.conn.WriteMessage(websocket.TextMessage, endPayload)
	s.writeMu.Unlock()

	closeErr := s.conn.Close()
	<-s.connDone

	if writeErr != nil && !isNormalWebSocketClose(writeErr) {
		return fmt.Errorf("send end to tencent asr websocket: %w", writeErr)
	}
	if closeErr != nil && !isNormalWebSocketClose(closeErr) {
		return fmt.Errorf("close tencent asr websocket: %w", closeErr)
	}
	return nil
}


func (s *tencentAsrWebSocketStream) readLoop() {
	defer close(s.connDone)
	defer func() { _ = s.conn.Close() }()

	for {
		msgType, payload, err := s.conn.ReadMessage()
		if err != nil {
			s.listener.OnRecognitionComplete(nil)

			s.stateMu.Lock()
			unexpectedClose := !s.expectClose
			s.stateMu.Unlock()
			if unexpectedClose && !isNormalWebSocketClose(err) {
				s.listener.OnFail(nil, fmt.Errorf("read tencent asr websocket: %w", err))
			}
			return
		}

		if msgType != websocket.TextMessage {
			continue
		}

		var response TencentAsrResponse
		if err := json.Unmarshal(payload, &response); err != nil {
			s.listener.OnFail(nil, fmt.Errorf("decode tencent asr response: %w", err))
			continue
		}

		// if s.config.Logger != nil {

		// 	if response.Result != nil {

		// 		s.config.Logger.Printf(
		// 			"[tencent-asr] session=%s code=%d message=%s voice_id=%s message_id=%s final=%d slice_type=%d index=%d start=%d end=%d text=%s word_size=%d word_list=%v",
		// 			s.sessionID,
		// 			response.Code,
		// 			response.Message,
		// 			response.VoiceID,
		// 			response.MessageID,
		// 			response.Final,
		// 			response.Result.SliceType,
		// 			response.Result.Index,
		// 			response.Result.StartTime,
		// 			response.Result.EndTime,
		// 			response.Result.VoiceTextStr,
		// 			response.Result.WordSize,
		// 			response.Result.WordList,
		// 		)

		// 	} else {

		// 		s.config.Logger.Printf(
		// 			"[tencent-asr] session=%s code=%d message=%s voice_id=%s message_id=%s final=%d result=nil",
		// 			s.sessionID,
		// 			response.Code,
		// 			response.Message,
		// 			response.VoiceID,
		// 			response.MessageID,
		// 			response.Final,
		// 		)
		// 	}
		// }

		s.dispatchResponse(&response)

	}
}

func (s *tencentAsrWebSocketStream) dispatchResponse(response *TencentAsrResponse) {
	if response == nil || response.Code != 0 {
		if response != nil {
			s.listener.OnFail(response, fmt.Errorf("tencent asr error: code=%d message=%s", response.Code, response.Message))
		}
		return
	}

	s.stateMu.Lock()
	if !s.started {
		s.started = true
		s.stateMu.Unlock()
		s.listener.OnRecognitionStart(response)
	} else {
		s.stateMu.Unlock()
	}

	if response.Result != nil {
		switch response.Result.SliceType {
		case tencentSliceSentenceBegin:
			s.listener.OnSentenceBegin(response)
		case tencentSliceResultChange:
			s.listener.OnRecognitionResultChange(response)
		case tencentSliceSentenceFinish:
			s.listener.OnSentenceEnd(response)
		default:
			if strings.TrimSpace(response.Result.VoiceTextStr) != "" {
				s.listener.OnRecognitionResultChange(response)
			}
		}
	}
}

// func (s *tencentAsrWebSocketStream) isClosed() bool {
// 	s.stateMu.Lock()
// 	defer s.stateMu.Unlock()
// 	return s.closed
// }

type tencentTranscriptListener struct {
	sessionID string
	sink      Sink
	logger    *log.Logger

	mu sync.Mutex

	activeIndex    int
	activeText     string
	lastFinalIndex int
	lastFinalText  string
	completed      bool
}

func newTencentTranscriptListener(sessionID string, sink Sink, logger *log.Logger) *tencentTranscriptListener {
	return &tencentTranscriptListener{
		sessionID:      sessionID,
		sink:           sink,
		logger:         logger,
		activeIndex:    -1,
		lastFinalIndex: -1,
	}
}

func (l *tencentTranscriptListener) OnRecognitionStart(_ *TencentAsrResponse) {}

func (l *tencentTranscriptListener) OnSentenceBegin(response *TencentAsrResponse) {
	result, ok := trimmedTencentResult(response)
	if !ok {
		return
	}
	l.mu.Lock()
	l.activeIndex = result.Index
	l.activeText = result.VoiceTextStr
	l.mu.Unlock()

	// if result.VoiceTextStr == "" {
	// 	return
	// }
	l.emitPartial(result.Index, result.VoiceTextStr)
}

func (l *tencentTranscriptListener) OnRecognitionResultChange(response *TencentAsrResponse) {
	result, ok := trimmedTencentResult(response)
	if !ok {
		return
	}
	l.mu.Lock()
	prevIndex := l.activeIndex
	prevText := strings.TrimSpace(l.activeText)
	l.activeIndex = result.Index
	l.activeText = result.VoiceTextStr
	l.mu.Unlock()

	// When the index advances, the previous segment is complete — emit a
	// final event before the new partial, mirroring Sherpa's behaviour.
	// prevIndex >= 0: skip the initial state where no segment is active yet.
	// result.Index > prevIndex: only fire when the server moved to a new segment.
	// prevText != "": avoid emitting empty final events.
	if prevIndex >= 0 && result.Index > prevIndex && prevText != "" {
		l.sink.HandleTranscriptEvent(context.Background(), TranscriptEvent{
			Kind:     EventFinal,
			Text:     prevText,
			Endpoint: true,
		})
	}

	l.emitPartial(result.Index, result.VoiceTextStr)
}

func (l *tencentTranscriptListener) OnSentenceEnd(response *TencentAsrResponse) {
	result, ok := trimmedTencentResult(response)
	if !ok {
		return
	}

	l.mu.Lock()
	text := result.VoiceTextStr
	if text == "" && result.Index == l.activeIndex {
		text = strings.TrimSpace(l.activeText)
	}
	l.activeIndex = -1
	l.activeText = ""
	shouldEmit := text != "" && (result.Index != l.lastFinalIndex || text != l.lastFinalText)
	if shouldEmit {
		l.lastFinalIndex = result.Index
		l.lastFinalText = text
	}
	l.mu.Unlock()

	if shouldEmit {
		l.sink.HandleTranscriptEvent(context.Background(), TranscriptEvent{
			Kind:     EventFinal,
			Text:     text,
			Endpoint: true,
		})
	}
}

func (l *tencentTranscriptListener) OnRecognitionComplete(_ *TencentAsrResponse) {
	l.mu.Lock()
	if l.completed {
		l.mu.Unlock()
		return
	}
	l.completed = true

	index := l.activeIndex
	text := strings.TrimSpace(l.activeText)
	l.activeIndex = -1
	l.activeText = ""

	shouldEmit := text != "" && (index != l.lastFinalIndex || text != l.lastFinalText)
	if shouldEmit {
		l.lastFinalIndex = index
		l.lastFinalText = text
	}
	l.mu.Unlock()

	if shouldEmit {
		l.sink.HandleTranscriptEvent(context.Background(), TranscriptEvent{
			Kind:     EventFinal,
			Text:     text,
			Endpoint: true,
		})
	}
}

func (l *tencentTranscriptListener) OnFail(_ *TencentAsrResponse, err error) {
	if err == nil {
		return
	}
	if l.logger != nil {
		l.logger.Printf("[tencent-asr] %s: %v", l.sessionID, err)
	}
	l.sink.HandleSTTError(context.Background(), err)
}

func (l *tencentTranscriptListener) emitPartial(index int, text string) {
	l.sink.HandleTranscriptEvent(context.Background(), TranscriptEvent{
		Kind: EventPartial,
		Text: text,
	})
}

func trimmedTencentResult(response *TencentAsrResponse) (TencentAsrResult, bool) {
	if response == nil || response.Result == nil {
		return TencentAsrResult{}, false
	}

	result := *response.Result
	result.VoiceTextStr = strings.TrimSpace(result.VoiceTextStr)
	return result, true
}

func normalizeTencentAsrConfig(cfg TencentAsrConfig) TencentAsrConfig {
	cfg.WSURL = strings.TrimSpace(cfg.WSURL)
	cfg.AppID = strings.TrimSpace(cfg.AppID)
	cfg.SecretID = strings.TrimSpace(cfg.SecretID)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.EngineType = strings.TrimSpace(cfg.EngineType)
	if cfg.EngineType == "" {
		cfg.EngineType = "16k_zh"
	}
	return cfg
}

func validateTencentAsrConfig(cfg TencentAsrConfig) error {
	if cfg.AppID == "" {
		return fmt.Errorf("tencent asr app id is empty")
	}
	if cfg.SecretID == "" {
		return fmt.Errorf("tencent asr secret id is empty")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("tencent asr secret key is empty")
	}
	return nil
}

func buildTencentSignedURL(cfg TencentAsrConfig, voiceID string, now time.Time) (string, error) {
	endpoint, err := resolveTencentEndpoint(cfg.WSURL, cfg.AppID)
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("engine_model_type", cfg.EngineType)
	query.Set("expired", strconv.FormatInt(now.Add(24*time.Hour).Unix(), 10))
	query.Set("needvad", strconv.Itoa(cfg.NeedVAD))
	query.Set("filter_empty_result", strconv.Itoa(cfg.NoEmptyResult))
	query.Set("nonce", strconv.FormatInt(now.UnixNano()%1e10, 10))
	query.Set("secretid", cfg.SecretID)
	query.Set("timestamp", strconv.FormatInt(now.Unix(), 10))
	query.Set("voice_format", tencentVoiceFormatPCM)
	query.Set("voice_id", voiceID)

	// if cfg.Logger != nil {
	// 	cfg.Logger.Printf("[tencent-asr] NeedVAD=%d, EngineType=%s, NoEmptyResult=%d",
	// 		cfg.NeedVAD, cfg.EngineType, cfg.NoEmptyResult)
	// }

	unsignedQuery := query.Encode()
	signatureSource := endpoint.Host + endpoint.Path + "?" + unsignedQuery
	mac := hmac.New(sha1.New, []byte(cfg.SecretKey))
	_, _ = mac.Write([]byte(signatureSource))
	query.Set("signature", base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	signedEndpoint := *endpoint
	signedEndpoint.RawQuery = query.Encode()
	return signedEndpoint.String(), nil
}

func resolveTencentEndpoint(rawURL string, appID string) (*url.URL, error) {
	base := strings.TrimSpace(rawURL)
	if base == "" {
		base = defaultTencentASRWSURL
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse tencent asr websocket url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid tencent asr websocket url %q", base)
	}

	path := strings.TrimRight(parsed.Path, "/")
	if path == "" || path == "/asr/v2" {
		parsed.Path = "/asr/v2/" + appID
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed, nil
}

func newTencentVoiceID() (string, error) {
	buf := make([]byte, tencentVoiceIDLength)
	if _, err := crand.Read(buf); err != nil {
		return "", err
	}

	voiceID := make([]byte, tencentVoiceIDLength)
	for i, b := range buf {
		voiceID[i] = tencentVoiceIDAlphabet[int(b)%len(tencentVoiceIDAlphabet)]
	}
	return string(voiceID), nil
}

// TencentAsrResponse represents the response from Tencent ASR.
type TencentAsrResponse struct {
	Code      int               `json:"code"`
	Message   string            `json:"message"`
	VoiceID   string            `json:"voice_id"`
	MessageID string            `json:"message_id"`
	Result    *TencentAsrResult `json:"result"`
	Final     int               `json:"final"`
}

// TencentAsrResult represents the recognition result from Tencent ASR.
type TencentAsrResult struct {
	SliceType    int    `json:"slice_type"`
	Index        int    `json:"index"`
	StartTime    int    `json:"start_time"`
	EndTime      int    `json:"end_time"`
	VoiceTextStr string `json:"voice_text_str"`
	WordSize     int    `json:"word_size"`
	WordList     []struct {
		Word       string `json:"word"`
		StartTime  int    `json:"start_time"`
		EndTime    int    `json:"end_time"`
		StableFlag int    `json:"stable_flag"`
	} `json:"word_list"`
}

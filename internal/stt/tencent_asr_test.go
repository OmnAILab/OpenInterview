package stt

import (
	"context"
	"errors"
	"testing"
)

type tencentTestSink struct {
	events []TranscriptEvent
	errs   []error
}

func (s *tencentTestSink) HandleTranscriptEvent(_ context.Context, event TranscriptEvent) {
	s.events = append(s.events, event)
}

func (s *tencentTestSink) HandleSTTError(_ context.Context, err error) {
	s.errs = append(s.errs, err)
}

func TestTencentTranscriptListenerSentenceLifecycle(t *testing.T) {
	sink := &tencentTestSink{}
	listener := newTencentTranscriptListener("session-1", sink, nil)

	listener.OnRecognitionStart(tencentResponse(tencentSliceSentenceBegin, 0, "hello"))
	listener.OnSentenceBegin(tencentResponse(tencentSliceSentenceBegin, 0, "hello"))
	listener.OnRecognitionResultChange(tencentResponse(tencentSliceResultChange, 0, "hello world"))
	listener.OnSentenceEnd(tencentResponse(tencentSliceSentenceFinish, 0, "hello world"))
	listener.OnRecognitionComplete(&TencentAsrResponse{Final: 1})

	if len(sink.errs) != 0 {
		t.Fatalf("errs = %d, want 0", len(sink.errs))
	}
	if len(sink.events) != 3 {
		t.Fatalf("events = %d, want 3", len(sink.events))
	}

	if got := sink.events[0]; got.Kind != EventPartial || got.Text != "hello" || got.Endpoint {
		t.Fatalf("event[0] = %#v", got)
	}
	if got := sink.events[1]; got.Kind != EventPartial || got.Text != "hello world" || got.Endpoint {
		t.Fatalf("event[1] = %#v", got)
	}
	if got := sink.events[2]; got.Kind != EventFinal || got.Text != "hello world" || !got.Endpoint {
		t.Fatalf("event[2] = %#v", got)
	}
}

func TestTencentTranscriptListenerCompletionFlushesPendingSentence(t *testing.T) {
	sink := &tencentTestSink{}
	listener := newTencentTranscriptListener("session-2", sink, nil)

	listener.OnSentenceBegin(tencentResponse(tencentSliceSentenceBegin, 3, "pending"))
	listener.OnRecognitionComplete(&TencentAsrResponse{Final: 1})

	if len(sink.events) != 2 {
		t.Fatalf("events = %d, want 2", len(sink.events))
	}
	if got := sink.events[0]; got.Kind != EventPartial || got.Text != "pending" {
		t.Fatalf("event[0] = %#v", got)
	}
	if got := sink.events[1]; got.Kind != EventFinal || got.Text != "pending" || !got.Endpoint {
		t.Fatalf("event[1] = %#v", got)
	}
}

func TestTencentTranscriptListenerIndexAdvanceEmitsFinalLikeSherpa(t *testing.T) {
	sink := &tencentTestSink{}
	listener := newTencentTranscriptListener("session-3", sink, nil)

	listener.OnRecognitionResultChange(tencentResponse(tencentSliceResultChange, 0, "first"))
	listener.OnRecognitionResultChange(tencentResponse(tencentSliceResultChange, 1, "second"))

	if len(sink.events) != 3 {
		t.Fatalf("events = %d, want 3", len(sink.events))
	}
	if got := sink.events[0]; got.Kind != EventPartial || got.Text != "first" || got.Endpoint {
		t.Fatalf("event[0] = %#v", got)
	}
	if got := sink.events[1]; got.Kind != EventFinal || got.Text != "first" || !got.Endpoint {
		t.Fatalf("event[1] = %#v", got)
	}
	if got := sink.events[2]; got.Kind != EventPartial || got.Text != "second" || got.Endpoint {
		t.Fatalf("event[2] = %#v", got)
	}
}

func TestTencentTranscriptListenerFailReportsError(t *testing.T) {
	sink := &tencentTestSink{}
	listener := newTencentTranscriptListener("session-4", sink, nil)
	want := errors.New("boom")

	listener.OnFail(nil, want)

	if len(sink.errs) != 1 {
		t.Fatalf("errs = %d, want 1", len(sink.errs))
	}
	if !errors.Is(sink.errs[0], want) {
		t.Fatalf("sink.errs[0] = %v, want %v", sink.errs[0], want)
	}
}

func tencentResponse(sliceType int, index int, text string) *TencentAsrResponse {
	return &TencentAsrResponse{
		Code: 0,
		Result: &TencentAsrResult{
			SliceType:    sliceType,
			Index:        index,
			VoiceTextStr: text,
		},
	}
}

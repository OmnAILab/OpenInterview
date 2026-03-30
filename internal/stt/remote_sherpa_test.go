package stt

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestMapSherpaMessage(t *testing.T) {
	segment := -1
	text := ""

	events := mapSherpaMessage(&segment, &text, sherpaServerMessage{Segment: 0, Text: "hello"})
	if len(events) != 1 || events[0].Kind != EventPartial || events[0].Text != "hello" {
		t.Fatalf("unexpected first events: %#v", events)
	}

	events = mapSherpaMessage(&segment, &text, sherpaServerMessage{Segment: 0, Text: "hello world"})
	if len(events) != 1 || events[0].Kind != EventPartial || events[0].Text != "hello world" {
		t.Fatalf("unexpected partial update: %#v", events)
	}

	events = mapSherpaMessage(&segment, &text, sherpaServerMessage{Segment: 1, Text: "next"})
	if len(events) != 2 {
		t.Fatalf("expected final + partial, got %#v", events)
	}
	if events[0].Kind != EventFinal || events[0].Text != "hello world" || !events[0].Endpoint {
		t.Fatalf("unexpected final event: %#v", events[0])
	}
	if events[1].Kind != EventPartial || events[1].Text != "next" {
		t.Fatalf("unexpected partial event: %#v", events[1])
	}
}

func TestPCM16ToFloat32Bytes(t *testing.T) {
	input := []byte{0x00, 0x00, 0xff, 0x7f, 0x00, 0x80}
	output, err := pcm16ToFloat32Bytes(input)
	if err != nil {
		t.Fatalf("pcm16ToFloat32Bytes returned error: %v", err)
	}

	if len(output) != 12 {
		t.Fatalf("unexpected output size: %d", len(output))
	}

	got0 := float32FromBytes(output[0:4])
	got1 := float32FromBytes(output[4:8])
	got2 := float32FromBytes(output[8:12])

	if got0 != 0 {
		t.Fatalf("sample 0 = %f, want 0", got0)
	}
	if got1 < 0.9999 {
		t.Fatalf("sample 1 = %f, want about 1.0", got1)
	}
	if got2 > -0.9999 {
		t.Fatalf("sample 2 = %f, want about -1.0", got2)
	}
}

func float32FromBytes(data []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(data))
}

package textdeal

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Snapshot struct {
	StableText  string `json:"stableText"`
	SentUntil   int    `json:"sentUntil"`
	Markers     []int  `json:"markers"`
	PendingText string `json:"pendingText"`
}

type Segment struct {
	Text  string `json:"text"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type Buffer struct {
	stableText string
	sentUntil  int
	markers    []int
}

func (b *Buffer) Reset() {
	b.stableText = ""
	b.sentUntil = 0
	b.markers = nil
}

func (b *Buffer) AppendStable(text string) Snapshot {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return b.Snapshot()
	}

	b.stableText = appendStableText(b.stableText, normalized)
	return b.Snapshot()
}

func (b *Buffer) AddStopMarker(stop int) (Snapshot, error) {
	total := utf8.RuneCountInString(b.stableText)
	switch {
	case total == 0:
		return b.Snapshot(), fmt.Errorf("stable transcript is empty")
	case stop < 0 || stop > total:
		return b.Snapshot(), fmt.Errorf("stop index %d out of range 0..%d", stop, total)
	case stop <= b.sentUntil:
		return b.Snapshot(), fmt.Errorf("stop index %d must be greater than last stop %d", stop, b.sentUntil)
	case b.hasMarker(stop):
		return b.Snapshot(), fmt.Errorf("stop marker %d already exists", stop)
	}

	b.markers = append(b.markers, stop)
	return b.Snapshot(), nil
}

func (b *Buffer) SubmitStop(stop int) (Segment, Snapshot, error) {
	total := utf8.RuneCountInString(b.stableText)
	switch {
	case total == 0:
		return Segment{}, b.Snapshot(), fmt.Errorf("stable transcript is empty")
	case stop < 0 || stop > total:
		return Segment{}, b.Snapshot(), fmt.Errorf("stop index %d out of range 0..%d", stop, total)
	case stop <= b.sentUntil:
		return Segment{}, b.Snapshot(), fmt.Errorf("stop index %d must be greater than last stop %d", stop, b.sentUntil)
	}

	segment := strings.TrimSpace(sliceRunes(b.stableText, b.sentUntil, stop))
	if segment == "" {
		return Segment{}, b.Snapshot(), fmt.Errorf("segment between stops is empty")
	}

	result := Segment{
		Text:  segment,
		Start: b.sentUntil,
		End:   stop,
	}

	b.sentUntil = stop
	if !b.hasMarker(stop) {
		b.markers = append(b.markers, stop)
	}

	return result, b.Snapshot(), nil
}

func (b Buffer) Snapshot() Snapshot {
	total := utf8.RuneCountInString(b.stableText)
	sentUntil := b.sentUntil
	if sentUntil < 0 {
		sentUntil = 0
	}
	if sentUntil > total {
		sentUntil = total
	}

	markers := make([]int, 0, len(b.markers))
	for _, marker := range b.markers {
		if marker <= 0 || marker > total {
			continue
		}
		markers = append(markers, marker)
	}

	return Snapshot{
		StableText:  b.stableText,
		SentUntil:   sentUntil,
		Markers:     markers,
		PendingText: strings.TrimSpace(sliceRunes(b.stableText, sentUntil, total)),
	}
}

func (b Buffer) hasMarker(stop int) bool {
	for _, marker := range b.markers {
		if marker == stop {
			return true
		}
	}
	return false
}

func appendStableText(existing, next string) string {
	if existing == "" {
		return next
	}

	last, _ := utf8.DecodeLastRuneInString(existing)
	first, _ := utf8.DecodeRuneInString(next)
	if unicode.IsSpace(last) || unicode.IsSpace(first) {
		return existing + next
	}
	if shouldInsertSpace(last, first) {
		return existing + " " + next
	}
	return existing + next
}

func shouldInsertSpace(last, first rune) bool {
	if isASCIIWord(last) && isASCIIWord(first) {
		return true
	}
	if isSentencePunctuation(last) && isASCIIWord(first) {
		return true
	}
	return false
}

func isASCIIWord(value rune) bool {
	return value <= unicode.MaxASCII && (unicode.IsLetter(value) || unicode.IsDigit(value))
}

func isSentencePunctuation(value rune) bool {
	switch value {
	case '.', ',', '!', '?', ';', ':':
		return true
	default:
		return false
	}
}

func sliceRunes(text string, start, end int) string {
	runes := []rune(text)
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	if start > end {
		start = end
	}
	return string(runes[start:end])
}

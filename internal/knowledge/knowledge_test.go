package knowledge

import "testing"

func TestParseSearchResponse(t *testing.T) {
	payload := []byte(`{"results":[{"title":"rag","content":"retrieval chunk","path":"knowledge/rag.md","score":0.92}]}`)

	docs, err := parseSearchResponse(payload)
	if err != nil {
		t.Fatalf("parseSearchResponse: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("len(docs) = %d, want 1", len(docs))
	}
	if got, want := docs[0].Title, "rag"; got != want {
		t.Fatalf("Title = %q, want %q", got, want)
	}
	if got, want := docs[0].Content, "retrieval chunk"; got != want {
		t.Fatalf("Content = %q, want %q", got, want)
	}
}

func TestParseSearchResponseWithNestedDocument(t *testing.T) {
	payload := []byte(`{"data":[{"document":{"text":"nested chunk","path":"knowledge/nested.md"},"similarity":0.73}]}`)

	docs, err := parseSearchResponse(payload)
	if err != nil {
		t.Fatalf("parseSearchResponse: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("len(docs) = %d, want 1", len(docs))
	}
	if got, want := docs[0].Title, "nested"; got != want {
		t.Fatalf("Title = %q, want %q", got, want)
	}
	if got, want := docs[0].Path, "knowledge/nested.md"; got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
	if got, want := docs[0].Score, 0.73; got != want {
		t.Fatalf("Score = %v, want %v", got, want)
	}
}

package llm

import "testing"

func TestParseOpenAIChunk(t *testing.T) {
	token, done, err := parseOpenAIChunk([]byte(`{"choices":[{"delta":{"content":"hello"},"finish_reason":""}]}`))
	if err != nil {
		t.Fatalf("parseOpenAIChunk returned error: %v", err)
	}
	if token != "hello" {
		t.Fatalf("unexpected token: %q", token)
	}
	if done {
		t.Fatal("expected done=false")
	}

	token, done, err = parseOpenAIChunk([]byte(`{"choices":[{"delta":{},"finish_reason":"stop"}]}`))
	if err != nil {
		t.Fatalf("parseOpenAIChunk returned error: %v", err)
	}
	if token != "" || !done {
		t.Fatalf("unexpected final state token=%q done=%v", token, done)
	}
}

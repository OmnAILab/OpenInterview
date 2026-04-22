package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Path              string
	SearchEndpoint    string
	MaxResults        int
	EmbeddingEndpoint string
	EmbeddingAPIKey   string
	EmbeddingModel    string
	Timeout           time.Duration
}

type Document struct {
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Path    string  `json:"path,omitempty"`
	Score   float64 `json:"score,omitempty"`
}

type Client interface {
	Retrieve(ctx context.Context, query string) ([]Document, error)
}

func NewClient(cfg Config, logger *log.Logger) Client {
	cfg.Path = strings.TrimSpace(cfg.Path)
	cfg.SearchEndpoint = strings.TrimSpace(cfg.SearchEndpoint)
	cfg.EmbeddingEndpoint = strings.TrimSpace(cfg.EmbeddingEndpoint)
	if cfg.MaxResults <= 0 {
		cfg.MaxResults = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	if cfg.SearchEndpoint != "" {
		return &remoteSearchClient{
			cfg:    cfg,
			client: &http.Client{Timeout: cfg.Timeout},
			logger: logger,
		}
	}
	if cfg.Path == "" || cfg.EmbeddingEndpoint == "" {
		return noopClient{}
	}
	return &localVectorClient{
		cfg:      cfg,
		embedder: remoteEmbedder{cfg: cfg, client: &http.Client{Timeout: cfg.Timeout}},
		logger:   logger,
	}
}

type noopClient struct{}

func (noopClient) Retrieve(context.Context, string) ([]Document, error) {
	return nil, nil
}

type remoteSearchClient struct {
	cfg    Config
	client *http.Client
	logger *log.Logger
}

func (c *remoteSearchClient) Retrieve(ctx context.Context, query string) ([]Document, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	body := map[string]any{
		"query":    query,
		"question": query,
		"text":     query,
		"top_k":    c.cfg.MaxResults,
		"topK":     c.cfg.MaxResults,
		"limit":    c.cfg.MaxResults,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.SearchEndpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.EmbeddingAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.EmbeddingAPIKey)
		req.Header.Set("X-API-Key", c.cfg.EmbeddingAPIKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("knowledge request failed: %s: %s", resp.Status, strings.TrimSpace(string(payload)))
	}

	results, err := parseSearchResponse(payload)
	if err != nil {
		return nil, err
	}
	if len(results) > c.cfg.MaxResults {
		results = results[:c.cfg.MaxResults]
	}
	if c.logger != nil {
		c.logger.Printf("knowledge retrieved %d chunks from %s", len(results), c.cfg.SearchEndpoint)
	}
	return results, nil
}

type localVectorClient struct {
	cfg      Config
	embedder remoteEmbedder
	logger   *log.Logger

	mu     sync.RWMutex
	loaded bool
	index  []indexedDocument
}

type indexedDocument struct {
	doc       Document
	embedding []float32
}

func (c *localVectorClient) Retrieve(ctx context.Context, query string) ([]Document, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	if err := c.ensureLoaded(ctx); err != nil {
		return nil, err
	}

	queryVectors, err := c.embedder.Embed(ctx, []string{query}, "query")
	if err != nil {
		return nil, err
	}
	if len(queryVectors) == 0 {
		return nil, nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	scored := make([]Document, 0, len(c.index))
	for _, item := range c.index {
		score := cosine(queryVectors[0], item.embedding)
		if score <= 0 {
			continue
		}
		doc := item.doc
		doc.Score = score
		scored = append(scored, doc)
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	if len(scored) > c.cfg.MaxResults {
		scored = scored[:c.cfg.MaxResults]
	}
	return scored, nil
}

func (c *localVectorClient) ensureLoaded(ctx context.Context) error {
	c.mu.RLock()
	loaded := c.loaded
	c.mu.RUnlock()
	if loaded {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loaded {
		return nil
	}
	index, err := c.load(ctx)
	if err != nil {
		return err
	}
	c.index = index
	c.loaded = true
	return nil
}

func (c *localVectorClient) load(ctx context.Context) ([]indexedDocument, error) {
	docs, err := loadDocuments(c.cfg.Path)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.Content)
	}
	vectors, err := c.embedder.Embed(ctx, texts, "document")
	if err != nil {
		return nil, err
	}
	if len(vectors) != len(docs) {
		return nil, fmt.Errorf("embedding count %d does not match document count %d", len(vectors), len(docs))
	}

	index := make([]indexedDocument, 0, len(docs))
	for i, doc := range docs {
		if len(vectors[i]) == 0 {
			continue
		}
		index = append(index, indexedDocument{doc: doc, embedding: vectors[i]})
	}

	if c.logger != nil {
		c.logger.Printf("knowledge indexed %d chunks from %s", len(index), c.cfg.Path)
	}
	return index, nil
}

func loadDocuments(root string) ([]Document, error) {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return loadDocumentFile(root)
	}

	var docs []Document
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !isKnowledgeFile(path) {
			return nil
		}
		fileDocs, err := loadDocumentFile(path)
		if err != nil {
			return err
		}
		docs = append(docs, fileDocs...)
		return nil
	})
	return docs, err
}

func loadDocumentFile(path string) ([]Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	chunks := splitChunks(string(content), 1200)
	docs := make([]Document, 0, len(chunks))
	for i, chunk := range chunks {
		docTitle := title
		if len(chunks) > 1 {
			docTitle = fmt.Sprintf("%s #%d", title, i+1)
		}
		docs = append(docs, Document{
			Title:   docTitle,
			Content: chunk,
			Path:    path,
		})
	}
	return docs, nil
}

func isKnowledgeFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".txt", ".md", ".markdown":
		return true
	default:
		return false
	}
}

func splitChunks(text string, maxRunes int) []string {
	parts := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n\n")
	var chunks []string
	var builder strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if builder.Len() > 0 && len([]rune(builder.String()))+len([]rune(part))+2 > maxRunes {
			chunks = append(chunks, builder.String())
			builder.Reset()
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(part)
	}
	if builder.Len() > 0 {
		chunks = append(chunks, builder.String())
	}
	return chunks
}

func cosine(left, right []float32) float64 {
	if len(left) == 0 || len(left) != len(right) {
		return 0
	}
	var dot, leftNorm, rightNorm float64
	for i := range left {
		l := float64(left[i])
		r := float64(right[i])
		dot += l * r
		leftNorm += l * l
		rightNorm += r * r
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

type remoteEmbedder struct {
	cfg    Config
	client *http.Client
}

func (e remoteEmbedder) Embed(ctx context.Context, texts []string, task string) ([][]float32, error) {
	body := map[string]any{
		"texts":     texts,
		"inputs":    texts,
		"sentences": texts,
		"task":      task,
		"normalize": true,
	}
	if e.cfg.EmbeddingModel != "" {
		body["model"] = e.cfg.EmbeddingModel
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.cfg.EmbeddingEndpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if e.cfg.EmbeddingAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.cfg.EmbeddingAPIKey)
		req.Header.Set("X-API-Key", e.cfg.EmbeddingAPIKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding request failed: %s: %s", resp.Status, strings.TrimSpace(string(payload)))
	}
	return parseEmbeddingResponse(payload)
}

func parseSearchResponse(payload []byte) ([]Document, error) {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return extractDocuments(raw), nil
}

func extractDocuments(raw any) []Document {
	switch value := raw.(type) {
	case []any:
		return extractDocumentList(value)
	case map[string]any:
		for _, key := range []string{"results", "data", "documents", "records", "hits"} {
			if docs := extractDocuments(value[key]); len(docs) > 0 {
				return docs
			}
		}
		if docs := extractDocumentCandidates(value); len(docs) > 0 {
			return docs
		}
	}
	return nil
}

func extractDocumentList(items []any) []Document {
	result := make([]Document, 0, len(items))
	for _, item := range items {
		switch value := item.(type) {
		case map[string]any:
			if doc, ok := toDocument(value); ok {
				result = append(result, doc)
				continue
			}
			if docs := extractDocuments(value); len(docs) > 0 {
				result = append(result, docs...)
			}
		case string:
			text := strings.TrimSpace(value)
			if text != "" {
				result = append(result, Document{Content: text, Title: "result"})
			}
		}
	}
	return result
}

func extractDocumentCandidates(value map[string]any) []Document {
	if doc, ok := toDocument(value); ok {
		return []Document{doc}
	}
	return nil
}

func toDocument(value map[string]any) (Document, bool) {
	if embedded, ok := value["document"].(map[string]any); ok {
		if doc, ok := toDocument(embedded); ok {
			if doc.Score == 0 {
				doc.Score = extractFloat(value, "score", "similarity", "relevance")
			}
			return doc, true
		}
	}

	content := firstNonEmptyString(value, "content", "text", "body", "snippet", "chunk")
	title := firstNonEmptyString(value, "title", "name", "heading")
	path := firstNonEmptyString(value, "path", "source", "file")
	if content == "" {
		return Document{}, false
	}
	if title == "" {
		title = inferTitle(path)
	}
	if title == "" {
		title = "result"
	}
	return Document{
		Title:   title,
		Content: content,
		Path:    path,
		Score:   extractFloat(value, "score", "similarity", "relevance"),
	}, true
}

func firstNonEmptyString(value map[string]any, keys ...string) string {
	for _, key := range keys {
		raw, ok := value[key]
		if !ok {
			continue
		}
		switch typed := raw.(type) {
		case string:
			if trimmed := strings.TrimSpace(typed); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func extractFloat(value map[string]any, keys ...string) float64 {
	for _, key := range keys {
		raw, ok := value[key]
		if !ok {
			continue
		}
		switch typed := raw.(type) {
		case float64:
			return typed
		case float32:
			return float64(typed)
		case int:
			return float64(typed)
		case int64:
			return float64(typed)
		}
	}
	return 0
}

func inferTitle(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func parseEmbeddingResponse(payload []byte) ([][]float32, error) {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return extractEmbeddings(raw), nil
}

func extractEmbeddings(raw any) [][]float32 {
	switch value := raw.(type) {
	case []any:
		if vectors := extractVectorList(value); len(vectors) > 0 {
			return vectors
		}
		var result [][]float32
		for _, item := range value {
			if obj, ok := item.(map[string]any); ok {
				if vector := extractVector(obj["embedding"]); len(vector) > 0 {
					result = append(result, vector)
				}
			}
		}
		return result
	case map[string]any:
		for _, key := range []string{"embeddings", "vectors"} {
			if vectors := extractEmbeddings(value[key]); len(vectors) > 0 {
				return vectors
			}
		}
		if vector := extractVector(value["embedding"]); len(vector) > 0 {
			return [][]float32{vector}
		}
		if data, ok := value["data"].([]any); ok {
			var result [][]float32
			for _, item := range data {
				if obj, ok := item.(map[string]any); ok {
					if vector := extractVector(obj["embedding"]); len(vector) > 0 {
						result = append(result, vector)
					}
				}
			}
			return result
		}
	}
	return nil
}

func extractVectorList(value []any) [][]float32 {
	result := make([][]float32, 0, len(value))
	for _, item := range value {
		vector := extractVector(item)
		if len(vector) == 0 {
			return nil
		}
		result = append(result, vector)
	}
	return result
}

func extractVector(raw any) []float32 {
	values, ok := raw.([]any)
	if !ok {
		return nil
	}
	vector := make([]float32, 0, len(values))
	for _, item := range values {
		number, ok := item.(float64)
		if !ok {
			return nil
		}
		vector = append(vector, float32(number))
	}
	return vector
}

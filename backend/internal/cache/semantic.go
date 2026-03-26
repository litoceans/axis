package cache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/godlabs/axis/pkg/types"
	"github.com/qdrant/go-client/qdrant"
	"github.com/rs/zerolog/log"
)

// SemanticConfig holds semantic cache configuration
type SemanticConfig struct {
	Enabled             bool          `mapstructure:"enabled"`
	QdrantURL           string        `mapstructure:"qdrant_url"`
	Collection          string        `mapstructure:"collection"`
	SimilarityThreshold float64       `mapstructure:"similarity_threshold"`
	TTL                 time.Duration `mapstructure:"ttl"`
	EmbedderModel       string        `mapstructure:"embedder_model"`
	EmbedderProvider    string        `mapstructure:"embedder_provider"`
	EmbedderAPIKey      string        `mapstructure:"embedder_api_key"`
	EmbedderBaseURL     string        `mapstructure:"embedder_base_url"`
	VectorSize          int           `mapstructure:"vector_size"`
}

// SemanticCache provides semantic caching using Qdrant
type SemanticCache struct {
	mu        sync.RWMutex
	client    *qdrant.Client
	config    SemanticConfig
	embedder  Embedder
	keyMu     sync.RWMutex
	vectorMap map[string]string // cacheKey -> point ID
}

// NewSemanticCache creates a new semantic cache instance
func NewSemanticCache(config SemanticConfig, embedder Embedder) (*SemanticCache, error) {
	if !config.Enabled {
		return &SemanticCache{}, nil
	}

	host, port, err := parseQdrantURL(config.QdrantURL)
	if err != nil {
		return nil, fmt.Errorf("invalid qdrant url: %w", err)
	}

	qdrantConfig := &qdrant.Config{
		Host: host,
		Port: port,
	}

	client, err := qdrant.NewClient(qdrantConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	sc := &SemanticCache{
		client:    client,
		config:    config,
		embedder:  embedder,
		vectorMap: make(map[string]string),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sc.ensureCollection(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to ensure semantic cache collection, cache disabled")
		return &SemanticCache{}, nil
	}

	log.Info().
		Str("collection", config.Collection).
		Float64("threshold", config.SimilarityThreshold).
		Msg("semantic cache initialized")

	return sc, nil
}

// IsEnabled returns whether the semantic cache is active
func (s *SemanticCache) IsEnabled() bool {
	return s != nil && s.client != nil && s.config.Enabled
}

// ensureCollection creates the Qdrant collection if it doesn't exist
func (s *SemanticCache) ensureCollection(ctx context.Context) error {
	exists, err := s.client.CollectionExists(ctx, s.config.Collection)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	vectorSize := uint64(s.config.VectorSize)
	if vectorSize == 0 {
		vectorSize = 768 // default for nomic-embed-text
	}

	collection := &qdrant.CreateCollection{
		CollectionName: s.config.Collection,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     vectorSize,
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	}

	return s.client.CreateCollection(ctx, collection)
}

// Search looks for a cached response similar to the given content.
// Returns the cached response bytes, cost, and whether a hit was found.
func (s *SemanticCache) Search(ctx context.Context, messages []types.ChatMessage, model string) ([]byte, float64, bool) {
	if !s.IsEnabled() || s.embedder == nil {
		return nil, 0, false
	}

	cacheKey := s.getCacheKey(messages, model)

	// Fast-path: check if we have a known point ID for this cache key
	s.keyMu.RLock()
	pointIDStr := s.vectorMap[cacheKey]
	s.keyMu.RUnlock()

	if pointIDStr != "" {
		points, err := s.client.Get(ctx, &qdrant.GetPoints{
			CollectionName: s.config.Collection,
			Ids:           []*qdrant.PointId{qdrant.NewID(pointIDStr)},
			WithPayload:   &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		})
		if err == nil && len(points) > 0 && points[0] != nil {
			payload := points[0].GetPayload()
			if keyField := payload["key"]; keyField != nil {
				if sv, ok := keyField.GetKind().(*qdrant.Value_StringValue); ok && sv.StringValue == cacheKey {
					if respField := payload["response"]; respField != nil {
						if rv, ok := respField.GetKind().(*qdrant.Value_StringValue); ok {
							resp, _ := hex.DecodeString(rv.StringValue)
							cost := 0.0
							if cf := payload["cost_usd"]; cf != nil {
								if dv, ok := cf.GetKind().(*qdrant.Value_DoubleValue); ok {
									cost = dv.DoubleValue
								}
							}
							return resp, cost, true
						}
					}
				}
			}
		}
	}

	// Generate embedding for the query
	text := normalizeMessagesForEmbedding(messages)
	embedding, err := s.embedder.Embed(ctx, text)
	if err != nil {
		log.Debug().Err(err).Msg("failed to generate embedding for semantic cache lookup")
		return nil, 0, false
	}

	// Search Qdrant
	threshold := float32(s.config.SimilarityThreshold)
	searchReq := &qdrant.SearchPoints{
		CollectionName: s.config.Collection,
		Vector:        embedding,
		Limit:         5,
		WithPayload:   &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		ScoreThreshold: &threshold,
	}

	results, err := s.client.GetPointsClient().Search(ctx, searchReq)
	if err != nil {
		log.Debug().Err(err).Msg("semantic cache search failed")
		return nil, 0, false
	}

	for _, result := range results.GetResult() {
		payload := result.GetPayload()
		if payload == nil {
			continue
		}

		keyField := payload["key"]
		if keyField == nil {
			continue
		}
		sv, ok := keyField.GetKind().(*qdrant.Value_StringValue)
		if !ok {
			continue
		}
		storedKey := sv.StringValue

		if storedKey != cacheKey {
			continue
		}

		respField := payload["response"]
		if respField == nil {
			continue
		}
		rv, ok := respField.GetKind().(*qdrant.Value_StringValue)
		if !ok {
			continue
		}

		resp, err := hex.DecodeString(rv.StringValue)
		if err != nil {
			continue
		}

		cost := 0.0
		if cf := payload["cost_usd"]; cf != nil {
			if dv, ok := cf.GetKind().(*qdrant.Value_DoubleValue); ok {
				cost = dv.DoubleValue
			}
		}

		// Store point ID mapping for fast retrieval
		pointIDStr := pointIDToString(result.Id)
		s.keyMu.Lock()
		s.vectorMap[cacheKey] = pointIDStr
		s.keyMu.Unlock()

		log.Debug().
			Str("cache_key", cacheKey[:16]).
			Float32("score", result.GetScore()).
			Msg("semantic cache hit")
		return resp, cost, true
	}

	return nil, 0, false
}

// Store saves a response in the semantic cache
func (s *SemanticCache) Store(ctx context.Context, messages []types.ChatMessage, model string, response []byte, costUSD float64) {
	if !s.IsEnabled() || s.embedder == nil {
		return
	}

	cacheKey := s.getCacheKey(messages, model)

	text := normalizeMessagesForEmbedding(messages)
	embedding, err := s.embedder.Embed(ctx, text)
	if err != nil {
		log.Debug().Err(err).Msg("failed to generate embedding for semantic cache store")
		return
	}

	// Use deterministic point ID based on cache key hash
	pointIDStr := hashToUUID(cacheKey)

	// Encode response as hex string for payload storage
	responseHex := hex.EncodeToString(response)

	ttlSeconds := uint64(s.config.TTL.Seconds())
	if ttlSeconds == 0 {
		ttlSeconds = 604800 // default 7 days
	}
	nowSec := int64(time.Now().Unix())

	point := &qdrant.PointStruct{
		Id: qdrant.NewID(pointIDStr),
		Vectors: qdrant.NewVectorsDense(embedding),
		Payload: map[string]*qdrant.Value{
			"key": {
				Kind: &qdrant.Value_StringValue{StringValue: cacheKey},
			},
			"response": {
				Kind: &qdrant.Value_StringValue{StringValue: responseHex},
			},
			"cost_usd": {
				Kind: &qdrant.Value_DoubleValue{DoubleValue: costUSD},
			},
			"created_at": {
				Kind: &qdrant.Value_IntegerValue{IntegerValue: nowSec},
			},
			"expires_at": {
				Kind: &qdrant.Value_IntegerValue{IntegerValue: nowSec + int64(ttlSeconds)},
			},
		},
	}

	upsert := &qdrant.UpsertPoints{
		CollectionName: s.config.Collection,
		Points:        []*qdrant.PointStruct{point},
	}

	if _, err := s.client.Upsert(ctx, upsert); err != nil {
		log.Debug().Err(err).Msg("failed to store in semantic cache")
		return
	}

	s.keyMu.Lock()
	s.vectorMap[cacheKey] = pointIDStr
	s.keyMu.Unlock()

	log.Debug().Str("cache_key", cacheKey[:16]).Msg("stored in semantic cache")
}

// getCacheKey generates a deterministic cache key from messages and model
func (s *SemanticCache) getCacheKey(messages []types.ChatMessage, model string) string {
	normalized := normalizeMessagesForEmbedding(messages)
	hash := sha256.Sum256(append([]byte(model), []byte(normalized)...))
	return hex.EncodeToString(hash[:])
}

// normalizeMessagesForEmbedding converts chat messages to a canonical string
func normalizeMessagesForEmbedding(messages []types.ChatMessage) string {
	var parts []string
	for _, m := range messages {
		parts = append(parts, fmt.Sprintf("%s:%s", m.Role, m.Content))
	}
	return strings.Join(parts, "\n")
}

// hashToUUID creates a UUID-like string from a hex string
func hashToUUID(hash string) string {
	return hash[:8] + "-" + hash[8:12] + "-" + hash[12:16] + "-" + hash[16:20] + "-" + hash[20:36]
}

// pointIDToString converts a Qdrant point ID to string
func pointIDToString(id *qdrant.PointId) string {
	if id == nil {
		return ""
	}
	if uuid := id.GetUuid(); uuid != "" {
		return uuid
	}
	return fmt.Sprintf("%d", id.GetNum())
}

// Embedder defines the interface for generating embeddings
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// OllamaEmbedder generates embeddings using Ollama's API
type OllamaEmbedder struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

// NewOllamaEmbedder creates a new Ollama embedder
func NewOllamaEmbedder(baseURL, model string) *OllamaEmbedder {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return &OllamaEmbedder{
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Embed generates an embedding for the given text using Ollama
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model":  e.Model,
		"prompt": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := e.BaseURL + "api/embeddings"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	if len(result.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return float64sToFloat32s(result.Embedding), nil
}

// OpenAIEmbedder generates embeddings using OpenAI's API
type OpenAIEmbedder struct {
	BaseURL string
	Model   string
	APIKey  string
	Client  *http.Client
}

// NewOpenAIEmbedder creates a new OpenAI embedder
func NewOpenAIEmbedder(baseURL, apiKey, model string) *OpenAIEmbedder {
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL += "/v1"
	}
	return &OpenAIEmbedder{
		BaseURL: baseURL,
		Model:   model,
		APIKey:  apiKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Embed generates an embedding for the given text using OpenAI
func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model": e.Model,
		"input": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := e.BaseURL + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return float64sToFloat32s(result.Data[0].Embedding), nil
}

// float64sToFloat32s converts []float64 to []float32
func float64sToFloat32s(floats []float64) []float32 {
	result := make([]float32, len(floats))
	for i, f := range floats {
		result[i] = float32(f)
	}
	return result
}

// parseQdrantURL parses a Qdrant URL into host and port
func parseQdrantURL(rawURL string) (string, int, error) {
	rawURL = strings.TrimPrefix(rawURL, "http://")
	rawURL = strings.TrimPrefix(rawURL, "https://")

	parts := strings.Split(rawURL, ":")
	if len(parts) != 2 {
		return rawURL, 6334, nil
	}

	var port int
	if _, err := fmt.Sscanf(parts[1], "%d", &port); err != nil {
		port = 6334
	}
	return parts[0], port, nil
}

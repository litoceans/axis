package types

import (
	"encoding/json"
	"time"
)

// ChatMessage represents a chat message in the OpenAI format
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatRequest represents an incoming chat completion request
type ChatRequest struct {
	Model       string          `json:"model"`
	Messages    []ChatMessage    `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Stop        json.RawMessage `json:"stop,omitempty"`
	PresencePenalty float64     `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64    `json:"frequency_penalty,omitempty"`
	User        string          `json:"user,omitempty"`
	ExtraBody   json.RawMessage `json:"extra_body,omitempty"`
}

// ChatResponseChoice represents a single choice in a chat response
type ChatResponseChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatUsage represents token usage in a chat response
type ChatUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	CostUSD          float64 `json:"cost_usd,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Provider string              `json:"provider,omitempty"`
	Choices []ChatResponseChoice `json:"choices"`
	Usage   ChatUsage            `json:"usage"`
	Cached  bool                 `json:"cached,omitempty"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Provider string `json:"provider,omitempty"`
	Choices []StreamChoice `json:"choices"`
	Usage   *ChatUsage `json:"usage,omitempty"`
}

// StreamChoice represents a single choice in a stream chunk
type StreamChoice struct {
	Index        int              `json:"index"`
	Delta        json.RawMessage  `json:"delta"`
	FinishReason string           `json:"finish_reason,omitempty"`
}

// EmbedRequest represents an embedding request
type EmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// Embedding represents a single embedding
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbedUsage represents usage stats for embeddings
type EmbedUsage struct {
	PromptTokens int     `json:"prompt_tokens"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
}

// EmbedResponse represents an embedding response
type EmbedResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  EmbedUsage  `json:"usage"`
}

// ModelInfo represents information about an available model
type ModelInfo struct {
	ID              string  `json:"id"`
	Object          string  `json:"object"`
	Created         int64   `json:"created"`
	OwnedBy         string  `json:"owned_by"`
	Provider        string  `json:"provider,omitempty"`
	InputCost       float64 `json:"input_cost_per_million,omitempty"`
	OutputCost      float64 `json:"output_cost_per_million,omitempty"`
	MaxContextTokens int    `json:"max_context_tokens,omitempty"`
}

// ModelsResponse represents the response for listing models
type ModelsResponse struct {
	Object string       `json:"object"`
	Data   []ModelInfo  `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Limit     int    `json:"limit,omitempty"`
	Current   int    `json:"current,omitempty"`
	ResetAt   string `json:"reset_at,omitempty"`
	RetryAfterMs int `json:"retry_after_ms,omitempty"`
}

// APIKeyResponse represents an API key in responses
type APIKeyResponse struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	KeyPrefix       string   `json:"keyPrefix"`
	OrgID           string   `json:"orgId"`
	TeamID          string   `json:"teamId,omitempty"`
	MemberID        string   `json:"memberId,omitempty"`
	Scopes          []string `json:"scopes"`
	Models          []string `json:"models,omitempty"`
	RPMLimit        int      `json:"rpmLimit,omitempty"`
	TPMLimit        int      `json:"tpmLimit,omitempty"`
	MonthlyBudgetUSD float64 `json:"monthlyBudgetUsd,omitempty"`
	Environments    []string `json:"environments"`
	ExpiresAt       string   `json:"expiresAt,omitempty"`
	LastUsedAt      string   `json:"lastUsedAt,omitempty"`
	RevokedAt       string   `json:"revokedAt,omitempty"`
	CreatedAt       string   `json:"createdAt"`
}

// CreateKeyRequest represents a request to create an API key
type CreateKeyRequest struct {
	Name            string   `json:"name"`
	OrgID           string   `json:"orgId,omitempty"`
	TeamID          string   `json:"teamId,omitempty"`
	Scopes          []string `json:"scopes"`
	Models          []string `json:"models,omitempty"`
	RPMLimit        int      `json:"rpmLimit,omitempty"`
	TPMLimit        int      `json:"tpmLimit,omitempty"`
	MonthlyBudgetUSD float64 `json:"monthlyBudgetUsd,omitempty"`
	ExpiresAt       string   `json:"expiresAt,omitempty"`
	Environments    []string `json:"environments,omitempty"`
}

// CreateKeyResponse represents the response when creating an API key (includes the raw key)
type CreateKeyResponse struct {
	APIKey *APIKeyResponse `json:"apiKey"`
	RawKey string          `json:"rawKey"` // Only returned once at creation
}

// UsageResponse represents usage data response
type UsageResponse struct {
	Usage []UsageData `json:"usage"`
	Totals struct {
		Requests  int     `json:"requests"`
		TokensIn  int     `json:"tokensIn"`
		TokensOut int     `json:"tokensOut"`
		CostUsd   float64 `json:"costUsd"`
	} `json:"totals"`
}

// UsageData represents a single usage data point
type UsageData struct {
	Date      string  `json:"date"`
	Requests  int     `json:"requests"`
	TokensIn  int     `json:"tokensIn"`
	TokensOut int     `json:"tokensOut"`
	CostUsd   float64 `json:"costUsd"`
}

// CostBreakdown represents cost breakdown by model/provider
type CostBreakdown struct {
	Model      string  `json:"model"`
	Provider   string  `json:"provider"`
	Requests   int     `json:"requests"`
	TokensIn   int     `json:"tokensIn"`
	TokensOut  int     `json:"tokensOut"`
	CostUsd    float64 `json:"costUsd"`
}

// RoutingChain represents a routing chain
type RoutingChain struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	IsDefault bool         `json:"isDefault"`
	Models    []ChainModel `json:"models"`
}

// ChainModel represents a model in a routing chain
type ChainModel struct {
	Model      string  `json:"model"`
	Provider   string  `json:"provider"`
	MaxLatency float64 `json:"maxLatency"`
	Retries    int     `json:"retries"`
	Weight     int     `json:"weight"`
	FailOpen   bool    `json:"failOpen,omitempty"`
}

// CreateChainRequest represents a request to create a routing chain
type CreateChainRequest struct {
	Name      string       `json:"name"`
	IsDefault bool         `json:"isDefault"`
	Models    []ChainModel `json:"models"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	HitRate         float64           `json:"hitRate"`
	TotalHits       int               `json:"totalHits"`
	TotalMisses     int               `json:"totalMisses"`
	Savings         float64           `json:"savings"`
	TopCachedPrompts []CachedPrompt   `json:"topCachedPrompts"`
}

// CachedPrompt represents a cached prompt entry
type CachedPrompt struct {
	Hash     string `json:"hash"`
	Count    int    `json:"count"`
	LastUsed string `json:"lastUsed"`
}

// ProviderHealth represents health status of a provider
type ProviderHealth struct {
	Provider   string  `json:"provider"`
	Status     string  `json:"status"` // "healthy", "degraded", "down"
	ErrorRate  float64 `json:"error_rate"`
	P99LatencyMs int64 `json:"p99_latency_ms"`
	Requests   int64   `json:"requests_last_window"`
}

// UsageLog represents a usage log entry
type UsageLog struct {
	ID                     string    `json:"id"`
	KeyID                  string    `json:"key_id"`
	OrgID                  string    `json:"org_id"`
	Model                  string    `json:"model"`
	Provider               string    `json:"provider"`
	Endpoint               string    `json:"endpoint"`
	InputTokens            int       `json:"input_tokens"`
	OutputTokens           int       `json:"output_tokens"`
	CachedTokens           int       `json:"cached_tokens"`
	CostUSD                float64   `json:"cost_usd"`
	LatencyMs              int       `json:"latency_ms"`
	StatusCode             int       `json:"status_code"`
	Cached                 bool      `json:"cached"`
	Error                  bool      `json:"error"`
	ErrorType              string    `json:"error_type,omitempty"`
	TraceID                string    `json:"trace_id,omitempty"`
	TokensInputReported    int       `json:"tokens_input_reported"`
	TokensOutputReported   int       `json:"tokens_output_reported"`
	TokensInputCalculated  int       `json:"tokens_input_calculated"`
	TokensOutputCalculated int       `json:"tokens_output_calculated"`
	CostReported           float64   `json:"cost_reported"`
	CostCalculated         float64   `json:"cost_calculated"`
	CreatedAt              time.Time `json:"created_at"`
}

// AxisConfig represents the runtime configuration
type AxisConfig struct {
	Server      ServerConfig       `mapstructure:"server"`
	Database    DatabaseConfig     `mapstructure:"database"`
	Cache       CacheConfig        `mapstructure:"cache"`
	Providers   ProvidersConfig    `mapstructure:"providers"`
	Routing     RoutingConfig      `mapstructure:"routing"`
	Keys        KeysConfig         `mapstructure:"keys"`
	RateLimits  RateLimitsConfig   `mapstructure:"rate_limits"`
	Costs       CostsConfig        `mapstructure:"costs"`
	Telemetry   TelemetryConfig    `mapstructure:"telemetry"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxConnections int           `mapstructure:"max_connections"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL         string `mapstructure:"url"`
	MaxOpenConns int   `mapstructure:"max_open_conns"`
	MaxIdleConns int   `mapstructure:"max_idle_conns"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	LRU       LRUConfig       `mapstructure:"lru"`
	Semantic  SemanticConfig   `mapstructure:"semantic"`
}

// LRUConfig holds in-memory LRU cache config
type LRUConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	MaxEntries int  `mapstructure:"max_entries"`
}

// SemanticConfig holds semantic cache config
type SemanticConfig struct {
	Enabled               bool    `mapstructure:"enabled"`
	Provider              string  `mapstructure:"provider"`
	URL                   string  `mapstructure:"url"`
	Collection            string  `mapstructure:"collection"`
	SimilarityThreshold   float64 `mapstructure:"similarity_threshold"`
	TTL                   time.Duration `mapstructure:"ttl"`
	MaxCostSaving         float64 `mapstructure:"max_cost_saving"`
}

// ProviderConfig holds individual provider configuration
type ProviderConfig struct {
	APIKey            string        `mapstructure:"api_key"`
	Organization      string        `mapstructure:"organization"`
	BaseURL           string        `mapstructure:"base_url"`
	Timeout           time.Duration `mapstructure:"timeout"`
	MaxRetries        int           `mapstructure:"max_retries"`
	RetryDelay        time.Duration `mapstructure:"retry_delay"`
	ConnectionPoolSize int          `mapstructure:"connection_pool_size"`
}

// ProvidersConfig holds all provider configurations
type ProvidersConfig struct {
	OpenAI    ProviderConfig `mapstructure:"openai"`
	Anthropic ProviderConfig `mapstructure:"anthropic"`
	Google    ProviderConfig `mapstructure:"google"`
	Ollama    ProviderConfig `mapstructure:"ollama"`
}

// RoutingConfig holds routing configuration
type RoutingConfig struct {
	DefaultChain  string            `mapstructure:"default_chain"`
	Chains        map[string][]ChainStep `mapstructure:"chains"`
	Health        HealthConfig       `mapstructure:"health"`
	Latency       LatencyConfig      `mapstructure:"latency"`
	StickySession bool               `mapstructure:"sticky_session"`
}

// ChainStep represents a single step in a fallback chain
type ChainStep struct {
	Model            string `mapstructure:"model"`
	Provider         string `mapstructure:"provider"`
	MaxLatencyMs     int    `mapstructure:"max_latency_ms"`
	MaxRetries       int    `mapstructure:"max_retries"`
	FailOpen         bool   `mapstructure:"fail_open"`
	Weight           int    `mapstructure:"weight"`
	MaxContextTokens int    `mapstructure:"max_context_tokens"`
}

// HealthConfig holds health monitoring configuration
type HealthConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	Window           time.Duration `mapstructure:"window"`
	ErrorThreshold   float64       `mapstructure:"error_threshold"`
	LatencyThresholdMs int64       `mapstructure:"latency_threshold_ms"`
}

// LatencyConfig holds latency-aware routing config
type LatencyConfig struct {
	Enabled         bool `mapstructure:"enabled"`
	P99ThresholdMs  int64 `mapstructure:"p99_threshold_ms"`
	SampleSize      int   `mapstructure:"sample_size"`
}

// KeysConfig holds API key configuration
type KeysConfig struct {
	BootstrapKey string `mapstructure:"bootstrap_key"`
	Storage      string `mapstructure:"storage"`
	MaxPerOrg    int    `mapstructure:"max_per_org"`
}

// RateLimitsConfig holds rate limiting configuration
type RateLimitsConfig struct {
	DefaultRPM        int           `mapstructure:"default_rpm"`
	DefaultTPM        int           `mapstructure:"default_tpm"`
	EnforceOnStreaming bool         `mapstructure:"enforce_on_streaming"`
	QueueRequests     bool          `mapstructure:"queue_requests"`
	QueueTimeout      time.Duration `mapstructure:"queue_timeout"`
}

// CostsConfig holds cost tracking configuration
type CostsConfig struct {
	Models map[string]ModelCost `mapstructure:"models"`
}

// ModelCost holds cost per million tokens for a model
type ModelCost struct {
	Input  float64 `mapstructure:"input"`
	Output float64 `mapstructure:"output"`
}

// TelemetryConfig holds telemetry configuration
type TelemetryConfig struct {
	LogRequests   bool           `mapstructure:"log_requests"`
	LogResponses  bool           `mapstructure:"log_responses"`
	Async         bool           `mapstructure:"async"`
	Prometheus    PrometheusConfig `mapstructure:"prometheus"`
	OTel          OTelConfig     `mapstructure:"otel"`
	Tracing       TracingConfig  `mapstructure:"tracing"`
}

// PrometheusConfig holds Prometheus telemetry config
type PrometheusConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// OTelConfig holds OpenTelemetry configuration
type OTelConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service_name"`
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	SampleRate         float64 `mapstructure:"sample_rate"`
	IncludeRequestBody bool    `mapstructure:"include_request_body"`
	IncludeResponseBody bool   `mapstructure:"include_response_body"`
}

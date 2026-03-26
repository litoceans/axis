package router

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/godlabs/axis/internal/providers"
	"github.com/godlabs/axis/internal/telemetry"
	"github.com/godlabs/axis/pkg/types"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Router handles request routing with fallback chains
type Router struct {
	mu        sync.RWMutex
	chains    map[string][]types.ChainStep
	providers map[string]providers.Provider
	health    *HealthMonitor
	cache     *Cache
	config    types.RoutingConfig
	tracer    *telemetry.Tracer

	// Sticky session state: sessionID -> {provider, lastUsed}
	stickySessions map[string]*stickyEntry
	stickyMu       sync.RWMutex
}

type stickyEntry struct {
	provider  string
	model    string
	lastUsed time.Time
}

// New creates a new router
func New(config types.RoutingConfig, health *HealthMonitor, cache *Cache, tracer *telemetry.Tracer) *Router {
	return &Router{
		chains:         config.Chains,
		health:         health,
		cache:          cache,
		providers:      make(map[string]providers.Provider),
		config:         config,
		tracer:         tracer,
		stickySessions: make(map[string]*stickyEntry),
	}
}

// RegisterProvider registers a provider with the router
func (r *Router) RegisterProvider(name string, provider providers.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

// Route routes a chat request through the fallback chain
func (r *Router) Route(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	// Get the chain name from extra_body or use default
	chainName := r.getChainName(req.ExtraBody)

	chain, exists := r.chains[chainName]
	if !exists {
		// Fall back to default chain
		chain = r.chains[r.config.DefaultChain]
		if chain == nil {
			return nil, fmt.Errorf("no routing chain available")
		}
	}

	// Extract session ID for sticky sessions
	sessionID := r.getSessionID(req.ExtraBody)

	log.Debug().Str("chain", chainName).Str("model", req.Model).Str("session_id", sessionID).Msg("routing request")

	// Try sticky session first if enabled and session exists
	var stickyProvider string
	var stickyModel string
	if r.config.StickySession && sessionID != "" {
		if entry := r.getStickySession(sessionID); entry != nil {
			stickyProvider = entry.provider
			stickyModel = entry.model
			log.Debug().Str("session_id", sessionID).Str("provider", stickyProvider).Msg("using sticky session")
		}
	}

	// Try each provider in the chain
	var lastErr error
	for _, step := range chain {
		// Estimate input tokens for context length check
		inputTokens := r.estimateTokens(req.Messages)

		// Check context window size
		if step.MaxContextTokens > 0 && inputTokens > step.MaxContextTokens {
			log.Debug().Str("model", step.Model).Int("input_tokens", inputTokens).
				Int("max_context", step.MaxContextTokens).
				Msg("skipping model: input exceeds context window")
			continue
		}

		// Check if provider is available
		if r.health.ShouldAvoid(step.Provider) {
			log.Debug().Str("provider", step.Provider).Msg("provider marked as unhealthy, skipping")
			continue
		}

		// Sticky session: if we have a preferred provider and this step matches, use it
		if stickyProvider != "" && step.Provider != stickyProvider {
			log.Debug().Str("provider", step.Provider).Str("sticky_provider", stickyProvider).Msg("skipping non-sticky provider")
			continue
		}

		provider, exists := r.providers[step.Provider]
		if !exists {
			log.Warn().Str("provider", step.Provider).Msg("provider not registered")
			continue
		}

		// Set model from chain step if not specified
		model := step.Model
		if model == "" {
			model = req.Model
		}
		// Use sticky model if available
		if stickyModel != "" {
			model = stickyModel
		}
		req.Model = model

		// Execute request with timeout and tracing
		timeout := time.Duration(step.MaxLatencyMs) * time.Millisecond
		if timeout == 0 {
			timeout = 30 * time.Second
		}

		subCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Start provider span
		var providerSpan trace.Span
		if r.tracer != nil && r.tracer.IsEnabled() {
			subCtx, providerSpan = r.tracer.StartSpan(subCtx, fmt.Sprintf("axis.provider.%s", step.Provider),
				trace.WithAttributes(
					attribute.String("provider", step.Provider),
					attribute.String("model", model),
					attribute.Int("timeout_ms", int(timeout.Milliseconds())),
				),
			)
		}

		start := time.Now()
		resp, err := provider.Chat(subCtx, req)
		latencyMs := time.Since(start).Milliseconds()

		if err != nil {
			if providerSpan != nil {
				providerSpan.RecordError(err)
				providerSpan.SetStatus(codes.Error, "request failed")
				providerSpan.End()
			}
			log.Warn().Str("provider", step.Provider).Str("model", model).Err(err).Msg("provider request failed")

			// Record error
			r.health.Record(step.Provider, 0, latencyMs, subCtx.Err() == context.DeadlineExceeded)
			lastErr = err
			// Clear sticky session on failure so we try other providers
			if r.config.StickySession && sessionID != "" {
				r.clearStickySession(sessionID)
			}
			continue
		}

		if providerSpan != nil {
			providerSpan.SetAttributes(
				attribute.Int64("latency_ms", latencyMs),
				attribute.Int("input_tokens", resp.Usage.PromptTokens),
				attribute.Int("output_tokens", resp.Usage.CompletionTokens),
			)
			providerSpan.SetStatus(codes.Ok, "")
			providerSpan.End()
		}

		// Record success
		r.health.Record(step.Provider, 200, latencyMs, false)

		// Update response with actual provider
		resp.Provider = step.Provider

		// Record sticky session mapping on success
		if r.config.StickySession && sessionID != "" {
			r.setStickySession(sessionID, step.Provider, model)
		}

		log.Debug().Str("provider", step.Provider).Str("model", model).Int64("latency_ms", latencyMs).Msg("request succeeded")
		return resp, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all providers in chain failed: %w", lastErr)
	}
	return nil, fmt.Errorf("no providers available in chain")
}

// RouteStream routes a streaming request through the fallback chain
func (r *Router) RouteStream(ctx context.Context, req types.ChatRequest) (<-chan types.StreamChunk, <-chan error, string, error) {
	chainName := r.getChainName(req.ExtraBody)

	chain, exists := r.chains[chainName]
	if !exists {
		chain = r.chains[r.config.DefaultChain]
		if chain == nil {
			return nil, nil, "", fmt.Errorf("no routing chain available")
		}
	}

	// Extract session ID for sticky sessions
	sessionID := r.getSessionID(req.ExtraBody)

	// Try sticky session first if enabled and session exists
	var stickyProvider string
	var stickyModel string
	if r.config.StickySession && sessionID != "" {
		if entry := r.getStickySession(sessionID); entry != nil {
			stickyProvider = entry.provider
			stickyModel = entry.model
		}
	}

	var lastErr error
	for _, step := range chain {
		// Estimate input tokens for context length check
		inputTokens := r.estimateTokens(req.Messages)

		// Check context window size
		if step.MaxContextTokens > 0 && inputTokens > step.MaxContextTokens {
			log.Debug().Str("model", step.Model).Int("input_tokens", inputTokens).
				Int("max_context", step.MaxContextTokens).
				Msg("skipping model: input exceeds context window")
			continue
		}

		if r.health.ShouldAvoid(step.Provider) {
			continue
		}

		// Sticky session: if we have a preferred provider and this step matches, use it
		if stickyProvider != "" && step.Provider != stickyProvider {
			continue
		}

		provider, exists := r.providers[step.Provider]
		if !exists {
			continue
		}

		model := step.Model
		if model == "" {
			model = req.Model
		}
		if stickyModel != "" {
			model = stickyModel
		}
		req.Model = model

		timeout := time.Duration(step.MaxLatencyMs) * time.Millisecond
		if timeout == 0 {
			timeout = 30 * time.Second
		}

		subCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		chunks, errCh := provider.ChatStream(subCtx, req)

		// Record sticky session on stream start
		if r.config.StickySession && sessionID != "" {
			r.setStickySession(sessionID, step.Provider, model)
		}

		return chunks, errCh, step.Provider, nil
	}

	if lastErr != nil {
		return nil, nil, "", fmt.Errorf("all providers in chain failed: %w", lastErr)
	}
	return nil, nil, "", fmt.Errorf("no providers available in chain")
}

// getChainName extracts the chain name from extra_body
func (r *Router) getChainName(extraBody json.RawMessage) string {
	if len(extraBody) == 0 {
		return r.config.DefaultChain
	}

	var body map[string]interface{}
	if err := json.Unmarshal(extraBody, &body); err != nil {
		return r.config.DefaultChain
	}

	if hint, ok := body["axis_model_hint"].(string); ok {
		return hint
	}

	return r.config.DefaultChain
}

// getSessionID extracts the session ID from extra_body
func (r *Router) getSessionID(extraBody json.RawMessage) string {
	if len(extraBody) == 0 {
		return ""
	}

	var body map[string]interface{}
	if err := json.Unmarshal(extraBody, &body); err != nil {
		return ""
	}

	if sessionID, ok := body["axis_session_id"].(string); ok {
		return sessionID
	}

	return ""
}

// getStickySession retrieves a sticky session entry if valid (not expired)
func (r *Router) getStickySession(sessionID string) *stickyEntry {
	r.stickyMu.RLock()
	defer r.stickyMu.RUnlock()

	entry, exists := r.stickySessions[sessionID]
	if !exists {
		return nil
	}

	// Check TTL (30 minutes)
	if time.Since(entry.lastUsed) > 30*time.Minute {
		return nil
	}

	return entry
}

// setStickySession records a session -> provider mapping
func (r *Router) setStickySession(sessionID, provider, model string) {
	r.stickyMu.Lock()
	defer r.stickyMu.Unlock()

	r.stickySessions[sessionID] = &stickyEntry{
		provider:  provider,
		model:    model,
		lastUsed: time.Now(),
	}
}

// clearStickySession removes a sticky session entry (on failure)
func (r *Router) clearStickySession(sessionID string) {
	r.stickyMu.Lock()
	defer r.stickyMu.Unlock()

	delete(r.stickySessions, sessionID)
}

// estimateTokens estimates token count for a message list
func (r *Router) estimateTokens(messages []types.ChatMessage) int {
	// Rough estimate: ~4 characters per token for English
	// Plus overhead for roles/content markers
	total := 0
	for _, m := range messages {
		// Approximate: each message has ~4 tokens overhead + content/4
		total += 4 + len(m.Content)/4
	}
	return total
}

// RouteEmbed routes an embedding request to the appropriate provider
func (r *Router) RouteEmbed(ctx context.Context, req types.EmbedRequest) (*types.EmbedResponse, error) {
	// Route based on model prefix
	model := req.Model
	providerName := r.getEmbedProvider(model)

	provider, exists := r.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider not found for model: %s", model)
	}

	resp, err := provider.Embed(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// getEmbedProvider determines which provider to use for an embedding model
func (r *Router) getEmbedProvider(model string) string {
	// Map model prefixes to providers
	switch {
	case containsAny(model, "text-embedding", "gpt-", "o1", "o3"):
		return "openai"
	case containsAny(model, "claude", "anthropic"):
		return "anthropic"
	case containsAny(model, "gemini", "embedding"):
		return "google"
	default:
		// Default to ollama for local models
		return "ollama"
	}
}

func containsAny(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if len(s) >= len(p) && s[:len(p)] == p {
			return true
		}
	}
	return false
}

// Close closes all providers
func (r *Router) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, provider := range r.providers {
		if err := provider.Close(); err != nil {
			log.Warn().Err(err).Msg("error closing provider")
		}
	}
	return nil
}

// GetChains returns the configured chains
func (r *Router) GetChains() map[string][]types.ChainStep {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]types.ChainStep)
	for k, v := range r.chains {
		result[k] = v
	}
	return result
}

// Health returns the health monitor
func (r *Router) Health() *HealthMonitor {
	return r.health
}

// Cache returns the cache instance
func (r *Router) Cache() *Cache {
	return r.cache
}

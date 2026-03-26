package gateway

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/godlabs/axis/internal/cache"
	"github.com/godlabs/axis/internal/cost"
	"github.com/godlabs/axis/internal/ratelimit"
	"github.com/godlabs/axis/internal/router"
	"github.com/godlabs/axis/internal/storage"
	"github.com/godlabs/axis/internal/telemetry"
	"github.com/godlabs/axis/pkg/types"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Handler handles HTTP requests
type Handler struct {
	router        *router.Router
	storage       *storage.Storage
	rateLimiter   *ratelimit.Limiter
	costTracker   *cost.Tracker
	tracer        *telemetry.Tracer
	models        []types.ModelInfo
	metrics       *Metrics
	semanticCache *cache.SemanticCache
}

// Metrics holds Prometheus metrics
type Metrics struct {
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	tokensTotal     *prometheus.CounterVec
	costTotal       *prometheus.CounterVec
	cacheHits       *prometheus.CounterVec
	cacheMisses     *prometheus.CounterVec
	providerErrors  *prometheus.CounterVec
	activeRequests  *prometheus.GaugeVec
	rateLimitHits   *prometheus.CounterVec
}

// NewHandler creates a new handler
func NewHandler(r *router.Router, s *storage.Storage, rl *ratelimit.Limiter, ct *cost.Tracker, tr *telemetry.Tracer) *Handler {
	m := &Metrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "axis_requests_total",
				Help: "Total number of requests",
			},
			[]string{"model", "provider", "status_code"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "axis_request_duration_seconds",
				Help:    "Request duration in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"model", "provider", "cached"},
		),
		tokensTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "axis_tokens_total",
				Help: "Total number of tokens",
			},
			[]string{"model", "provider", "type"},
		),
		costTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "axis_cost_total_usd",
				Help: "Total cost in USD",
			},
			[]string{"model", "provider"},
		),
		cacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "axis_cache_hits_total",
				Help: "Total cache hits",
			},
			[]string{"model"},
		),
		cacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "axis_cache_misses_total",
				Help: "Total cache misses",
			},
			[]string{"model"},
		),
		providerErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "axis_provider_errors_total",
				Help: "Total provider errors",
			},
			[]string{"provider", "error_type"},
		),
		activeRequests: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "axis_active_requests",
				Help: "Number of active requests",
			},
			[]string{"model"},
		),
		rateLimitHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "axis_rate_limit_hits_total",
				Help: "Total rate limit hits",
			},
			[]string{"key_id"},
		),
	}

	// Default models list
	models := []types.ModelInfo{
		{ID: "gpt-4o", Object: "model", Created: 1715727360, OwnedBy: "openai", Provider: "openai", MaxContextTokens: 128000},
		{ID: "gpt-4o-mini", Object: "model", Created: 1715727360, OwnedBy: "openai", Provider: "openai", MaxContextTokens: 128000},
		{ID: "gpt-4-turbo", Object: "model", Created: 1712361441, OwnedBy: "openai", Provider: "openai", MaxContextTokens: 128000},
		{ID: "claude-3-5-sonnet", Object: "model", Created: 1712361441, OwnedBy: "anthropic", Provider: "anthropic", MaxContextTokens: 200000},
		{ID: "claude-3-5-haiku", Object: "model", Created: 1712361441, OwnedBy: "anthropic", Provider: "anthropic", MaxContextTokens: 200000},
		{ID: "gemini-1.5-pro", Object: "model", Created: 1712361441, OwnedBy: "google", Provider: "google", MaxContextTokens: 1000000},
		{ID: "gemini-1.5-flash", Object: "model", Created: 1712361441, OwnedBy: "google", Provider: "google", MaxContextTokens: 1000000},
		{ID: "llama3.3", Object: "model", Created: 1712361441, OwnedBy: "ollama", Provider: "ollama", MaxContextTokens: 128000},
	}

	return &Handler{
		router:        r,
		storage:       s,
		rateLimiter:   rl,
		costTracker:   ct,
		tracer:        tr,
		models:        models,
		metrics:       m,
		semanticCache: nil, // Set via SetSemanticCache
	}
}

// SetSemanticCache sets the semantic cache on the handler
func (h *Handler) SetSemanticCache(sc *cache.SemanticCache) {
	h.semanticCache = sc
}

// HandleChatCompletions handles chat completion requests
func (h *Handler) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := uuid.New().String()
	
	// Start root span
	var span trace.Span
	ctx := r.Context()
	if h.tracer != nil && h.tracer.IsEnabled() {
		ctx, span = h.tracer.StartSpan(ctx, "axis.request",
			trace.WithAttributes(
				attribute.String("request_id", requestID),
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
			),
		)
		defer span.End()
	}
	ctx = context.WithValue(ctx, "request_id", requestID)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "failed to read request body")
		return
	}

	// Parse request
	var req types.ChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("failed to parse request: %v", err))
		return
	}

	// Validate request
	if len(req.Messages) == 0 {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "messages are required")
		return
	}

	// Authenticate
	keyHash := h.extractAPIKey(r)
	if keyHash == "" {
		h.writeError(w, http.StatusUnauthorized, "authentication_required", "API key is required")
		return
	}

	key, err := h.storage.GetKeyByHash(ctx, keyHash)
	if err != nil || key == nil {
		h.writeError(w, http.StatusUnauthorized, "invalid_api_key", "invalid API key")
		return
	}

	// Budget enforcement check
	if key.MonthlyBudgetUSD > 0 {
		spent, err := h.storage.GetKeySpend(ctx, key.ID)
		if err == nil && spent >= key.MonthlyBudgetUSD {
			h.writeError(w, http.StatusForbidden, "budget_exceeded",
				fmt.Sprintf("Monthly budget exceeded. Limit: $%.2f, Used: $%.2f", key.MonthlyBudgetUSD, spent))
			return
		}
	}

	// Rate limit check
	estimatedTokens := h.estimateTokens(req.Messages)
	if err := h.rateLimiter.Check(key.ID, estimatedTokens); err != nil {
		if rle, ok := err.(*ratelimit.RateLimitError); ok {
			h.metrics.rateLimitHits.WithLabelValues(key.ID).Inc()
			h.writeRateLimitError(w, rle)
			return
		}
	}

	// Cache lookup (non-streaming only)
	var cachedResp *types.ChatResponse
	var cacheHit bool
	if !req.Stream {
		cacheKey := h.router.Cache().GenerateKey(key.OrgID, req.Model, h.messagesToMap(req.Messages))
		if cachedRespBytes, costUSD, found := h.router.Cache().Get(ctx, cacheKey); found {
			if err := json.Unmarshal(cachedRespBytes, &cachedResp); err == nil {
				cachedResp.Cached = true
				cachedResp.Usage.CostUSD = costUSD
				cacheHit = true
				h.metrics.cacheHits.WithLabelValues(req.Model).Inc()
			}
		}
		if !cacheHit {
			h.metrics.cacheMisses.WithLabelValues(req.Model).Inc()
		}

		// Semantic cache lookup (only on LRU cache miss)
		if !cacheHit && h.semanticCache != nil && h.semanticCache.IsEnabled() {
			if semRespBytes, costUSD, found := h.semanticCache.Search(ctx, req.Messages, req.Model); found {
				if err := json.Unmarshal(semRespBytes, &cachedResp); err == nil {
					cachedResp.Cached = true
					cachedResp.Usage.CostUSD = costUSD
					cacheHit = true
					h.metrics.cacheHits.WithLabelValues(req.Model).Inc()
					log.Debug().Str("model", req.Model).Msg("semantic cache hit")
				}
			}
		}
	}

	// Route request with tracing
	h.metrics.activeRequests.WithLabelValues(req.Model).Inc()
	defer h.metrics.activeRequests.WithLabelValues(req.Model).Dec()

	var resp *types.ChatResponse
	var routeSpan trace.Span
	if h.tracer != nil && h.tracer.IsEnabled() {
		ctx, routeSpan = h.tracer.StartSpan(ctx, "axis.route",
			trace.WithAttributes(
				attribute.String("model", req.Model),
				attribute.Bool("stream", req.Stream),
				attribute.Bool("cached", cacheHit),
			),
		)
		defer routeSpan.End()
	}

	if cacheHit && cachedResp != nil {
		resp = cachedResp
	} else if req.Stream {
		// Streaming path
		chunks, errCh, provider, err := h.router.RouteStream(ctx, req)
		if err != nil {
			if routeSpan != nil {
				routeSpan.RecordError(err)
				routeSpan.SetStatus(codes.Error, "routing failed")
			}
			h.writeError(w, http.StatusBadGateway, "provider_error", err.Error())
			return
		}

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Stream response and aggregate tokens
		reportedInput, reportedOutput, reportedCost := h.streamResponse(w, chunks, errCh, req.Model, provider, start)

		// Calculate cost from aggregated tokens
		calculatedCost := h.costTracker.CalculateCost(req.Model, reportedInput, reportedOutput)

		// Build response for logging with aggregated usage
		resp := &types.ChatResponse{
			Model:    req.Model,
			Provider: provider,
			Usage: types.ChatUsage{
				PromptTokens:     reportedInput,
				CompletionTokens: reportedOutput,
				CostUSD:          reportedCost,
			},
		}

		// Log usage with reconciliation data
		go h.logUsageWithReconciliation(ctx, requestID, key, req.Model, resp, start, 200, false,
			reportedInput, reportedOutput, reportedInput, reportedOutput, reportedCost, calculatedCost)

		// Update metrics
		h.metrics.requestsTotal.WithLabelValues(req.Model, provider, "200").Inc()
		h.metrics.requestDuration.WithLabelValues(req.Model, provider, "false").Observe(time.Since(start).Seconds())
		h.metrics.tokensTotal.WithLabelValues(req.Model, provider, "input").Add(float64(reportedInput))
		h.metrics.tokensTotal.WithLabelValues(req.Model, provider, "output").Add(float64(reportedOutput))
		h.metrics.costTotal.WithLabelValues(req.Model, provider).Add(calculatedCost)

		if routeSpan != nil {
			routeSpan.SetAttributes(attribute.String("provider", provider))
		}
		return
	} else {
		// Non-streaming path
		var err error
		resp, err = h.router.Route(ctx, req)
		if err != nil {
			if routeSpan != nil {
				routeSpan.RecordError(err)
				routeSpan.SetStatus(codes.Error, "routing failed")
			}
			h.metrics.providerErrors.WithLabelValues(req.Model, "request_failed").Inc()
			h.writeError(w, http.StatusBadGateway, "provider_error", err.Error())
			return
		}
	}

	// Calculate cost
	if resp != nil && !cacheHit {
		resp.Usage.CostUSD = h.costTracker.UpdateFromResponse(resp.Model, resp.Usage)
	}

	// Log usage asynchronously and check budget alerts
	go func() {
		h.logUsage(ctx, requestID, key, req.Model, resp, start, 200, false)
		h.checkBudgetAlerts(ctx, key)
	}()

	// Update metrics
	h.metrics.requestsTotal.WithLabelValues(req.Model, resp.Provider, "200").Inc()
	h.metrics.requestDuration.WithLabelValues(req.Model, resp.Provider, strconv.FormatBool(cacheHit)).Observe(time.Since(start).Seconds())
	if resp != nil {
		h.metrics.tokensTotal.WithLabelValues(resp.Model, resp.Provider, "input").Add(float64(resp.Usage.PromptTokens))
		h.metrics.tokensTotal.WithLabelValues(resp.Model, resp.Provider, "output").Add(float64(resp.Usage.CompletionTokens))
		h.metrics.costTotal.WithLabelValues(resp.Model, resp.Provider).Add(resp.Usage.CostUSD)
	}

	// Cache response (non-streaming, non-error)
	if !cacheHit && resp != nil && !req.Stream {
		cacheKey := h.router.Cache().GenerateKey(key.OrgID, req.Model, h.messagesToMap(req.Messages))
		if respBytes, err := json.Marshal(resp); err == nil {
			go h.router.Cache().Set(ctx, cacheKey, respBytes, resp.Usage.CostUSD)
			// Also store in semantic cache
			if h.semanticCache != nil && h.semanticCache.IsEnabled() {
				go h.semanticCache.Store(ctx, req.Messages, req.Model, respBytes, resp.Usage.CostUSD)
			}
		}
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleEmbeddings handles embedding requests
func (h *Handler) HandleEmbeddings(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := uuid.New().String()
	ctx := context.WithValue(r.Context(), "request_id", requestID)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "failed to read request body")
		return
	}

	// Parse request
	var req types.EmbedRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("failed to parse request: %v", err))
		return
	}

	// Authenticate
	keyHash := h.extractAPIKey(r)
	if keyHash == "" {
		h.writeError(w, http.StatusUnauthorized, "authentication_required", "API key is required")
		return
	}

	key, err := h.storage.GetKeyByHash(ctx, keyHash)
	if err != nil || key == nil {
		h.writeError(w, http.StatusUnauthorized, "invalid_api_key", "invalid API key")
		return
	}

	// Route request
	resp, err := h.router.RouteEmbed(ctx, req)
	if err != nil {
		h.writeError(w, http.StatusBadGateway, "provider_error", err.Error())
		return
	}

	// Calculate cost
	resp.Usage.CostUSD = h.costTracker.CalculateEmbeddingCost(resp.Model, resp.Usage.PromptTokens)

	// Update metrics
	h.metrics.requestsTotal.WithLabelValues(req.Model, "embeddings", "200").Inc()

	// Log usage asynchronously
	go h.logUsage(ctx, requestID, key, req.Model, nil, start, 200, false)

	// Write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleListModels handles listing available models
func (h *Handler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	resp := types.ModelsResponse{
		Object: "list",
		Data:   h.models,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleHealth handles health check requests
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	healthScores := h.router.Health().GetAllHealthScores()

	// Determine overall status
	overallStatus := "healthy"
	for _, score := range healthScores {
		if score.Status == "down" {
			overallStatus = "degraded"
		}
		if score.Status == "degraded" && overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}

	resp := map[string]interface{}{
		"status":   overallStatus,
		"providers": healthScores,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleListKeys handles listing API keys for an org
func (h *Handler) HandleListKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org ID from query param
	orgID := r.URL.Query().Get("orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "orgId is required")
		return
	}

	// List keys
	keys, err := h.storage.ListKeys(ctx, orgID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// Convert to response format
	response := make([]*types.APIKeyResponse, len(keys))
	for i, key := range keys {
		response[i] = h.keyToResponse(key)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"keys": response})
}

// HandleCreateKey handles creating a new API key
func (h *Handler) HandleCreateKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "failed to read request body")
		return
	}

	var req types.CreateKeyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("failed to parse request: %v", err))
		return
	}

	// Validate
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}

	// Generate raw key
	rawKey := "sk_axis_" + generateRandomKey(32)

	// Hash the key for storage
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	// Create key prefix (first 12 chars after sk_axis_)
	keyPrefix := rawKey[:15]

	// Parse expires at
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	// Set defaults
	if req.Environments == nil {
		req.Environments = []string{"production"}
	}

	// Create key object
	key := &storage.APIKey{
		ID:               uuid.New().String(),
		KeyHash:          keyHash,
		KeyPrefix:        keyPrefix,
		KeyName:          req.Name,
		OrgID:            req.OrgID,
		TeamID:           req.TeamID,
		Scopes:           req.Scopes,
		Models:           req.Models,
		RPMLimit:         req.RPMLimit,
		TPMLimit:         req.TPMLimit,
		MonthlyBudgetUSD: req.MonthlyBudgetUSD,
		Environments:     req.Environments,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
	}

	// Store key
	if err := h.storage.CreateKey(ctx, key); err != nil {
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// Return response with raw key (only time it's returned)
	resp := types.CreateKeyResponse{
		APIKey: h.keyToResponse(key),
		RawKey: rawKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleDeleteKey handles deleting/revoking an API key
func (h *Handler) HandleDeleteKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract key ID from path
	keyID := extractPathParam(r.URL.Path, "/v1/keys/")
	if keyID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "key_id is required")
		return
	}

	// Delete key
	if err := h.storage.DeleteKey(ctx, keyID); err != nil {
		if err.Error() == "key not found" {
			h.writeError(w, http.StatusNotFound, "not_found", "key not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleRotateKey handles rotating an API key
func (h *Handler) HandleRotateKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract key ID from path
	keyID := extractPathParam(r.URL.Path, "/v1/keys/")
	if keyID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "key_id is required")
		return
	}

	// Revoke old key
	if err := h.storage.RotateKey(ctx, keyID); err != nil {
		if err.Error() == "key not found" {
			h.writeError(w, http.StatusNotFound, "not_found", "key not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// HandleGetUsage handles getting usage data
func (h *Handler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query params
	orgID := r.URL.Query().Get("orgId")
	keyID := r.URL.Query().Get("keyId")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	granularity := r.URL.Query().Get("granularity")

	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "orgId is required")
		return
	}

	// Parse time range
	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		fromTime = time.Now().AddDate(0, 0, -30) // Default to last 30 days
	}

	toTime, err := time.Parse(time.RFC3339, to)
	if err != nil {
		toTime = time.Now()
	}

	// Get usage data
	var logs []*types.UsageLog
	if keyID != "" {
		logs, err = h.storage.GetUsage(ctx, keyID, orgID, fromTime, toTime)
	} else {
		logs, err = h.storage.GetUsageDetailed(ctx, orgID, fromTime, toTime)
	}

	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// Aggregate by date/granularity
	usageByDate := make(map[string]*types.UsageData)
	var totals struct {
		Requests  int
		TokensIn  int
		TokensOut int
		CostUsd   float64
	}

	for _, log := range logs {
		dateKey := log.CreatedAt.Format("2006-01-02")
		if granularity == "hour" {
			dateKey = log.CreatedAt.Format("2006-01-02T15:00:00Z")
		}

		if _, ok := usageByDate[dateKey]; !ok {
			usageByDate[dateKey] = &types.UsageData{Date: dateKey}
		}

		usageByDate[dateKey].Requests++
		usageByDate[dateKey].TokensIn += log.InputTokens
		usageByDate[dateKey].TokensOut += log.OutputTokens
		usageByDate[dateKey].CostUsd += log.CostUSD

		totals.Requests++
		totals.TokensIn += log.InputTokens
		totals.TokensOut += log.OutputTokens
		totals.CostUsd += log.CostUSD
	}

	// Build response
	usageSlice := make([]types.UsageData, 0, len(usageByDate))
	for _, data := range usageByDate {
		usageSlice = append(usageSlice, *data)
	}

	resp := types.UsageResponse{
		Usage: usageSlice,
	}
	resp.Totals.Requests = totals.Requests
	resp.Totals.TokensIn = totals.TokensIn
	resp.Totals.TokensOut = totals.TokensOut
	resp.Totals.CostUsd = totals.CostUsd

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleGetCosts handles getting cost breakdown
func (h *Handler) HandleGetCosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query params
	orgID := r.URL.Query().Get("orgId")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	by := r.URL.Query().Get("by") // "model" or "provider"

	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "orgId is required")
		return
	}

	// Parse time range
	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		fromTime = time.Now().AddDate(0, 0, -30)
	}

	toTime, err := time.Parse(time.RFC3339, to)
	if err != nil {
		toTime = time.Now()
	}

	// Get costs by model
	costs, err := h.storage.GetCostsByModel(ctx, orgID, fromTime, toTime)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// If grouping by provider, aggregate
	if by == "provider" {
		byProvider := make(map[string]*types.CostBreakdown)
		for _, c := range costs {
			if existing, ok := byProvider[c.Provider]; ok {
				existing.Requests += c.Requests
				existing.TokensIn += c.TokensIn
				existing.TokensOut += c.TokensOut
				existing.CostUsd += c.CostUsd
			} else {
				byProvider[c.Provider] = &types.CostBreakdown{
					Provider:   c.Provider,
					Model:      c.Provider, // Use provider as "model" name
					Requests:   c.Requests,
					TokensIn:   c.TokensIn,
					TokensOut:  c.TokensOut,
					CostUsd:    c.CostUsd,
				}
			}
		}

		costs = make([]*types.CostBreakdown, 0, len(byProvider))
		for _, c := range byProvider {
			costs = append(costs, c)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"costs": costs})
}

// HandleListChains handles listing routing chains
func (h *Handler) HandleListChains(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID := r.URL.Query().Get("orgId")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "orgId is required")
		return
	}

	chains, err := h.storage.ListRoutingChains(ctx, orgID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"chains": chains})
}

// HandleCreateChain handles creating a routing chain
func (h *Handler) HandleCreateChain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "failed to read request body")
		return
	}

	var req types.CreateChainRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("failed to parse request: %v", err))
		return
	}

	if req.Name == "" || len(req.Models) == 0 {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "name and models are required")
		return
	}

	// Get orgID from query or request
	orgID := r.URL.Query().Get("orgId")
	if orgID == "" {
		orgID = req.Models[0].Model // Fallback (not ideal but matches DB schema)
	}

	chain := &types.RoutingChain{
		ID:        uuid.New().String(),
		Name:      req.Name,
		IsDefault: req.IsDefault,
		Models:    req.Models,
	}

	if err := h.storage.CreateRoutingChain(ctx, chain); err != nil {
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chain)
}

// HandleGetCacheStats handles getting cache statistics
func (h *Handler) HandleGetCacheStats(w http.ResponseWriter, r *http.Request) {
	cache := h.router.Cache()
	if cache == nil {
		h.writeError(w, http.StatusServiceUnavailable, "cache_disabled", "cache is not enabled")
		return
	}

	size, maxSize, _, _ := cache.Stats()

	// For a more detailed hit rate, we'd need to track hits/misses separately
	// For now, return basic stats
	hitRate := 0.0
	if size > 0 {
		hitRate = float64(size) / float64(maxSize)
	}

	stats := types.CacheStats{
		HitRate:     hitRate,
		TotalHits:   size,
		TotalMisses: maxSize - size,
		Savings:     0, // Would need to track actual savings
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleGetBudget handles getting budget status for a key
func (h *Handler) HandleGetBudget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract key ID from path
	keyID := extractPathParam(r.URL.Path, "/v1/keys/")
	if keyID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "key_id is required")
		return
	}

	// Get the key
	key, err := h.storage.GetKeyByID(ctx, keyID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}
	if key == nil {
		h.writeError(w, http.StatusNotFound, "not_found", "key not found")
		return
	}

	// Get current month spend
	spent, err := h.storage.GetKeySpend(ctx, keyID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// Calculate remaining
	limit := key.MonthlyBudgetUSD
	remaining := limit - spent
	if remaining < 0 {
		remaining = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key_id":    keyID,
		"limit":     limit,
		"spent":     spent,
		"remaining": remaining,
		"month":     time.Now().UTC().Format("2006-01"),
	})
}

// HandleClearCache handles clearing the cache
func (h *Handler) HandleClearCache(w http.ResponseWriter, r *http.Request) {
	cache := h.router.Cache()
	if cache == nil {
		h.writeError(w, http.StatusServiceUnavailable, "cache_disabled", "cache is not enabled")
		return
	}

	cache.Clear(context.Background())

	w.WriteHeader(http.StatusNoContent)
}

// HandleListAlerts handles listing budget alerts
func (h *Handler) HandleListAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID := r.URL.Query().Get("orgId")
	keyID := r.URL.Query().Get("keyId")

	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "orgId is required")
		return
	}

	alerts, err := h.storage.ListBudgetAlerts(ctx, orgID, keyID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// Convert to response format
	response := make([]map[string]interface{}, len(alerts))
	for i, alert := range alerts {
		response[i] = map[string]interface{}{
			"id":                 alert.ID,
			"key_id":             alert.KeyID,
			"org_id":             alert.OrgID,
			"threshold_percent":  alert.ThresholdPercent,
			"spent_usd":          alert.SpentUSD,
			"limit_usd":          alert.LimitUSD,
			"triggered_at":       alert.TriggeredAt.Format(time.RFC3339),
			"acknowledged":       alert.Acknowledged,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"alerts": response})
}

// HandleAcknowledgeAlert handles acknowledging a budget alert
func (h *Handler) HandleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract alert ID from path
	alertID := extractPathParam(r.URL.Path, "/v1/alerts/")
	if alertID == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "alert_id is required")
		return
	}

	if err := h.storage.AcknowledgeAlert(ctx, alertID); err != nil {
		if err.Error() == "alert not found" {
			h.writeError(w, http.StatusNotFound, "not_found", "alert not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// keyToResponse converts a storage.APIKey to types.APIKeyResponse
func (h *Handler) keyToResponse(key *storage.APIKey) *types.APIKeyResponse {
	resp := &types.APIKeyResponse{
		ID:              key.ID,
		Name:            key.KeyName,
		KeyPrefix:       key.KeyPrefix,
		OrgID:           key.OrgID,
		TeamID:          key.TeamID,
		MemberID:        key.MemberID,
		Scopes:          key.Scopes,
		Models:          key.Models,
		RPMLimit:        key.RPMLimit,
		TPMLimit:        key.TPMLimit,
		MonthlyBudgetUSD: key.MonthlyBudgetUSD,
		Environments:    key.Environments,
		CreatedAt:       key.CreatedAt.Format(time.RFC3339),
	}

	if key.ExpiresAt != nil {
		resp.ExpiresAt = key.ExpiresAt.Format(time.RFC3339)
	}
	if !key.LastUsedAt.IsZero() {
		resp.LastUsedAt = key.LastUsedAt.Format(time.RFC3339)
	}
	if key.RevokedAt != nil {
		resp.RevokedAt = key.RevokedAt.Format(time.RFC3339)
	}

	return resp
}

// HandleMetrics handles Prometheus metrics
func (h *Handler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	// Prometheus handler is registered separately
	http.Error(w, "metrics endpoint should use promhttp.Handler()", http.StatusInternalServerError)
}

// streamResponse streams a response to the client and aggregates token counts
func (h *Handler) streamResponse(w http.ResponseWriter, chunks <-chan types.StreamChunk, errCh <-chan error, model, provider string, start time.Time) (int, int, float64) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return 0, 0, 0
	}

	var totalInputTokens, totalOutputTokens int
	var reportedCost float64

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				// Done streaming
				flusher.Flush()
				return totalInputTokens, totalOutputTokens, reportedCost
			}

			// Aggregate tokens from usage if present in chunk
			if chunk.Usage != nil {
				totalInputTokens += chunk.Usage.PromptTokens
				totalOutputTokens += chunk.Usage.CompletionTokens
				reportedCost = chunk.Usage.CostUSD
			}

			// Write SSE event
			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

		case err := <-errCh:
			if err != nil {
				log.Error().Err(err).Msg("stream error")
				fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
				flusher.Flush()
			}
			return totalInputTokens, totalOutputTokens, reportedCost
		}
	}
}

// logUsage logs request usage to storage
func (h *Handler) logUsage(ctx context.Context, requestID string, key *storage.APIKey, model string, resp *types.ChatResponse, start time.Time, statusCode int, isError bool) {
	if h.storage == nil {
		return
	}

	var inputTokens, outputTokens, cachedTokens int
	var costUSD float64
	var cached bool

	if resp != nil {
		inputTokens = resp.Usage.PromptTokens
		outputTokens = resp.Usage.CompletionTokens
		costUSD = resp.Usage.CostUSD
		cached = resp.Cached
		if cached {
			cachedTokens = inputTokens + outputTokens
		}
	}

	log := &types.UsageLog{
		ID:           requestID,
		KeyID:        key.ID,
		OrgID:        key.OrgID,
		Model:        model,
		Provider:     "openai",
		Endpoint:     "chat_completions",
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		CachedTokens: cachedTokens,
		CostUSD:      costUSD,
		LatencyMs:    int(time.Since(start).Milliseconds()),
		StatusCode:   statusCode,
		Cached:       cached,
		Error:        isError,
		CreatedAt:    time.Now(),
	}

	h.storage.LogUsage(ctx, log)
}

// logUsageWithReconciliation logs usage with token reconciliation data for streaming responses
func (h *Handler) logUsageWithReconciliation(ctx context.Context, requestID string, key *storage.APIKey, model string, resp *types.ChatResponse, start time.Time, statusCode int, isError bool,
	tokensInputReported, tokensOutputReported, tokensInputCalculated, tokensOutputCalculated int,
	costReported, costCalculated float64) {
	if h.storage == nil {
		return
	}

	var cachedTokens int
	var cached bool

	if resp != nil {
		cached = resp.Cached
		if cached {
			cachedTokens = tokensInputReported + tokensOutputReported
		}
	}

	// Check for discrepancy > 1%
	discrepancy := 0.0
	if costCalculated > 0 {
		discrepancy = abs(costReported-costCalculated) / costCalculated
	}
	if discrepancy > 0.01 {
		log.Warn().
			Float64("discrepancy_percent", discrepancy*100).
			Float64("cost_reported", costReported).
			Float64("cost_calculated", costCalculated).
			Str("model", model).
			Msg("cost reconciliation discrepancy detected")
	}

	logEntry := &types.UsageLog{
		ID:                     requestID,
		KeyID:                  key.ID,
		OrgID:                  key.OrgID,
		Model:                  model,
		Provider:               resp.Provider,
		Endpoint:               "chat_completions",
		InputTokens:            tokensInputReported,
		OutputTokens:           tokensOutputReported,
		CachedTokens:           cachedTokens,
		CostUSD:                costCalculated, // Use calculated cost as canonical
		LatencyMs:              int(time.Since(start).Milliseconds()),
		StatusCode:             statusCode,
		Cached:                 cached,
		Error:                  isError,
		TokensInputReported:    tokensInputReported,
		TokensOutputReported:   tokensOutputReported,
		TokensInputCalculated:  tokensInputCalculated,
		TokensOutputCalculated: tokensOutputCalculated,
		CostReported:           costReported,
		CostCalculated:         costCalculated,
		CreatedAt:              time.Now(),
	}

	h.storage.LogUsage(ctx, logEntry)
}

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// checkBudgetAlerts checks if budget thresholds have been crossed and creates alerts
func (h *Handler) checkBudgetAlerts(ctx context.Context, key *storage.APIKey) {
	if key.MonthlyBudgetUSD <= 0 {
		return // No budget limit set
	}

	thresholds := []float64{0.50, 0.80, 1.00} // 50%, 80%, 100%

	spent, err := h.storage.GetKeySpend(ctx, key.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get key spend for budget alert check")
		return
	}

	spentPercent := spent / key.MonthlyBudgetUSD

	for _, threshold := range thresholds {
		if spentPercent >= threshold {
			// Check if alert already exists for this threshold this month
			existing, err := h.storage.GetAlertForThreshold(ctx, key.ID, threshold*100)
			if err != nil {
				log.Error().Err(err).Float64("threshold", threshold*100).Msg("failed to check existing alert")
				continue
			}
			if existing != nil {
				continue // Alert already triggered this month
			}

			// Create new alert
			alert := &storage.BudgetAlert{
				ID:               uuid.New().String(),
				KeyID:            key.ID,
				OrgID:            key.OrgID,
				ThresholdPercent: threshold * 100,
				SpentUSD:         spent,
				LimitUSD:         key.MonthlyBudgetUSD,
				TriggeredAt:      time.Now(),
				Acknowledged:     false,
			}

			if err := h.storage.CreateBudgetAlert(ctx, alert); err != nil {
				log.Error().Err(err).Float64("threshold", threshold*100).Msg("failed to create budget alert")
				continue
			}

			log.Warn().
				Str("key_id", key.ID).
				Str("org_id", key.OrgID).
				Float64("threshold_percent", threshold*100).
				Float64("spent_usd", spent).
				Float64("limit_usd", key.MonthlyBudgetUSD).
				Msg("budget alert triggered")
		}
	}
}

// extractAPIKey extracts the API key from the Authorization header
func (h *Handler) extractAPIKey(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	// Hash the API key for lookup
	hash := sha256.Sum256([]byte(parts[1]))
	return hex.EncodeToString(hash[:])
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := types.ErrorResponse{
		Error: types.ErrorDetail{
			Type:    errType,
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// writeRateLimitError writes a rate limit error response
func (h *Handler) writeRateLimitError(w http.ResponseWriter, err *ratelimit.RateLimitError) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(err.RetryAfterMs/1000))
	w.WriteHeader(http.StatusTooManyRequests)

	resp := types.ErrorResponse{
		Error: types.ErrorDetail{
			Type:         "rate_limit_exceeded",
			Message:      fmt.Sprintf("rate limit exceeded. Retry after %dms", err.RetryAfterMs),
			Limit:        err.Limit,
			RetryAfterMs: err.RetryAfterMs,
			ResetAt:      strconv.FormatInt(err.ResetAt, 10),
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// estimateTokens estimates token count for a message
func (h *Handler) estimateTokens(messages []types.ChatMessage) int {
	// Rough estimate: 4 chars per token
	total := 0
	for _, m := range messages {
		total += len(m.Content) / 4
	}
	return total
}

// messagesToMap converts chat messages to map for caching
func (h *Handler) messagesToMap(messages []types.ChatMessage) []map[string]string {
	result := make([]map[string]string, len(messages))
	for i, m := range messages {
		result[i] = map[string]string{
			"role":    m.Role,
			"content": m.Content,
		}
	}
	return result
}

// generateRandomKey generates a random key string
func generateRandomKey(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	if _, err := rand.Read(result); err != nil {
		// Fallback to simpler random
		for i := range result {
			result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		}
	} else {
		for i, b := range result {
			result[i] = chars[int(b)%len(chars)]
		}
	}
	return hex.EncodeToString(result)[:length]
}

// extractPathParam extracts a path parameter from the URL path
func extractPathParam(path, prefix string) string {
	// Remove prefix and return the next path segment
	if len(path) > len(prefix) {
		remainder := path[len(prefix):]
		// Remove trailing slash if present
		if len(remainder) > 0 && remainder[0] == '/' {
			remainder = remainder[1:]
		}
		// Stop at next slash if present
		if idx := strings.Index(remainder, "/"); idx != -1 {
			remainder = remainder[:idx]
		}
		return remainder
	}
	return ""
}

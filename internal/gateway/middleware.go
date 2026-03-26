package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/godlabs/axis/internal/ratelimit"
	"github.com/godlabs/axis/internal/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Middleware provides HTTP middleware functions
type Middleware struct {
	storage     *storage.Storage
	rateLimiter *ratelimit.Limiter
	logger      zerolog.Logger
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(s *storage.Storage, rl *ratelimit.Limiter) *Middleware {
	return &Middleware{
		storage:     s,
		rateLimiter: rl,
		logger:      log.With().Str("component", "middleware").Logger(),
	}
}

// AuthMiddleware validates API keys
func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health and metrics endpoints
		if r.URL.Path == "/v1/health" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract API key
		auth := r.Header.Get("Authorization")
		if auth == "" {
			// Check for API key in query params (not recommended but allowed)
			apiKey := r.URL.Query().Get("api_key")
			if apiKey != "" {
				auth = "Bearer " + apiKey
			}
		}

		if auth == "" {
			writeJSONError(w, http.StatusUnauthorized, "authentication_required", "API key is required")
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeJSONError(w, http.StatusUnauthorized, "invalid_authorization", "Invalid authorization header format")
			return
		}

		apiKey := parts[1]

		// Hash the key for lookup
		hash := sha256.Sum256([]byte(apiKey))
		keyHash := hex.EncodeToString(hash[:])

		// Look up key in storage
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		key, err := m.storage.GetKeyByHash(ctx, keyHash)
		if err != nil {
			m.logger.Error().Err(err).Msg("failed to lookup API key")
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to validate API key")
			return
		}

		if key == nil {
			writeJSONError(w, http.StatusUnauthorized, "invalid_api_key", "invalid API key")
			return
		}

		// Check if key is expired
		if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			writeJSONError(w, http.StatusUnauthorized, "expired_api_key", "API key has expired")
			return
		}

		// Check if key is revoked
		if key.RevokedAt != nil {
			writeJSONError(w, http.StatusUnauthorized, "revoked_api_key", "API key has been revoked")
			return
		}

		// Add key to context
		ctx = context.WithValue(r.Context(), "api_key", key)
		ctx = context.WithValue(ctx, "api_key_id", key.ID)
		ctx = context.WithValue(ctx, "org_id", key.OrgID)

		// Update last used timestamp (async)
		go func() {
			if err := m.storage.UpdateKeyLastUsed(context.Background(), key.ID); err != nil {
				m.logger.Warn().Err(err).Str("key_id", key.ID).Msg("failed to update key last used")
			}
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimitMiddleware enforces rate limits
func (m *Middleware) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health and metrics endpoints
		if r.URL.Path == "/v1/health" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Get key from context (set by auth middleware)
		keyVal := r.Context().Value("api_key")
		if keyVal == nil {
			// No key in context, auth middleware should have rejected
			next.ServeHTTP(w, r)
			return
		}

		key := keyVal.(*storage.APIKey)

		// Get custom limits or use defaults
		rpmLimit := key.RPMLimit
		if rpmLimit == 0 {
			rpmLimit = 1000 // Default
		}

		tpmLimit := key.TPMLimit
		if tpmLimit == 0 {
			tpmLimit = 10000000 // Default
		}

		// Set custom limits for this key
		m.rateLimiter.SetLimits(key.ID, rpmLimit, tpmLimit)

		// Estimate tokens (simplified)
		estimatedTokens := estimateTokens(r)

		// Check rate limit
		if err := m.rateLimiter.Check(key.ID, estimatedTokens); err != nil {
			if rle, ok := err.(*ratelimit.RateLimitError); ok {
				m.logger.Warn().
					Str("key_id", key.ID).
					Int("retry_after_ms", rle.RetryAfterMs).
					Msg("rate limit exceeded")

				w.Header().Set("Retry-After", fmt.Sprintf("%d", rle.RetryAfterMs/1000))
				writeJSONError(w, http.StatusTooManyRequests, "rate_limit_exceeded",
					fmt.Sprintf("rate limit exceeded. Retry after %dms", rle.RetryAfterMs))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs requests
func (m *Middleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response wrapper to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		// Log request
		duration := time.Since(start)
		requestID := r.Context().Value("request_id")

		m.logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.statusCode).
			Dur("duration", duration).
			Str("request_id", fmt.Sprintf("%v", requestID)).
			Str("remote_addr", r.RemoteAddr).
			Msg("request")
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// estimateTokens estimates token count from request
func estimateTokens(r *http.Request) int {
	// This is a simplified estimate
	// In production, you'd want to parse the actual request body
	return 100 // Default estimate
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResp := map[string]interface{}{
		"error": map[string]string{
			"type":    errType,
			"message": message,
		},
	}

	json.NewEncoder(w).Encode(errorResp)
}

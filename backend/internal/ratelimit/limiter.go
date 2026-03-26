package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	mu       sync.RWMutex
	tokens   float64
	maxTokens float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	rpmLimit int
	tpmLimit int
	currentTokens int
	tokensUsed int
	windowStart time.Time
}

// Limiter manages rate limits for API keys
type Limiter struct {
	mu       sync.RWMutex
	buckets  map[string]*TokenBucket
	rpmLimit int
	tpmLimit int
	window   time.Duration
}

// New creates a new rate limiter
func New(rpmLimit, tpmLimit int) *Limiter {
	limiter := &Limiter{
		buckets:  make(map[string]*TokenBucket),
		rpmLimit: rpmLimit,
		tpmLimit: tpmLimit,
		window:   time.Minute,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// cleanup removes stale buckets periodically
func (l *Limiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, bucket := range l.buckets {
			bucket.mu.Lock()
			if now.Sub(bucket.windowStart) > 10*time.Minute && bucket.tokensUsed == 0 {
				delete(l.buckets, key)
			}
			bucket.mu.Unlock()
		}
		l.mu.Unlock()
	}
}

// getBucket retrieves or creates a bucket for an API key
func (l *Limiter) getBucket(key string) *TokenBucket {
	l.mu.Lock()
	defer l.mu.Unlock()

	if bucket, exists := l.buckets[key]; exists {
		return bucket
	}

	bucket := &TokenBucket{
		tokens:       float64(l.rpmLimit),
		maxTokens:    float64(l.rpmLimit),
		refillRate:   float64(l.rpmLimit) / 60.0, // per second
		lastRefill:   time.Now(),
		rpmLimit:     l.rpmLimit,
		tpmLimit:     l.tpmLimit,
		tokensUsed:   0,
		windowStart:  time.Now(),
	}

	l.buckets[key] = bucket
	return bucket
}

// Allow checks if a request is allowed
func (l *Limiter) Allow(key string, tokens int) (allowed bool, remaining int, retryAfterMs int) {
	bucket := l.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.lastRefill = now

	// Refill tokens based on elapsed time
	bucket.tokens += elapsed * bucket.refillRate
	if bucket.tokens > bucket.maxTokens {
		bucket.tokens = bucket.maxTokens
	}

	// Reset window if expired
	if now.Sub(bucket.windowStart) >= time.Minute {
		bucket.windowStart = now
		bucket.tokensUsed = 0
		bucket.tokens = float64(bucket.rpmLimit)
	}

	// Check if request is allowed
	if bucket.tokens >= 1 && bucket.tokensUsed < bucket.rpmLimit {
		bucket.tokens--
		bucket.tokensUsed++
		remaining = int(bucket.tokens)
		return true, remaining, 0
	}

	// Calculate retry after
	retryAfter := (1 - bucket.tokens) / bucket.refillRate
	if retryAfter < 0 {
		retryAfter = 0
	}

	return false, 0, int(retryAfter * 1000)
}

// CheckRPM checks if key is within RPM limit
func (l *Limiter) CheckRPM(key string) (allowed bool, remaining int, retryAfterMs int) {
	bucket := l.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()

	// Reset window if expired
	if now.Sub(bucket.windowStart) >= time.Minute {
		bucket.windowStart = now
		bucket.tokensUsed = 0
		bucket.tokens = float64(bucket.rpmLimit)
	}

	// Check RPM
	if bucket.tokensUsed < bucket.rpmLimit {
		return true, bucket.rpmLimit - bucket.tokensUsed, 0
	}

	// Calculate retry after
	retryAfter := time.Minute - now.Sub(bucket.windowStart)
	return false, 0, int(retryAfter.Milliseconds())
}

// CheckTPM checks if key is within TPM limit
func (l *Limiter) CheckTPM(key string, tokens int) (allowed bool, remaining int, retryAfterMs int) {
	bucket := l.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()

	// Reset window if expired
	if now.Sub(bucket.windowStart) >= time.Minute {
		bucket.windowStart = now
		bucket.tokensUsed = 0
	}

	// Check TPM
	if bucket.tokensUsed+tokens <= bucket.tpmLimit {
		bucket.tokensUsed += tokens
		return true, bucket.tpmLimit - bucket.tokensUsed, 0
	}

	// Calculate retry after
	retryAfter := time.Minute - now.Sub(bucket.windowStart)
	return false, 0, int(retryAfter.Milliseconds())
}

// RecordUsage records token usage for a key
func (l *Limiter) RecordUsage(key string, tokens int) {
	bucket := l.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.tokensUsed += tokens
}

// GetRemainingRPM returns remaining RPM for a key
func (l *Limiter) GetRemainingRPM(key string) int {
	bucket := l.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	remaining := bucket.rpmLimit - bucket.tokensUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetRemainingTPM returns remaining TPM for a key
func (l *Limiter) GetRemainingTPM(key string) int {
	bucket := l.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	remaining := bucket.tpmLimit - bucket.tokensUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// SetLimits updates rate limits for a specific key (for custom limits per key)
func (l *Limiter) SetLimits(key string, rpm, tpm int) {
	bucket := l.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.rpmLimit = rpm
	bucket.tpmLimit = tpm
	bucket.maxTokens = float64(rpm)
	bucket.refillRate = float64(rpm) / 60.0
}

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct {
	Limit       int   `json:"limit"`
	Current     int   `json:"current"`
	ResetAt     int64 `json:"reset_at"`
	RetryAfterMs int  `json:"retry_after_ms"`
}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded"
}

// Check checks both RPM and TPM limits
func (l *Limiter) Check(key string, tokens int) error {
	// Check RPM
	allowed, _, retryAfterMs := l.CheckRPM(key)
	if !allowed {
		return &RateLimitError{
			Limit:        l.rpmLimit,
			RetryAfterMs: retryAfterMs,
			ResetAt:      time.Now().Add(time.Duration(retryAfterMs) * time.Millisecond).Unix(),
		}
	}

	// Check TPM
	allowed, _, retryAfterMs = l.CheckTPM(key, tokens)
	if !allowed {
		return &RateLimitError{
			Limit:        l.tpmLimit,
			RetryAfterMs: retryAfterMs,
			ResetAt:      time.Now().Add(time.Duration(retryAfterMs) * time.Millisecond).Unix(),
		}
	}

	return nil
}

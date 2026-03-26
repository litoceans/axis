package router

import (
	"sync"
	"time"
)

// HealthMonitor tracks provider health metrics
type HealthMonitor struct {
	mu       sync.RWMutex
	requests map[string]*HealthWindow // keyed by provider
	config   HealthConfig
}

// HealthWindow maintains a rolling window of health metrics
type HealthWindow struct {
	mu         sync.RWMutex
	Requests   int64
	Errors     int64
	Timeouts   int64
	RateLimited int64
	TotalLatencyMs int64
	Latencies  []int64 // for percentile calculation
	windowSize time.Duration
	startTime  time.Time
}

// HealthConfig holds health monitoring configuration
type HealthConfig struct {
	Window           time.Duration
	ErrorThreshold   float64
	LatencyThresholdMs int64
}

// HealthScore represents a provider's health status
type HealthScore struct {
	Provider     string
	Status       string // "healthy", "degraded", "down"
	ErrorRate    float64
	P99LatencyMs int64
	Requests     int64
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(config HealthConfig) *HealthMonitor {
	if config.Window == 0 {
		config.Window = 5 * time.Minute
	}
	
	return &HealthMonitor{
		requests: make(map[string]*HealthWindow),
		config:   config,
	}
}

// GetWindow returns or creates a health window for a provider
func (h *HealthMonitor) GetWindow(provider string) *HealthWindow {
	h.mu.Lock()
	defer h.mu.Unlock()

	if window, exists := h.requests[provider]; exists {
		return window
	}

	window := &HealthWindow{
		windowSize: h.config.Window,
		startTime:  time.Now(),
	}
	h.requests[provider] = window
	return window
}

// RecordSuccess records a successful request
func (h *HealthMonitor) RecordSuccess(provider string, latencyMs int64) {
	window := h.GetWindow(provider)
	window.mu.Lock()
	defer window.mu.Unlock()

	window.Requests++
	window.TotalLatencyMs += latencyMs
	window.Latencies = append(window.Latencies, latencyMs)

	// Keep only last 1000 latencies for percentile calculation
	if len(window.Latencies) > 1000 {
		window.Latencies = window.Latencies[len(window.Latencies)-1000:]
	}

	// Cleanup old window
	h.cleanupWindow(provider)
}

// RecordError records an error
func (h *HealthMonitor) RecordError(provider string, isTimeout bool) {
	window := h.GetWindow(provider)
	window.mu.Lock()
	defer window.mu.Unlock()

	window.Requests++
	if isTimeout {
		window.Timeouts++
	} else {
		window.Errors++
	}

	// Cleanup old window
	h.cleanupWindow(provider)
}

// RecordRateLimited records a rate limited response
func (h *HealthMonitor) RecordRateLimited(provider string) {
	window := h.GetWindow(provider)
	window.mu.Lock()
	defer window.mu.Unlock()

	window.Requests++
	window.RateLimited++

	// Cleanup old window
	h.cleanupWindow(provider)
}

// cleanupWindow removes old entries outside the window
func (h *HealthMonitor) cleanupWindow(provider string) {
	// For simplicity, we reset the window periodically
	// In production, you'd want more precise time-based cleanup
	if time.Since(h.requests[provider].startTime) > h.config.Window {
		h.requests[provider].Requests = 0
		h.requests[provider].Errors = 0
		h.requests[provider].Timeouts = 0
		h.requests[provider].RateLimited = 0
		h.requests[provider].TotalLatencyMs = 0
		h.requests[provider].Latencies = nil
		h.requests[provider].startTime = time.Now()
	}
}

// GetHealthScore returns the health score for a provider
func (h *HealthMonitor) GetHealthScore(provider string) HealthScore {
	window := h.GetWindow(provider)
	window.mu.RLock()
	defer window.mu.RUnlock()

	score := HealthScore{
		Provider: provider,
		Requests: window.Requests,
	}

	if window.Requests == 0 {
		score.Status = "healthy"
		return score
	}

	// Calculate error rate
	errorRate := float64(window.Errors+window.Timeouts) / float64(window.Requests)
	score.ErrorRate = errorRate

	// Calculate P99 latency
	if len(window.Latencies) > 0 {
		score.P99LatencyMs = calculateP99(window.Latencies)
	}

	// Determine status
	if errorRate > 0.1 || score.P99LatencyMs > h.config.LatencyThresholdMs {
		score.Status = "down"
	} else if errorRate > h.config.ErrorThreshold || score.P99LatencyMs > h.config.LatencyThresholdMs/2 {
		score.Status = "degraded"
	} else {
		score.Status = "healthy"
	}

	return score
}

// GetAllHealthScores returns health scores for all providers
func (h *HealthMonitor) GetAllHealthScores() []HealthScore {
	h.mu.RLock()
	providers := make([]string, 0, len(h.requests))
	for p := range h.requests {
		providers = append(providers, p)
	}
	h.mu.RUnlock()

	scores := make([]HealthScore, len(providers))
	for i, p := range providers {
		scores[i] = h.GetHealthScore(p)
	}
	return scores
}

// ShouldAvoid returns true if provider should be avoided
func (h *HealthMonitor) ShouldAvoid(provider string) bool {
	score := h.GetHealthScore(provider)
	return score.Status == "down"
}

// calculateP99 calculates the 99th percentile latency
func calculateP99(latencies []int64) int64 {
	if len(latencies) == 0 {
		return 0
	}

	// Simple sorted approach for small arrays
	sorted := make([]int64, len(latencies))
	copy(sorted, latencies)
	
	// Sort
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Get 99th percentile
	index := int(float64(len(sorted)) * 0.99)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

// Record records a request outcome
func (h *HealthMonitor) Record(provider string, statusCode int, latencyMs int64, isTimeout bool) {
	switch {
	case statusCode == 429:
		h.RecordRateLimited(provider)
	case statusCode >= 500:
		h.RecordError(provider, false)
	case isTimeout:
		h.RecordError(provider, true)
	default:
		h.RecordSuccess(provider, latencyMs)
	}
}

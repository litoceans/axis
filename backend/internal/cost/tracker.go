package cost

import (
	"sync"

	"github.com/godlabs/axis/pkg/types"
)

// Tracker tracks token usage and calculates costs
type Tracker struct {
	mu     sync.RWMutex
	costs  map[string]ModelCost
}

// ModelCost represents cost per million tokens
type ModelCost struct {
	Input  float64
	Output float64
}

// New creates a new cost tracker
func New() *Tracker {
	return &Tracker{
		costs: defaultCosts(),
	}
}

// defaultCosts returns default costs for common models
func defaultCosts() map[string]ModelCost {
	return map[string]ModelCost{
		// OpenAI
		"gpt-4o":                 {Input: 2.50, Output: 10.00},
		"gpt-4o-mini":            {Input: 0.15, Output: 0.60},
		"gpt-4-turbo":            {Input: 10.00, Output: 30.00},
		"gpt-3.5-turbo":          {Input: 0.50, Output: 1.50},
		"gpt-4":                  {Input: 30.00, Output: 60.00},
		"text-embedding-3-small":  {Input: 0.02, Output: 0},
		"text-embedding-3-large": {Input: 0.13, Output: 0},
		"text-embedding-ada-002": {Input: 0.10, Output: 0},

		// Anthropic
		"claude-3-5-sonnet":  {Input: 3.00, Output: 15.00},
		"claude-3-5-haiku":   {Input: 0.80, Output: 4.00},
		"claude-3-opus":      {Input: 15.00, Output: 75.00},
		"claude-3-sonnet":    {Input: 3.00, Output: 15.00},
		"claude-3-haiku":     {Input: 0.25, Output: 1.25},
		"claude-sonnet-4":    {Input: 3.00, Output: 15.00},

		// Google
		"gemini-1.5-pro":  {Input: 1.25, Output: 5.00},
		"gemini-1.5-flash": {Input: 0.075, Output: 0.30},
		"gemini-2.0-flash": {Input: 0.00, Output: 0.00}, // Free
		"gemini-1.0-pro":  {Input: 1.25, Output: 5.00},

		// Ollama (local models - free)
		"llama3.3":           {Input: 0, Output: 0},
		"llama3.2":           {Input: 0, Output: 0},
		"llama3.1":           {Input: 0, Output: 0},
		"mistral":            {Input: 0, Output: 0},
		"mixtral":            {Input: 0, Output: 0},
		"qwen2.5":            {Input: 0, Output: 0},
		"phi4":               {Input: 0, Output: 0},
		"deepseek-r1":        {Input: 0, Output: 0},
		"nomic-embed-text":   {Input: 0, Output: 0},

		// Cohere
		"command-r":        {Input: 3.50, Output: 15.00},
		"command-r-plus":   {Input: 3.50, Output: 15.00},
		"embed-english-v3": {Input: 0.10, Output: 0},
	}
}

// SetCost sets the cost for a model
func (t *Tracker) SetCost(model string, input, output float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.costs[model] = ModelCost{Input: input, Output: output}
}

// GetCost returns the cost for a model
func (t *Tracker) GetCost(model string) ModelCost {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.costs[model]
}

// CalculateCost calculates the cost for given token usage
func (t *Tracker) CalculateCost(model string, inputTokens, outputTokens int) float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cost, exists := t.costs[model]
	if !exists {
		// Default to a reasonable estimate
		return float64(inputTokens+outputTokens) / 1_000_000 * 1.0
	}

	totalCost := float64(inputTokens)/1_000_000*cost.Input + float64(outputTokens)/1_000_000*cost.Output
	return totalCost
}

// CalculateEmbeddingCost calculates cost for embeddings
func (t *Tracker) CalculateEmbeddingCost(model string, tokens int) float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cost, exists := t.costs[model]
	if !exists {
		return 0.00001 * float64(tokens) / 1000 // Default estimate
	}

	return float64(tokens) / 1_000_000 * cost.Input
}

// UpdateFromResponse updates costs from a provider response
func (t *Tracker) UpdateFromResponse(model string, usage types.ChatUsage) float64 {
	return t.CalculateCost(model, usage.PromptTokens, usage.CompletionTokens)
}

// SetCostsFromConfig sets costs from configuration
func (t *Tracker) SetCostsFromConfig(config map[string]types.ModelCost) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for model, cost := range config {
		t.costs[model] = ModelCost{
			Input:  cost.Input,
			Output: cost.Output,
		}
	}
}

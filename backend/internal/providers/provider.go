package providers

import (
	"context"

	"github.com/godlabs/axis/pkg/types"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Name returns the provider identifier
	Name() string
	
	// Chat performs a chat completion request
	Chat(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error)
	
	// ChatStream performs a streaming chat completion
	ChatStream(ctx context.Context, req types.ChatRequest) (<-chan types.StreamChunk, <-chan error)
	
	// Embed performs an embedding request
	Embed(ctx context.Context, req types.EmbedRequest) (*types.EmbedResponse, error)
	
	// Close releases provider resources
	Close() error
}

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	NameValue   string
	APIKey      string
	BaseURL     string
	Timeout     int
	MaxRetries  int
}

// Name returns the provider name
func (b *BaseProvider) Name() string {
	return b.NameValue
}

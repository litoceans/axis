package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/godlabs/axis/pkg/types"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	BaseProvider
	client *http.Client
}

// ModelInfo represents model metadata
type ModelInfo struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	ContextWindow    int     `json:"context_window"`
	MaxOutputTokens  int     `json:"max_output_tokens"`
	InputPriceUSD    float64 `json:"input_price_usd"`  // per 1M tokens
	OutputPriceUSD   float64 `json:"output_price_usd"` // per 1M tokens
	SupportsVision   bool    `json:"supports_vision"`
	SupportsFunction bool    `json:"supports_function"`
	KnowledgeCutoff  string  `json:"knowledge_cutoff"`
}

// OpenAIModels lists current OpenAI models (2025-2026)
var OpenAIModels = []ModelInfo{
	// GPT-4.1 Series (April 2025)
	{ID: "gpt-4.1", Name: "GPT-4.1", ContextWindow: 1048576, MaxOutputTokens: 32768, InputPriceUSD: 2.00, OutputPriceUSD: 8.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	{ID: "gpt-4.1-mini", Name: "GPT-4.1 Mini", ContextWindow: 1048576, MaxOutputTokens: 32768, InputPriceUSD: 0.40, OutputPriceUSD: 1.60, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	{ID: "gpt-4.1-nano", Name: "GPT-4.1 Nano", ContextWindow: 1048576, MaxOutputTokens: 32768, InputPriceUSD: 0.10, OutputPriceUSD: 0.40, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	
	// GPT-4o 2025 Series
	{ID: "gpt-4o-2025", Name: "GPT-4o 2025", ContextWindow: 128000, MaxOutputTokens: 16384, InputPriceUSD: 2.50, OutputPriceUSD: 10.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "gpt-4o-mini-2025", Name: "GPT-4o Mini 2025", ContextWindow: 128000, MaxOutputTokens: 16384, InputPriceUSD: 0.15, OutputPriceUSD: 0.60, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	
	// o3/o4 Reasoning Series (April 2025)
	{ID: "o3", Name: "o3", ContextWindow: 200000, MaxOutputTokens: 100000, InputPriceUSD: 10.00, OutputPriceUSD: 40.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	{ID: "o3-mini", Name: "o3 Mini", ContextWindow: 200000, MaxOutputTokens: 100000, InputPriceUSD: 1.10, OutputPriceUSD: 4.40, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	{ID: "o4-mini", Name: "o4 Mini", ContextWindow: 200000, MaxOutputTokens: 100000, InputPriceUSD: 1.10, OutputPriceUSD: 4.40, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	{ID: "o3-pro", Name: "o3 Pro", ContextWindow: 200000, MaxOutputTokens: 100000, InputPriceUSD: 15.00, OutputPriceUSD: 60.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	
	// GPT-5 Series (Late 2025/Early 2026)
	{ID: "gpt-5", Name: "GPT-5", ContextWindow: 400000, MaxOutputTokens: 128000, InputPriceUSD: 5.00, OutputPriceUSD: 20.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-09"},
	{ID: "gpt-5-mini", Name: "GPT-5 Mini", ContextWindow: 400000, MaxOutputTokens: 128000, InputPriceUSD: 1.00, OutputPriceUSD: 4.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-09"},
	{ID: "gpt-5-nano", Name: "GPT-5 Nano", ContextWindow: 400000, MaxOutputTokens: 128000, InputPriceUSD: 0.25, OutputPriceUSD: 1.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-09"},
	
	// GPT-5.4 Series (March 2026)
	{ID: "gpt-5.4", Name: "GPT-5.4", ContextWindow: 400000, MaxOutputTokens: 128000, InputPriceUSD: 5.00, OutputPriceUSD: 20.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2026-02"},
	{ID: "gpt-5.4-mini", Name: "GPT-5.4 Mini", ContextWindow: 400000, MaxOutputTokens: 128000, InputPriceUSD: 1.00, OutputPriceUSD: 4.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2026-02"},
	
	// Embedding Models
	{ID: "text-embedding-3-small", Name: "Text Embedding 3 Small", ContextWindow: 8191, MaxOutputTokens: 8191, InputPriceUSD: 0.02, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: false, KnowledgeCutoff: "2024-01"},
	{ID: "text-embedding-3-large", Name: "Text Embedding 3 Large", ContextWindow: 8191, MaxOutputTokens: 8191, InputPriceUSD: 0.13, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: false, KnowledgeCutoff: "2024-01"},
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, baseURL string, timeout time.Duration, maxRetries int) *OpenAIProvider {
	return &OpenAIProvider{
		BaseProvider: BaseProvider{
			NameValue:   "openai",
			APIKey:      apiKey,
			BaseURL:     baseURL,
			Timeout:     int(timeout.Seconds()),
			MaxRetries: maxRetries,
		},
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetModels returns the list of available OpenAI models
func (p *OpenAIProvider) GetModels() []ModelInfo {
	return OpenAIModels
}

// Chat performs a chat completion request
func (p *OpenAIProvider) Chat(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	url := p.BaseURL + "/chat/completions"
	
	openAIReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      false,
	}
	
	if req.MaxTokens > 0 {
		openAIReq["max_tokens"] = req.MaxTokens
	}
	if req.TopP > 0 {
		openAIReq["top_p"] = req.TopP
	}
	
	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}
	
	var chatResp types.ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	chatResp.Provider = "openai"
	return &chatResp, nil
}

// ChatStream performs a streaming chat completion
func (p *OpenAIProvider) ChatStream(ctx context.Context, req types.ChatRequest) (<-chan types.StreamChunk, <-chan error) {
	chunks := make(chan types.StreamChunk)
	errCh := make(chan error, 1)
	
	go func() {
		defer close(chunks)
		defer close(errCh)
		
		url := p.BaseURL + "/chat/completions"
		
		openAIReq := map[string]interface{}{
			"model":       req.Model,
			"messages":    req.Messages,
			"temperature": req.Temperature,
			"stream":      true,
		}
		
		if req.MaxTokens > 0 {
			openAIReq["max_tokens"] = req.MaxTokens
		}
		
		body, err := json.Marshal(openAIReq)
		if err != nil {
			errCh <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}
		
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			errCh <- fmt.Errorf("failed to create request: %w", err)
			return
		}
		
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
		httpReq.Header.Set("Content-Type", "application/json")
		
		resp, err := p.client.Do(httpReq)
		if err != nil {
			errCh <- fmt.Errorf("request failed: %w", err)
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
			return
		}
		
		reader := resp.Body
		decoder := json.NewDecoder(reader)
		
		for decoder.More() {
			var line []byte
			if err := decoder.Decode(&line); err != nil {
				if err == io.EOF {
					break
				}
				errCh <- fmt.Errorf("failed to decode stream: %w", err)
				return
			}
			
			// SSE format: "data: {...}"
			if len(line) < 6 || string(line[:6]) != "data: " {
				continue
			}
			
			data := line[6:]
			if string(data) == "[DONE]" {
				break
			}
			
			var chunk types.StreamChunk
			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}
			chunk.Provider = "openai"
			
			select {
			case chunks <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return chunks, errCh
}

// Embed performs an embedding request
func (p *OpenAIProvider) Embed(ctx context.Context, req types.EmbedRequest) (*types.EmbedResponse, error) {
	url := p.BaseURL + "/embeddings"
	
	embedReq := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}
	
	body, err := json.Marshal(embedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}
	
	var embedResp types.EmbedResponse
	if err := json.Unmarshal(respBody, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &embedResp, nil
}

// Close releases provider resources
func (p *OpenAIProvider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

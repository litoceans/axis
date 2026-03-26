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

// XAIProvider implements the Provider interface for XAI/Grok
type XAIProvider struct {
	BaseProvider
	client *http.Client
}

// XAIModels lists current XAI/Grok models (2025-2026)
var XAIModels = []ModelInfo{
	// Grok-2 Series
	{ID: "grok-2", Name: "Grok-2", ContextWindow: 131072, MaxOutputTokens: 16384, InputPriceUSD: 5.00, OutputPriceUSD: 15.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "grok-2-vision", Name: "Grok-2 Vision", ContextWindow: 131072, MaxOutputTokens: 16384, InputPriceUSD: 5.00, OutputPriceUSD: 15.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "grok-2-beta", Name: "Grok-2 Beta", ContextWindow: 131072, MaxOutputTokens: 16384, InputPriceUSD: 5.00, OutputPriceUSD: 15.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	
	// Grok-3 Series (Late 2025)
	{ID: "grok-3", Name: "Grok-3", ContextWindow: 256000, MaxOutputTokens: 32768, InputPriceUSD: 6.00, OutputPriceUSD: 18.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-09"},
	{ID: "grok-3-mini", Name: "Grok-3 Mini", ContextWindow: 128000, MaxOutputTokens: 16384, InputPriceUSD: 2.00, OutputPriceUSD: 6.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-09"},
	{ID: "grok-3-mini-fast", Name: "Grok-3 Mini Fast", ContextWindow: 128000, MaxOutputTokens: 16384, InputPriceUSD: 1.50, OutputPriceUSD: 4.50, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-09"},
	{ID: "grok-3-fast", Name: "Grok-3 Fast", ContextWindow: 256000, MaxOutputTokens: 32768, InputPriceUSD: 4.00, OutputPriceUSD: 12.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-09"},
	
	// Grok-4 Series (Early 2026)
	{ID: "grok-4", Name: "Grok-4", ContextWindow: 512000, MaxOutputTokens: 65536, InputPriceUSD: 8.00, OutputPriceUSD: 24.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-12"},
	{ID: "grok-4-mini", Name: "Grok-4 Mini", ContextWindow: 256000, MaxOutputTokens: 32768, InputPriceUSD: 3.00, OutputPriceUSD: 9.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-12"},
}

// NewXAIProvider creates a new XAI/Grok provider
func NewXAIProvider(apiKey, baseURL string, timeout time.Duration, maxRetries int) *XAIProvider {
	return &XAIProvider{
		BaseProvider: BaseProvider{
			NameValue:   "xai",
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

// GetModels returns the list of available XAI models
func (p *XAIProvider) GetModels() []ModelInfo {
	return XAIModels
}

// Chat performs a chat completion request
func (p *XAIProvider) Chat(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	url := p.BaseURL + "/v1/chat/completions"
	
	xaiReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      false,
	}
	
	if req.MaxTokens > 0 {
		xaiReq["max_tokens"] = req.MaxTokens
	}
	if req.TopP > 0 {
		xaiReq["top_p"] = req.TopP
	}
	
	body, err := json.Marshal(xaiReq)
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
	
	chatResp.Provider = "xai"
	return &chatResp, nil
}

// ChatStream performs a streaming chat completion
func (p *XAIProvider) ChatStream(ctx context.Context, req types.ChatRequest) (<-chan types.StreamChunk, <-chan error) {
	chunks := make(chan types.StreamChunk)
	errCh := make(chan error, 1)
	
	go func() {
		defer close(chunks)
		defer close(errCh)
		
		url := p.BaseURL + "/v1/chat/completions"
		
		xaiReq := map[string]interface{}{
			"model":       req.Model,
			"messages":    req.Messages,
			"temperature": req.Temperature,
			"stream":      true,
		}
		
		if req.MaxTokens > 0 {
			xaiReq["max_tokens"] = req.MaxTokens
		}
		
		body, err := json.Marshal(xaiReq)
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
			chunk.Provider = "xai"
			
			select {
			case chunks <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return chunks, errCh
}

// Embed performs an embedding request (not supported by XAI)
func (p *XAIProvider) Embed(ctx context.Context, req types.EmbedRequest) (*types.EmbedResponse, error) {
	return nil, fmt.Errorf("embeddings not supported by XAI provider")
}

// Close releases provider resources
func (p *XAIProvider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

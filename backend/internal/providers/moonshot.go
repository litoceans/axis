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

// MoonshotProvider implements the Provider interface for Moonshot/Kimi
type MoonshotProvider struct {
	BaseProvider
	client *http.Client
}

// MoonshotModels lists current Moonshot/Kimi models (2025-2026)
var MoonshotModels = []ModelInfo{
	// Kimi K2 Series
	{ID: "kimi-k2", Name: "Kimi K2", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.60, OutputPriceUSD: 2.40, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "kimi-k2-base", Name: "Kimi K2 Base", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.40, OutputPriceUSD: 1.60, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "kimi-k2-0711", Name: "Kimi K2 0711", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.60, OutputPriceUSD: 2.40, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-06"},
	
	// Kimi K2.5 Series (January 2026)
	{ID: "kimi-k2.5", Name: "Kimi K2.5", ContextWindow: 262144, MaxOutputTokens: 16384, InputPriceUSD: 0.60, OutputPriceUSD: 2.40, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-12"},
	{ID: "kimi-k2.5-0127", Name: "Kimi K2.5 0127", ContextWindow: 262144, MaxOutputTokens: 16384, InputPriceUSD: 0.60, OutputPriceUSD: 2.40, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-12"},
	
	// Kimi VL Series (Vision-Language)
	{ID: "kimi-vl", Name: "Kimi VL", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.80, OutputPriceUSD: 3.20, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-06"},
}

// NewMoonshotProvider creates a new Moonshot/Kimi provider
func NewMoonshotProvider(apiKey, baseURL string, timeout time.Duration, maxRetries int) *MoonshotProvider {
	return &MoonshotProvider{
		BaseProvider: BaseProvider{
			NameValue:   "moonshot",
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

// GetModels returns the list of available Moonshot models
func (p *MoonshotProvider) GetModels() []ModelInfo {
	return MoonshotModels
}

// Chat performs a chat completion request
func (p *MoonshotProvider) Chat(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	url := p.BaseURL + "/v1/chat/completions"
	
	moonshotReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      false,
	}
	
	if req.MaxTokens > 0 {
		moonshotReq["max_tokens"] = req.MaxTokens
	}
	if req.TopP > 0 {
		moonshotReq["top_p"] = req.TopP
	}
	
	body, err := json.Marshal(moonshotReq)
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
	
	chatResp.Provider = "moonshot"
	return &chatResp, nil
}

// ChatStream performs a streaming chat completion
func (p *MoonshotProvider) ChatStream(ctx context.Context, req types.ChatRequest) (<-chan types.StreamChunk, <-chan error) {
	chunks := make(chan types.StreamChunk)
	errCh := make(chan error, 1)
	
	go func() {
		defer close(chunks)
		defer close(errCh)
		
		url := p.BaseURL + "/v1/chat/completions"
		
		moonshotReq := map[string]interface{}{
			"model":       req.Model,
			"messages":    req.Messages,
			"temperature": req.Temperature,
			"stream":      true,
		}
		
		if req.MaxTokens > 0 {
			moonshotReq["max_tokens"] = req.MaxTokens
		}
		
		body, err := json.Marshal(moonshotReq)
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
			chunk.Provider = "moonshot"
			
			select {
			case chunks <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return chunks, errCh
}

// Embed performs an embedding request (not supported by Moonshot)
func (p *MoonshotProvider) Embed(ctx context.Context, req types.EmbedRequest) (*types.EmbedResponse, error) {
	return nil, fmt.Errorf("embeddings not supported by Moonshot provider")
}

// Close releases provider resources
func (p *MoonshotProvider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

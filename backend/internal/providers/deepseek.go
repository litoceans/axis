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

// DeepSeekProvider implements the Provider interface for DeepSeek
type DeepSeekProvider struct {
	BaseProvider
	client *http.Client
}

// DeepSeekModels lists current DeepSeek models (2025-2026)
var DeepSeekModels = []ModelInfo{
	// DeepSeek V3 Series
	{ID: "deepseek-v3", Name: "DeepSeek V3", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.27, OutputPriceUSD: 1.10, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "deepseek-v3.2", Name: "DeepSeek V3.2", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.27, OutputPriceUSD: 1.10, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-06"},
	{ID: "deepseek-v3.5", Name: "DeepSeek V3.5", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.40, OutputPriceUSD: 1.60, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-12"},
	
	// DeepSeek R1 Reasoning Series
	{ID: "deepseek-r1", Name: "DeepSeek R1", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.55, OutputPriceUSD: 2.19, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "deepseek-r1-distill", Name: "DeepSeek R1 Distill", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.30, OutputPriceUSD: 1.20, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "deepseek-r1-distill-llama", Name: "DeepSeek R1 Distill Llama", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.30, OutputPriceUSD: 1.20, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "deepseek-r1-distill-qwen", Name: "DeepSeek R1 Distill Qwen", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.30, OutputPriceUSD: 1.20, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	
	// DeepSeek Coder Series
	{ID: "deepseek-coder-v2", Name: "DeepSeek Coder V2", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.20, OutputPriceUSD: 0.80, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-12"},
	{ID: "deepseek-coder-v2.5", Name: "DeepSeek Coder V2.5", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.30, OutputPriceUSD: 1.20, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-06"},
}

// NewDeepSeekProvider creates a new DeepSeek provider
func NewDeepSeekProvider(apiKey, baseURL string, timeout time.Duration, maxRetries int) *DeepSeekProvider {
	return &DeepSeekProvider{
		BaseProvider: BaseProvider{
			NameValue:   "deepseek",
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

// GetModels returns the list of available DeepSeek models
func (p *DeepSeekProvider) GetModels() []ModelInfo {
	return DeepSeekModels
}

// Chat performs a chat completion request
func (p *DeepSeekProvider) Chat(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	url := p.BaseURL + "/v1/chat/completions"
	
	deepseekReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      false,
	}
	
	if req.MaxTokens > 0 {
		deepseekReq["max_tokens"] = req.MaxTokens
	}
	if req.TopP > 0 {
		deepseekReq["top_p"] = req.TopP
	}
	
	body, err := json.Marshal(deepseekReq)
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
	
	chatResp.Provider = "deepseek"
	return &chatResp, nil
}

// ChatStream performs a streaming chat completion
func (p *DeepSeekProvider) ChatStream(ctx context.Context, req types.ChatRequest) (<-chan types.StreamChunk, <-chan error) {
	chunks := make(chan types.StreamChunk)
	errCh := make(chan error, 1)
	
	go func() {
		defer close(chunks)
		defer close(errCh)
		
		url := p.BaseURL + "/v1/chat/completions"
		
		deepseekReq := map[string]interface{}{
			"model":       req.Model,
			"messages":    req.Messages,
			"temperature": req.Temperature,
			"stream":      true,
		}
		
		if req.MaxTokens > 0 {
			deepseekReq["max_tokens"] = req.MaxTokens
		}
		
		body, err := json.Marshal(deepseekReq)
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
			chunk.Provider = "deepseek"
			
			select {
			case chunks <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return chunks, errCh
}

// Embed performs an embedding request (not supported by DeepSeek)
func (p *DeepSeekProvider) Embed(ctx context.Context, req types.EmbedRequest) (*types.EmbedResponse, error) {
	return nil, fmt.Errorf("embeddings not supported by DeepSeek provider")
}

// Close releases provider resources
func (p *DeepSeekProvider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

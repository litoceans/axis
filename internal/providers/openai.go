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

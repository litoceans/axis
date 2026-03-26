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

// AnthropicProvider implements the Provider interface for Anthropic
type AnthropicProvider struct {
	BaseProvider
	client *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, baseURL string, timeout time.Duration, maxRetries int) *AnthropicProvider {
	return &AnthropicProvider{
		BaseProvider: BaseProvider{
			NameValue:   "anthropic",
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

// Anthropic has a different API format - convert to their format
func convertToAnthropicFormat(req types.ChatRequest) map[string]interface{} {
	// Extract system message if present
	var systemPrompt string
	var messages []map[string]string
	
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			messages = append(messages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}
	
	result := map[string]interface{}{
		"model": req.Model,
		"messages": messages,
	}
	
	if systemPrompt != "" {
		result["system"] = systemPrompt
	}
	
	if req.Temperature > 0 {
		result["temperature"] = req.Temperature
	}
	
	if req.MaxTokens > 0 {
		result["max_tokens"] = req.MaxTokens
	} else {
		result["max_tokens"] = 4096 // Anthropic requires this
	}
	
	return result
}

// Chat performs a chat completion request
func (p *AnthropicProvider) Chat(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	url := p.BaseURL + "/messages"
	
	anthropicReq := convertToAnthropicFormat(req)
	
	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("x-api-key", p.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
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
	
	var anthropicResp struct {
		ID        string `json:"id"`
		Type      string `json:"type"`
		Role      string `json:"role"`
		Content   []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	response := &types.ChatResponse{
		ID:      anthropicResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Provider: "anthropic",
		Choices: []types.ChatResponseChoice{
			{
				Index: 0,
				Message: types.ChatMessage{
					Role:    "assistant",
					Content: anthropicResp.Content[0].Text,
				},
				FinishReason: "stop",
			},
		},
		Usage: types.ChatUsage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}
	
	return response, nil
}

// ChatStream performs a streaming chat completion
func (p *AnthropicProvider) ChatStream(ctx context.Context, req types.ChatRequest) (<-chan types.StreamChunk, <-chan error) {
	chunks := make(chan types.StreamChunk)
	errCh := make(chan error, 1)
	
	go func() {
		defer close(chunks)
		defer close(errCh)
		
		url := p.BaseURL + "/messages"
		
		// Anthropic streaming requires "stream": true and uses text/event-stream
		anthropicReq := convertToAnthropicFormat(req)
		anthropicReq["stream"] = true
		
		body, err := json.Marshal(anthropicReq)
		if err != nil {
			errCh <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}
		
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			errCh <- fmt.Errorf("failed to create request: %w", err)
			return
		}
		
		httpReq.Header.Set("x-api-key", p.APIKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
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
		
		for {
			line := make([]byte, 0)
			for {
				b := make([]byte, 1)
				n, err := reader.Read(b)
				if err != nil {
					if err.Error() != "EOF" {
						errCh <- err
					}
					return
				}
				if n == 0 {
					continue
				}
				if b[0] == '\n' {
					break
				}
				line = append(line, b...)
			}
			
			if len(line) < 6 {
				continue
			}
			
			// SSE format
			if string(line[:6]) != "data: " {
				continue
			}
			
			data := string(line[6:])
			if data == "[DONE]" {
				break
			}
			
			var event struct {
				Type string `json:"type"`
				Index int `json:"index"`
				ContentBlock *struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content_block"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			}
			
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}
			
			chunk := types.StreamChunk{
				ID:      "anthropic-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   req.Model,
				Provider: "anthropic",
				Choices: []types.StreamChoice{
					{
						Index:        event.Index,
						Delta:        json.RawMessage(`{"content":"` + event.Delta.Text + `"}`),
						FinishReason: "",
					},
				},
			}
			
			select {
			case chunks <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return chunks, errCh
}

// Embed performs an embedding request (Anthropic doesn't support embeddings natively)
// This is a placeholder that returns an error
func (p *AnthropicProvider) Embed(ctx context.Context, req types.EmbedRequest) (*types.EmbedResponse, error) {
	return nil, fmt.Errorf("embeddings not supported by Anthropic provider")
}

// Close releases provider resources
func (p *AnthropicProvider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

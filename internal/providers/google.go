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

// GoogleProvider implements the Provider interface for Google AI
type GoogleProvider struct {
	BaseProvider
	client *http.Client
}

// NewGoogleProvider creates a new Google AI provider
func NewGoogleProvider(apiKey, baseURL string, timeout time.Duration, maxRetries int) *GoogleProvider {
	return &GoogleProvider{
		BaseProvider: BaseProvider{
			NameValue:   "google",
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
func (p *GoogleProvider) Chat(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	// Convert messages to Google format
	var contents []map[string]interface{}
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": msg.Content},
			},
		})
	}
	
	googleReq := map[string]interface{}{
		"contents": contents,
	}
	
	// Map model name to Google format
	modelPath := req.Model
	if req.Model == "gemini-1.5-pro" {
		modelPath = "gemini-1.5-pro"
	} else if req.Model == "gemini-1.5-flash" {
		modelPath = "gemini-1.5-flash"
	} else if req.Model == "gemini-2.0-flash" {
		modelPath = "gemini-2.0-flash"
	}
	
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.BaseURL, modelPath, p.APIKey)
	
	body, err := json.Marshal(googleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
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
	
	var googleResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}
	
	if err := json.Unmarshal(respBody, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	if len(googleResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}
	
	response := &types.ChatResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Provider: "google",
		Choices: []types.ChatResponseChoice{
			{
				Index: 0,
				Message: types.ChatMessage{
					Role:    "assistant",
					Content: googleResp.Candidates[0].Content.Parts[0].Text,
				},
				FinishReason: googleResp.Candidates[0].FinishReason,
			},
		},
		Usage: types.ChatUsage{
			PromptTokens:     googleResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: googleResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      googleResp.UsageMetadata.TotalTokenCount,
		},
	}
	
	return response, nil
}

// ChatStream performs a streaming chat completion
func (p *GoogleProvider) ChatStream(ctx context.Context, req types.ChatRequest) (<-chan types.StreamChunk, <-chan error) {
	chunks := make(chan types.StreamChunk)
	errCh := make(chan error, 1)
	
	go func() {
		defer close(chunks)
		defer close(errCh)
		
		var contents []map[string]interface{}
		for _, msg := range req.Messages {
			role := "user"
			if msg.Role == "assistant" {
				role = "model"
			}
			contents = append(contents, map[string]interface{}{
				"role": role,
				"parts": []map[string]string{
					{"text": msg.Content},
				},
			})
		}
		
		modelPath := req.Model
		if req.Model == "gemini-1.5-pro" {
			modelPath = "gemini-1.5-pro"
		} else if req.Model == "gemini-1.5-flash" {
			modelPath = "gemini-1.5-flash"
		} else if req.Model == "gemini-2.0-flash" {
			modelPath = "gemini-2.0-flash"
		}
		
		url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", p.BaseURL, modelPath, p.APIKey)
		
		googleReq := map[string]interface{}{
			"contents": contents,
		}
		
		body, err := json.Marshal(googleReq)
		if err != nil {
			errCh <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}
		
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			errCh <- fmt.Errorf("failed to create request: %w", err)
			return
		}
		
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
		
		index := 0
		for decoder.More() {
			var googleEvent struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
					FinishReason string `json:"finishReason"`
				} `json:"candidates"`
			}
			
			if err := decoder.Decode(&googleEvent); err != nil {
				if err.Error() == "EOF" {
					break
				}
				continue
			}
			
			if len(googleEvent.Candidates) == 0 {
				continue
			}
			
			text := ""
			if len(googleEvent.Candidates[0].Content.Parts) > 0 {
				text = googleEvent.Candidates[0].Content.Parts[0].Text
			}
			
			chunk := types.StreamChunk{
				ID:      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   req.Model,
				Provider: "google",
				Choices: []types.StreamChoice{
					{
						Index:        index,
						Delta:        json.RawMessage(`{"content":"` + text + `"}`),
						FinishReason: googleEvent.Candidates[0].FinishReason,
					},
				},
			}
			
			index++
			
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
func (p *GoogleProvider) Embed(ctx context.Context, req types.EmbedRequest) (*types.EmbedResponse, error) {
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("no input provided")
	}
	
	modelPath := "text-embedding-004"
	if req.Model == "embedding-001" {
		modelPath = "embedding-001"
	}
	
	url := fmt.Sprintf("%s/models/%s:embedContent?key=%s", p.BaseURL, modelPath, p.APIKey)
	
	googleReq := map[string]interface{}{
		"model": modelPath,
		"content": map[string]interface{}{
			"parts": []map[string]string{
				{"text": req.Input[0]},
			},
		},
	}
	
	body, err := json.Marshal(googleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
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
	
	var googleResp struct {
		Embedding struct {
			Values []float64 `json:"values"`
		} `json:"embedding"`
	}
	
	if err := json.Unmarshal(respBody, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	embeddings := make([]types.Embedding, len(req.Input))
	for i := range req.Input {
		embeddings[i] = types.Embedding{
			Object:    "embedding",
			Embedding: googleResp.Embedding.Values,
			Index:     i,
		}
	}
	
	return &types.EmbedResponse{
		Object: "list",
		Data:   embeddings,
		Model:  req.Model,
		Usage: types.EmbedUsage{
			PromptTokens: len(req.Input[0]),
		},
	}, nil
}

// Close releases provider resources
func (p *GoogleProvider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

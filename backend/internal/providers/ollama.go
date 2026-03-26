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

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	BaseProvider
	client *http.Client
}

// OllamaModels lists current Ollama models (2025-2026)
var OllamaModels = []ModelInfo{
	// Llama Series
	{ID: "llama3.3", Name: "Llama 3.3 70B", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-12"},
	{ID: "llama3.3-70b-instruct", Name: "Llama 3.3 70B Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-12"},
	{ID: "llama3.2", Name: "Llama 3.2", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-09"},
	{ID: "llama3.2-3b", Name: "Llama 3.2 3B", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-09"},
	{ID: "llama3.2-1b", Name: "Llama 3.2 1B", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-09"},
	{ID: "llama3.1", Name: "Llama 3.1 405B", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	{ID: "llama3.1-405b", Name: "Llama 3.1 405B Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	{ID: "llama3.1-70b", Name: "Llama 3.1 70B Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	{ID: "llama3.1-8b", Name: "Llama 3.1 8B Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	
	// Llama 4 Series (2025)
	{ID: "llama4", Name: "Llama 4 400B", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	{ID: "llama4-400b", Name: "Llama 4 400B Instruct", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	{ID: "llama4-70b", Name: "Llama 4 70B Instruct", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-03"},
	
	// Qwen Series
	{ID: "qwen3", Name: "Qwen 3", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-06"},
	{ID: "qwen3-72b", Name: "Qwen 3 72B", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-06"},
	{ID: "qwen3-32b", Name: "Qwen 3 32B", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2025-06"},
	{ID: "qwen2.5", Name: "Qwen 2.5 72B", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-12"},
	{ID: "qwen2.5-72b-instruct", Name: "Qwen 2.5 72B Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-12"},
	{ID: "qwen2.5-32b-instruct", Name: "Qwen 2.5 32B Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-12"},
	{ID: "qwen2.5-coder", Name: "Qwen 2.5 Coder 32B", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-12"},
	{ID: "qwen2.5-vl", Name: "Qwen 2.5 VL", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: true, SupportsFunction: true, KnowledgeCutoff: "2024-12"},
	
	// Mistral Series
	{ID: "mistral-large", Name: "Mistral Large 2", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "mistral-small", Name: "Mistral Small 3", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "mistral-nemo", Name: "Mistral Nemo", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	{ID: "codestral", Name: "Codestral", ContextWindow: 256000, MaxOutputTokens: 16384, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	
	// DeepSeek Series
	{ID: "deepseek-v3", Name: "DeepSeek V3", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "deepseek-r1", Name: "DeepSeek R1", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	{ID: "deepseek-r1-distill", Name: "DeepSeek R1 Distill", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2025-01"},
	
	// Phi Series
	{ID: "phi4", Name: "Phi-4", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-12"},
	{ID: "phi3.5", Name: "Phi-3.5 Mini", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	{ID: "phi3.5-mini", Name: "Phi-3.5 Mini Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	
	// Gemma Series
	{ID: "gemma2", Name: "Gemma 2 27B", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	{ID: "gemma2-27b", Name: "Gemma 2 27B Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	{ID: "gemma2-9b", Name: "Gemma 2 9B Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	{ID: "gemma2-2b", Name: "Gemma 2 2B Instruct", ContextWindow: 128000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-07"},
	
	// Other Models
	{ID: "mixtral", Name: "Mixtral 8x7B", ContextWindow: 32000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-01"},
	{ID: "mixtral-8x7b", Name: "Mixtral 8x7B Instruct", ContextWindow: 32000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-01"},
	{ID: "mixtral-8x22b", Name: "Mixtral 8x22B Instruct", ContextWindow: 64000, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: true, KnowledgeCutoff: "2024-04"},
	
	// Embedding Models
	{ID: "nomic-embed-text", Name: "Nomic Embed Text", ContextWindow: 8192, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: false, KnowledgeCutoff: "2024-01"},
	{ID: "mxbai-embed-large", Name: "mxbai-embed-large", ContextWindow: 8192, MaxOutputTokens: 8192, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: false, KnowledgeCutoff: "2024-01"},
	{ID: "all-minilm", Name: "all-MiniLM-L6-v2", ContextWindow: 512, MaxOutputTokens: 512, InputPriceUSD: 0.00, OutputPriceUSD: 0.00, SupportsVision: false, SupportsFunction: false, KnowledgeCutoff: "2023-01"},
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(apiKey, baseURL string, timeout time.Duration, maxRetries int) *OllamaProvider {
	return &OllamaProvider{
		BaseProvider: BaseProvider{
			NameValue:   "ollama",
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

// GetModels returns the list of available Ollama models
func (p *OllamaProvider) GetModels() []ModelInfo {
	return OllamaModels
}

// Chat performs a chat completion request
func (p *OllamaProvider) Chat(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	url := p.BaseURL + "/chat"
	
	// Convert to Ollama format
	var messages []map[string]string
	for _, msg := range req.Messages {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	
	ollamaReq := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
		"stream":   false,
	}
	
	if req.Temperature > 0 {
		ollamaReq["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		ollamaReq["options"] = map[string]int{
			"num_predict": req.MaxTokens,
		}
	}
	
	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if p.APIKey != "" && p.APIKey != "ollama" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
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
	
	var ollamaResp struct {
		Model       string `json:"model"`
		Message     struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Done bool `json:"done"`
		Usage struct {
			PromptEvalCount     int `json:"prompt_eval_count"`
			EvalCount           int `json:"eval_count"`
			PromptEvalDuration  int `json:"prompt_eval_duration"`
			EvalDuration        int `json:"eval_duration"`
		} `json:"usage"`
	}
	
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	response := &types.ChatResponse{
		ID:      fmt.Sprintf("ollama-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   ollamaResp.Model,
		Provider: "ollama",
		Choices: []types.ChatResponseChoice{
			{
				Index: 0,
				Message: types.ChatMessage{
					Role:    ollamaResp.Message.Role,
					Content: ollamaResp.Message.Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: types.ChatUsage{
			PromptTokens:     ollamaResp.Usage.PromptEvalCount,
			CompletionTokens: ollamaResp.Usage.EvalCount,
			TotalTokens:      ollamaResp.Usage.PromptEvalCount + ollamaResp.Usage.EvalCount,
		},
	}
	
	return response, nil
}

// ChatStream performs a streaming chat completion
func (p *OllamaProvider) ChatStream(ctx context.Context, req types.ChatRequest) (<-chan types.StreamChunk, <-chan error) {
	chunks := make(chan types.StreamChunk)
	errCh := make(chan error, 1)
	
	go func() {
		defer close(chunks)
		defer close(errCh)
		
		url := p.BaseURL + "/chat"
		
		var messages []map[string]string
		for _, msg := range req.Messages {
			messages = append(messages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
		
		ollamaReq := map[string]interface{}{
			"model":    req.Model,
			"messages": messages,
			"stream":   true,
		}
		
		if req.Temperature > 0 {
			ollamaReq["temperature"] = req.Temperature
		}
		
		body, err := json.Marshal(ollamaReq)
		if err != nil {
			errCh <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}
		
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			errCh <- fmt.Errorf("failed to create request: %w", err)
			return
		}
		
		if p.APIKey != "" && p.APIKey != "ollama" {
			httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
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
			var ollamaEvent struct {
				Model   string `json:"model"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				Done           bool `json:"done"`
				Complete       bool `json:"complete"`
			}
			
			if err := decoder.Decode(&ollamaEvent); err != nil {
				if err.Error() == "EOF" {
					break
				}
				continue
			}
			
			chunk := types.StreamChunk{
				ID:      fmt.Sprintf("ollama-%d", time.Now().UnixNano()),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   ollamaEvent.Model,
				Provider: "ollama",
				Choices: []types.StreamChoice{
					{
						Index: index,
						Delta: json.RawMessage(`{"content":"` + ollamaEvent.Message.Content + `"}`),
					},
				},
			}
			
			if ollamaEvent.Done {
				chunk.Choices[0].FinishReason = "stop"
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
func (p *OllamaProvider) Embed(ctx context.Context, req types.EmbedRequest) (*types.EmbedResponse, error) {
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("no input provided")
	}
	
	url := p.BaseURL + "/embeddings"
	
	ollamaReq := map[string]interface{}{
		"model": req.Model,
		"prompt": req.Input[0],
	}
	
	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if p.APIKey != "" && p.APIKey != "ollama" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
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
	
	var ollamaResp struct {
		Embedding []float64 `json:"embedding"`
	}
	
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	embeddings := make([]types.Embedding, len(req.Input))
	for i := range req.Input {
		embeddings[i] = types.Embedding{
			Object:    "embedding",
			Embedding: ollamaResp.Embedding,
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
func (p *OllamaProvider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

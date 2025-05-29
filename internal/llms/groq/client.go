package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dskuldeep/gateway/internal/llms"
	"github.com/dskuldeep/gateway/internal/types"
)

// Client implements the LLMClient interface for Groq
type Client struct {
	httpClient *http.Client
	models     []llms.Model
}

// NewClient creates a new Groq client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		models: []llms.Model{
			{
				ID:          "llama2-70b-4096",
				Provider:    types.ProviderGroq,
				Name:        "Llama2 70B",
				Description: "High-performance Llama2 70B model",
				MaxTokens:   4096,
				CostPer1K:   0.0007,
			},
			{
				ID:          "mixtral-8x7b-32768",
				Provider:    types.ProviderGroq,
				Name:        "Mixtral 8x7B",
				Description: "High-performance Mixtral 8x7B model",
				MaxTokens:   32768,
				CostPer1K:   0.00027,
			},
		},
	}
}

// Query implements the LLMClient interface
func (c *Client) Query(ctx context.Context, req llms.Request, apiKey string) (*llms.Response, error) {
	// Prepare request body
	body := map[string]interface{}{
		"model":       req.Model,
		"messages":    []map[string]string{{"role": "user", "content": req.Prompt}},
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}

	if len(req.Stop) > 0 {
		body["stop"] = req.Stop
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		ID      string `json:"id"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Create response
	return &llms.Response{
		ID:           result.ID,
		Provider:     req.Provider,
		Model:        req.Model,
		Text:         result.Choices[0].Message.Content,
		Usage: types.TokenUsage{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		},
		FinishReason: result.Choices[0].FinishReason,
		Metadata:     req.Metadata,
		Latency:      time.Since(start),
	}, nil
}

// GetModels implements the LLMClient interface
func (c *Client) GetModels() []llms.Model {
	return c.models
}

// GetProvider implements the LLMClient interface
func (c *Client) GetProvider() types.Provider {
	return types.ProviderGroq
} 
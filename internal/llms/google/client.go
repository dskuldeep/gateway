package google

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

// Client implements the LLMClient interface for Google AI Studio
type Client struct {
	httpClient *http.Client
	models     []llms.Model
}

// NewClient creates a new Google AI Studio client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		models: []llms.Model{
			{
				ID:          "gemini-pro",
				Provider:    types.ProviderGoogle,
				Name:        "Gemini Pro",
				Description: "Google's most capable model for text generation",
				MaxTokens:   32768,
				CostPer1K:   0.00025,
			},
			{
				ID:          "gemini-pro-vision",
				Provider:    types.ProviderGoogle,
				Name:        "Gemini Pro Vision",
				Description: "Google's model for text and image generation",
				MaxTokens:   32768,
				CostPer1K:   0.00025,
			},
		},
	}
}

// Query implements the LLMClient interface
func (c *Client) Query(ctx context.Context, req llms.Request, apiKey string) (*llms.Response, error) {
	// Prepare request body
	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{
						"text": req.Prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": req.MaxTokens,
			"temperature":     req.Temperature,
		},
	}

	if len(req.Stop) > 0 {
		body["generationConfig"].(map[string]interface{})["stopSequences"] = req.Stop
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", req.Model, apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		PromptFeedback struct {
			TokenCount struct {
				PromptTokens     int `json:"promptTokenCount"`
				CompletionTokens int `json:"completionTokenCount"`
				TotalTokens      int `json:"totalTokenCount"`
			} `json:"tokenCount"`
		} `json:"promptFeedback"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Create response
	return &llms.Response{
		ID:    fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Provider: req.Provider,
		Model:  req.Model,
		Text:   result.Candidates[0].Content.Parts[0].Text,
		Usage: types.TokenUsage{
			PromptTokens:     result.PromptFeedback.TokenCount.PromptTokens,
			CompletionTokens: result.PromptFeedback.TokenCount.CompletionTokens,
			TotalTokens:      result.PromptFeedback.TokenCount.TotalTokens,
		},
		FinishReason: result.Candidates[0].FinishReason,
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
	return types.ProviderGoogle
} 
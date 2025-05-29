package llms

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/dskuldeep/gateway/internal/types"
)

// Provider represents an LLM provider (e.g., Groq, Google)
type Provider string

const (
	ProviderGroq   Provider = "groq"
	ProviderGoogle Provider = "google"
	ProviderOpenAI   Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderMistral  Provider = "mistral"
)

// Model represents a specific model from a provider
type Model struct {
	ID          string        `json:"id"`
	Provider    types.Provider `json:"provider"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	MaxTokens   int          `json:"max_tokens"`
	CostPer1K   float64      `json:"cost_per_1k"`
}

// Request represents a query to an LLM
type Request struct {
	Provider    types.Provider      `json:"provider" binding:"required"`
	Model       string             `json:"model" binding:"required"`
	Prompt      string             `json:"prompt" binding:"required"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	Stop        []string           `json:"stop,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
}

// Response represents a response from an LLM
type Response struct {
	ID           string            `json:"id"`
	Provider     types.Provider    `json:"provider"`
	Model        string            `json:"model"`
	Text         string            `json:"text"`
	Usage        types.TokenUsage  `json:"usage"`
	FinishReason string            `json:"finish_reason"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Latency      time.Duration     `json:"latency"`
}

// TokenUsage tracks token usage for a request
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// APIKey represents an API key for an LLM provider
type APIKey struct {
	gorm.Model
	Provider    types.Provider `json:"provider" gorm:"not null"`
	Key         string        `json:"key" gorm:"not null"`
	Description string        `json:"description"`
	IsActive    bool          `json:"is_active" gorm:"default:true"`
	LastUsed    time.Time     `json:"last_used"`
	Usage       []KeyUsage    `json:"usage,omitempty" gorm:"foreignKey:APIKeyID"`
}

// KeyUsage tracks usage of an API key
type KeyUsage struct {
	gorm.Model
	APIKeyID     uint      `json:"api_key_id" gorm:"not null"`
	APIKey       APIKey    `json:"api_key,omitempty"`
	RequestCount int64     `json:"request_count"`
	TokenCount   int64     `json:"token_count"`
	Cost         float64   `json:"cost"`
	Timestamp    time.Time `json:"timestamp"`
}

// LLMClient defines the interface for LLM providers
type LLMClient interface {
	// Query sends a request to the LLM and returns the response
	Query(ctx context.Context, req Request, apiKey string) (*Response, error)
	
	// GetModels returns available models for this provider
	GetModels() []Model
	
	// GetProvider returns the provider type
	GetProvider() types.Provider
} 
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

// Service handles LLM operations
type Service struct {
	clients map[types.Provider]LLMClient
	models  map[string]Model
	db      *gorm.DB
}

// NewService creates a new LLM service
func NewService(db *gorm.DB) *Service {
	return &Service{
		clients: make(map[types.Provider]LLMClient),
		models:  make(map[string]Model),
		db:      db,
	}
}

// RegisterClient registers a new LLM client
func (s *Service) RegisterClient(client LLMClient) {
	s.clients[client.GetProvider()] = client
	for _, model := range client.GetModels() {
		s.models[model.ID] = model
	}
}

// GetClient returns the client for a specific provider
func (s *Service) GetClient(provider types.Provider) (LLMClient, bool) {
	client, ok := s.clients[provider]
	return client, ok
}

// GetModel returns model information by ID
func (s *Service) GetModel(id string) (Model, bool) {
	model, ok := s.models[id]
	return model, ok
}

// ListModels returns all available models
func (s *Service) ListModels() []Model {
	models := make([]Model, 0, len(s.models))
	for _, model := range s.models {
		models = append(models, model)
	}
	return models
}

// GetAPIKey returns an active API key for a provider
func (s *Service) GetAPIKey(provider types.Provider) (*APIKey, error) {
	var key APIKey
	err := s.db.Where("provider = ? AND is_active = ?", provider, true).
		Order("last_used ASC").
		First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// RecordAPIKeyUsage records usage of an API key
func (s *Service) RecordAPIKeyUsage(keyID uint, usage types.TokenUsage, cost float64) error {
	usage := KeyUsage{
		APIKeyID:     keyID,
		RequestCount: 1,
		TokenCount:   int64(usage.TotalTokens),
		Cost:         cost,
		Timestamp:    time.Now(),
	}
	return s.db.Create(&usage).Error
}

// UpdateAPIKeyLastUsed updates the last used timestamp of an API key
func (s *Service) UpdateAPIKeyLastUsed(keyID uint) error {
	return s.db.Model(&APIKey{}).Where("id = ?", keyID).
		Update("last_used", time.Now()).Error
} 
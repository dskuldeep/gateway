package llms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dskuldeep/gateway/internal/metrics"
	"github.com/dskuldeep/gateway/internal/orgs"
	"github.com/dskuldeep/gateway/internal/types"
	"gorm.io/gorm"
)

// QueryRequest represents the incoming request body
type QueryRequest struct {
	Provider    types.Provider     `json:"provider" binding:"required"`
	Model       string            `json:"model" binding:"required"`
	Prompt      string            `json:"prompt" binding:"required"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	Stop        []string          `json:"stop,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Service handles LLM requests and manages providers
type Service struct {
	clients    map[types.Provider]LLMClient
	models     map[string]Model
	orgService *orgs.Service
	db         *gorm.DB
	mu         sync.RWMutex
}

// NewService creates a new LLM service
func NewService(orgService *orgs.Service, db *gorm.DB) *Service {
	return &Service{
		clients:    make(map[types.Provider]LLMClient),
		models:     make(map[string]Model),
		orgService: orgService,
		db:         db,
	}
}

// RegisterClient registers a new LLM client
func (s *Service) RegisterClient(client LLMClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.GetProvider()] = client
	for _, model := range client.GetModels() {
		s.models[model.ID] = model
	}
}

// GetClient returns the client for a given provider
func (s *Service) GetClient(provider types.Provider) (LLMClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, ok := s.clients[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", provider)
	}
	return client, nil
}

// GetModel returns model information by ID
func (s *Service) GetModel(id string) (Model, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	model, ok := s.models[id]
	return model, ok
}

// ListModels returns all available models
func (s *Service) ListModels() []Model {
	s.mu.RLock()
	defer s.mu.RUnlock()
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
	keyUsage := KeyUsage{
		APIKeyID:     keyID,
		RequestCount: 1,
		TokenCount:   int64(usage.TotalTokens),
		Cost:         cost,
		Timestamp:    time.Now(),
	}
	return s.db.Create(&keyUsage).Error
}

// UpdateAPIKeyLastUsed updates the last used timestamp of an API key
func (s *Service) UpdateAPIKeyLastUsed(keyID uint) error {
	return s.db.Model(&APIKey{}).Where("id = ?", keyID).
		Update("last_used", time.Now()).Error
}

// Query sends a request to the appropriate LLM provider
func (s *Service) Query(ctx context.Context, orgID, projectID string, req Request) (*Response, error) {
	// Get API key for the provider
	apiKey, err := s.GetAPIKey(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Get client for the provider
	client, err := s.GetClient(req.Provider)
	if err != nil {
		return nil, err
	}

	// Send request
	start := time.Now()
	resp, err := client.Query(ctx, req, apiKey.Key)
	if err != nil {
		metrics.RecordLLMError(req.Provider, req.Model)
		return nil, fmt.Errorf("failed to query LLM: %w", err)
	}

	// Record metrics
	metrics.RecordLLMLatency(req.Provider, req.Model, time.Since(start).Seconds())
	metrics.RecordTokenUsage(req.Provider, req.Model, resp.Usage)

	return resp, nil
}

// RegisterRoutes registers LLM-related routes
func (s *Service) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/v1/llm")
	{
		v1.POST("/query", s.HandleQuery)
		v1.GET("/models", func(c *gin.Context) {
			c.JSON(http.StatusOK, s.ListModels())
		})
	}
}

// HandleQuery processes an LLM query request
func (s *Service) HandleQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get client for provider
	client, err := s.GetClient(req.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get model info
	model, ok := s.GetModel(req.Model)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model"})
		return
	}

	// Get API key for provider
	apiKey, err := s.GetAPIKey(req.Provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no API key available"})
		return
	}

	// Create LLM request
	llmReq := Request{
		Provider:    req.Provider,
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stop:        req.Stop,
		Stream:      req.Stream,
		Metadata:    req.Metadata,
	}

	// Add project and user info to metadata
	if llmReq.Metadata == nil {
		llmReq.Metadata = make(map[string]string)
	}
	llmReq.Metadata["project_id"] = c.GetString("project_id")
	llmReq.Metadata["user_id"] = c.GetString("user_id")

	// Record start time for latency tracking
	start := time.Now()

	// Make request to LLM
	resp, err := client.Query(c.Request.Context(), llmReq, apiKey.Key)
	if err != nil {
		metrics.RecordLLMError(req.Provider, req.Model)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Record metrics
	metrics.RecordLLMLatency(req.Provider, req.Model, time.Since(start).Seconds())
	metrics.RecordTokenUsage(req.Provider, req.Model, resp.Usage)

	// Record API key usage
	cost := float64(resp.Usage.TotalTokens) * model.CostPer1K / 1000
	if err := s.RecordAPIKeyUsage(apiKey.ID, resp.Usage, cost); err != nil {
		// Log error but don't fail the request
		c.Error(err)
	}

	// Update API key last used timestamp
	if err := s.UpdateAPIKeyLastUsed(apiKey.ID); err != nil {
		// Log error but don't fail the request
		c.Error(err)
	}

	// Return response
	c.JSON(http.StatusOK, resp)
} 
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
)

// QueryRequest represents the incoming request body
type QueryRequest struct {
	Provider    Provider          `json:"provider" binding:"required"`
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
	clients    map[Provider]Client
	orgService *orgs.Service
	metrics    *metrics.Metrics
	mu         sync.RWMutex
}

// NewService creates a new LLM service
func NewService(orgService *orgs.Service, metrics *metrics.Metrics) *Service {
	return &Service{
		clients:    make(map[Provider]Client),
		orgService: orgService,
		metrics:    metrics,
	}
}

// RegisterClient registers a new LLM client
func (s *Service) RegisterClient(client Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.GetProvider()] = client
}

// GetClient returns the client for a given provider
func (s *Service) GetClient(provider Provider) (Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, ok := s.clients[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", provider)
	}
	return client, nil
}

// Query sends a request to the appropriate LLM provider
func (s *Service) Query(ctx context.Context, orgID, projectID string, req Request) (*Response, error) {
	// Get organization and project
	org, err := s.orgService.GetOrganization(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	project, err := s.orgService.GetProject(orgID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Get API key for the provider
	apiKey, err := s.orgService.GetAPIKey(orgID, projectID, req.Provider)
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
	resp, err := client.Query(ctx, req, apiKey)
	if err != nil {
		s.metrics.RecordLLMError(req.Provider, req.Model)
		return nil, fmt.Errorf("failed to query LLM: %w", err)
	}

	// Record metrics
	s.metrics.RecordLLMLatency(req.Provider, req.Model, time.Since(start))
	s.metrics.RecordLLMTokens(req.Provider, req.Model, resp.Usage)

	// Update usage in organization
	if err := s.orgService.UpdateUsage(orgID, projectID, req.Provider, resp.Usage); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to update usage: %v\n", err)
	}

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
	client, ok := s.GetClient(req.Provider)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider not supported"})
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
	metrics.RecordLLMLatency(req.Provider, req.Model, time.Since(start))
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

// RegisterRoutes registers HTTP handlers for the LLM service
func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/llm/query", s.handleQuery)
	mux.HandleFunc("/api/v1/llm/models", s.handleGetModels)
}

func (s *Service) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get organization and project IDs from context (set by auth middleware)
	orgID := r.Context().Value("org_id").(string)
	projectID := r.Context().Value("project_id").(string)

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := s.Query(r.Context(), orgID, projectID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Service) handleGetModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	models := make(map[Provider][]Model)
	for provider, client := range s.clients {
		models[provider] = client.GetModels()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
} 
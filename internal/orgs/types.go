package orgs

import (
	"time"
	"github.com/dskuldeep/gateway/internal/types"
)

// Organization represents an organization in the system
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Project represents a project within an organization
type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// APIKey represents an API key for a specific LLM provider in a project
type APIKey struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ProjectID      string    `json:"project_id"`
	Provider       types.Provider `json:"provider"`
	APIKey         string    `json:"api_key"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UsageMetrics represents token usage metrics for a project
type UsageMetrics struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ProjectID      string    `json:"project_id"`
	Provider       types.Provider `json:"provider"`
	PromptTokens   int       `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens    int       `json:"total_tokens"`
	CreatedAt      time.Time `json:"created_at"`
}

// Member represents a member of an organization
type Member struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateOrganizationRequest represents a request to create an organization
type CreateOrganizationRequest struct {
	Name string `json:"name"`
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name string `json:"name"`
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Provider types.Provider `json:"provider"`
	APIKey   string        `json:"api_key"`
}

// UpdateOrganizationRequest represents a request to update an organization
type UpdateOrganizationRequest struct {
	Name string `json:"name"`
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Name string `json:"name"`
}

// ListProjectsRequest represents a request to list projects
type ListProjectsRequest struct {
	OrganizationID string `json:"organization_id"`
}

// ListAPIKeysRequest represents a request to list API keys
type ListAPIKeysRequest struct {
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
}

// DeleteAPIKeyRequest represents a request to delete an API key
type DeleteAPIKeyRequest struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
} 
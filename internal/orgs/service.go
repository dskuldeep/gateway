package orgs

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dskuldeep/gateway/internal/types"
	"gorm.io/gorm"
)

// Service handles organization and project operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new organization service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetAPIKeyForProvider retrieves an API key for a specific provider
func (s *Service) GetAPIKeyForProvider(orgID, projectID string, provider types.Provider) (string, error) {
	var key APIKey
	err := s.db.Where("organization_id = ? AND project_id = ? AND provider = ?", orgID, projectID, provider).
		First(&key).Error
	if err != nil {
		return "", err
	}
	return key.APIKey, nil
}

// UpdateUsage updates usage metrics for a project
func (s *Service) UpdateUsage(orgID, projectID string, provider types.Provider, usage types.TokenUsage) error {
	metrics := &UsageMetrics{
		OrganizationID:   orgID,
		ProjectID:        projectID,
		Provider:         provider,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		CreatedAt:        time.Now(),
	}
	return s.db.Create(metrics).Error
}

// RegisterRoutes registers HTTP handlers for the organization service
func (s *Service) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/v1")
	{
		// Organization routes
		orgs := v1.Group("/orgs")
		{
			orgs.POST("", s.CreateOrganization)
			orgs.GET("", s.ListOrganizations)
			orgs.GET("/:id", s.GetOrganization)
			orgs.PUT("/:id", s.UpdateOrganization)
			orgs.DELETE("/:id", s.DeleteOrganization)
		}

		// Project routes
		projects := v1.Group("/projects")
		{
			projects.POST("", s.CreateProject)
			projects.GET("", s.ListProjects)
			projects.GET("/:id", s.GetProject)
			projects.PUT("/:id", s.UpdateProject)
			projects.DELETE("/:id", s.DeleteProject)
		}

		// API key routes
		apiKeys := v1.Group("/api-keys")
		{
			apiKeys.POST("", s.CreateAPIKey)
			apiKeys.GET("", s.ListAPIKeys)
			apiKeys.DELETE("/:id", s.RevokeAPIKey)
		}
	}
}

// CreateOrganization creates a new organization
func (s *Service) CreateOrganization(c *gin.Context) {
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org := &Organization{
		Name: req.Name,
	}

	if err := s.db.Create(org).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		return
	}

	c.JSON(http.StatusCreated, org)
}

// ListOrganizations lists all organizations
func (s *Service) ListOrganizations(c *gin.Context) {
	var orgs []Organization
	if err := s.db.Find(&orgs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list organizations"})
		return
	}

	c.JSON(http.StatusOK, orgs)
}

// GetOrganization retrieves an organization by ID
func (s *Service) GetOrganization(c *gin.Context) {
	id := c.Param("id")
	var org Organization
	if err := s.db.First(&org, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// UpdateOrganization updates an organization
func (s *Service) UpdateOrganization(c *gin.Context) {
	id := c.Param("id")
	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var org Organization
	if err := s.db.First(&org, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	org.Name = req.Name
	if err := s.db.Save(&org).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// DeleteOrganization deletes an organization
func (s *Service) DeleteOrganization(c *gin.Context) {
	id := c.Param("id")
	if err := s.db.Delete(&Organization{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization"})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateProject creates a new project
func (s *Service) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := &Project{
		Name: req.Name,
	}

	if err := s.db.Create(project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// ListProjects lists all projects
func (s *Service) ListProjects(c *gin.Context) {
	var projects []Project
	if err := s.db.Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProject retrieves a project by ID
func (s *Service) GetProject(c *gin.Context) {
	id := c.Param("id")
	var project Project
	if err := s.db.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject updates a project
func (s *Service) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project Project
	if err := s.db.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	project.Name = req.Name
	if err := s.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject deletes a project
func (s *Service) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if err := s.db.Delete(&Project{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateAPIKey creates a new API key
func (s *Service) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := &APIKey{
		Provider: req.Provider,
		APIKey:   req.APIKey,
	}

	if err := s.db.Create(key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, key)
}

// ListAPIKeys lists all API keys
func (s *Service) ListAPIKeys(c *gin.Context) {
	var keys []APIKey
	if err := s.db.Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list API keys"})
		return
	}

	c.JSON(http.StatusOK, keys)
}

// RevokeAPIKey revokes an API key
func (s *Service) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")
	if err := s.db.Delete(&APIKey{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke API key"})
		return
	}

	c.Status(http.StatusNoContent)
} 
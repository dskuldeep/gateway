package orgs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dskuldeep/gateway/internal/metrics"
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

// CreateOrganization creates a new organization
func (s *Service) CreateOrganization(org *Organization) error {
	return s.db.Create(org).Error
}

// GetOrganization retrieves an organization by ID
func (s *Service) GetOrganization(id uint) (*Organization, error) {
	var org Organization
	err := s.db.Preload("Projects").Preload("Members").First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// ListOrganizations lists all organizations
func (s *Service) ListOrganizations() ([]Organization, error) {
	var orgs []Organization
	err := s.db.Find(&orgs).Error
	return orgs, err
}

// UpdateOrganization updates an organization
func (s *Service) UpdateOrganization(org *Organization) error {
	return s.db.Save(org).Error
}

// DeleteOrganization deletes an organization
func (s *Service) DeleteOrganization(id uint) error {
	return s.db.Delete(&Organization{}, id).Error
}

// CreateProject creates a new project
func (s *Service) CreateProject(project *Project) error {
	return s.db.Create(project).Error
}

// GetProject retrieves a project by ID
func (s *Service) GetProject(id uint) (*Project, error) {
	var project Project
	err := s.db.Preload("APIKeys").Preload("Usage").First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListProjects lists all projects for an organization
func (s *Service) ListProjects(orgID uint) ([]Project, error) {
	var projects []Project
	err := s.db.Where("organization_id = ?", orgID).Find(&projects).Error
	return projects, err
}

// UpdateProject updates a project
func (s *Service) UpdateProject(project *Project) error {
	return s.db.Save(project).Error
}

// DeleteProject deletes a project
func (s *Service) DeleteProject(id uint) error {
	return s.db.Delete(&Project{}, id).Error
}

// CreateAPIKey creates a new API key
func (s *Service) CreateAPIKey(key *APIKey) error {
	return s.db.Create(key).Error
}

// GetAPIKey retrieves an API key by value
func (s *Service) GetAPIKey(key string) (*APIKey, error) {
	var apiKey APIKey
	err := s.db.Where("key = ?", key).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// ListAPIKeys lists all API keys for a project
func (s *Service) ListAPIKeys(projectID uint) ([]APIKey, error) {
	var keys []APIKey
	err := s.db.Where("project_id = ?", projectID).Find(&keys).Error
	return keys, err
}

// RevokeAPIKey revokes an API key
func (s *Service) RevokeAPIKey(id uint) error {
	return s.db.Model(&APIKey{}).Where("id = ?", id).Update("is_active", false).Error
}

// RecordUsage records API usage
func (s *Service) RecordUsage(usage *Usage) error {
	return s.db.Create(usage).Error
}

// GetAPIKey retrieves an API key for a specific provider
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
func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/orgs", s.handleOrganizations)
	mux.HandleFunc("/api/v1/orgs/", s.handleOrganization)
	mux.HandleFunc("/api/v1/projects", s.handleProjects)
	mux.HandleFunc("/api/v1/projects/", s.handleProject)
	mux.HandleFunc("/api/v1/api-keys", s.handleAPIKeys)
}

func (s *Service) handleOrganizations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListOrganizations(w, r)
	case http.MethodPost:
		s.handleCreateOrganization(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleOrganization(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetOrganization(w, r)
	case http.MethodPut:
		s.handleUpdateOrganization(w, r)
	case http.MethodDelete:
		s.handleDeleteOrganization(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListProjects(w, r)
	case http.MethodPost:
		s.handleCreateProject(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleProject(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetProject(w, r)
	case http.MethodPut:
		s.handleUpdateProject(w, r)
	case http.MethodDelete:
		s.handleDeleteProject(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleAPIKeys(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListAPIKeys(w, r)
	case http.MethodPost:
		s.handleCreateAPIKey(w, r)
	case http.MethodDelete:
		s.handleDeleteAPIKey(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper functions for HTTP handlers
func (s *Service) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
	var orgs []Organization
	err := s.db.Find(&orgs).Error
	if err != nil {
		http.Error(w, "failed to list organizations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

func (s *Service) handleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	var org Organization
	if err := json.NewDecoder(r.Body).Decode(&org); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.db.Create(&org).Error; err != nil {
		http.Error(w, "failed to create organization", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

func (s *Service) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.URL.Path[len("/api/v1/orgs/"):], 10, 32)
	if err != nil {
		http.Error(w, "invalid organization ID", http.StatusBadRequest)
		return
	}

	org, err := s.GetOrganization(uint(id))
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

func (s *Service) handleUpdateOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.URL.Path[len("/api/v1/orgs/"):], 10, 32)
	if err != nil {
		http.Error(w, "invalid organization ID", http.StatusBadRequest)
		return
	}

	var org Organization
	if err := json.NewDecoder(r.Body).Decode(&org); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	org.ID = id
	org.UpdatedAt = time.Now()

	_, err = s.db.Exec(`
		UPDATE organizations
		SET name = $1, updated_at = $2
		WHERE id = $3
	`, org.Name, org.UpdatedAt, org.ID)
	if err != nil {
		http.Error(w, "failed to update organization", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

func (s *Service) handleDeleteOrganization(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/orgs/"):]
	_, err := s.db.Exec("DELETE FROM organizations WHERE id = $1", id)
	if err != nil {
		http.Error(w, "failed to delete organization", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) handleListProjects(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		http.Error(w, "organization_id is required", http.StatusBadRequest)
		return
	}

	rows, err := s.db.Query(`
		SELECT id, organization_id, name, created_at, updated_at
		FROM projects
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		http.Error(w, "failed to list projects", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.OrganizationID, &project.Name, &project.CreatedAt, &project.UpdatedAt); err != nil {
			http.Error(w, "failed to scan project", http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (s *Service) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var project Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO projects (id, organization_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, project.ID, project.OrganizationID, project.Name, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (s *Service) handleGetProject(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/projects/"):]
	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		http.Error(w, "organization_id is required", http.StatusBadRequest)
		return
	}

	project, err := s.GetProject(uint(id))
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (s *Service) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/projects/"):]
	var project Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	project.ID = id
	project.UpdatedAt = time.Now()

	_, err := s.db.Exec(`
		UPDATE projects
		SET name = $1, updated_at = $2
		WHERE id = $3 AND organization_id = $4
	`, project.Name, project.UpdatedAt, project.ID, project.OrganizationID)
	if err != nil {
		http.Error(w, "failed to update project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (s *Service) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/projects/"):]
	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		http.Error(w, "organization_id is required", http.StatusBadRequest)
		return
	}

	_, err := s.db.Exec("DELETE FROM projects WHERE id = $1 AND organization_id = $2", id, orgID)
	if err != nil {
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	projectID := r.URL.Query().Get("project_id")
	if orgID == "" || projectID == "" {
		http.Error(w, "organization_id and project_id are required", http.StatusBadRequest)
		return
	}

	rows, err := s.db.Query(`
		SELECT id, organization_id, project_id, provider, created_at, updated_at
		FROM api_keys
		WHERE organization_id = $1 AND project_id = $2
		ORDER BY created_at DESC
	`, orgID, projectID)
	if err != nil {
		http.Error(w, "failed to list API keys", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		if err := rows.Scan(&key.ID, &key.OrganizationID, &key.ProjectID, &key.Provider, &key.CreatedAt, &key.UpdatedAt); err != nil {
			http.Error(w, "failed to scan API key", http.StatusInternalServerError)
			return
		}
		keys = append(keys, key)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (s *Service) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var key APIKey
	if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	now := time.Now()
	key.CreatedAt = now
	key.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO api_keys (id, organization_id, project_id, provider, api_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, key.ID, key.OrganizationID, key.ProjectID, key.Provider, key.APIKey, key.CreatedAt, key.UpdatedAt)
	if err != nil {
		http.Error(w, "failed to create API key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(key)
}

func (s *Service) handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	orgID := r.URL.Query().Get("organization_id")
	projectID := r.URL.Query().Get("project_id")
	if id == "" || orgID == "" || projectID == "" {
		http.Error(w, "id, organization_id, and project_id are required", http.StatusBadRequest)
		return
	}

	_, err := s.db.Exec(`
		DELETE FROM api_keys
		WHERE id = $1 AND organization_id = $2 AND project_id = $3
	`, id, orgID, projectID)
	if err != nil {
		http.Error(w, "failed to delete API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to create database tables
func createTables(db *sql.DB) error {
	// Create organizations table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create organizations table: %w", err)
	}

	// Create projects table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL REFERENCES organizations(id),
			name TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create projects table: %w", err)
	}

	// Create api_keys table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL REFERENCES organizations(id),
			project_id TEXT NOT NULL REFERENCES projects(id),
			provider TEXT NOT NULL,
			api_key TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create api_keys table: %w", err)
	}

	// Create usage_metrics table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS usage_metrics (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL REFERENCES organizations(id),
			project_id TEXT NOT NULL REFERENCES projects(id),
			provider TEXT NOT NULL,
			prompt_tokens INTEGER NOT NULL,
			completion_tokens INTEGER NOT NULL,
			total_tokens INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create usage_metrics table: %w", err)
	}

	return nil
}

// CreateOrganization handles organization creation
func (s *Service) CreateOrganization(c *gin.Context) {
	var org Organization
	if err := c.ShouldBindJSON(&org); err != nil {
		metrics.RecordAPIError("organizations", "POST", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add current user as admin
	userID := c.GetString("user_id")
	org.Members = []Member{
		{
			UserID: userID,
			Role:   RoleAdmin,
		},
	}

	if err := s.CreateOrganization(&org); err != nil {
		metrics.RecordAPIError("organizations", "POST", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, org)
}

// ListOrganizations handles listing organizations
func (s *Service) ListOrganizations(c *gin.Context) {
	orgs, err := s.ListOrganizations()
	if err != nil {
		metrics.RecordAPIError("organizations", "GET", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orgs)
}

// GetOrganization handles retrieving an organization
func (s *Service) GetOrganization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		metrics.RecordAPIError("organizations", "GET", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	org, err := s.GetOrganization(uint(id))
	if err != nil {
		metrics.RecordAPIError("organizations", "GET", http.StatusNotFound)
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// UpdateOrganization handles updating an organization
func (s *Service) UpdateOrganization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		metrics.RecordAPIError("organizations", "PUT", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	var org Organization
	if err := c.ShouldBindJSON(&org); err != nil {
		metrics.RecordAPIError("organizations", "PUT", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org.ID = uint(id)
	if err := s.UpdateOrganization(&org); err != nil {
		metrics.RecordAPIError("organizations", "PUT", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// DeleteOrganization handles deleting an organization
func (s *Service) DeleteOrganization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		metrics.RecordAPIError("organizations", "DELETE", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	if err := s.DeleteOrganization(uint(id)); err != nil {
		metrics.RecordAPIError("organizations", "DELETE", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateProject handles project creation
func (s *Service) CreateProject(c *gin.Context) {
	var project Project
	if err := c.ShouldBindJSON(&project); err != nil {
		metrics.RecordAPIError("projects", "POST", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set organization ID from context
	orgID, _ := strconv.ParseUint(c.GetString("organization_id"), 10, 32)
	project.OrganizationID = uint(orgID)

	if err := s.CreateProject(&project); err != nil {
		metrics.RecordAPIError("projects", "POST", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// ListProjects handles listing projects
func (s *Service) ListProjects(c *gin.Context) {
	orgID, _ := strconv.ParseUint(c.GetString("organization_id"), 10, 32)
	projects, err := s.ListProjects(uint(orgID))
	if err != nil {
		metrics.RecordAPIError("projects", "GET", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProject handles retrieving a project
func (s *Service) GetProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		metrics.RecordAPIError("projects", "GET", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := s.GetProject(uint(id))
	if err != nil {
		metrics.RecordAPIError("projects", "GET", http.StatusNotFound)
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject handles updating a project
func (s *Service) UpdateProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		metrics.RecordAPIError("projects", "PUT", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var project Project
	if err := c.ShouldBindJSON(&project); err != nil {
		metrics.RecordAPIError("projects", "PUT", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project.ID = uint(id)
	if err := s.UpdateProject(&project); err != nil {
		metrics.RecordAPIError("projects", "PUT", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject handles deleting a project
func (s *Service) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		metrics.RecordAPIError("projects", "DELETE", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	if err := s.DeleteProject(uint(id)); err != nil {
		metrics.RecordAPIError("projects", "DELETE", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateAPIKey handles API key creation
func (s *Service) CreateAPIKey(c *gin.Context) {
	var key APIKey
	if err := c.ShouldBindJSON(&key); err != nil {
		metrics.RecordAPIError("api-keys", "POST", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set project ID from context
	projectID, _ := strconv.ParseUint(c.GetString("project_id"), 10, 32)
	key.ProjectID = uint(projectID)

	if err := s.CreateAPIKey(&key); err != nil {
		metrics.RecordAPIError("api-keys", "POST", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, key)
}

// ListAPIKeys handles listing API keys
func (s *Service) ListAPIKeys(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.GetString("project_id"), 10, 32)
	keys, err := s.ListAPIKeys(uint(projectID))
	if err != nil {
		metrics.RecordAPIError("api-keys", "GET", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, keys)
}

// RevokeAPIKey handles revoking an API key
func (s *Service) RevokeAPIKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		metrics.RecordAPIError("api-keys", "DELETE", http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid API key ID"})
		return
	}

	if err := s.RevokeAPIKey(uint(id)); err != nil {
		metrics.RecordAPIError("api-keys", "DELETE", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
} 
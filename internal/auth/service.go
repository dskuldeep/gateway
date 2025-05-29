package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/dskuldeep/gateway/internal/metrics"
)

// Service handles authentication and authorization
type Service struct {
	clerkSecretKey string
}

// NewService creates a new auth service
func NewService() *Service {
	return &Service{
		clerkSecretKey: os.Getenv("CLERK_SECRET_KEY"),
	}
}

// AuthMiddleware validates Clerk JWT tokens
func (s *Service) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			metrics.RecordAPIError("auth", "GET", http.StatusUnauthorized)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			metrics.RecordAPIError("auth", "GET", http.StatusUnauthorized)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}
		token := parts[1]

		// Verify token with Clerk
		claims, err := s.verifyClerkToken(token)
		if err != nil {
			metrics.RecordAPIError("auth", "GET", http.StatusUnauthorized)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.Subject)
		c.Set("organization_id", claims.OrganizationID)
		c.Set("project_id", claims.ProjectID)

		c.Next()
	}
}

// ClerkClaims represents the JWT claims from Clerk
type ClerkClaims struct {
	Subject        string `json:"sub"`
	OrganizationID string `json:"org_id"`
	ProjectID      string `json:"project_id"`
}

// verifyClerkToken verifies a JWT token with Clerk
func (s *Service) verifyClerkToken(token string) (*ClerkClaims, error) {
	// Make request to Clerk's verify endpoint
	req, err := http.NewRequest("GET", "https://api.clerk.dev/v1/verify", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.clerkSecretKey)
	req.Header.Set("X-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token verification failed: %s", resp.Status)
	}

	// Parse response
	var claims ClerkClaims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &claims, nil
}

// RequireRole middleware ensures the user has the required role
func (s *Service) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Example: If userID is not used, remove it
		// userID := c.GetString("user_id")

		// Example: If orgID is not used, remove it
		// orgID := c.GetString("organization_id")

		// TODO: Check user's role in the organization
		// This would typically involve querying the database

		c.Next()
	}
}

// RequireProjectAccess middleware ensures the user has access to the project
func (s *Service) RequireProjectAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Example: If projectID is not used, remove it
		// projectID := c.GetString("project_id")

		// TODO: Check user's access to the project
		// This would typically involve querying the database

		c.Next()
	}
} 
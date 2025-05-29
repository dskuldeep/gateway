package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/dskuldeep/gateway/internal/auth"
	"github.com/dskuldeep/gateway/internal/llms"
	"github.com/dskuldeep/gateway/internal/orgs"
	"github.com/dskuldeep/gateway/internal/metrics"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize metrics
	metrics.Init()

	// Initialize database
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize router
	router := gin.Default()

	// Initialize services
	authService := auth.NewService()
	orgService := orgs.NewService(db)
	llmService := llms.NewService(orgService, db)

	// Setup routes
	setupRoutes(router, authService, llmService, orgService)

	// Create server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func setupRoutes(router *gin.Engine, authService *auth.Service, llmService *llms.Service, orgService *orgs.Service) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(metrics.Handler()))

	// API routes
	v1 := router.Group("/v1")
	{
		// LLM routes
		llm := v1.Group("/llm")
		{
			llm.POST("/query", authService.AuthMiddleware(), llmService.HandleQuery)
		}

		// Organization routes
		orgs := v1.Group("/orgs")
		{
			orgs.Use(authService.AuthMiddleware())
			orgs.POST("", orgService.CreateOrganization)
			orgs.GET("", orgService.ListOrganizations)
			orgs.GET("/:id", orgService.GetOrganization)
			orgs.PUT("/:id", orgService.UpdateOrganization)
			orgs.DELETE("/:id", orgService.DeleteOrganization)
		}

		// Project routes
		projects := v1.Group("/projects")
		{
			projects.Use(authService.AuthMiddleware())
			projects.POST("", orgService.CreateProject)
			projects.GET("", orgService.ListProjects)
			projects.GET("/:id", orgService.GetProject)
			projects.PUT("/:id", orgService.UpdateProject)
			projects.DELETE("/:id", orgService.DeleteProject)
		}

		// API key routes
		apiKeys := v1.Group("/api-keys")
		{
			apiKeys.Use(authService.AuthMiddleware())
			apiKeys.POST("", orgService.CreateAPIKey)
			apiKeys.GET("", orgService.ListAPIKeys)
			apiKeys.DELETE("/:id", orgService.RevokeAPIKey)
		}
	}
} 
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dskuldeep/gateway/internal/auth"
	"github.com/dskuldeep/gateway/internal/llms"
	"github.com/dskuldeep/gateway/internal/metrics"
	"github.com/dskuldeep/gateway/internal/orgs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
	db, err := initDB()
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
	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(metrics.Handler()))

	// API routes
	v1 := router.Group("/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.GET("/me", authService.AuthMiddleware(), func(c *gin.Context) {
				c.JSON(200, gin.H{
					"user_id":         c.GetString("user_id"),
					"organization_id": c.GetString("organization_id"),
					"project_id":      c.GetString("project_id"),
				})
			})
		}

		// LLM routes
		llm := v1.Group("/llm")
		{
			llm.POST("/query", authService.AuthMiddleware(), llmService.HandleQuery)
			llm.GET("/models", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"models": []gin.H{
						{"id": "gpt-4", "name": "GPT-4", "provider": "openai"},
						{"id": "gpt-3.5-turbo", "name": "GPT-3.5 Turbo", "provider": "openai"},
						{"id": "claude-3-opus", "name": "Claude 3 Opus", "provider": "anthropic"},
						{"id": "gemini-pro", "name": "Gemini Pro", "provider": "google"},
					},
				})
			})
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

func initDB() (*gorm.DB, error) {
	// Get database URL from environment
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		// Build database URL from individual components
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "postgres")
		dbname := getEnv("DB_NAME", "gateway")
		sslmode := getEnv("DB_SSL_MODE", "disable")

		databaseURL = "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode
	}

	// Open database connection with GORM
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

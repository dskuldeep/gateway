package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/dskuldeep/gateway/internal/auth"
	"github.com/dskuldeep/gateway/internal/intelligence"
	"github.com/dskuldeep/gateway/internal/llms"
	"github.com/dskuldeep/gateway/internal/llms/google"
	"github.com/dskuldeep/gateway/internal/llms/groq"
	"github.com/dskuldeep/gateway/internal/metrics"
	"github.com/dskuldeep/gateway/internal/orgs"
)

func main() {
	// Initialize database connection
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize services
	metrics := metrics.NewMetrics()
	orgService, err := orgs.NewService(db)
	if err != nil {
		log.Fatalf("Failed to initialize organization service: %v", err)
	}

	authService := auth.NewService()
	llmService := llms.NewService(orgService, metrics)

	// Register LLM providers
	llmService.RegisterClient(groq.NewClient())
	llmService.RegisterClient(google.NewClient())

	// Initialize intelligence layer
	evaluator := intelligence.NewDefaultEvaluator()
	if err := evaluator.LoadConfig(); err != nil {
		log.Printf("Warning: Failed to load intelligence config: %v", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Register routes
	authService.RegisterRoutes(mux)
	orgService.RegisterRoutes(mux)
	llmService.RegisterRoutes(mux)

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add metrics endpoint
	mux.Handle("/metrics", metrics.Handler())

	// Create server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

func initDB() (*sql.DB, error) {
	// Get database connection details from environment
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "gateway")
	sslmode := getEnv("DB_SSL_MODE", "disable")

	// Create connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
} 
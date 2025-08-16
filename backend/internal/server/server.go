package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alex/reading-companion/internal/api"
	"github.com/alex/reading-companion/internal/llmclient"
	"github.com/alex/reading-companion/internal/repository"
	"github.com/alex/reading-companion/internal/service"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server struct holds the HTTP server instance
type Server struct {
	httpServer *http.Server
	db         *pgxpool.Pool
}

// NewServer creates a new HTTP server instance
func NewServer(port string) *Server {
	// Use port 9090 if no port is specified
	if port == "" {
		port = "9090"
	}

	// Database connection with retry mechanism
	db := connectToDatabaseWithRetry()

	// Run database migrations
	runMigrations()

	// Create LLM client
	var llmClient llmclient.Client

	maxTokensStr := getEnv("LLM_MAX_TOKENS", "100")
	maxTokens, err := strconv.Atoi(maxTokensStr)
	if err != nil {
		fmt.Println("Invalid number for LLM_MAX_TOKENS, using default 100")
		maxTokens = 100
	}
	// Check if LLM API key is provided
	if os.Getenv("LLM_API_KEY") != "" {
		// Use real LLM client when API key is provided
		llmConfig := llmclient.Config{
			APIEndpoint: getEnv("LLM_API_ENDPOINT", "https://api.openai.com/v1/chat/completions"),
			APIKey:      os.Getenv("LLM_API_KEY"),
			Model:       getEnv("LLM_MODEL", "gpt-3.5-turbo"),
			Timeout:     30, // seconds
			MaxTokens:   maxTokens,
		}
		llmClient = llmclient.New(llmConfig)
	} else {
		// Use mock client when no API key is provided
		llmClient = llmclient.NewMock()
	}

	// Create repository and services
	repo := repository.New(db)
	sessionService := service.NewSessionService(repo, llmClient)
	sessionHandler := api.NewSessionHandler(sessionService)

	// Create a new HTTP server instance with gorilla/mux router
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/api/healthcheck", api.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/api/sessions", sessionHandler.ListSessions).Methods("GET")
	router.HandleFunc("/api/sessions", sessionHandler.CreateSession).Methods("POST")
	router.HandleFunc("/api/sessions/{session_id}", sessionHandler.GetSession).Methods("GET")
	router.HandleFunc("/api/sessions/{session_id}", sessionHandler.UpdateSessionName).Methods("PATCH")
	router.HandleFunc("/api/sessions/{session_id}", sessionHandler.DeleteSession).Methods("DELETE")
	router.HandleFunc("/api/sessions/{session_id}/export", sessionHandler.ExportSession).Methods("GET")
	router.HandleFunc("/api/sessions/{session_id}/content", sessionHandler.GetSessionContent).Methods("GET")
	router.HandleFunc("/api/sessions/{session_id}/markdown", sessionHandler.GetSessionMarkdown).Methods("GET")
	router.HandleFunc("/api/sessions/{session_id}/process", sessionHandler.ProcessAnswer).Methods("POST")
	router.HandleFunc("/api/sessions/{session_id}/regenerate_question", sessionHandler.RegenerateQuestion).Methods("POST")

	// Wrap the router with CORS middleware
	corsHandler := corsMiddleware(router)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: corsHandler,
	}

	return &Server{
		httpServer: httpServer,
		db:         db,
	}
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from localhost:3000 (frontend)
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

		// Allow common HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Allow common headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Allow credentials
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Start starts the HTTP server
func (s *Server) Start() {
	log.Println("Server starting on port", s.httpServer.Addr)

	// Debug output to see if we can access the router
	fmt.Println("Server routes registered")

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on port %s: %v", s.httpServer.Addr, err)
	}
}

// connectToDatabaseWithRetry attempts to connect to the database with retries
func connectToDatabaseWithRetry() *pgxpool.Pool {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	// Retry connection up to 10 times with 3 second delays
	var db *pgxpool.Pool
	var err error

	for i := 0; i < 10; i++ {
		db, err = pgxpool.New(context.Background(), dbURL)
		if err != nil {
			log.Printf("Unable to connect to database (attempt %d): %v", i+1, err)
			time.Sleep(3 * time.Second)
			continue
		}

		// Test the connection
		err = db.Ping(context.Background())
		if err != nil {
			log.Printf("Unable to ping database (attempt %d): %v", i+1, err)
			db.Close()
			time.Sleep(3 * time.Second)
			continue
		}

		log.Println("Successfully connected to database")
		return db
	}

	log.Fatalf("Unable to connect to database after 10 attempts: %v", err)
	return nil
}

// runMigrations applies database migrations
func runMigrations() {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	// Use the absolute path to migrations directory
	// When running in Docker, the migrations will be in /root/migrations
	// When running locally, they will be in backend/migrations
	migrationPath := "file:///root/migrations"
	if _, err := os.Stat("backend/migrations"); err == nil {
		migrationPath = "file://backend/migrations"
	}

	m, err := migrate.New(migrationPath, dbURL)
	if err != nil {
		log.Fatalf("Unable to create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Unable to apply migrations: %v", err)
	}

	log.Println("Database migrations applied successfully")
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop() error {
	// Close database connection
	s.db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/alex/reading-companion/internal/api"
)

// Server struct holds the HTTP server instance
type Server struct {
	httpServer *http.Server
}

// NewServer creates a new HTTP server instance
func NewServer(port string) *Server {
	// Create a new HTTP server instance
	httpServer := &http.Server{
		Addr: ":" + port,
	}

	// Register routes
	http.HandleFunc("/api/healthcheck", api.HealthCheckHandler)

	return &Server{
		httpServer: httpServer,
	}
}

// Start starts the HTTP server
func (s *Server) Start() {
	log.Println("Server starting on port", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on port %s: %v", s.httpServer.Addr, err)
	}
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}
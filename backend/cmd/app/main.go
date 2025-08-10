package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alex/reading-companion/internal/server"
)

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create a new server instance
	s := server.NewServer(port)

	// Run server in a goroutine
	go func() {
		s.Start()
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := s.Stop(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
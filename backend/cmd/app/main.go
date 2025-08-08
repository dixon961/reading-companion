package main

import (
"fmt"
"log"
"net/http"
)

func main() {
// Create a new HTTP server
mux := http.NewServeMux()

// Define the healthcheck endpoint
mux.HandleFunc("GET /api/healthcheck", func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
fmt.Fprintf(w, `{"status": "healthy"}`)
})

// Start the server
fmt.Println("Starting backend server on :8080")
log.Fatal(http.ListenAndServe(":8080", mux))
}

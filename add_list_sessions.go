package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	// Read the original file
	content, err := ioutil.ReadFile("backend/internal/api/session.go")
	if err != nil {
		panic(err)
	}

	// Convert to string and split into lines
	lines := strings.Split(string(content), "\n")
	
	// Find the closing brace of the last function
	insertIndex := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "}" {
			insertIndex = i + 1
			break
		}
	}
	
	// Create the new function
	newFunction := []string{
		"",
		"// ListSessions handles GET /api/sessions",
		"func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {",
		"\t// Get sessions using service",
		"\tsessions, err := h.sessionService.ListSessions()",
		"\tif err != nil {",
		"\t\thttp.Error(w, fmt.Sprintf(\"failed to list sessions: %v\", err), http.StatusInternalServerError)",
		"\t\treturn",
		"\t}",
		"",
		"\t// Send successful response",
		"\tw.Header().Set(\"Content-Type\", \"application/json\")",
		"\tw.WriteHeader(http.StatusOK)",
		"\tjson.NewEncoder(w).Encode(sessions)",
		"}",
	}
	
	// Insert the new function at the correct position
	result := make([]string, 0, len(lines)+len(newFunction))
	result = append(result, lines[:insertIndex]...)
	result = append(result, newFunction...)
	result = append(result, lines[insertIndex:]...)
	
	// Join and write back
	output := strings.Join(result, "\n")
	err = ioutil.WriteFile("backend/internal/api/session.go", []byte(output), 0644)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Successfully appended ListSessions method")
}

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/alex/reading-companion/internal/models"
	"github.com/alex/reading-companion/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// SessionHandler handles session-related HTTP requests
type SessionHandler struct {
	sessionService *service.SessionService
}

// NewSessionHandler creates a new SessionHandler
func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// CreateSession handles POST /api/sessions
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with max memory of 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse multipart form: %v", err), http.StatusBadRequest)
		return
	}

	// Get session name from form data (optional)
	sessionName := r.FormValue("session_name")

	// Get file from form data
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get file from form: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create session using service
	response, err := h.sessionService.CreateSession(file, sessionName)
	if err != nil {
		// Handle validation errors
		http.Error(w, fmt.Sprintf("failed to create session: %v", err), http.StatusBadRequest)
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetSession handles GET /api/sessions/{session_id}
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from URL path parameters
	vars := mux.Vars(r)
	sessionIDStr, ok := vars["session_id"]
	if !ok {
		http.Error(w, "session ID is required", http.StatusBadRequest)
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid session ID: %v", err), http.StatusBadRequest)
		return
	}

	// Get session using service
	session, err := h.sessionService.GetSession(sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get session: %v", err), http.StatusNotFound)
		return
	}

	// Get highlights for this session
	highlights, err := h.sessionService.GetHighlightsBySession(sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get highlights: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the first highlight for the initial step
	var nextStep *models.NextStep
	if len(highlights) > 0 {
		firstHighlight := highlights[0]
		
		// For now, we'll use a placeholder question
		// In a real implementation, this would come from the database or be generated
		nextStep = &models.NextStep{
			HighlightIndex: 0,
			HighlightText:  firstHighlight.Text,
			Question:       "What are your thoughts on this highlight?",
		}
	}

	// Prepare response
	response := struct {
		ID              uuid.UUID      `json:"id"`
		Name            string         `json:"name"`
		Status          string         `json:"status"`
		TotalHighlights int            `json:"total_highlights"`
		NextStep        *models.NextStep `json:"next_step,omitempty"`
	}{
		ID:              session.ID,
		Name:            session.Name,
		Status:          session.Status,
		TotalHighlights: len(highlights),
		NextStep:        nextStep,
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ProcessAnswer handles POST /api/sessions/{session_id}/process
func (h *SessionHandler) ProcessAnswer(w http.ResponseWriter, r *http.Request) {
	// Debug output to see if we're reaching this handler
	fmt.Printf("ProcessAnswer handler called\n")
	
	// Extract session ID from URL path parameters
	vars := mux.Vars(r)
	sessionIDStr, ok := vars["session_id"]
	if !ok {
		http.Error(w, "session ID is required", http.StatusBadRequest)
		return
	}
	
	fmt.Printf("Session ID: %s\n", sessionIDStr)
	
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid session ID: %v", err), http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var req models.ProcessAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Process answer using service
	response, err := h.sessionService.ProcessAnswerAndGetNextStep(sessionID, req.HighlightIndex, req.UserAnswer)
	if err != nil {
		// Handle not found errors
		if strings.Contains(err.Error(), "invalid highlight index") {
			http.Error(w, fmt.Sprintf("session or highlight not found: %v", err), http.StatusNotFound)
			return
		}
		// Handle other errors
		http.Error(w, fmt.Sprintf("failed to process answer: %v", err), http.StatusBadRequest)
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RegenerateQuestion handles POST /api/sessions/{session_id}/regenerate_question
func (h *SessionHandler) RegenerateQuestion(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from URL path parameters
	vars := mux.Vars(r)
	sessionIDStr, ok := vars["session_id"]
	if !ok {
		http.Error(w, "session ID is required", http.StatusBadRequest)
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid session ID: %v", err), http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var req struct {
		HighlightIndex int `json:"highlight_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Regenerate question using service
	newQuestion, err := h.sessionService.RegenerateQuestion(sessionID, req.HighlightIndex)
	if err != nil {
		// Handle not found errors
		if strings.Contains(err.Error(), "invalid highlight index") {
			http.Error(w, fmt.Sprintf("session or highlight not found: %v", err), http.StatusNotFound)
			return
		}
		// Handle other errors
		http.Error(w, fmt.Sprintf("failed to regenerate question: %v", err), http.StatusBadRequest)
		return
	}

	// Send successful response
	response := struct {
		NewQuestion string `json:"new_question"`
	}{
		NewQuestion: newQuestion,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
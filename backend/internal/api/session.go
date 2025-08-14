package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
		http.Error(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get session name from form data (optional)
	sessionName := r.FormValue("session_name")

	// Get file from form data
	file, header, err := r.FormFile("file")
	if err != nil {
		// Check if file was not provided
		if err == http.ErrMissingFile {
			http.Error(w, "Файл пуст или имеет неверный формат", http.StatusBadRequest)
		} else {
			http.Error(w, "failed to get file from form: "+err.Error(), http.StatusBadRequest)
		}
		return
	}
	defer file.Close()

	// Check if file is empty
	if header.Size == 0 {
		http.Error(w, "Файл пуст или имеет неверный формат", http.StatusBadRequest)
		return
	}

	// Create session using service
	response, err := h.sessionService.CreateSession(file, sessionName)
	if err != nil {
		// Handle validation errors with specific messages
		if strings.Contains(err.Error(), "minimum 3 highlights required") {
			http.Error(w, "Для начала сессии необходимо как минимум 3 пометки", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to create session: "+err.Error(), http.StatusBadRequest)
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
		ID              uuid.UUID        `json:"id"`
		Name            string           `json:"name"`
		Status          string           `json:"status"`
		TotalHighlights int              `json:"total_highlights"`
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
		// Handle LLM service unavailable errors
		if strings.Contains(err.Error(), "LLM service unavailable") {
			http.Error(w, fmt.Sprintf("LLM service unavailable: %v", err), http.StatusServiceUnavailable)
			return
		}
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

// UpdateSessionName handles PATCH /api/sessions/{session_id}
func (h *SessionHandler) UpdateSessionName(w http.ResponseWriter, r *http.Request) {
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
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Update session name using service
	updatedSession, err := h.sessionService.UpdateSessionName(sessionID, req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update session name: %v", err), http.StatusInternalServerError)
		return
	}

	// Send successful response
	response := struct {
		ID        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		ID:        updatedSession.ID,
		Name:      updatedSession.Name,
		Status:    updatedSession.Status,
		CreatedAt: updatedSession.CreatedAt,
		UpdatedAt: updatedSession.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteSession handles DELETE /api/sessions/{session_id}
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
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

	// Delete session using service
	err = h.sessionService.DeleteSession(sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete session: %v", err), http.StatusInternalServerError)
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// ExportSession handles GET /api/sessions/{session_id}/export
func (h *SessionHandler) ExportSession(w http.ResponseWriter, r *http.Request) {
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

	// Export session using service
	markdownContent, err := h.sessionService.ExportSessionAsMarkdown(sessionID)
	if err != nil {
		// Handle not found errors
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, fmt.Sprintf("session not found: %v", err), http.StatusNotFound)
			return
		}
		// Handle status errors
		if strings.Contains(err.Error(), "not completed") {
			http.Error(w, fmt.Sprintf("session not completed: %v", err), http.StatusBadRequest)
			return
		}
		// Handle other errors
		http.Error(w, fmt.Sprintf("failed to export session: %v", err), http.StatusInternalServerError)
		return
	}

	// Send successful response with markdown content
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"session_export.md\"")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(markdownContent))
}

// ListSessions handles GET /api/sessions
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Get sessions using service
	sessions, err := h.sessionService.ListSessions()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list sessions: %v", err), http.StatusInternalServerError)
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sessions)
}

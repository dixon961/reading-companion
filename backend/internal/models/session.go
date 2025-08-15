package models

import (
	"github.com/google/uuid"
	"time"
)

// Session represents a reading session
type Session struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Highlight represents a single highlight from a book
type Highlight struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Text      string    `json:"text"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateSessionRequest represents the request body for creating a session
type CreateSessionRequest struct {
	SessionName string `json:"session_name"`
}

// NextStep represents the next step in the session
type NextStep struct {
	HighlightIndex int    `json:"highlight_index"`
	HighlightText  string `json:"highlight_text"`
	Question       string `json:"question"`
}

// CreateSessionResponse represents the response body for creating a session
type CreateSessionResponse struct {
	SessionID       uuid.UUID `json:"session_id"`
	Name            string    `json:"name"`
	TotalHighlights int       `json:"total_highlights"`
	NextStep        NextStep  `json:"next_step"`
}

// ProcessAnswerRequest represents the request body for processing an answer
type ProcessAnswerRequest struct {
	HighlightIndex int    `json:"highlight_index"`
	UserAnswer     string `json:"user_answer"`
}

// ProcessAnswerResponse represents the response body for processing an answer
type ProcessAnswerResponse struct {
	Status   string    `json:"status,omitempty"`
	Message  string    `json:"message,omitempty"`
	NextStep *NextStep `json:"next_step,omitempty"`
}

// SessionMetadata represents session information for listing
type SessionMetadata struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

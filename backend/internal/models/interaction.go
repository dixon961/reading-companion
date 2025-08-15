package models

import (
	"database/sql"
	"github.com/google/uuid"
)

// Interaction represents a question-answer interaction for a highlight
type Interaction struct {
	ID          uuid.UUID      `json:"id"`
	HighlightID uuid.UUID      `json:"highlight_id"`
	Question    string         `json:"question"`
	Answer      sql.NullString `json:"answer"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

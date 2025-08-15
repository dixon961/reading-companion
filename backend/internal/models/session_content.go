package models

// SessionContent represents the content of a completed session for display
type SessionContent struct {
	Session    SessionContentInfo   `json:"session"`
	Highlights []HighlightContent    `json:"highlights"`
}

// SessionContentInfo represents session information for content display
type SessionContentInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// HighlightContent represents a highlight with its interaction for content display
type HighlightContent struct {
	Text       string      `json:"text"`
	Question   string      `json:"question"`
	Answer     string      `json:"answer"`
	Answered   bool        `json:"answered"`
}
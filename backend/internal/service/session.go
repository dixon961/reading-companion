package service

import (
	"bufio"
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/alex/reading-companion/internal/models"
	"github.com/alex/reading-companion/internal/repository"
	"github.com/alex/reading-companion/internal/llmclient"
	"github.com/jackc/pgx/v5/pgtype"
)

// SessionService provides session management operations
type SessionService struct {
	repo repository.Querier
	llm  llmclient.Client
}

// NewSessionService creates a new SessionService
func NewSessionService(repo repository.Querier, llm llmclient.Client) *SessionService {
	return &SessionService{
		repo: repo,
		llm:  llm,
	}
}

// CreateSession creates a new session from uploaded file
func (s *SessionService) CreateSession(file multipart.File, sessionName string) (*models.CreateSessionResponse, error) {
	// Parse the file into highlights
	highlights, err := s.parseHighlights(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse highlights: %w", err)
	}

	// Validate the highlights
	if err := s.validateHighlights(highlights); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate session name if not provided
	if sessionName == "" {
		sessionName = fmt.Sprintf("Session %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	// Create session in database
	sessionID := uuid.New()
	sessionParams := repository.CreateSessionParams{
		ID:     pgtype.UUID{Bytes: sessionID, Valid: true},
		Name:   sessionName,
		Status: "in_progress",
	}

	createdSession, err := s.repo.CreateSession(context.Background(), sessionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Save highlights to database
	err = s.saveHighlights(sessionID, highlights)
	if err != nil {
		return nil, fmt.Errorf("failed to save highlights: %w", err)
	}

	// Prepare response
	response := &models.CreateSessionResponse{
		SessionID:       sessionID,
		Name:            createdSession.Name,
		TotalHighlights: len(highlights),
	}

	// Add first highlight data if available
	if len(highlights) > 0 {
		response.NextStep.HighlightIndex = 0
		response.NextStep.HighlightText = highlights[0]
		
		// Generate question using LLM
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		question, err := s.llm.GenerateQuestion(ctx, highlights[0])
		if err != nil {
			// Fallback to placeholder question if LLM fails
			question = "What are your thoughts on this highlight?"
		}
		response.NextStep.Question = question
	}

	return response, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(sessionID uuid.UUID) (*models.Session, error) {
	session, err := s.repo.GetSession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &models.Session{
		ID:        uuid.UUID(session.ID.Bytes),
		Name:      session.Name,
		Status:    session.Status,
		CreatedAt: session.CreatedAt.Time,
		UpdatedAt: session.UpdatedAt.Time,
	}, nil
}

// GetHighlightsBySession retrieves all highlights for a session
func (s *SessionService) GetHighlightsBySession(sessionID uuid.UUID) ([]*models.Highlight, error) {
	highlights, err := s.repo.GetHighlightsBySession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get highlights: %w", err)
	}

	var result []*models.Highlight
	for _, highlight := range highlights {
		result = append(result, &models.Highlight{
			ID:        uuid.UUID(highlight.ID.Bytes),
			SessionID: uuid.UUID(highlight.SessionID.Bytes),
			Text:      highlight.Text,
			Position:  int(highlight.Position),
			CreatedAt: highlight.CreatedAt.Time,
		})
	}

	return result, nil
}

// parseHighlights parses the uploaded file into a slice of highlight strings
func (s *SessionService) parseHighlights(file multipart.File) ([]string, error) {
	var highlights []string

	// Reset file pointer to beginning
	file.Seek(0, 0)

	scanner := bufio.NewScanner(file)
	var currentHighlight strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			// If we have content in current highlight, save it
			if currentHighlight.Len() > 0 {
				highlights = append(highlights, currentHighlight.String())
				currentHighlight.Reset()
			}
			continue
		}

		// Add line to current highlight
		if currentHighlight.Len() > 0 {
			currentHighlight.WriteString(" ")
		}
		currentHighlight.WriteString(line)
	}

	// Don't forget the last highlight if file doesn't end with empty line
	if currentHighlight.Len() > 0 {
		highlights = append(highlights, currentHighlight.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return highlights, nil
}

// validateHighlights validates the parsed highlights
func (s *SessionService) validateHighlights(highlights []string) error {
	// Check if we have enough highlights
	if len(highlights) < 3 {
		return fmt.Errorf("minimum 3 highlights required, got %d", len(highlights))
	}

	// Check if highlights are not empty
	for i, highlight := range highlights {
		if strings.TrimSpace(highlight) == "" {
			return fmt.Errorf("highlight %d is empty", i)
		}
	}

	return nil
}

// saveHighlights saves the parsed highlights to the database
func (s *SessionService) saveHighlights(sessionID uuid.UUID, highlights []string) error {
	for i, text := range highlights {
		highlightID := uuid.New()
		highlightParams := repository.CreateHighlightParams{
			ID:        pgtype.UUID{Bytes: highlightID, Valid: true},
			SessionID: pgtype.UUID{Bytes: sessionID, Valid: true},
			Text:      text,
			Position:  int32(i),
		}

		_, err := s.repo.CreateHighlight(context.Background(), highlightParams)
		if err != nil {
			return fmt.Errorf("failed to save highlight %d: %w", i, err)
		}
	}

	return nil
}

// ListSessions retrieves all sessions
func (s *SessionService) ListSessions() ([]*models.SessionMetadata, error) {
	sessions, err := s.repo.ListSessions(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var result []*models.SessionMetadata
	for _, session := range sessions {
		result = append(result, &models.SessionMetadata{
			ID:        uuid.UUID(session.ID.Bytes),
			Name:      session.Name,
			Status:    session.Status,
			CreatedAt: session.CreatedAt.Time,
			UpdatedAt: session.UpdatedAt.Time,
		})
	}

	return result, nil
}

// UpdateSessionName updates the name of a session
func (s *SessionService) UpdateSessionName(sessionID uuid.UUID, name string) (*models.Session, error) {
	params := repository.UpdateSessionNameParams{
		ID:   pgtype.UUID{Bytes: sessionID, Valid: true},
		Name: name,
	}

	session, err := s.repo.UpdateSessionName(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("failed to update session name: %w", err)
	}

	return &models.Session{
		ID:        uuid.UUID(session.ID.Bytes),
		Name:      session.Name,
		Status:    session.Status,
		CreatedAt: session.CreatedAt.Time,
		UpdatedAt: session.UpdatedAt.Time,
	}, nil
}

// DeleteSession deletes a session and all related data
func (s *SessionService) DeleteSession(sessionID uuid.UUID) error {
	// First delete all highlights for this session (cascading will delete interactions)
	err := s.repo.DeleteHighlightsBySession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to delete session highlights: %w", err)
	}

	// Then delete the session itself
	err = s.repo.DeleteSession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// ProcessAnswerAndGetNextStep processes a user's answer and returns the next step
func (s *SessionService) ProcessAnswerAndGetNextStep(sessionID uuid.UUID, highlightIndex int, userAnswer string) (*models.ProcessAnswerResponse, error) {
	// Get the highlight by index
	highlights, err := s.repo.GetHighlightsBySession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get highlights: %w", err)
	}

	// Validate highlight index
	if highlightIndex < 0 || highlightIndex >= len(highlights) {
		return nil, fmt.Errorf("invalid highlight index: %d", highlightIndex)
	}

	// Get the current highlight
	currentHighlight := highlights[highlightIndex]

	// Create a new interaction for this highlight
	interactionID := uuid.New()
	interactionParams := repository.CreateInteractionParams{
		ID:          pgtype.UUID{Bytes: interactionID, Valid: true},
		HighlightID: currentHighlight.ID,
		Question:    "What are your thoughts on this highlight?", // Placeholder question, will be updated
		Answer:      pgtype.Text{String: userAnswer, Valid: true},
	}

	_, err = s.repo.CreateInteraction(context.Background(), interactionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create interaction: %w", err)
	}

	// Determine next step
	nextIndex := highlightIndex + 1
	if nextIndex >= len(highlights) {
		// No more highlights, mark session as completed
		err = s.repo.UpdateSessionStatus(context.Background(), repository.UpdateSessionStatusParams{
			ID:     pgtype.UUID{Bytes: sessionID, Valid: true},
			Status: "completed",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update session status: %w", err)
		}

		// Return completion response
		return &models.ProcessAnswerResponse{
			Status:  "completed",
			Message: "Session successfully completed.",
		}, nil
	}

	// Return next step
	nextHighlight := highlights[nextIndex]
	
	// Generate question using LLM
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	question, err := s.llm.GenerateQuestion(ctx, nextHighlight.Text)
	if err != nil {
		// Fallback to placeholder question if LLM fails
		question = "What are your thoughts on this highlight?"
	}
	
	return &models.ProcessAnswerResponse{
		NextStep: &models.NextStep{
			HighlightIndex: nextIndex,
			HighlightText:  nextHighlight.Text,
			Question:       question,
		},
	}, nil
}

// RegenerateQuestion generates a new question for a highlight
func (s *SessionService) RegenerateQuestion(sessionID uuid.UUID, highlightIndex int) (string, error) {
	// Get the highlight by index
	highlights, err := s.repo.GetHighlightsBySession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return "", fmt.Errorf("failed to get highlights: %w", err)
	}

	// Validate highlight index
	if highlightIndex < 0 || highlightIndex >= len(highlights) {
		return "", fmt.Errorf("invalid highlight index: %d", highlightIndex)
	}

	// Get the current highlight
	currentHighlight := highlights[highlightIndex]

	// Generate new question using LLM
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	question, err := s.llm.RegenerateQuestion(ctx, currentHighlight.Text, "What are your thoughts on this highlight?")
	if err != nil {
		return "", fmt.Errorf("failed to regenerate question: %w", err)
	}

	return question, nil
}
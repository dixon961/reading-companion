package service

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/alex/reading-companion/internal/llmclient"
	"github.com/alex/reading-companion/internal/models"
	"github.com/alex/reading-companion/internal/repository"
	"github.com/alex/reading-companion/pkg/markdown"
	"github.com/google/uuid"
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
			// Check if the error is due to LLM service unavailability
			if strings.Contains(err.Error(), "failed to get response after retries") {
				// Return a 503 Service Unavailable error
				return nil, fmt.Errorf("LLM service unavailable: %w", err)
			}
			// Fallback to placeholder question if LLM fails
			question = "What are your thoughts on this highlight?"
		}
		response.NextStep.Question = question

		// Create interaction for first highlight with generated question and null answer
		// First get the highlight ID from database
		dbHighlights, err := s.repo.GetHighlightsBySession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
		if err == nil && len(dbHighlights) > 0 {
			firstHighlight := dbHighlights[0]
			interactionID := uuid.New()
			interactionParams := repository.CreateInteractionParams{
				ID:          pgtype.UUID{Bytes: interactionID, Valid: true},
				HighlightID: firstHighlight.ID,
				Question:    question,
				Answer:      pgtype.Text{Valid: false}, // Null answer for now
			}

			_, err = s.repo.CreateInteraction(context.Background(), interactionParams)
			if err != nil {
				// Log error but don't fail the request, since we can still return the question
				fmt.Printf("Warning: failed to create interaction for first highlight: %v\n", err)
			}
		}
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
	// Reset file pointer to beginning
	file.Seek(0, 0)

	// Try to detect the file format by reading the first few lines
	scanner := bufio.NewScanner(file)
	var firstLines []string
	lineCount := 0
	
	// Read first 10 lines to determine format
	for scanner.Scan() && lineCount < 10 {
		firstLines = append(firstLines, scanner.Text())
		lineCount++
	}
	
	// Reset file pointer again for actual parsing
	file.Seek(0, 0)
	
	// Check if it's the kon-tiki format (has "время：" marker)
	isKonTikiFormat := false
	for _, line := range firstLines {
		if strings.Contains(line, "время：") {
			isKonTikiFormat = true
			break
		}
	}
	
	if isKonTikiFormat {
		return s.parseKonTikiFormat(file)
	}
	return s.parseSimpleFormat(file)
}

// parseSimpleFormat parses the simple format where highlights are separated by empty lines
func (s *SessionService) parseSimpleFormat(file multipart.File) ([]string, error) {
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

// parseKonTikiFormat parses the kon-tiki format which has structured entries
func (s *SessionService) parseKonTikiFormat(file multipart.File) ([]string, error) {
	var highlights []string

	// Reset file pointer to beginning
	file.Seek(0, 0)

	scanner := bufio.NewScanner(file)
	var currentContent strings.Builder
	inContentSection := false

	for scanner.Scan() {
		line := scanner.Text()

		// Check if we're starting a new entry
		if strings.Contains(line, "время：") {
			// If we have content from previous entry, save it
			if currentContent.Len() > 0 {
				content := strings.TrimSpace(currentContent.String())
				if content != "" {
					highlights = append(highlights, content)
				}
				currentContent.Reset()
			}
			inContentSection = false
			continue
		}

		// Check for content section start marker
		if strings.Contains(line, "【Контент】") {
			inContentSection = true
			// Remove the section marker and add the rest of the line if it exists
			contentStart := strings.Index(line, "【Контент】") + len("【Контент】")
			if contentStart < len(line) {
				trimmedLine := strings.TrimSpace(line[contentStart:])
				if trimmedLine != "" {
					currentContent.WriteString(trimmedLine)
					// Add space if content is not empty
					if trimmedLine != "" {
						currentContent.WriteString(" ")
					}
				}
			}
			continue
		}

		// Check for content section end markers
		if strings.Contains(line, "【Заметки】") || strings.HasPrefix(line, "-------------------") {
			// End of content section, save current content
			if currentContent.Len() > 0 {
				content := strings.TrimSpace(currentContent.String())
				if content != "" {
					highlights = append(highlights, content)
				}
				currentContent.Reset()
			}
			inContentSection = false
			continue
		}

		// Add line to content if we're in content section
		if inContentSection {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" {
				currentContent.WriteString(trimmedLine)
				// Add space after each line for proper joining
				currentContent.WriteString(" ")
			}
		}
	}

	// Don't forget the last highlight if there's content and we haven't saved it yet
	if currentContent.Len() > 0 {
		content := strings.TrimSpace(currentContent.String())
		if content != "" {
			highlights = append(highlights, content)
		}
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

	// Generate question for current highlight (if not already generated)
	// First check if there's already an interaction for this highlight
	interactions, err := s.repo.GetInteractionsByHighlight(context.Background(), currentHighlight.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get interactions: %w", err)
	}

	var question string
	if len(interactions) > 0 {
		// Use the question from the existing interaction
		question = interactions[0].Question
	} else {
		// This shouldn't happen in normal flow, but just in case
		question = "What are your thoughts on this highlight?"
	}

	// Create or update interaction for this highlight with the answer
	if len(interactions) > 0 {
		// Update existing interaction with answer
		interaction := interactions[0]
		_, err = s.repo.UpdateInteractionAnswer(context.Background(), repository.UpdateInteractionAnswerParams{
			ID:     interaction.ID,
			Answer: pgtype.Text{String: userAnswer, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update interaction: %w", err)
		}
	} else {
		// Create new interaction (this is for the first call)
		interactionID := uuid.New()
		interactionParams := repository.CreateInteractionParams{
			ID:          pgtype.UUID{Bytes: interactionID, Valid: true},
			HighlightID: currentHighlight.ID,
			Question:    question,
			Answer:      pgtype.Text{String: userAnswer, Valid: true},
		}

		_, err = s.repo.CreateInteraction(context.Background(), interactionParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create interaction: %w", err)
		}
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

	// Generate question for next highlight
	nextHighlight := highlights[nextIndex]

	// Generate question using LLM
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	question, err = s.llm.GenerateQuestion(ctx, nextHighlight.Text)
	if err != nil {
		// Check if the error is due to LLM service unavailability
		if strings.Contains(err.Error(), "failed to get response after retries") {
			// Return a 503 Service Unavailable error
			return nil, fmt.Errorf("LLM service unavailable: %w", err)
		}
		// Fallback to placeholder question if LLM fails
		question = "What are your thoughts on this highlight?"
	}

	// Create interaction for next highlight with generated question and null answer
	nextInteractionID := uuid.New()
	nextInteractionParams := repository.CreateInteractionParams{
		ID:          pgtype.UUID{Bytes: nextInteractionID, Valid: true},
		HighlightID: nextHighlight.ID,
		Question:    question,
		Answer:      pgtype.Text{Valid: false}, // Null answer for now
	}

	_, err = s.repo.CreateInteraction(context.Background(), nextInteractionParams)
	if err != nil {
		// Log error but don't fail the request, since we can still return the question
		fmt.Printf("Warning: failed to create interaction for next highlight: %v\n", err)
	}

	return &models.ProcessAnswerResponse{
		NextStep: &models.NextStep{
			HighlightIndex: nextIndex,
			HighlightText:  nextHighlight.Text,
			Question:       question,
		},
	}, nil
}

// RegenerateQuestion generates a new question for a highlight and updates the database
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
		// Check if the error is due to LLM service unavailability
		if strings.Contains(err.Error(), "failed to get response after retries") {
			// Return a 503 Service Unavailable error
			return "", fmt.Errorf("LLM service unavailable: %w", err)
		}
		return "", fmt.Errorf("failed to regenerate question: %w", err)
	}

	// Update the interaction with the new question
	// First get existing interaction for this highlight
	interactions, err := s.repo.GetInteractionsByHighlight(context.Background(), currentHighlight.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get interactions: %w", err)
	}

	if len(interactions) > 0 {
		// Update existing interaction with new question
		interaction := interactions[0]
		_, err = s.repo.UpdateInteractionQuestion(context.Background(), repository.UpdateInteractionQuestionParams{
			ID:       interaction.ID,
			Question: question,
		})
		if err != nil {
			return "", fmt.Errorf("failed to update interaction: %w", err)
		}
	} else {
		// Create new interaction with the regenerated question
		interactionID := uuid.New()
		interactionParams := repository.CreateInteractionParams{
			ID:          pgtype.UUID{Bytes: interactionID, Valid: true},
			HighlightID: currentHighlight.ID,
			Question:    question,
			Answer:      pgtype.Text{Valid: false}, // Null answer for now
		}

		_, err = s.repo.CreateInteraction(context.Background(), interactionParams)
		if err != nil {
			return "", fmt.Errorf("failed to create interaction: %w", err)
		}
	}

	return question, nil
}

// ExportSessionAsMarkdown exports a completed session as a markdown document
func (s *SessionService) ExportSessionAsMarkdown(sessionID uuid.UUID) (string, error) {
	// Get the session
	session, err := s.repo.GetSession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return "", fmt.Errorf("session not found: %w", err)
	}

	// Check if session is completed
	if session.Status != "completed" {
		return "", fmt.Errorf("session not completed: only completed sessions can be exported")
	}

	// Get highlights for this session
	highlights, err := s.repo.GetHighlightsBySession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return "", fmt.Errorf("failed to get highlights: %w", err)
	}

	// Get interactions for all highlights
	interactions := make(map[string]*models.Interaction)
	for _, highlight := range highlights {
		interaction, err := s.repo.GetInteractionByHighlight(context.Background(), highlight.ID)
		if err != nil {
			// Skip if no interaction found for this highlight
			continue
		}

		interactions[uuid.UUID(highlight.ID.Bytes).String()] = &models.Interaction{
			ID:          uuid.UUID(interaction.ID.Bytes),
			HighlightID: uuid.UUID(interaction.HighlightID.Bytes),
			Question:    interaction.Question,
			Answer: sql.NullString{
				String: interaction.Answer.String,
				Valid:  interaction.Answer.Valid,
			},
			CreatedAt: interaction.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: interaction.UpdatedAt.Time.Format(time.RFC3339),
		}
	}

	// Convert repository highlights to models highlights
	modelHighlights := make([]*models.Highlight, len(highlights))
	for i, highlight := range highlights {
		modelHighlights[i] = &models.Highlight{
			ID:        uuid.UUID(highlight.ID.Bytes),
			SessionID: uuid.UUID(highlight.SessionID.Bytes),
			Text:      highlight.Text,
			Position:  int(highlight.Position),
			CreatedAt: highlight.CreatedAt.Time,
		}
	}

	// Convert repository session to models session
	modelSession := &models.Session{
		ID:        uuid.UUID(session.ID.Bytes),
		Name:      session.Name,
		Status:    session.Status,
		CreatedAt: session.CreatedAt.Time,
		UpdatedAt: session.UpdatedAt.Time,
	}

	// Generate markdown
	return markdown.GenerateMarkdown(modelSession, modelHighlights, interactions), nil
}

// GetSessionContentAsJSON returns the content of a completed session in JSON format
func (s *SessionService) GetSessionContentAsJSON(sessionID uuid.UUID) (*models.SessionContent, error) {
	// Get the session
	session, err := s.repo.GetSession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check if session is completed
	if session.Status != "completed" {
		return nil, fmt.Errorf("session not completed: only completed sessions can be viewed")
	}

	// Get highlights for this session
	highlights, err := s.repo.GetHighlightsBySession(context.Background(), pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get highlights: %w", err)
	}

	// Get interactions for all highlights
	interactions := make(map[string]*models.Interaction)
	for _, highlight := range highlights {
		interaction, err := s.repo.GetInteractionByHighlight(context.Background(), highlight.ID)
		if err != nil {
			// Skip if no interaction found for this highlight
			continue
		}

		interactions[uuid.UUID(highlight.ID.Bytes).String()] = &models.Interaction{
			ID:          uuid.UUID(interaction.ID.Bytes),
			HighlightID: uuid.UUID(interaction.HighlightID.Bytes),
			Question:    interaction.Question,
			Answer: sql.NullString{
				String: interaction.Answer.String,
				Valid:  interaction.Answer.Valid,
			},
			CreatedAt: interaction.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: interaction.UpdatedAt.Time.Format(time.RFC3339),
		}
	}

	// Create highlight content
	highlightContents := make([]models.HighlightContent, len(highlights))
	for i, highlight := range highlights {
		highlightContent := models.HighlightContent{
			Text:     highlight.Text,
			Question: "No question available",
			Answer:   "",
			Answered: false,
		}

		// Add interaction data if it exists
		if interaction, exists := interactions[uuid.UUID(highlight.ID.Bytes).String()]; exists {
			highlightContent.Question = interaction.Question
			if interaction.Answer.Valid {
				highlightContent.Answer = interaction.Answer.String
				highlightContent.Answered = true
			}
		}

		highlightContents[i] = highlightContent
	}

	// Create session content
	sessionContent := &models.SessionContent{
		Session: models.SessionContentInfo{
			ID:        uuid.UUID(session.ID.Bytes).String(),
			Name:      session.Name,
			CreatedAt: session.CreatedAt.Time.Format(time.RFC3339),
		},
		Highlights: highlightContents,
	}

	return sessionContent, nil
}

package llmclient

import (
	"context"
)

// Client defines the interface for interacting with an LLM
type Client interface {
	// GenerateQuestion generates a question for a given highlight text
	GenerateQuestion(ctx context.Context, highlightText string) (string, error)
	
	// RegenerateQuestion generates an alternative question for a given highlight text
	RegenerateQuestion(ctx context.Context, highlightText string, previousQuestion string) (string, error)
}
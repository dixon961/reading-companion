package llmclient

import (
	"context"
)

// mockClient implements the Client interface for testing purposes
type mockClient struct {
	// We don't need any fields for the mock client
}

// NewMock creates a new mock LLM client that returns placeholder responses immediately
func NewMock() Client {
	return &mockClient{}
}

// GenerateQuestion returns a placeholder question immediately
func (c *mockClient) GenerateQuestion(ctx context.Context, highlightText string) (string, error) {
	// Return a placeholder question immediately
	return "What are your thoughts on this highlight?", nil
}

// RegenerateQuestion returns a placeholder question immediately
func (c *mockClient) RegenerateQuestion(ctx context.Context, highlightText string, previousQuestion string) (string, error) {
	// Return a different placeholder question immediately
	return "Can you elaborate on your understanding of this point?", nil
}

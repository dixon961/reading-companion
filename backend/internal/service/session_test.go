package service

import (
	"strings"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

// mockFile is a simple implementation of multipart.File for testing
type mockFile struct {
	*strings.Reader
}

func (m *mockFile) Close() error {
	return nil
}

// TestSessionService_validateHighlights tests the validateHighlights method
func TestSessionService_validateHighlights(t *testing.T) {
	service := &SessionService{}
	
	// Test with valid highlights (more than 3)
	validHighlights := []string{"Highlight 1", "Highlight 2", "Highlight 3", "Highlight 4"}
	err := service.validateHighlights(validHighlights)
	assert.NoError(t, err)
	
	// Test with invalid highlights (less than 3)
	invalidHighlights := []string{"Highlight 1", "Highlight 2"}
	err = service.validateHighlights(invalidHighlights)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum 3 highlights required")
	
	// Test with empty highlights
	emptyHighlights := []string{}
	err = service.validateHighlights(emptyHighlights)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum 3 highlights required")
}

// TestSessionService_parseHighlights tests the parseHighlights method
func TestSessionService_parseHighlights(t *testing.T) {
	service := &SessionService{}
	
	// Test with valid highlights
	validContent := "First highlight\n\nSecond highlight\n\nThird highlight"
	reader := &mockFile{strings.NewReader(validContent)}
	highlights, err := service.parseHighlights(reader)
	
	assert.NoError(t, err)
	assert.Len(t, highlights, 3)
	assert.Equal(t, "First highlight", highlights[0])
	assert.Equal(t, "Second highlight", highlights[1])
	assert.Equal(t, "Third highlight", highlights[2])
	
	// Test with empty content
	emptyReader := &mockFile{strings.NewReader("")}
	highlights, err = service.parseHighlights(emptyReader)
	
	assert.NoError(t, err)
	assert.Len(t, highlights, 0)
	
	// Test with content that has extra newlines
	extraNewlines := "\n\nFirst highlight\n\n\n\nSecond highlight\n\n\n"
	reader = &mockFile{strings.NewReader(extraNewlines)}
	highlights, err = service.parseHighlights(reader)
	
	assert.NoError(t, err)
	assert.Len(t, highlights, 2)
	assert.Equal(t, "First highlight", highlights[0])
	assert.Equal(t, "Second highlight", highlights[1])
}
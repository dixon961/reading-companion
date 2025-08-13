package llmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Config holds the configuration for the LLM client
type Config struct {
	APIEndpoint string `mapstructure:"api_endpoint"`
	APIKey      string `mapstructure:"api_key"`
	Model       string `mapstructure:"model"`
	Timeout     int    `mapstructure:"timeout"` // in seconds
}

// openAIRequest represents the request structure for OpenAI API
type openAIRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

// message represents a single message in the conversation
type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIResponse represents the response structure from OpenAI API
type openAIResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// client implements the Client interface
type client struct {
	config Config
	client *http.Client
}

// New creates a new LLM client
func New(config Config) Client {
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second // default timeout
	}
	
	return &client{
		config: config,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// GenerateQuestion generates a question for a given highlight text
func (c *client) GenerateQuestion(ctx context.Context, highlightText string) (string, error) {
	// Create the prompt for the LLM
	prompt := fmt.Sprintf(`You are an intelligent reading assistant that helps users deeply understand and reflect on book highlights. 
Your task is to generate a thoughtful, open-ended question about the following highlight that will encourage the user to think critically and make connections:

"%s"

The question should:
1. Be open-ended (not yes/no)
2. Encourage deeper thinking about the concept
3. Help the user connect the idea to their own experience or other knowledge
4. Be clear and concise

Return only the question text, nothing else.`, highlightText)

	return c.callLLM(ctx, prompt)
}

// RegenerateQuestion generates an alternative question for a given highlight text
func (c *client) RegenerateQuestion(ctx context.Context, highlightText string, previousQuestion string) (string, error) {
	// Create the prompt for regenerating a question
	prompt := fmt.Sprintf(`You are an intelligent reading assistant that helps users deeply understand and reflect on book highlights. 
Your task is to generate an alternative question about the following highlight. The user was not satisfied with the previous question and wants a different perspective.

Highlight: "%s"
Previous question: "%s"

Generate a new question that:
1. Is open-ended (not yes/no)
2. Offers a different angle or perspective than the previous question
3. Encourages deeper thinking about the concept
4. Helps the user connect the idea to their own experience or other knowledge
5. Is clear and concise

Return only the new question text, nothing else.`, highlightText, previousQuestion)

	return c.callLLM(ctx, prompt)
}

// callLLM makes the actual HTTP request to the LLM API
func (c *client) callLLM(ctx context.Context, prompt string) (string, error) {
	// Prepare the request
	requestBody := openAIRequest{
		Model: c.config.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	// Make the request with retry logic
	var lastErr error
	for i := 0; i < 3; i++ { // Retry up to 3 times
		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to make request: %w", err)
			time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		// Parse response
		var response openAIResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			lastErr = fmt.Errorf("failed to decode response: %w", err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Check for API errors
		if response.Error != nil {
			lastErr = fmt.Errorf("API error: %s", response.Error.Message)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Return the generated question
		if len(response.Choices) > 0 {
			return response.Choices[0].Message.Content, nil
		}

		lastErr = fmt.Errorf("no choices in response")
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return "", fmt.Errorf("failed to get response after retries: %w", lastErr)
}
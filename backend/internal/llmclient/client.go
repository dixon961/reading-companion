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
	Timeout     int    `mapstructure:"timeout"`    // in seconds
	MaxTokens   int    `mapstructure:"max_tokens"` // Optional: for completion endpoint
}

// message represents a single message in the chat
type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest represents the request structure for a chat completions endpoint
type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

// chatResponse represents the response structure for a chat completions endpoint
type chatResponse struct {
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
	// Create the system prompt for the LLM
	systemPrompt := `Вы — умный помощник для чтения, который помогает пользователям глубже понять и осмыслить выделенные фрагменты из книг.
Ваша задача — сгенерировать содержательный, открытый вопрос к следующему фрагменту, который побудит пользователя критически мыслить и устанавливать связи.
Вопрос должен:
1. Быть открытым (не предполагать ответа «да» или «нет»).
2. Побуждать глубже задуматься о концепции.
3. Помогать пользователю связать идею с собственным опытом или другими знаниями.
4. Быть ясным и лаконичным.
Верните только текст вопроса, без лишнего. Только одно предложение! Строго!`

	// Create the user message with the highlight text
	userMessage := fmt.Sprintf(`Сгенерируйте вопрос для следующего фрагмента:
«%s»`, highlightText)

	return c.callLLM(ctx, systemPrompt, userMessage)
}

// RegenerateQuestion generates an alternative question for a given highlight text
func (c *client) RegenerateQuestion(ctx context.Context, highlightText string, previousQuestion string) (string, error) {
	// Create the system prompt for regenerating a question
	systemPrompt := `Вы — умный помощник для чтения, который помогает пользователям глубже понять и осмыслить выделенные фрагменты из книг.
Ваша задача — сгенерировать альтернативный вопрос к следующему фрагменту. Пользователя не устроил предыдущий вопрос, и он хочет получить иной взгляд.
Сгенерируйте новый вопрос, который:
1. Является открытым (не предполагает ответа «да» или «нет»).
2. Предлагает иной ракурс или перспективу, чем предыдущий вопрос.
3. Побуждает глубже задуматься о концепции.
4. Помогает пользователю связать идею с собственным опытом или другими знаниями.
5. Ясный и лаконичный.
Верните только текст нового вопроса, без лишнего. Только одно предложение! Строго!`

	// Create the user message with the highlight text and previous question
	userMessage := fmt.Sprintf(`Выделенный фрагмент: «%s»
Предыдущий вопрос: «%s»
Сгенерируйте альтернативный вопрос.`, highlightText, previousQuestion)

	return c.callLLM(ctx, systemPrompt, userMessage)
}

// callLLM makes the actual HTTP request to the LLM API using chat completions format
func (c *client) callLLM(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	// Prepare the request for a chat completions endpoint
	requestBody := chatRequest{
		Model: c.config.Model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
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
		var response chatResponse
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

		// Return the generated text
		if len(response.Choices) > 0 {
			return response.Choices[0].Message.Content, nil
		}

		lastErr = fmt.Errorf("no choices in response")
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return "", fmt.Errorf("failed to get response after retries: %w", lastErr)
}

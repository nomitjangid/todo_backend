package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"todo-backend/internal/config"
)

const openAIExtractionPrompt = `
You are a highly efficient task extraction AI. Your sole purpose is to parse user-provided text and extract structured tasks in a strict JSON array format.

Current Date: November 19, 2025

Here are the rules:
- ALWAYS respond with a JSON array of tasks. Do not include any other prose, explanations, or text outside the JSON array.
- If no tasks can be extracted, return an empty JSON array: []
- Each task object must adhere to the following strict JSON schema:
  {
    "title": "string",            // Required: A concise summary of the task.
    "description": "string",      // Required: A detailed description of the task. If not explicitly provided, infer from the title.
    "due_date": "string",         // Required: The due date of the task in ISO 8601 format (e.g., "2025-11-23T10:00:00Z"). If no specific time is given, default to 00:00:00Z on the specified date. If no date is mentioned, use null.
    "priority": "string",         // Required: The priority of the task. Must be one of: "low", "medium", "high". Default to "medium" if not specified.
    "subtasks": ["string"]        // Required: An array of strings, where each string is a subtask. If no subtasks, return an empty array [].
  }
- Handle natural date expressions (e.g., "tomorrow", "next week", "Monday morning", "in 3 days"). Convert them to the appropriate ISO 8601 timestamp relative to the current date and time.
- Detect multiple tasks within a single input text.
- Ensure all required fields are present. Infer if necessary.
- On failure to extract or parse, return an empty array [].
`

// OpenAIExtractor implements the TaskExtractor interface using OpenAI's API.
type OpenAIExtractor struct {
	apiKey      string
	apiBaseURL  string
	httpClient  *http.Client
}

// NewOpenAIExtractor creates a new OpenAIExtractor.
func NewOpenAIExtractor(cfg *config.Config) *OpenAIExtractor {
	return &OpenAIExtractor{
		apiKey:     cfg.OpenAPIKey,
		apiBaseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{},
	}
}

// NewOpenAIExtractorWithClient creates a new OpenAIExtractor with a custom HTTP client and base URL (for testing).
func NewOpenAIExtractorWithClient(apiKey, apiBaseURL string, client *http.Client) *OpenAIExtractor {
	return &OpenAIExtractor{
		apiKey:     apiKey,
		apiBaseURL: apiBaseURL,
		httpClient: client,
	}
}

// ExtractTasks extracts tasks from text using OpenAI's GPT model.
func (e *OpenAIExtractor) ExtractTasks(ctx context.Context, text string) ([]Task, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "gpt-3.5-turbo", // or "gpt-4" for better results
		"messages": []map[string]string{
			{"role": "system", "content": openAIExtractionPrompt},
			{"role": "user", "content": text},
		},
		"response_format": map[string]string{"type": "json_object"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.apiBaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api error: status %d, body: %s", resp.StatusCode, respBody)
	}

	var openaiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openaiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	if len(openaiResponse.Choices) == 0 {
		return []Task{}, nil
	}

	var tasks []Task
	// The content from OpenAI might be a string containing the JSON array.
	// We need to unmarshal that string.
	err = json.Unmarshal([]byte(openaiResponse.Choices[0].Message.Content), &tasks)
	if err != nil {
		// If unmarshaling fails, it means the LLM didn't return strict JSON,
		// or there was another parsing error. As per requirement, return empty array.
		return []Task{}, fmt.Errorf("failed to unmarshal tasks from LLM response: %w. Response content: %s", err, openaiResponse.Choices[0].Message.Content)
	}

	return tasks, nil
}

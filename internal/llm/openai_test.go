package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOpenAIExtractor_ExtractTasks(t *testing.T) {
	// Mock OpenAI API Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Contains(t, reqBody, "model")
		assert.Contains(t, reqBody, "messages")
		assert.Contains(t, reqBody, "response_format")

		// Simulate different responses based on input text
		userContent := reqBody["messages"].([]interface{})[1].(map[string]interface{})["content"].(string)

		if strings.Contains(userContent, "buy milk") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"choices": [
					{
						"message": {
							"content": "[{\"title\": \"Buy milk\", \"description\": \"Buy milk tomorrow\", \"due_date\": \"2025-11-20T08:00:00Z\", \"priority\": \"medium\", \"subtasks\": []}]"
						}
					}
				]
			}`))
		} else if strings.Contains(userContent, "invalid json") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"choices": [
					{
						"message": {
							"content": "This is not JSON"
						}
					}
				]
			}`))
		} else if strings.Contains(userContent, "failed API call") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "internal server error"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"choices": [
					{
						"message": {
							"content": "[]"
						}
					}
				]
			}`))
		}
	}))
	defer server.Close()

	// Create an extractor instance that uses the mock server
	extractor := NewOpenAIExtractorWithClient("test-api-key", server.URL+"/v1", server.Client())

	t.Run("should extract a single task", func(t *testing.T) {
		text := "Tomorrow buy milk and call Raj"
		tasks, err := extractor.ExtractTasks(context.Background(), text)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, "Buy milk", tasks[0].Title)
		assert.Equal(t, "medium", tasks[0].Priority)
		expectedDate, _ := time.Parse(time.RFC3339, "2025-11-20T08:00:00Z") // Adjusted for "tomorrow" based on test prompt's current date
		assert.Equal(t, expectedDate, tasks[0].DueDate)
	})

	t.Run("should return empty array for no tasks", func(t *testing.T) {
		text := "Just some random text."
		tasks, err := extractor.ExtractTasks(context.Background(), text)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("should handle invalid JSON from LLM gracefully", func(t *testing.T) {
		text := "This is invalid json"
		tasks, err := extractor.ExtractTasks(context.Background(), text)
		assert.Error(t, err) // Should return an error indicating unmarshal failure
		assert.Empty(t, tasks)
		assert.Contains(t, err.Error(), "failed to unmarshal tasks from LLM response")
	})

	t.Run("should handle API errors", func(t *testing.T) {
		text := "failed API call"
		tasks, err := extractor.ExtractTasks(context.Background(), text)
		assert.Error(t, err)
		assert.Empty(t, tasks)
		assert.Contains(t, err.Error(), "openai api error")
	})
}

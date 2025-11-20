package llm

import (
	"context"
	"time"
)

// Task represents a structured task extracted by the LLM
type Task struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Priority    string    `json:"priority"` // low|medium|high
	Subtasks    []string  `json:"subtasks"`
}

// TaskExtractor defines the interface for LLM-based task extraction
type TaskExtractor interface {
	ExtractTasks(ctx context.Context, text string) ([]Task, error)
}

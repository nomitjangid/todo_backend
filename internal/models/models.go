package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the database
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Task represents a task in the database
type Task struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Priority    string    `json:"priority"`
	RawText     string    `json:"raw_text"`
	CreatedAt   time.Time `json:"created_at"`
}

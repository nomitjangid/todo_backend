package models

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	Priority    string     `json:"priority" gorm:"default:'medium'"`
	RawText     string     `json:"raw_text"`
	Completed   bool       `json:"completed" gorm:"default:false"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	Priority    string     `json:"priority"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	Priority    *string    `json:"priority"`
	Completed   *bool      `json:"completed"`
}

type ExtractTasksRequest struct {
	Text string `json:"text" binding:"required"`
}

type ExtractedTask struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	Priority    string     `json:"priority"`
}

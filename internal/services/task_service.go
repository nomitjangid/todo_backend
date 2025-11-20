package services

import (
	"context"
	"errors"
	"fmt"
	"todo-backend/internal/llm"
	"todo-backend/internal/models"
	"todo-backend/internal/repositories"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskService handles task-related business logic
type TaskService struct {
	taskRepo    repositories.TaskRepositoryInterface
	llmExtractor llm.TaskExtractor
}

// NewTaskService creates a new TaskService
func NewTaskService(taskRepo repositories.TaskRepositoryInterface, llmExtractor llm.TaskExtractor) *TaskService {
	return &TaskService{
		taskRepo:    taskRepo,
		llmExtractor: llmExtractor,
	}
}

// CreateTask creates a new task
func (s *TaskService) CreateTask(task *models.Task) error {
	return s.taskRepo.CreateTask(task)
}

// GetTaskByID retrieves a task by its ID
func (s *TaskService) GetTaskByID(id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetTaskByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found or unauthorized")
		}
		return nil, err
	}
	return task, nil
}

// GetTasksByUserID retrieves all tasks for a given user ID
func (s *TaskService) GetTasksByUserID(userID uuid.UUID) ([]models.Task, error) {
	return s.taskRepo.GetTasksByUserID(userID)
}

// UpdateTask updates an existing task
func (s *TaskService) UpdateTask(task *models.Task, userID uuid.UUID) error {
	// Ensure the user owns the task
	existingTask, err := s.taskRepo.GetTaskByID(task.ID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("task not found or unauthorized")
		}
		return err
	}

	// Update fields
	existingTask.Title = task.Title
	existingTask.Description = task.Description
	existingTask.DueDate = task.DueDate
	existingTask.Priority = task.Priority
	existingTask.RawText = task.RawText

	return s.taskRepo.UpdateTask(existingTask)
}

// DeleteTask deletes a task
func (s *TaskService) DeleteTask(id uuid.UUID, userID uuid.UUID) error {
	err := s.taskRepo.DeleteTask(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("task not found or unauthorized")
		}
		return err
	}
	return nil
}

// ExtractAndCreateTasks extracts tasks from text and creates them in the database
func (s *TaskService) ExtractAndCreateTasks(ctx context.Context, text string, userID uuid.UUID) ([]models.Task, error) {
	extractedLLMTasks, err := s.llmExtractor.ExtractTasks(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tasks with LLM: %w", err)
	}

	var createdTasks []models.Task
	for _, llmTask := range extractedLLMTasks {
		task := &models.Task{
			ID:          uuid.New(),
			UserID:      userID,
			Title:       llmTask.Title,
			Description: llmTask.Description,
			DueDate: func() *time.Time {
				if llmTask.DueDate.IsZero() {
					return nil
				}
				return &llmTask.DueDate
			}(),
			Priority:    llmTask.Priority,
			RawText:     text, // Store the raw text that led to this task
		}
		if err := s.taskRepo.CreateTask(task); err != nil {
			// Log the error but try to continue with other tasks
			// Or decide if you want to fail all if one fails
			continue
		}
		createdTasks = append(createdTasks, *task)
	}
	return createdTasks, nil
}

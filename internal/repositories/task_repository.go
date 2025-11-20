package repositories

import (
	"todo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskRepositoryInterface defines the methods for interacting with task data
type TaskRepositoryInterface interface {
	CreateTask(task *models.Task) error
	GetTaskByID(id uuid.UUID, userID uuid.UUID) (*models.Task, error)
	GetTasksByUserID(userID uuid.UUID) ([]models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id uuid.UUID, userID uuid.UUID) error
}

// TaskRepository handles database operations for tasks
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new TaskRepository
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// CreateTask creates a new task in the database
func (r *TaskRepository) CreateTask(task *models.Task) error {
	return r.db.Create(task).Error
}

// GetTaskByID retrieves a task by its ID
func (r *TaskRepository) GetTaskByID(id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	return &task, err
}

// GetTasksByUserID retrieves all tasks for a given user ID
func (r *TaskRepository) GetTasksByUserID(userID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("user_id = ?", userID).Find(&tasks).Error
	return tasks, err
}

// UpdateTask updates an existing task in the database
func (r *TaskRepository) UpdateTask(task *models.Task) error {
	return r.db.Save(task).Error
}

// DeleteTask deletes a task from the database
func (r *TaskRepository) DeleteTask(id uuid.UUID, userID uuid.UUID) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Task{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

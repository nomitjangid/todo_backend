package services

import (
	"context"
	"errors"
	"testing"
	"time"
	"todo-backend/internal/llm"
	"todo-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockTaskRepository is a mock implementation of TaskRepositoryInterface
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) CreateTask(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetTaskByID(id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) GetTasksByUserID(userID uuid.UUID) ([]models.Task, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskRepository) UpdateTask(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) DeleteTask(id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

// MockLLMExtractor is a mock implementation of llm.TaskExtractor
type MockLLMExtractor struct {
	mock.Mock
}

func (m *MockLLMExtractor) ExtractTasks(ctx context.Context, text string) ([]llm.Task, error) {
	args := m.Called(ctx, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]llm.Task), args.Error(1)
}

func TestTaskService_CreateTask(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockLLMExtractor := new(MockLLMExtractor)
	taskService := NewTaskService(mockTaskRepo, mockLLMExtractor)

	t.Run("successfully creates a task", func(t *testing.T) {
		task := &models.Task{ID: uuid.New(), UserID: uuid.New(), Title: "Test Task"}
		mockTaskRepo.On("CreateTask", task).Return(nil).Once()

		err := taskService.CreateTask(task)
		assert.NoError(t, err)
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("returns error if CreateTask fails", func(t *testing.T) {
		task := &models.Task{ID: uuid.New(), UserID: uuid.New(), Title: "Test Task"}
		mockTaskRepo.On("CreateTask", task).Return(errors.New("db error")).Once()

		err := taskService.CreateTask(task)
		assert.Error(t, err)
		assert.EqualError(t, err, "db error")
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetTaskByID(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockLLMExtractor := new(MockLLMExtractor)
	taskService := NewTaskService(mockTaskRepo, mockLLMExtractor)

	userID := uuid.New()
	taskID := uuid.New()
	testTask := &models.Task{ID: taskID, UserID: userID, Title: "Test Task"}

	t.Run("successfully retrieves a task by ID", func(t *testing.T) {
		mockTaskRepo.On("GetTaskByID", taskID, userID).Return(testTask, nil).Once()

		task, err := taskService.GetTaskByID(taskID, userID)
		assert.NoError(t, err)
		assert.Equal(t, testTask, task)
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("returns error if task not found", func(t *testing.T) {
		mockTaskRepo.On("GetTaskByID", taskID, userID).Return(nil, gorm.ErrRecordNotFound).Once()

		task, err := taskService.GetTaskByID(taskID, userID)
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.EqualError(t, err, "task not found or unauthorized")
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetTasksByUserID(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockLLMExtractor := new(MockLLMExtractor)
	taskService := NewTaskService(mockTaskRepo, mockLLMExtractor)

	userID := uuid.New()
	testTasks := []models.Task{
		{ID: uuid.New(), UserID: userID, Title: "Task 1"},
		{ID: uuid.New(), UserID: userID, Title: "Task 2"},
	}

	t.Run("successfully retrieves tasks by user ID", func(t *testing.T) {
		mockTaskRepo.On("GetTasksByUserID", userID).Return(testTasks, nil).Once()

		tasks, err := taskService.GetTasksByUserID(userID)
		assert.NoError(t, err)
		assert.Equal(t, testTasks, tasks)
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("returns empty slice if no tasks found", func(t *testing.T) {
		mockTaskRepo.On("GetTasksByUserID", userID).Return([]models.Task{}, nil).Once()

		tasks, err := taskService.GetTasksByUserID(userID)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestTaskService_UpdateTask(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockLLMExtractor := new(MockLLMExtractor)
	taskService := NewTaskService(mockTaskRepo, mockLLMExtractor)

	userID := uuid.New()
	taskID := uuid.New()
	originalTask := &models.Task{ID: taskID, UserID: userID, Title: "Original", Description: "Desc", Priority: "medium"}
	updatedTaskInput := &models.Task{ID: taskID, UserID: userID, Title: "Updated", Description: "New Desc", Priority: "high"}

	t.Run("successfully updates a task", func(t *testing.T) {
		mockTaskRepo.On("GetTaskByID", taskID, userID).Return(originalTask, nil).Once()
		mockTaskRepo.On("UpdateTask", mock.AnythingOfType("*models.Task")).Return(nil).Once()

		err := taskService.UpdateTask(updatedTaskInput, userID)
		assert.NoError(t, err)

		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("returns error if task not found for update", func(t *testing.T) {
		mockTaskRepo.On("GetTaskByID", taskID, userID).Return(nil, gorm.ErrRecordNotFound).Once()

		err := taskService.UpdateTask(updatedTaskInput, userID)
		assert.Error(t, err)
		assert.EqualError(t, err, "task not found or unauthorized")
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockLLMExtractor := new(MockLLMExtractor)
	taskService := NewTaskService(mockTaskRepo, mockLLMExtractor)

	userID := uuid.New()
	taskID := uuid.New()

	t.Run("successfully deletes a task", func(t *testing.T) {
		mockTaskRepo.On("DeleteTask", taskID, userID).Return(nil).Once()

		err := taskService.DeleteTask(taskID, userID)
		assert.NoError(t, err)
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("returns error if task not found for deletion", func(t *testing.T) {
		mockTaskRepo.On("DeleteTask", taskID, userID).Return(gorm.ErrRecordNotFound).Once()

		err := taskService.DeleteTask(taskID, userID)
		assert.Error(t, err)
		assert.EqualError(t, err, "task not found or unauthorized")
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestTaskService_ExtractAndCreateTasks(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockLLMExtractor := new(MockLLMExtractor)
	taskService := NewTaskService(mockTaskRepo, mockLLMExtractor)

	userID := uuid.New()
	inputText := "Buy milk tomorrow"
	extractedLLMTasks := []llm.Task{
		{
			Title:       "Buy milk",
			Description: "Buy milk tomorrow",
			DueDate:     time.Date(2025, time.November, 20, 0, 0, 0, 0, time.UTC),
			Priority:    "medium",
			Subtasks:    []string{},
		},
	}

	t.Run("successfully extracts and creates tasks", func(t *testing.T) {
		mockLLMExtractor.On("ExtractTasks", mock.AnythingOfType("context.backgroundCtx"), inputText).Return(extractedLLMTasks, nil).Once()
		mockTaskRepo.On("CreateTask", mock.AnythingOfType("*models.Task")).Return(nil).Once()

		createdTasks, err := taskService.ExtractAndCreateTasks(context.Background(), inputText, userID)
		assert.NoError(t, err)
		assert.Len(t, createdTasks, 1)
		assert.Equal(t, "Buy milk", createdTasks[0].Title)
		mockLLMExtractor.AssertExpectations(t)
		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("returns error if LLM extraction fails", func(t *testing.T) {
		mockLLMExtractor.On("ExtractTasks", mock.AnythingOfType("context.backgroundCtx"), inputText).Return(nil, errors.New("llm error")).Once()

		createdTasks, err := taskService.ExtractAndCreateTasks(context.Background(), inputText, userID)
		assert.Error(t, err)
		assert.Nil(t, createdTasks)
		assert.Contains(t, err.Error(), "failed to extract tasks with LLM")
		mockLLMExtractor.AssertExpectations(t)
		mockTaskRepo.AssertNotCalled(t, "CreateTask") // Ensure no DB interaction if extraction fails
	})

	t.Run("handles multiple tasks and skips failed creations", func(t *testing.T) {
		extractedMultiLLMTasks := []llm.Task{
			{Title: "Task 1"},
			{Title: "Task 2"},
		}
		mockLLMExtractor.On("ExtractTasks", mock.AnythingOfType("context.backgroundCtx"), inputText).Return(extractedMultiLLMTasks, nil).Once()
		// Simulate one task creation failure
		mockTaskRepo.On("CreateTask", mock.AnythingOfType("*models.Task")).Return(errors.New("db error")).Once().
			On("CreateTask", mock.AnythingOfType("*models.Task")).Return(nil).Once()


		createdTasks, err := taskService.ExtractAndCreateTasks(context.Background(), inputText, userID)
		assert.NoError(t, err) // No error because it logs and continues
		assert.Len(t, createdTasks, 1) // Only one task successfully created
		assert.Equal(t, "Task 2", createdTasks[0].Title) // The second one was successful
		mockLLMExtractor.AssertExpectations(t)
		mockTaskRepo.AssertExpectations(t)
	})
}

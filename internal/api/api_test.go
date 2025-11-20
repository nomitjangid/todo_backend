package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"todo-backend/internal/config"
	"todo-backend/internal/llm"
	"todo-backend/internal/models"
	"todo-backend/internal/repositories"
	"todo-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

// setupTestEnvironment sets up an in-memory SQLite database and all services/repositories for testing
func setupTestEnvironment() (*gin.Engine, *gorm.DB, error) {
	// 1. Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	// Migrate schema
	db.AutoMigrate(&models.User{}, &models.Task{})

	// 2. Load test config (or mock it)
	cfg := &config.Config{
		JWTSecret:  "test-secret",
		OpenAPIKey: "test-openai-key", // dummy key for extractor init
	}
	_ = cfg // cfg is not directly used after service init but might be used by LLM extractor

	// 3. Initialize Repositories
	userRepo := repositories.NewUserRepository(db)
	taskRepo := repositories.NewTaskRepository(db)

	// 4. Initialize LLM Service (mock if needed, for integration test, we might use a dummy or real)
	// For API integration tests, we can use a mock LLM Extractor
	mockLLMExtractor := &MockLLMExtractor{}
	mockLLMExtractor.On("ExtractTasks", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).Return([]llm.Task{
		{
			Title:       "Buy groceries",
			Description: "Buy milk and eggs",
			DueDate:     time.Now().Add(24 * time.Hour),
			Priority:    "medium",
			Subtasks:    []string{},
		},
	}, nil)

	// 5. Initialize Services
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	taskService := services.NewTaskService(taskRepo, mockLLMExtractor)

	// 6. Inject services into API handlers
	SetAuthService(authService)
	SetUserService(userService)
	SetTaskService(taskService)

	// 7. Setup router
	router := SetupRouter()
	return router, db, nil
}

func TestAuthEndpoints(t *testing.T) {
	router, db, err := setupTestEnvironment()
	assert.NoError(t, err)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Register Test
	t.Run("POST /auth/register should register a new user", func(t *testing.T) {
		w := httptest.NewRecorder()
		reqBody := bytes.NewBufferString(`{"email": "test@example.com", "password": "password123"}`)
		req, _ := http.NewRequest("POST", "/auth/register", reqBody)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "id")
		assert.Equal(t, "test@example.com", response["email"])
	})

	t.Run("POST /auth/register should return 400 for invalid input", func(t *testing.T) {
		w := httptest.NewRecorder()
		reqBody := bytes.NewBufferString(`{"email": "invalid", "password": ""}`)
		req, _ := http.NewRequest("POST", "/auth/register", reqBody)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST /auth/register should return 500 if user already exists", func(t *testing.T) {
		w := httptest.NewRecorder()
		reqBody := bytes.NewBufferString(`{"email": "test@example.com", "password": "password123"}`)
		req, _ := http.NewRequest("POST", "/auth/register", reqBody)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code) // Service returns 500 for "user already exists"
	})

	// Login Test
	var authToken string
	t.Run("POST /auth/login should log in an existing user", func(t *testing.T) {
		w := httptest.NewRecorder()
		reqBody := bytes.NewBufferString(`{"email": "test@example.com", "password": "password123"}`)
		req, _ := http.NewRequest("POST", "/auth/login", reqBody)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "token")
		authToken = response["token"]
	})

	t.Run("POST /auth/login should return 401 for invalid credentials", func(t *testing.T) {
		w := httptest.NewRecorder()
		reqBody := bytes.NewBufferString(`{"email": "test@example.com", "password": "wrongpassword"}`)
		req, _ := http.NewRequest("POST", "/auth/login", reqBody)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Me Test
	t.Run("GET /auth/me should return authenticated user's info", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "id")
		assert.Equal(t, "test@example.com", response["email"])
	})

	t.Run("GET /auth/me should return 401 for unauthenticated request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth/me", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestTaskEndpoints(t *testing.T) {
	router, db, err := setupTestEnvironment()
	assert.NoError(t, err)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Register a user and get a token
	w := httptest.NewRecorder()
	reqBody := bytes.NewBufferString(`{"email": "taskuser@example.com", "password": "password123"}`)
	req, _ := http.NewRequest("POST", "/auth/register", reqBody)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	reqBody = bytes.NewBufferString(`{"email": "taskuser@example.com", "password": "password123"}`)
	req, _ = http.NewRequest("POST", "/auth/login", reqBody)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	authToken := loginResponse["token"]

	// Get authenticated user's ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+authToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var meResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &meResponse)
	userIDStr := meResponse["id"].(string)
	userID, _ := uuid.Parse(userIDStr)

	t.Run("POST /tasks should create a new task", func(t *testing.T) {
		w := httptest.NewRecorder()
		taskReqBody := bytes.NewBufferString(`{"title": "New Task", "description": "Details", "priority": "medium"}`)
		req, _ := http.NewRequest("POST", "/tasks/", taskReqBody)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var taskResponse models.Task
		json.Unmarshal(w.Body.Bytes(), &taskResponse)
		assert.NotEmpty(t, taskResponse.ID)
		assert.Equal(t, "New Task", taskResponse.Title)
		assert.Equal(t, userID, taskResponse.UserID)
	})

	t.Run("GET /tasks should return all tasks for the user", func(t *testing.T) {
		// Create a task first
		taskToCreate := models.Task{
			ID:          uuid.New(),
			UserID:      userID,
			Title:       "Another Task",
			Description: "More details",
			Priority:    "low",
			CreatedAt:   time.Now(),
		}
		db.Create(&taskToCreate)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tasks/", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var tasksResponse []models.Task
		json.Unmarshal(w.Body.Bytes(), &tasksResponse)
		assert.GreaterOrEqual(t, len(tasksResponse), 2) // At least the one created above and the current one

		found := false
		for _, task := range tasksResponse {
			if task.ID == taskToCreate.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected task not found in response")
	})

	t.Run("GET /tasks/:id should return a single task", func(t *testing.T) {
		taskToCreate := models.Task{
			ID:      uuid.New(),
			UserID:  userID,
			Title:   "Single Task",
			Description: "Details",
			Priority: "high",
			CreatedAt: time.Now(),
		}
		db.Create(&taskToCreate)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tasks/"+taskToCreate.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var taskResponse models.Task
		json.Unmarshal(w.Body.Bytes(), &taskResponse)
		assert.Equal(t, taskToCreate.ID, taskResponse.ID)
		assert.Equal(t, taskToCreate.Title, taskResponse.Title)
	})

	t.Run("PUT /tasks/:id should update an existing task", func(t *testing.T) {
		taskToUpdate := models.Task{
			ID:      uuid.New(),
			UserID:  userID,
			Title:   "Task to Update",
			Description: "Old Desc",
			Priority: "low",
			CreatedAt: time.Now(),
		}
		db.Create(&taskToUpdate)

		w := httptest.NewRecorder()
		updateReqBody := bytes.NewBufferString(`{"title": "Updated Title", "description": "New Description", "priority": "high"}`)
		req, _ := http.NewRequest("PUT", "/tasks/"+taskToUpdate.ID.String(), updateReqBody)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var taskResponse models.Task
		json.Unmarshal(w.Body.Bytes(), &taskResponse)
		assert.Equal(t, "Updated Title", taskResponse.Title)
		assert.Equal(t, "New Description", taskResponse.Description)
		assert.Equal(t, "high", taskResponse.Priority)
	})

	t.Run("DELETE /tasks/:id should delete a task", func(t *testing.T) {
		taskToDelete := models.Task{
			ID:      uuid.New(),
			UserID:  userID,
			Title:   "Task to Delete",
			CreatedAt: time.Now(),
		}
		db.Create(&taskToDelete)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/tasks/"+taskToDelete.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify task is deleted
		var deletedTask models.Task
		err := db.First(&deletedTask, "id = ?", taskToDelete.ID).Error
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("POST /tasks/from-text should extract and create tasks from text", func(t *testing.T) {
		w := httptest.NewRecorder()
		textReqBody := bytes.NewBufferString(`{"text": "Buy groceries and call mom tomorrow"}`)
		req, _ := http.NewRequest("POST", "/tasks/from-text", textReqBody)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var tasksResponse []models.Task
		json.Unmarshal(w.Body.Bytes(), &tasksResponse)
		assert.NotEmpty(t, tasksResponse)
		assert.Equal(t, "Buy groceries", tasksResponse[0].Title)
	})
}

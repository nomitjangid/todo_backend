package api

import (
	"fmt"
	"net/http"
	"time"
	"todo-backend/internal/models"
	"todo-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var taskService *services.TaskService // Will be initialized in main

// SetTaskService initializes the taskService
func SetTaskService(service *services.TaskService) {
	taskService = service
}

// CreateTaskRequest defines the request body for creating a task
type CreateTaskRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	DueDate     *string   `json:"due_date"` // Use *string to allow null for omitempty
	Priority    string    `json:"priority"`
	RawText     string    `json:"raw_text"`
}

// UpdateTaskRequest defines the request body for updating a task
type UpdateTaskRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     *string   `json:"due_date"`
	Priority    string    `json:"priority"`
	RawText     string    `json:"raw_text"`
}

// ExtractTasksFromTextRequest defines the request body for extracting tasks from text
type ExtractTasksFromTextRequest struct {
	Text string `json:"text" binding:"required"`
}

// GetTasks handles fetching all tasks for the authenticated user
func GetTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
		return
	}
	userIDUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format in context"})
		return
	}
	tasks, err := taskService.GetTasksByUserID(userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// CreateTask handles creating a new task
func CreateTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
		return
	}
	userIDUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format in context"})
		return
	}

	task := &models.Task{
		ID:          uuid.New(),
		UserID:      userIDUUID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		RawText:     req.RawText,
	}

	if req.DueDate != nil && *req.DueDate != "" {
		parsedTime, err := parseDueDate(*req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format"})
			return
		}
		task.DueDate = &parsedTime
	}

	if err := taskService.CreateTask(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTaskByID handles fetching a single task by ID
func GetTaskByID(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
		return
	}
	userIDUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format in context"})
		return
	}
	task, err := taskService.GetTaskByID(taskID, userIDUUID)
	if err != nil {
		if err.Error() == "task not found or unauthorized" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask handles updating an existing task
func UpdateTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
		return
	}
	userIDUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format in context"})
		return
	}

	task := &models.Task{
		ID: taskID,
		UserID: userIDUUID, // Important for authorization in service layer
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		RawText:     req.RawText,
	}

	if req.DueDate != nil && *req.DueDate != "" {
		parsedTime, err := parseDueDate(*req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format"})
			return
		}
		task.DueDate = &parsedTime
	}

	if err := taskService.UpdateTask(task, userIDUUID); err != nil {
		if err.Error() == "task not found or unauthorized" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask handles deleting a task
func DeleteTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
		return
	}
	userIDUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format in context"})
		return
	}

	if err := taskService.DeleteTask(taskID, userIDUUID); err != nil {
		if err.Error() == "task not found or unauthorized" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ExtractTasksFromText handles extracting tasks from provided text using LLM
func ExtractTasksFromText(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req ExtractTasksFromTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
		return
	}
	userIDUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format in context"})
		return
	}

	tasks, err := taskService.ExtractAndCreateTasks(c.Request.Context(), req.Text, userIDUUID)
	if err != nil {
		// Differentiate between LLM extraction error and database creation error if needed
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tasks)
}

// Helper to parse date strings from requests
func parseDueDate(dateStr string) (time.Time, error) {
	// Attempt to parse ISO 8601
	parsedTime, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return parsedTime, nil
	}

	// Add more date formats here if necessary for user input flexibility
	// For now, strict ISO 8601 is expected for API input
	return time.Time{}, fmt.Errorf("unsupported date format: %s", dateStr)
}

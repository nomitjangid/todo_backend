package main

import (
	"fmt"
	"log"
	"todo-backend/internal/api"
	"todo-backend/internal/config"
	"todo-backend/internal/database"
	"todo-backend/internal/llm"
	"todo-backend/internal/middleware"
	"todo-backend/internal/repositories"
	"todo-backend/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	middleware.InitLogger()

	// Set up database connection
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Repositories
	userRepo := repositories.NewUserRepository(db)
	taskRepo := repositories.NewTaskRepository(db)

	// Set up LLM service
	llmService := llm.NewOpenAIExtractor(cfg)
	
	// Set up Task service
	taskService := services.NewTaskService(taskRepo, llmService)
	api.SetTaskService(taskService)

	// Initialize Auth Service
	authService := services.NewAuthService(userRepo)
	api.SetAuthService(authService)

	// Initialize User Service
	userService := services.NewUserService(userRepo)
	api.SetUserService(userService)

	// Set up router
	router := api.SetupRouter()
	router.Use(middleware.RecoveryMiddleware()) // Use the recovery middleware
	router.Use(middleware.LoggerMiddleware())   // Use the logger middleware
	
	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	router.Run(fmt.Sprintf(":%s", cfg.Port))
}

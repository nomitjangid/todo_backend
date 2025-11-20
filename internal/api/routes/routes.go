package routes

import (
	"todo-backend/internal/api/handlers"
	"todo-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, taskHandler *handlers.TaskHandler, jwtSecret string) {
	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.RecoveryMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(jwtSecret), authHandler.Me)
	}

	// Task routes (protected)
	tasks := router.Group("/tasks")
	tasks.Use(middleware.AuthMiddleware(jwtSecret))
	{
		tasks.POST("", taskHandler.CreateTask)
		tasks.POST("/from-text", taskHandler.ExtractAndCreateTasks)
		tasks.GET("", taskHandler.GetTasks)
		tasks.GET("/:id", taskHandler.GetTask)
		tasks.PUT("/:id", taskHandler.UpdateTask)
		tasks.DELETE("/:id", taskHandler.DeleteTask)
	}
}

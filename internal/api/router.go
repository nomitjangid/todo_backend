package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the Gin router and defines the API routes
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins for development
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/register", Register)
		auth.POST("/login", Login)
		auth.GET("/me", AuthMiddleware(), Me)
	}

	tasks := r.Group("/tasks")
	tasks.Use(AuthMiddleware()) // Apply JWT middleware to all task routes
	{
		tasks.GET("/", GetTasks)
		tasks.POST("/", CreateTask)
		tasks.GET("/:id", GetTaskByID)
		tasks.PUT("/:id", UpdateTask)
		tasks.DELETE("/:id", DeleteTask)
		tasks.POST("/from-text", ExtractTasksFromText)
	}

	return r
}

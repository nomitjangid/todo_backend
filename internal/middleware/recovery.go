package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RecoveryMiddleware returns a Gin middleware that recovers from panics
// and logs the stack trace.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				log.Error().
					Interface("panic", r).
					Str("stack", string(debug.Stack())).
					Msg("Recovered from panic")

				// Return a 500 Internal Server Error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Internal Server Error: %v", r),
				})
			}
		}()
		c.Next()
	}
}

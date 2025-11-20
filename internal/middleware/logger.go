package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes zerolog for structured logging.
func InitLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// You might want to set different levels for production.
	// zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

// LoggerMiddleware returns a Gin middleware that logs requests using zerolog.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		param := map[string]string{}
		if raw != "" {
			path = path + "?" + raw
		}

		log.Info().
			Int("status", c.Writer.Status()).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("ip", c.ClientIP()).
			Dur("latency", time.Since(start)).
			Str("user_agent", c.Request.UserAgent()).
			RawJSON("param", []byte(fmt.Sprintf("%v", param))). // Or parse c.Params for cleaner output
			Msg("Request")
	}
}

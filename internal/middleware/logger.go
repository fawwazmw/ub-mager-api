package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if path == "/health" || path == "/ready" {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		event := log.Info()
		if status >= 400 {
			event = log.Warn()
		}
		if status >= 500 {
			event = log.Error()
		}

		reqID, _ := c.Get("request_id")
		reqIDStr, _ := reqID.(string)

		logger := event.
			Str("request_id", reqIDStr).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", c.Request.URL.RawQuery).
			Int("status", status).
			Dur("latency", latency).
			Int("bytes", c.Writer.Size()).
			Str("ip", c.ClientIP())

		if userID, exists := c.Get("user_id"); exists {
			logger = logger.Str("user_id", fmt.Sprintf("%v", userID))
		}

		logger.Msg("request")
	}
}

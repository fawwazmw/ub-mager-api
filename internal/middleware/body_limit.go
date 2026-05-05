package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wardayadev/ub-mager-api/internal/handler"
)

func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()

		if c.Errors.Last() != nil && c.Errors.Last().Err.Error() == "http: request body too large" {
			handler.Error(c, http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE", "Request body exceeds maximum allowed size")
		}
	}
}

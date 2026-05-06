package middleware

import (
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	htmlTagPattern  = regexp.MustCompile(`<[^>]*>`)
	scriptPattern   = regexp.MustCompile(`(?i)(javascript|on\w+\s*=|<script|<\/script)`)
	sqlInjPattern   = regexp.MustCompile(`(?i)(;\s*(DROP|DELETE|UPDATE|INSERT|ALTER|EXEC)\s)`)
	nullBytePattern = regexp.MustCompile(`\x00`)
)

func Sanitize() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

		sanitized := sanitizeBytes(body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(sanitized))
		c.Request.ContentLength = int64(len(sanitized))

		c.Next()
	}
}

func sanitizeBytes(data []byte) []byte {
	s := string(data)
	s = nullBytePattern.ReplaceAllString(s, "")
	s = htmlTagPattern.ReplaceAllString(s, "")
	s = scriptPattern.ReplaceAllString(s, "")
	s = sqlInjPattern.ReplaceAllString(s, "")
	return []byte(s)
}

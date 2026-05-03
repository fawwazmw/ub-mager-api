package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	allowedOrigins := os.Getenv("CORS_ORIGINS")
	origins := []string{"*"}
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := false

		if origins[0] == "*" {
			allowed = true
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			for _, o := range origins {
				if strings.TrimSpace(o) == origin {
					allowed = true
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		if !allowed {
			c.Header("Access-Control-Allow-Origin", origins[0])
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

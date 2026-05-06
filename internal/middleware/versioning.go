package middleware

import (
	"github.com/gin-gonic/gin"
)

func APIVersion(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-Version", version)
		c.Header("X-Supported-Versions", "v1")
		c.Next()
	}
}

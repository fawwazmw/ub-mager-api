package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wardayadev/ub-mager-api/internal/handler"
	jwtpkg "github.com/wardayadev/ub-mager-api/internal/pkg/jwt"
)

func AuthRequired(jwtService *jwtpkg.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			handler.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
			return
		}

		var tokenString string
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			tokenString = parts[1]
		} else {
			tokenString = authHeader
		}

		if tokenString == "" {
			handler.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization format")
			return
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			handler.Error(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "Token is invalid or expired")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			handler.Error(c, http.StatusForbidden, "FORBIDDEN", "Access denied")
			return
		}

		role, ok := userRole.(string)
		if !ok {
			handler.Error(c, http.StatusForbidden, "FORBIDDEN", "Access denied")
			return
		}

		for _, allowed := range roles {
			if role == allowed {
				c.Next()
				return
			}
		}

		handler.Error(c, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
	}
}

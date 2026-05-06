package middleware

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

func AuditLog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() >= 400 {
			return
		}

		method := c.Request.Method
		if method == "GET" {
			return
		}

		userIDVal, exists := c.Get("user_id")
		if !exists {
			return
		}
		adminID, ok := userIDVal.(uuid.UUID)
		if !ok {
			return
		}

		action := method + " " + c.FullPath()
		targetID := c.Param("id")

		details, _ := json.Marshal(map[string]string{
			"path":   c.Request.URL.Path,
			"method": method,
		})

		log := &model.AuditLog{
			ID:       uuid.New(),
			AdminID:  adminID,
			Action:   action,
			TargetID: targetID,
			Details:  string(details),
			IP:       c.ClientIP(),
		}

		db.Create(log)
	}
}

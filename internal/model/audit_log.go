package model

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	AdminID   uuid.UUID `gorm:"type:uuid;not null;index" json:"admin_id"`
	Action    string    `gorm:"type:varchar(50);not null;index" json:"action"`
	TargetID  string    `gorm:"type:varchar(36)" json:"target_id,omitempty"`
	Details   string    `gorm:"type:jsonb" json:"details,omitempty"`
	IP        string    `gorm:"type:varchar(45)" json:"ip"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

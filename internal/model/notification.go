package model

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_notif_user_read" json:"user_id"`
	Type      string    `gorm:"type:varchar(30);not null" json:"type"`
	Title     string    `gorm:"type:varchar(200);not null" json:"title"`
	Body      string    `gorm:"type:text" json:"body"`
	Data      string    `gorm:"type:jsonb" json:"data,omitempty"`
	IsRead    bool      `gorm:"not null;default:false;index:idx_notif_user_read" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

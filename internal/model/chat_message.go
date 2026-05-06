package model

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RideID    uuid.UUID `gorm:"type:uuid;not null;index:idx_chat_ride_created" json:"ride_id"`
	SenderID  uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"index:idx_chat_ride_created" json:"created_at"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

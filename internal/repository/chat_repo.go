package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Create(ctx context.Context, msg *model.ChatMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

type ChatMessageItem struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	SenderRole string    `json:"sender_role"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

func (r *ChatRepository) FindByRideID(ctx context.Context, rideID uuid.UUID, limit int) ([]ChatMessageItem, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}

	var results []ChatMessageItem
	err := r.db.WithContext(ctx).
		Table("chat_messages").
		Select("chat_messages.id, chat_messages.sender_id, users.full_name as sender_name, users.role as sender_role, chat_messages.content, chat_messages.created_at").
		Joins("JOIN users ON users.id = chat_messages.sender_id").
		Where("chat_messages.ride_id = ?", rideID).
		Order("chat_messages.created_at ASC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func (r *ChatRepository) FindByTaskID(ctx context.Context, taskID uuid.UUID, limit int) ([]ChatMessageItem, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}

	var results []ChatMessageItem
	err := r.db.WithContext(ctx).
		Table("chat_messages").
		Select("chat_messages.id, chat_messages.sender_id, users.full_name as sender_name, users.role as sender_role, chat_messages.content, chat_messages.created_at").
		Joins("JOIN users ON users.id = chat_messages.sender_id").
		Where("chat_messages.task_id = ?", taskID).
		Order("chat_messages.created_at ASC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func (r *ChatRepository) CountByRideID(ctx context.Context, rideID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where("ride_id = ?", rideID).
		Count(&count).Error
	return count, err
}

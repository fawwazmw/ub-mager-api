package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notif *model.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

func (r *NotificationRepository) FindByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]model.Notification, int64, error) {
	var total int64
	var results []model.Notification

	query := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&results).Error
	return results, total, err
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, notifID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

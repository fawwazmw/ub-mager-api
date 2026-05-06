package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) DB() *gorm.DB {
	return r.db
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *TaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, taskID uuid.UUID, status model.TaskStatus, updates map[string]any) error {
	updates["status"] = status
	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", taskID).
		Updates(updates).Error
}

type TaskFeedItem struct {
	ID              string  `json:"id"`
	CreatorID       string  `json:"creator_id"`
	CreatorName     string  `json:"creator_name"`
	Category        string  `json:"category"`
	Status          string  `json:"status"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Fee             float64 `json:"fee"`
	PickupAddress   string  `json:"pickup_address"`
	PickupZone      string  `json:"pickup_zone"`
	DeliveryAddress string  `json:"delivery_address"`
	DeliveryZone    string  `json:"delivery_zone"`
	CreatedAt       string  `json:"created_at"`
}

func (r *TaskRepository) Feed(ctx context.Context, page, perPage int, category, zone string) ([]TaskFeedItem, int64, error) {
	var total int64
	var results []TaskFeedItem

	query := r.db.WithContext(ctx).
		Table("tasks").
		Select("tasks.id, tasks.creator_id, users.full_name as creator_name, tasks.category, tasks.status, tasks.title, tasks.description, tasks.fee, tasks.pickup_address, tasks.pickup_zone, tasks.delivery_address, tasks.delivery_zone, tasks.created_at").
		Joins("JOIN users ON users.id = tasks.creator_id").
		Where("tasks.deleted_at IS NULL AND tasks.status = ?", model.TaskStatusOpen)

	if category != "" {
		query = query.Where("tasks.category = ?", category)
	}
	if zone != "" {
		query = query.Where("tasks.pickup_zone = ? OR tasks.delivery_zone = ?", zone, zone)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("tasks.created_at DESC").Offset(offset).Limit(perPage).Scan(&results).Error

	return results, total, err
}

func (r *TaskRepository) FindByCreator(ctx context.Context, creatorID uuid.UUID, page, perPage int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Task{}).Where("creator_id = ?", creatorID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&tasks).Error
	return tasks, total, err
}

func (r *TaskRepository) FindByHelper(ctx context.Context, helperID uuid.UUID, page, perPage int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Task{}).Where("helper_id = ?", helperID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&tasks).Error
	return tasks, total, err
}

type AdminTaskItem struct {
	ID          string     `json:"id"`
	CreatorName string     `json:"creator_name"`
	HelperName  *string    `json:"helper_name"`
	Category    string     `json:"category"`
	Status      string     `json:"status"`
	Title       string     `json:"title"`
	Fee         float64    `json:"fee"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (r *TaskRepository) AdminList(ctx context.Context, page, perPage int, category, status, search string) ([]AdminTaskItem, int64, error) {
	var total int64
	var results []AdminTaskItem

	query := r.db.WithContext(ctx).
		Table("tasks").
		Select(`tasks.id, creator.full_name as creator_name, helper.full_name as helper_name,
			tasks.category, tasks.status, tasks.title, tasks.fee, tasks.created_at, tasks.completed_at`).
		Joins("JOIN users AS creator ON creator.id = tasks.creator_id").
		Joins("LEFT JOIN users AS helper ON helper.id = tasks.helper_id").
		Where("tasks.deleted_at IS NULL")

	if category != "" {
		query = query.Where("tasks.category = ?", category)
	}
	if status != "" {
		query = query.Where("tasks.status = ?", status)
	}
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("(tasks.title ILIKE ? OR creator.full_name ILIKE ?)", pattern, pattern)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("tasks.created_at DESC").Offset(offset).Limit(perPage).Scan(&results).Error

	return results, total, err
}

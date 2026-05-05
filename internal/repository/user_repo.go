package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
}

func (r *UserRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("is_active", active).Error
}

type UserListItem struct {
	ID                string `json:"id"`
	FullName          string `json:"full_name"`
	Phone             string `json:"phone"`
	Email             string `json:"email"`
	Role              string `json:"role"`
	IsActive          bool   `json:"is_active"`
	IsStudentVerified bool   `json:"is_student_verified"`
	CreatedAt         string `json:"created_at"`
}

func (r *UserRepository) List(ctx context.Context, page, perPage int, role, search string) ([]UserListItem, int64, error) {
	var total int64
	var results []UserListItem

	query := r.db.WithContext(ctx).
		Table("users").
		Select("id, full_name, phone, email, role, is_active, is_student_verified, created_at").
		Where("deleted_at IS NULL")

	if role != "" {
		query = query.Where("role = ?", role)
	}

	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("(full_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?)", pattern, pattern, pattern)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Scan(&results).Error

	return results, total, err
}

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RolePassenger UserRole = "PASSENGER"
	RoleDriver    UserRole = "DRIVER"
	RoleAdmin     UserRole = "ADMIN"
)

type User struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	FullName          string         `gorm:"type:text;not null" json:"full_name"`
	Phone             string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`
	Email             string         `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	PasswordHash      string         `gorm:"type:text;not null" json:"-"`
	Role              UserRole       `gorm:"type:varchar(20);not null;default:'PASSENGER'" json:"role"`
	AvatarURL         *string        `gorm:"type:text" json:"avatar_url"`
	IsActive          bool           `gorm:"not null;default:true" json:"is_active"`
	IsStudentVerified bool           `gorm:"not null;default:false" json:"is_student_verified"`
	LastLoginAt       *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

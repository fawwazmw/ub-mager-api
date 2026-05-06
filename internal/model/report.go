package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "PENDING"
	ReportStatusReviewed ReportStatus = "REVIEWED"
	ReportStatusResolved ReportStatus = "RESOLVED"
)

type ReportCategory string

const (
	ReportCategoryRude   ReportCategory = "RUDE_BEHAVIOR"
	ReportCategorySafety ReportCategory = "SAFETY_CONCERN"
	ReportCategoryFraud  ReportCategory = "FRAUD"
	ReportCategorySpam   ReportCategory = "SPAM"
	ReportCategoryOther  ReportCategory = "OTHER"
)

type Report struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ReporterID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"reporter_id"`
	ReportedUserID uuid.UUID      `gorm:"type:uuid;not null;index" json:"reported_user_id"`
	RideID         *uuid.UUID     `gorm:"type:uuid;index" json:"ride_id,omitempty"`
	Category       ReportCategory `gorm:"type:varchar(30);not null" json:"category"`
	Description    string         `gorm:"type:text;not null" json:"description"`
	Status         ReportStatus   `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`
	AdminNote      string         `gorm:"type:text" json:"admin_note,omitempty"`
	ResolvedAt     *time.Time     `json:"resolved_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Report) TableName() string {
	return "reports"
}

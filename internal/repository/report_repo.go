package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(ctx context.Context, report *model.Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

type ReportListItem struct {
	ID             string    `json:"id"`
	ReporterName   string    `json:"reporter_name"`
	ReportedName   string    `json:"reported_name"`
	Category       string    `json:"category"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	RideID         *string   `json:"ride_id"`
	CreatedAt      time.Time `json:"created_at"`
}

func (r *ReportRepository) List(ctx context.Context, page, perPage int, status string) ([]ReportListItem, int64, error) {
	var total int64
	var results []ReportListItem

	query := r.db.WithContext(ctx).
		Table("reports").
		Select(`
			reports.id,
			reporter.full_name as reporter_name,
			reported.full_name as reported_name,
			reports.category,
			reports.description,
			reports.status,
			reports.ride_id::text as ride_id,
			reports.created_at
		`).
		Joins("JOIN users AS reporter ON reporter.id = reports.reporter_id").
		Joins("JOIN users AS reported ON reported.id = reports.reported_user_id").
		Where("reports.deleted_at IS NULL")

	if status != "" {
		query = query.Where("reports.status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("reports.created_at DESC").Offset(offset).Limit(perPage).Scan(&results).Error

	return results, total, err
}

func (r *ReportRepository) Resolve(ctx context.Context, reportID uuid.UUID, adminNote string, status model.ReportStatus) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Report{}).
		Where("id = ?", reportID).
		Updates(map[string]any{
			"status":      status,
			"admin_note":  adminNote,
			"resolved_at": now,
		}).Error
}

func (r *ReportRepository) CountPending(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Report{}).
		Where("status = ?", model.ReportStatusPending).
		Count(&count).Error
	return count, err
}

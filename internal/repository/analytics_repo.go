package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var ErrRideNotFoundOrTerminal = errors.New("ride not found or already completed/cancelled")

type AnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func startOfDay() time.Time {
	return time.Now().Truncate(24 * time.Hour)
}

func parsePeriod(period string) (time.Time, string) {
	now := time.Now()
	switch period {
	case "today":
		return startOfDay(), "today"
	case "week":
		return now.AddDate(0, 0, -7), "week"
	case "month":
		return now.AddDate(0, -1, 0), "month"
	default:
		return startOfDay(), "today"
	}
}

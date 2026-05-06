package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

type EarningsRepository struct {
	db *gorm.DB
}

func NewEarningsRepository(db *gorm.DB) *EarningsRepository {
	return &EarningsRepository{db: db}
}

type EarningsSummary struct {
	TotalEarnings float64 `json:"total_earnings"`
	TotalRides    int64   `json:"total_rides"`
	TotalTasks    int64   `json:"total_tasks"`
	AvgPerRide    float64 `json:"avg_per_ride"`
	Period        string  `json:"period"`
}

type DailyEarning struct {
	Date    string  `json:"date"`
	Rides   int64   `json:"rides"`
	Tasks   int64   `json:"tasks"`
	Earning float64 `json:"earning"`
}

func (r *EarningsRepository) GetSummary(ctx context.Context, driverProfileID uuid.UUID, period string) (*EarningsSummary, error) {
	var since time.Time
	now := time.Now()

	switch period {
	case "today":
		since = now.Truncate(24 * time.Hour)
	case "week":
		since = now.AddDate(0, 0, -7)
	case "month":
		since = now.AddDate(0, -1, 0)
	default:
		since = now.Truncate(24 * time.Hour)
		period = "today"
	}

	summary := &EarningsSummary{Period: period}

	var rideEarnings struct {
		Total *float64
		Count int64
	}
	r.db.WithContext(ctx).
		Model(&model.Ride{}).
		Select("COALESCE(SUM(total_fare), 0) as total, COUNT(*) as count").
		Where("driver_id = ? AND status = ? AND completed_at >= ?", driverProfileID, model.RideStatusCompleted, since).
		Scan(&rideEarnings)

	if rideEarnings.Total != nil {
		summary.TotalEarnings = *rideEarnings.Total
	}
	summary.TotalRides = rideEarnings.Count

	var taskEarnings struct {
		Total *float64
		Count int64
	}
	r.db.WithContext(ctx).
		Model(&model.Task{}).
		Select("COALESCE(SUM(fee), 0) as total, COUNT(*) as count").
		Where("helper_id = ? AND status = ? AND completed_at >= ?", driverProfileID, model.TaskStatusCompleted, since).
		Scan(&taskEarnings)

	if taskEarnings.Total != nil {
		summary.TotalEarnings += *taskEarnings.Total
	}
	summary.TotalTasks = taskEarnings.Count

	totalJobs := summary.TotalRides + summary.TotalTasks
	if totalJobs > 0 {
		summary.AvgPerRide = summary.TotalEarnings / float64(totalJobs)
	}

	return summary, nil
}

func (r *EarningsRepository) GetDaily(ctx context.Context, driverProfileID uuid.UUID, days int) ([]DailyEarning, error) {
	if days < 1 || days > 90 {
		days = 7
	}

	since := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -(days - 1))

	var rideResults []struct {
		Date    string
		Earning float64
		Count   int64
	}
	r.db.WithContext(ctx).
		Table("rides").
		Select("TO_CHAR(completed_at, 'YYYY-MM-DD') as date, COALESCE(SUM(total_fare), 0) as earning, COUNT(*) as count").
		Where("driver_id = ? AND status = ? AND completed_at >= ?", driverProfileID, model.RideStatusCompleted, since).
		Group("date").
		Scan(&rideResults)

	var taskResults []struct {
		Date    string
		Earning float64
		Count   int64
	}
	r.db.WithContext(ctx).
		Table("tasks").
		Select("TO_CHAR(completed_at, 'YYYY-MM-DD') as date, COALESCE(SUM(fee), 0) as earning, COUNT(*) as count").
		Where("helper_id = ? AND status = ? AND completed_at >= ?", driverProfileID, model.TaskStatusCompleted, since).
		Group("date").
		Scan(&taskResults)

	rideMap := make(map[string]struct {
		earning float64
		count   int64
	})
	for _, r := range rideResults {
		rideMap[r.Date] = struct {
			earning float64
			count   int64
		}{r.Earning, r.Count}
	}

	taskMap := make(map[string]struct {
		earning float64
		count   int64
	})
	for _, t := range taskResults {
		taskMap[t.Date] = struct {
			earning float64
			count   int64
		}{t.Earning, t.Count}
	}

	var results []DailyEarning
	for i := 0; i < days; i++ {
		date := since.AddDate(0, 0, i).Format("2006-01-02")
		ride := rideMap[date]
		task := taskMap[date]
		results = append(results, DailyEarning{
			Date:    date,
			Rides:   ride.count,
			Tasks:   task.count,
			Earning: ride.earning + task.earning,
		})
	}

	return results, nil
}

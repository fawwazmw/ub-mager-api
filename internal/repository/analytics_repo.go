package repository

import (
	"context"
	"time"

	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

type AnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

type DashboardStats struct {
	TotalUsers       int64   `json:"total_users"`
	TotalDrivers     int64   `json:"total_drivers"`
	OnlineDrivers    int64   `json:"online_drivers"`
	TotalRides       int64   `json:"total_rides"`
	ActiveRides      int64   `json:"active_rides"`
	CompletedToday   int64   `json:"completed_today"`
	RevenueToday     float64 `json:"revenue_today"`
	CancelledToday   int64   `json:"cancelled_today"`
}

func (r *AnalyticsRepository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var stats DashboardStats
	today := time.Now().Truncate(24 * time.Hour)

	r.db.WithContext(ctx).Model(&model.User{}).Count(&stats.TotalUsers)
	r.db.WithContext(ctx).Model(&model.DriverProfile{}).Count(&stats.TotalDrivers)
	r.db.WithContext(ctx).Model(&model.DriverProfile{}).Where("is_online = ?", true).Count(&stats.OnlineDrivers)
	r.db.WithContext(ctx).Model(&model.Ride{}).Count(&stats.TotalRides)
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status NOT IN ?", []model.RideStatus{model.RideStatusCompleted, model.RideStatusCancelled}).
		Count(&stats.ActiveRides)
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, today).
		Count(&stats.CompletedToday)
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status = ? AND cancelled_at >= ?", model.RideStatusCancelled, today).
		Count(&stats.CancelledToday)

	// Revenue today
	var revenue struct{ Total *float64 }
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Select("COALESCE(SUM(total_fare), 0) as total").
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, today).
		Scan(&revenue)
	if revenue.Total != nil {
		stats.RevenueToday = *revenue.Total
	}

	return &stats, nil
}

type RevenueStats struct {
	Period       string  `json:"period"`
	TotalRevenue float64 `json:"total_revenue"`
	TotalRides   int64   `json:"total_rides"`
	AvgFare      float64 `json:"avg_fare"`
	Currency     string  `json:"currency"`
}

func (r *AnalyticsRepository) GetRevenueStats(ctx context.Context, period string) (*RevenueStats, error) {
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

	stats := &RevenueStats{Period: period, Currency: "IDR"}

	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, since).
		Count(&stats.TotalRides)

	var agg struct {
		Total *float64
		Avg   *float64
	}
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Select("COALESCE(SUM(total_fare), 0) as total, COALESCE(AVG(total_fare), 0) as avg").
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, since).
		Scan(&agg)

	if agg.Total != nil {
		stats.TotalRevenue = *agg.Total
	}
	if agg.Avg != nil {
		stats.AvgFare = *agg.Avg
	}

	return stats, nil
}

type RideStats struct {
	Period         string `json:"period"`
	Total          int64  `json:"total"`
	Completed      int64  `json:"completed"`
	Cancelled      int64  `json:"cancelled"`
	CompletionRate float64 `json:"completion_rate"`
}

func (r *AnalyticsRepository) GetRideStats(ctx context.Context, period string) (*RideStats, error) {
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

	stats := &RideStats{Period: period}

	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("created_at >= ?", since).
		Count(&stats.Total)

	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, since).
		Count(&stats.Completed)

	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status = ? AND cancelled_at >= ?", model.RideStatusCancelled, since).
		Count(&stats.Cancelled)

	if stats.Total > 0 {
		stats.CompletionRate = float64(stats.Completed) / float64(stats.Total) * 100
	}

	return stats, nil
}

type DriverListItem struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	FullName     string  `json:"full_name"`
	Phone        string  `json:"phone"`
	VehicleType  string  `json:"vehicle_type"`
	LicensePlate string  `json:"license_plate"`
	IsOnline     bool    `json:"is_online"`
	IsVerified   bool    `json:"is_verified"`
	Rating       float64 `json:"rating"`
	TotalTrips   int     `json:"total_trips"`
	CreatedAt    time.Time `json:"created_at"`
}

func (r *AnalyticsRepository) ListDrivers(ctx context.Context, page, perPage int, status string) ([]DriverListItem, int64, error) {
	var total int64
	var results []DriverListItem

	query := r.db.WithContext(ctx).
		Table("driver_profiles").
		Select("driver_profiles.id, driver_profiles.user_id, users.full_name, users.phone, driver_profiles.vehicle_type, driver_profiles.license_plate, driver_profiles.is_online, driver_profiles.is_verified, driver_profiles.rating_avg as rating, driver_profiles.total_trips, driver_profiles.created_at").
		Joins("JOIN users ON users.id = driver_profiles.user_id").
		Where("driver_profiles.deleted_at IS NULL")

	switch status {
	case "online":
		query = query.Where("driver_profiles.is_online = ?", true)
	case "verified":
		query = query.Where("driver_profiles.is_verified = ?", true)
	case "pending":
		query = query.Where("driver_profiles.is_verified = ?", false)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("driver_profiles.created_at DESC").Offset(offset).Limit(perPage).Scan(&results).Error

	return results, total, err
}

func (r *AnalyticsRepository) VerifyDriver(ctx context.Context, driverID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("driver_profiles").
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"is_verified": true,
			"verified_at": now,
		}).Error
}

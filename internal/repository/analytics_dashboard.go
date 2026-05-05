package repository

import (
	"context"

	"github.com/wardayadev/ub-mager-api/internal/model"
)

type DashboardStats struct {
	TotalUsers     int64   `json:"total_users"`
	TotalDrivers   int64   `json:"total_drivers"`
	OnlineDrivers  int64   `json:"online_drivers"`
	PendingDrivers int64   `json:"pending_drivers"`
	TotalRides     int64   `json:"total_rides"`
	ActiveRides    int64   `json:"active_rides"`
	CompletedToday int64   `json:"completed_today"`
	RevenueToday   float64 `json:"revenue_today"`
	CancelledToday int64   `json:"cancelled_today"`
	PendingReports int64   `json:"pending_reports"`
}

func (r *AnalyticsRepository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var stats DashboardStats
	today := startOfDay()

	r.db.WithContext(ctx).Model(&model.User{}).Count(&stats.TotalUsers)
	r.db.WithContext(ctx).Model(&model.DriverProfile{}).Count(&stats.TotalDrivers)
	r.db.WithContext(ctx).Model(&model.DriverProfile{}).Where("is_online = ?", true).Count(&stats.OnlineDrivers)
	r.db.WithContext(ctx).Model(&model.DriverProfile{}).Where("is_verified = ?", false).Count(&stats.PendingDrivers)
	r.db.WithContext(ctx).Model(&model.Ride{}).Count(&stats.TotalRides)
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status NOT IN ?", model.TerminalStatuses).
		Count(&stats.ActiveRides)
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, today).
		Count(&stats.CompletedToday)
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status = ? AND cancelled_at >= ?", model.RideStatusCancelled, today).
		Count(&stats.CancelledToday)

	var revenue struct{ Total *float64 }
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Select("COALESCE(SUM(total_fare), 0) as total").
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, today).
		Scan(&revenue)
	if revenue.Total != nil {
		stats.RevenueToday = *revenue.Total
	}

	r.db.WithContext(ctx).Model(&model.Report{}).
		Where("status = ?", model.ReportStatusPending).
		Count(&stats.PendingReports)

	return &stats, nil
}

package handler

import (
	"context"

	"github.com/wardayadev/ub-mager-api/internal/repository"
)

type AnalyticsRepo interface {
	GetDashboardStats(ctx context.Context) (*repository.DashboardStats, error)
	GetRevenueStats(ctx context.Context, period string) (*repository.RevenueStats, error)
	GetDailyRevenue(ctx context.Context, days int) ([]repository.DailyRevenue, error)
	GetRideStats(ctx context.Context, period string) (*repository.RideStats, error)
	ListDrivers(ctx context.Context, page, perPage int, status, search string) ([]repository.DriverListItem, int64, error)
	VerifyDriver(ctx context.Context, driverID string) error
	ToggleDriverOnline(ctx context.Context, driverID string, online bool) error
	GetDriverDetail(ctx context.Context, driverID string) (*repository.DriverDetail, error)
	GetDriverRides(ctx context.Context, driverProfileID string, limit int) ([]repository.DriverRideItem, error)
	GetDriverLeaderboard(ctx context.Context, limit int) ([]repository.DriverPerformance, error)
	GetRideCountsByStatus(ctx context.Context) ([]repository.RideCountByStatus, error)
	GetRideDetail(ctx context.Context, rideID string) (*repository.RideDetail, error)
	ListRides(ctx context.Context, page, perPage int, status, search string) ([]repository.RideListItem, int64, error)
	AdminCancelRide(ctx context.Context, rideID, reason string) error
	BulkCancelStuckRides(ctx context.Context, reason string) (int64, error)
	GetRecentActivity(ctx context.Context, limit int) ([]repository.ActivityItem, error)
	GetPeakHours(ctx context.Context, days int) ([]repository.PeakHourItem, error)
}

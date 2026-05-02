package repository

import (
	"context"
	"fmt"
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

func (r *AnalyticsRepository) ListDrivers(ctx context.Context, page, perPage int, status, search string) ([]DriverListItem, int64, error) {
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

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"(users.full_name ILIKE ? OR users.phone ILIKE ? OR driver_profiles.license_plate ILIKE ?)",
			searchPattern, searchPattern, searchPattern,
		)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("driver_profiles.created_at DESC").Offset(offset).Limit(perPage).Scan(&results).Error

	return results, total, err
}

type RideListItem struct {
	ID                 string     `json:"id"`
	PassengerName      string     `json:"passenger_name"`
	PassengerPhone     string     `json:"passenger_phone"`
	DriverName         *string    `json:"driver_name"`
	Status             string     `json:"status"`
	VehicleType        string     `json:"vehicle_type"`
	PickupAddress      string     `json:"pickup_address"`
	DropoffAddress     string     `json:"dropoff_address"`
	EstimatedDistanceM float64    `json:"estimated_distance_m"`
	EstimatedDurationS int        `json:"estimated_duration_s"`
	TotalFare          float64    `json:"total_fare"`
	RequestedAt        time.Time  `json:"requested_at"`
	CompletedAt        *time.Time `json:"completed_at"`
}

func (r *AnalyticsRepository) ListRides(ctx context.Context, page, perPage int, status, search string) ([]RideListItem, int64, error) {
	var total int64
	var results []RideListItem

	query := r.db.WithContext(ctx).
		Table("rides").
		Select(`
			rides.id,
			passenger.full_name as passenger_name,
			passenger.phone as passenger_phone,
			driver_user.full_name as driver_name,
			rides.status,
			rides.vehicle_type,
			rides.pickup_address,
			rides.dropoff_address,
			rides.estimated_distance_m,
			rides.estimated_duration_s,
			rides.total_fare,
			rides.requested_at,
			rides.completed_at
		`).
		Joins("JOIN users AS passenger ON passenger.id = rides.passenger_id").
		Joins("LEFT JOIN driver_profiles ON driver_profiles.id = rides.driver_id").
		Joins("LEFT JOIN users AS driver_user ON driver_user.id = driver_profiles.user_id").
		Where("rides.deleted_at IS NULL")

	if status != "" {
		query = query.Where("rides.status = ?", status)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"(passenger.full_name ILIKE ? OR passenger.phone ILIKE ? OR rides.pickup_address ILIKE ? OR rides.dropoff_address ILIKE ? OR driver_user.full_name ILIKE ?)",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("rides.requested_at DESC").Offset(offset).Limit(perPage).Scan(&results).Error

	return results, total, err
}

type RideDetail struct {
	ID               string     `json:"id"`
	PassengerName    string     `json:"passenger_name"`
	PassengerPhone   string     `json:"passenger_phone"`
	DriverName       *string    `json:"driver_name"`
	DriverPhone      *string    `json:"driver_phone"`
	DriverPlate      *string    `json:"driver_plate"`
	DriverVehicle    *string    `json:"driver_vehicle"`
	Status           string     `json:"status"`
	VehicleType      string     `json:"vehicle_type"`
	PickupAddress    string     `json:"pickup_address"`
	PickupLat        float64    `json:"pickup_lat"`
	PickupLng        float64    `json:"pickup_lng"`
	DropoffAddress   string     `json:"dropoff_address"`
	DropoffLat       float64    `json:"dropoff_lat"`
	DropoffLng       float64    `json:"dropoff_lng"`
	EstimatedDistanceM float64  `json:"estimated_distance_m"`
	EstimatedDurationS int      `json:"estimated_duration_s"`
	ActualDistanceM  float64    `json:"actual_distance_m"`
	ActualDurationS  int        `json:"actual_duration_s"`
	BaseFare         float64    `json:"base_fare"`
	SurgeMultiplier  float64    `json:"surge_multiplier"`
	TotalFare        float64    `json:"total_fare"`
	PaymentMethod    string     `json:"payment_method"`
	Notes            string     `json:"notes"`
	RequestedAt      time.Time  `json:"requested_at"`
	MatchedAt        *time.Time `json:"matched_at"`
	DriverArrivedAt  *time.Time `json:"driver_arrived_at"`
	PickedUpAt       *time.Time `json:"picked_up_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	CancelledAt      *time.Time `json:"cancelled_at"`
	CancellationReason string   `json:"cancellation_reason"`
}

func (r *AnalyticsRepository) GetRideDetail(ctx context.Context, rideID string) (*RideDetail, error) {
	var result RideDetail
	err := r.db.WithContext(ctx).
		Table("rides").
		Select(`
			rides.id,
			passenger.full_name as passenger_name,
			passenger.phone as passenger_phone,
			driver_user.full_name as driver_name,
			driver_user.phone as driver_phone,
			driver_profiles.license_plate as driver_plate,
			CONCAT(driver_profiles.vehicle_brand, ' ', driver_profiles.vehicle_model) as driver_vehicle,
			rides.status,
			rides.vehicle_type,
			rides.pickup_address,
			rides.pickup_lat,
			rides.pickup_lng,
			rides.dropoff_address,
			rides.dropoff_lat,
			rides.dropoff_lng,
			rides.estimated_distance_m,
			rides.estimated_duration_s,
			rides.actual_distance_m,
			rides.actual_duration_s,
			rides.base_fare,
			rides.surge_multiplier,
			rides.total_fare,
			rides.payment_method,
			rides.notes,
			rides.requested_at,
			rides.matched_at,
			rides.driver_arrived_at,
			rides.picked_up_at,
			rides.completed_at,
			rides.cancelled_at,
			rides.cancellation_reason
		`).
		Joins("JOIN users AS passenger ON passenger.id = rides.passenger_id").
		Joins("LEFT JOIN driver_profiles ON driver_profiles.id = rides.driver_id").
		Joins("LEFT JOIN users AS driver_user ON driver_user.id = driver_profiles.user_id").
		Where("rides.id = ?", rideID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, nil
	}
	return &result, nil
}

type DriverPerformance struct {
	ID           string  `json:"id"`
	FullName     string  `json:"full_name"`
	Phone        string  `json:"phone"`
	VehicleType  string  `json:"vehicle_type"`
	LicensePlate string  `json:"license_plate"`
	Rating       float64 `json:"rating"`
	TotalTrips   int     `json:"total_trips"`
	TotalRevenue float64 `json:"total_revenue"`
	AvgFare      float64 `json:"avg_fare"`
}

func (r *AnalyticsRepository) GetDriverLeaderboard(ctx context.Context, limit int) ([]DriverPerformance, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var results []DriverPerformance
	err := r.db.WithContext(ctx).
		Table("driver_profiles").
		Select(`
			driver_profiles.id,
			users.full_name,
			users.phone,
			driver_profiles.vehicle_type,
			driver_profiles.license_plate,
			driver_profiles.rating_avg as rating,
			driver_profiles.total_trips,
			COALESCE(rev.total_revenue, 0) as total_revenue,
			CASE WHEN driver_profiles.total_trips > 0 
				THEN COALESCE(rev.total_revenue, 0) / driver_profiles.total_trips 
				ELSE 0 END as avg_fare
		`).
		Joins("JOIN users ON users.id = driver_profiles.user_id").
		Joins(`LEFT JOIN (
			SELECT driver_id, SUM(total_fare) as total_revenue 
			FROM rides WHERE status = 'COMPLETED' 
			GROUP BY driver_id
		) rev ON rev.driver_id = driver_profiles.id`).
		Where("driver_profiles.deleted_at IS NULL AND driver_profiles.is_verified = true").
		Order("total_revenue DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

type DailyRevenue struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Rides   int64   `json:"rides"`
}

func (r *AnalyticsRepository) GetDailyRevenue(ctx context.Context, days int) ([]DailyRevenue, error) {
	if days < 1 || days > 90 {
		days = 7
	}

	// Include today: go back (days-1) days from start of today
	since := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -(days - 1))

	var results []DailyRevenue
	err := r.db.WithContext(ctx).
		Table("rides").
		Select("TO_CHAR(completed_at, 'YYYY-MM-DD') as date, COALESCE(SUM(total_fare), 0) as revenue, COUNT(*) as rides").
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, since).
		Group("TO_CHAR(completed_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Fill in missing days with zero values
	filled := fillMissingDays(results, since, days)
	return filled, nil
}

func fillMissingDays(data []DailyRevenue, since time.Time, days int) []DailyRevenue {
	dateMap := make(map[string]DailyRevenue)
	for _, d := range data {
		dateMap[d.Date] = d
	}

	var result []DailyRevenue
	for i := 0; i < days; i++ {
		date := since.AddDate(0, 0, i).Format("2006-01-02")
		if d, ok := dateMap[date]; ok {
			result = append(result, d)
		} else {
			result = append(result, DailyRevenue{Date: date, Revenue: 0, Rides: 0})
		}
	}
	return result
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

type RideCountByStatus struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

func (r *AnalyticsRepository) GetRideCountsByStatus(ctx context.Context) ([]RideCountByStatus, error) {
	var results []RideCountByStatus
	err := r.db.WithContext(ctx).
		Table("rides").
		Select("status, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("status").
		Scan(&results).Error
	return results, err
}

func (r *AnalyticsRepository) AdminCancelRide(ctx context.Context, rideID, reason string) error {
	// Only cancel rides that are not already completed/cancelled
	result := r.db.WithContext(ctx).
		Table("rides").
		Where("id = ? AND status NOT IN ?", rideID, []string{"COMPLETED", "CANCELLED"}).
		Updates(map[string]interface{}{
			"status":              "CANCELLED",
			"cancelled_at":        time.Now(),
			"cancellation_reason": reason,
		})

	if result.RowsAffected == 0 {
		return fmt.Errorf("ride not found or already completed/cancelled")
	}
	return result.Error
}

func (r *AnalyticsRepository) ToggleDriverOnline(ctx context.Context, driverID string, online bool) error {
	updates := map[string]interface{}{
		"is_online": online,
	}
	if !online {
		updates["is_available"] = true
	}
	return r.db.WithContext(ctx).
		Table("driver_profiles").
		Where("id = ?", driverID).
		Updates(updates).Error
}

func (r *AnalyticsRepository) GetDriverDetail(ctx context.Context, driverID string) (map[string]interface{}, error) {
	var result struct {
		ID             string     `json:"id"`
		UserID         string     `json:"user_id"`
		FullName       string     `json:"full_name"`
		Phone          string     `json:"phone"`
		Email          string     `json:"email"`
		VehicleType    string     `json:"vehicle_type"`
		VehicleBrand   string     `json:"vehicle_brand"`
		VehicleModel   string     `json:"vehicle_model"`
		VehicleYear    int        `json:"vehicle_year"`
		VehicleColor   string     `json:"vehicle_color"`
		LicensePlate   string     `json:"license_plate"`
		LicenseNumber  string     `json:"license_number"`
		IsOnline       bool       `json:"is_online"`
		IsAvailable    bool       `json:"is_available"`
		IsVerified     bool       `json:"is_verified"`
		Rating         float64    `json:"rating"`
		TotalTrips     int        `json:"total_trips"`
		AcceptanceRate float64    `json:"acceptance_rate"`
		TotalRevenue   float64    `json:"total_revenue"`
		CreatedAt      time.Time  `json:"created_at"`
		VerifiedAt     *time.Time `json:"verified_at"`
	}

	err := r.db.WithContext(ctx).
		Table("driver_profiles").
		Select(`
			driver_profiles.id,
			driver_profiles.user_id,
			users.full_name,
			users.phone,
			users.email,
			driver_profiles.vehicle_type,
			driver_profiles.vehicle_brand,
			driver_profiles.vehicle_model,
			driver_profiles.vehicle_year,
			driver_profiles.vehicle_color,
			driver_profiles.license_plate,
			driver_profiles.license_number,
			driver_profiles.is_online,
			driver_profiles.is_available,
			driver_profiles.is_verified,
			driver_profiles.rating_avg as rating,
			driver_profiles.total_trips,
			driver_profiles.acceptance_rate,
			COALESCE(rev.total_revenue, 0) as total_revenue,
			driver_profiles.created_at,
			driver_profiles.verified_at
		`).
		Joins("JOIN users ON users.id = driver_profiles.user_id").
		Joins(`LEFT JOIN (
			SELECT driver_id, SUM(total_fare) as total_revenue
			FROM rides WHERE status = 'COMPLETED'
			GROUP BY driver_id
		) rev ON rev.driver_id = driver_profiles.id`).
		Where("driver_profiles.id = ?", driverID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	// Convert to map for flexible JSON
	return map[string]interface{}{
		"id": result.ID, "user_id": result.UserID, "full_name": result.FullName,
		"phone": result.Phone, "email": result.Email, "vehicle_type": result.VehicleType,
		"vehicle_brand": result.VehicleBrand, "vehicle_model": result.VehicleModel,
		"vehicle_year": result.VehicleYear, "vehicle_color": result.VehicleColor,
		"license_plate": result.LicensePlate, "license_number": result.LicenseNumber,
		"is_online": result.IsOnline, "is_available": result.IsAvailable,
		"is_verified": result.IsVerified, "rating": result.Rating,
		"total_trips": result.TotalTrips, "acceptance_rate": result.AcceptanceRate,
		"total_revenue": result.TotalRevenue, "created_at": result.CreatedAt,
		"verified_at": result.VerifiedAt,
	}, nil
}

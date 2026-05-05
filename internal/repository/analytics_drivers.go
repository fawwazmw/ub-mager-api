package repository

import (
	"context"
	"time"
)

type DriverListItem struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	VehicleType  string    `json:"vehicle_type"`
	LicensePlate string    `json:"license_plate"`
	IsOnline     bool      `json:"is_online"`
	IsVerified   bool      `json:"is_verified"`
	Rating       float64   `json:"rating"`
	TotalTrips   int       `json:"total_trips"`
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

type DriverDetail struct {
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

func (r *AnalyticsRepository) GetDriverDetail(ctx context.Context, driverID string) (*DriverDetail, error) {
	var result DriverDetail

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

	if result.ID == "" {
		return nil, nil
	}

	return &result, nil
}

func (r *AnalyticsRepository) VerifyDriver(ctx context.Context, driverID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("driver_profiles").
		Where("id = ?", driverID).
		Updates(map[string]any{
			"is_verified": true,
			"verified_at": now,
		}).Error
}

func (r *AnalyticsRepository) ToggleDriverOnline(ctx context.Context, driverID string, online bool) error {
	updates := map[string]any{
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

type DriverRideItem struct {
	ID             string    `json:"id"`
	Status         string    `json:"status"`
	PassengerName  string    `json:"passenger_name"`
	PickupAddress  string    `json:"pickup_address"`
	DropoffAddress string    `json:"dropoff_address"`
	TotalFare      float64   `json:"total_fare"`
	RequestedAt    time.Time `json:"requested_at"`
}

func (r *AnalyticsRepository) GetDriverRides(ctx context.Context, driverProfileID string, limit int) ([]DriverRideItem, error) {
	var rides []DriverRideItem
	err := r.db.WithContext(ctx).
		Table("rides").
		Select(`rides.id, rides.status, 
			users.full_name as passenger_name,
			rides.pickup_address, rides.dropoff_address,
			rides.total_fare, rides.requested_at`).
		Joins("JOIN users ON users.id = rides.passenger_id").
		Where("rides.driver_id = ?", driverProfileID).
		Order("rides.requested_at DESC").
		Limit(limit).
		Scan(&rides).Error

	return rides, err
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

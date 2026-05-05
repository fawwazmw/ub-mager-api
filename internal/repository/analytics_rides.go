package repository

import (
	"context"
	"time"

	"github.com/wardayadev/ub-mager-api/internal/model"
)

type VehicleTypeCount struct {
	VehicleType string `json:"vehicle_type"`
	Count       int64  `json:"count"`
}

type RideStats struct {
	Period         string             `json:"period"`
	Total          int64              `json:"total"`
	Completed      int64              `json:"completed"`
	Cancelled      int64              `json:"cancelled"`
	CompletionRate float64            `json:"completion_rate"`
	AvgDistanceKm  float64            `json:"avg_distance_km"`
	AvgDurationMin float64            `json:"avg_duration_min"`
	ByVehicleType  []VehicleTypeCount `json:"by_vehicle_type"`
}

func (r *AnalyticsRepository) GetRideStats(ctx context.Context, period string) (*RideStats, error) {
	since, period := parsePeriod(period)

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

	var avgMetrics struct {
		AvgDistance  *float64
		AvgDuration *float64
	}
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Select("COALESCE(AVG(estimated_distance_m), 0) / 1000.0 as avg_distance, COALESCE(AVG(estimated_duration_s), 0) / 60.0 as avg_duration").
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, since).
		Scan(&avgMetrics)
	if avgMetrics.AvgDistance != nil {
		stats.AvgDistanceKm = *avgMetrics.AvgDistance
	}
	if avgMetrics.AvgDuration != nil {
		stats.AvgDurationMin = *avgMetrics.AvgDuration
	}

	var vehicleCounts []VehicleTypeCount
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Select("vehicle_type, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("vehicle_type").
		Scan(&vehicleCounts)
	stats.ByVehicleType = vehicleCounts

	return stats, nil
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
	ID                 string     `json:"id"`
	PassengerName      string     `json:"passenger_name"`
	PassengerPhone     string     `json:"passenger_phone"`
	DriverName         *string    `json:"driver_name"`
	DriverPhone        *string    `json:"driver_phone"`
	DriverPlate        *string    `json:"driver_plate"`
	DriverVehicle      *string    `json:"driver_vehicle"`
	Status             string     `json:"status"`
	VehicleType        string     `json:"vehicle_type"`
	PickupAddress      string     `json:"pickup_address"`
	PickupLat          float64    `json:"pickup_lat"`
	PickupLng          float64    `json:"pickup_lng"`
	DropoffAddress     string     `json:"dropoff_address"`
	DropoffLat         float64    `json:"dropoff_lat"`
	DropoffLng         float64    `json:"dropoff_lng"`
	EstimatedDistanceM float64    `json:"estimated_distance_m"`
	EstimatedDurationS int        `json:"estimated_duration_s"`
	ActualDistanceM    float64    `json:"actual_distance_m"`
	ActualDurationS    int        `json:"actual_duration_s"`
	BaseFare           float64    `json:"base_fare"`
	SurgeMultiplier    float64    `json:"surge_multiplier"`
	TotalFare          float64    `json:"total_fare"`
	PaymentMethod      string     `json:"payment_method"`
	Notes              string     `json:"notes"`
	RequestedAt        time.Time  `json:"requested_at"`
	MatchedAt          *time.Time `json:"matched_at"`
	DriverArrivedAt    *time.Time `json:"driver_arrived_at"`
	PickedUpAt         *time.Time `json:"picked_up_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	CancelledAt        *time.Time `json:"cancelled_at"`
	CancellationReason string     `json:"cancellation_reason"`
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
	result := r.db.WithContext(ctx).
		Table("rides").
		Where("id = ? AND status NOT IN ?", rideID, model.TerminalStatuses).
		Updates(map[string]any{
			"status":              model.RideStatusCancelled,
			"cancelled_at":        time.Now(),
			"cancellation_reason": reason,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRideNotFoundOrTerminal
	}
	return nil
}

func (r *AnalyticsRepository) BulkCancelStuckRides(ctx context.Context, reason string) (int64, error) {
	cutoff := time.Now().Add(-30 * time.Minute)
	result := r.db.WithContext(ctx).
		Model(&model.Ride{}).
		Where("status = ? AND requested_at < ?", model.RideStatusSearching, cutoff).
		Updates(map[string]any{
			"status":              model.RideStatusCancelled,
			"cancelled_at":        time.Now(),
			"cancellation_reason": reason,
		})
	return result.RowsAffected, result.Error
}

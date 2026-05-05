package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/wardayadev/ub-mager-api/internal/model"
)

type ActivityItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (r *AnalyticsRepository) GetRecentActivity(ctx context.Context, limit int) ([]ActivityItem, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}

	var rides []struct {
		ID             string
		Status         string
		PassengerName  string
		DriverName     *string
		PickupAddress  string
		DropoffAddress string
		RequestedAt    time.Time
		CompletedAt    *time.Time
		CancelledAt    *time.Time
	}

	err := r.db.WithContext(ctx).
		Table("rides").
		Select(`rides.id, rides.status, 
			pu.full_name as passenger_name, 
			du.full_name as driver_name,
			rides.pickup_address, rides.dropoff_address,
			rides.requested_at, rides.completed_at, rides.cancelled_at`).
		Joins("JOIN users pu ON pu.id = rides.passenger_id").
		Joins("LEFT JOIN driver_profiles dp ON dp.id = rides.driver_id").
		Joins("LEFT JOIN users du ON du.id = dp.user_id").
		Order("rides.updated_at DESC").
		Limit(limit).
		Scan(&rides).Error

	if err != nil {
		return nil, err
	}

	var newDrivers []struct {
		ID        string
		FullName  string
		CreatedAt time.Time
	}

	r.db.WithContext(ctx).
		Table("driver_profiles").
		Select("driver_profiles.id, users.full_name, driver_profiles.created_at").
		Joins("JOIN users ON users.id = driver_profiles.user_id").
		Order("driver_profiles.created_at DESC").
		Limit(5).
		Scan(&newDrivers)

	var activities []ActivityItem

	for _, ride := range rides {
		var actType, msg string
		var ts time.Time

		switch model.RideStatus(ride.Status) {
		case model.RideStatusCompleted:
			actType = "ride_completed"
			msg = fmt.Sprintf("%s completed ride to %s", ride.PassengerName, ride.DropoffAddress)
			if ride.CompletedAt != nil {
				ts = *ride.CompletedAt
			} else {
				ts = ride.RequestedAt
			}
		case model.RideStatusCancelled:
			actType = "ride_cancelled"
			msg = fmt.Sprintf("%s cancelled ride from %s", ride.PassengerName, ride.PickupAddress)
			if ride.CancelledAt != nil {
				ts = *ride.CancelledAt
			} else {
				ts = ride.RequestedAt
			}
		default:
			actType = "ride_active"
			driverStr := "waiting for driver"
			if ride.DriverName != nil {
				driverStr = *ride.DriverName
			}
			msg = fmt.Sprintf("%s → %s (%s)", ride.PassengerName, ride.DropoffAddress, driverStr)
			ts = ride.RequestedAt
		}

		activities = append(activities, ActivityItem{
			ID:        ride.ID,
			Type:      actType,
			Message:   msg,
			Timestamp: ts,
		})
	}

	for _, d := range newDrivers {
		activities = append(activities, ActivityItem{
			ID:        d.ID,
			Type:      "driver_registered",
			Message:   fmt.Sprintf("%s registered as driver", d.FullName),
			Timestamp: d.CreatedAt,
		})
	}

	return activities, nil
}

type PeakHourItem struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

func (r *AnalyticsRepository) GetPeakHours(ctx context.Context, days int) ([]PeakHourItem, error) {
	if days < 1 || days > 90 {
		days = 7
	}

	since := time.Now().AddDate(0, 0, -days)

	var results []PeakHourItem
	err := r.db.WithContext(ctx).
		Table("rides").
		Select("EXTRACT(HOUR FROM requested_at)::int as hour, COUNT(*) as count").
		Where("requested_at >= ?", since).
		Group("hour").
		Order("hour ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	hourMap := make(map[int]int64)
	for _, r := range results {
		hourMap[r.Hour] = r.Count
	}

	full := make([]PeakHourItem, 24)
	for h := 0; h < 24; h++ {
		full[h] = PeakHourItem{Hour: h, Count: hourMap[h]}
	}

	return full, nil
}

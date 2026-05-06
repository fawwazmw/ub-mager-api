package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

type DriverRepository struct {
	db *gorm.DB
}

func NewDriverRepository(db *gorm.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

func (r *DriverRepository) Create(ctx context.Context, profile *model.DriverProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *DriverRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.DriverProfile, error) {
	var profile model.DriverProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *DriverRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.DriverProfile, error) {
	var profile model.DriverProfile
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *DriverRepository) Update(ctx context.Context, profile *model.DriverProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *DriverRepository) UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lng, heading, speed float64) error {
	return r.db.WithContext(ctx).
		Model(&model.DriverProfile{}).
		Where("id = ?", driverID).
		Updates(map[string]any{
			"latitude":         lat,
			"longitude":        lng,
			"heading":          heading,
			"speed":            speed,
			"last_location_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *DriverRepository) SetOnlineStatus(ctx context.Context, driverID uuid.UUID, isOnline bool) error {
	updates := map[string]any{
		"is_online": isOnline,
	}
	if !isOnline {
		updates["is_available"] = true
	}
	return r.db.WithContext(ctx).
		Model(&model.DriverProfile{}).
		Where("id = ?", driverID).
		Updates(updates).Error
}

func (r *DriverRepository) IncrementAcceptanceStats(ctx context.Context, driverID uuid.UUID, accepted bool) error {
	if accepted {
		return r.db.WithContext(ctx).
			Model(&model.DriverProfile{}).
			Where("id = ?", driverID).
			UpdateColumn("acceptance_rate",
				gorm.Expr("LEAST(100, (acceptance_rate * total_trips + 100) / (total_trips + 1))"),
			).Error
	}
	return r.db.WithContext(ctx).
		Model(&model.DriverProfile{}).
		Where("id = ?", driverID).
		UpdateColumn("acceptance_rate",
			gorm.Expr("GREATEST(0, (acceptance_rate * total_trips) / (total_trips + 1))"),
		).Error
}

func (r *DriverRepository) FindNearbyDrivers(ctx context.Context, lat, lng, radiusKm float64, vehicleType model.VehicleType, limit int) ([]model.DriverProfile, error) {
	var drivers []model.DriverProfile

	query := r.db.WithContext(ctx).
		Where("is_online = ? AND is_available = ? AND is_verified = ?", true, true, true).
		Where("latitude IS NOT NULL AND longitude IS NOT NULL")

	if vehicleType != "" {
		query = query.Where("vehicle_type = ?", vehicleType)
	}

	query = query.Where(`
		(6371 * acos(
			cos(radians(?)) * cos(radians(latitude)) *
			cos(radians(longitude) - radians(?)) +
			sin(radians(?)) * sin(radians(latitude))
		)) <= ?`, lat, lng, lat, radiusKm)

	query = query.Order(gorm.Expr(`
		(6371 * acos(
			cos(radians(?)) * cos(radians(latitude)) *
			cos(radians(longitude) - radians(?)) +
			sin(radians(?)) * sin(radians(latitude))
		))`, lat, lng, lat))

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&drivers).Error
	return drivers, err
}

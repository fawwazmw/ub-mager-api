package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

type RideRepository struct {
	db *gorm.DB
}

func NewRideRepository(db *gorm.DB) *RideRepository {
	return &RideRepository{db: db}
}

func (r *RideRepository) Create(ctx context.Context, ride *model.Ride) error {
	return r.db.WithContext(ctx).Create(ride).Error
}

func (r *RideRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Ride, error) {
	var ride model.Ride
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&ride).Error
	if err != nil {
		return nil, err
	}
	return &ride, nil
}

func (r *RideRepository) FindActiveByPassenger(ctx context.Context, passengerID uuid.UUID) (*model.Ride, error) {
	var ride model.Ride
	err := r.db.WithContext(ctx).
		Where("passenger_id = ? AND status NOT IN ?", passengerID, model.TerminalStatuses).
		Order("created_at DESC").
		First(&ride).Error
	if err != nil {
		return nil, err
	}
	return &ride, nil
}

func (r *RideRepository) FindActiveByDriver(ctx context.Context, driverID uuid.UUID) (*model.Ride, error) {
	var ride model.Ride
	err := r.db.WithContext(ctx).
		Where("driver_id = ? AND status NOT IN ?", driverID, model.TerminalStatuses).
		Order("created_at DESC").
		First(&ride).Error
	if err != nil {
		return nil, err
	}
	return &ride, nil
}

func (r *RideRepository) Update(ctx context.Context, ride *model.Ride) error {
	return r.db.WithContext(ctx).Save(ride).Error
}

func (r *RideRepository) UpdateStatus(ctx context.Context, rideID uuid.UUID, status model.RideStatus, updates map[string]any) error {
	updates["status"] = status
	return r.db.WithContext(ctx).
		Model(&model.Ride{}).
		Where("id = ?", rideID).
		Updates(updates).Error
}

func (r *RideRepository) FindHistory(ctx context.Context, userID uuid.UUID, role string, page, perPage int) ([]model.Ride, int64, error) {
	var rides []model.Ride
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Ride{})
	if role == string(model.RolePassenger) {
		query = query.Where("passenger_id = ?", userID)
	} else {
		query = query.Where("driver_id = ?", userID)
	}
	query = query.Where("status IN ?", model.TerminalStatuses)

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&rides).Error
	return rides, total, err
}

func (r *RideRepository) CreateRating(ctx context.Context, rating *model.Rating) error {
	return r.db.WithContext(ctx).Create(rating).Error
}

func (r *RideRepository) FindRatingByRideAndRater(ctx context.Context, rideID, raterID uuid.UUID) (*model.Rating, error) {
	var rating model.Rating
	err := r.db.WithContext(ctx).
		Where("ride_id = ? AND rater_id = ?", rideID, raterID).
		First(&rating).Error
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

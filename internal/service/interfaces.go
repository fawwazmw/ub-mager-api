package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
)

type UserRepo interface {
	Create(ctx context.Context, user *model.User) error
	FindByPhone(ctx context.Context, phone string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

type DriverRepo interface {
	Create(ctx context.Context, profile *model.DriverProfile) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.DriverProfile, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.DriverProfile, error)
	Update(ctx context.Context, profile *model.DriverProfile) error
	UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lng, heading, speed float64) error
	SetOnlineStatus(ctx context.Context, driverID uuid.UUID, isOnline bool) error
	IncrementAcceptanceStats(ctx context.Context, driverID uuid.UUID, accepted bool) error
	FindNearbyDrivers(ctx context.Context, lat, lng, radiusKm float64, vehicleType model.VehicleType, limit int) ([]model.DriverProfile, error)
}

type RideRepo interface {
	Create(ctx context.Context, ride *model.Ride) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Ride, error)
	FindActiveByPassenger(ctx context.Context, passengerID uuid.UUID) (*model.Ride, error)
	FindActiveByDriver(ctx context.Context, driverID uuid.UUID) (*model.Ride, error)
	Update(ctx context.Context, ride *model.Ride) error
	UpdateStatus(ctx context.Context, rideID uuid.UUID, status model.RideStatus, updates map[string]any) error
	FindHistory(ctx context.Context, userID uuid.UUID, role string, page, perPage int) ([]model.Ride, int64, error)
	CreateRating(ctx context.Context, rating *model.Rating) error
	FindRatingByRideAndRater(ctx context.Context, rideID, raterID uuid.UUID) (*model.Rating, error)
}

type GeoCache interface {
	UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lng float64) error
	RemoveDriver(ctx context.Context, driverID uuid.UUID) error
	SetDriverStatus(ctx context.Context, driverID uuid.UUID, status model.DriverStatus) error
}

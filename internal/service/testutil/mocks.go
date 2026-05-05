package testutil

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/wardayadev/ub-mager-api/internal/model"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockDriverRepo struct {
	mock.Mock
}

func (m *MockDriverRepo) Create(ctx context.Context, profile *model.DriverProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockDriverRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.DriverProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DriverProfile), args.Error(1)
}

func (m *MockDriverRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.DriverProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DriverProfile), args.Error(1)
}

func (m *MockDriverRepo) Update(ctx context.Context, profile *model.DriverProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockDriverRepo) UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lng, heading, speed float64) error {
	args := m.Called(ctx, driverID, lat, lng, heading, speed)
	return args.Error(0)
}

func (m *MockDriverRepo) SetOnlineStatus(ctx context.Context, driverID uuid.UUID, isOnline bool) error {
	args := m.Called(ctx, driverID, isOnline)
	return args.Error(0)
}

func (m *MockDriverRepo) IncrementAcceptanceStats(ctx context.Context, driverID uuid.UUID, accepted bool) error {
	args := m.Called(ctx, driverID, accepted)
	return args.Error(0)
}

func (m *MockDriverRepo) FindNearbyDrivers(ctx context.Context, lat, lng, radiusKm float64, vehicleType model.VehicleType, limit int) ([]model.DriverProfile, error) {
	args := m.Called(ctx, lat, lng, radiusKm, vehicleType, limit)
	return args.Get(0).([]model.DriverProfile), args.Error(1)
}

type MockRideRepo struct {
	mock.Mock
}

func (m *MockRideRepo) Create(ctx context.Context, ride *model.Ride) error {
	args := m.Called(ctx, ride)
	return args.Error(0)
}

func (m *MockRideRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Ride, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Ride), args.Error(1)
}

func (m *MockRideRepo) FindActiveByPassenger(ctx context.Context, passengerID uuid.UUID) (*model.Ride, error) {
	args := m.Called(ctx, passengerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Ride), args.Error(1)
}

func (m *MockRideRepo) FindActiveByDriver(ctx context.Context, driverID uuid.UUID) (*model.Ride, error) {
	args := m.Called(ctx, driverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Ride), args.Error(1)
}

func (m *MockRideRepo) Update(ctx context.Context, ride *model.Ride) error {
	args := m.Called(ctx, ride)
	return args.Error(0)
}

func (m *MockRideRepo) UpdateStatus(ctx context.Context, rideID uuid.UUID, status model.RideStatus, updates map[string]any) error {
	args := m.Called(ctx, rideID, status, updates)
	return args.Error(0)
}

func (m *MockRideRepo) FindHistory(ctx context.Context, userID uuid.UUID, role string, page, perPage int) ([]model.Ride, int64, error) {
	args := m.Called(ctx, userID, role, page, perPage)
	return args.Get(0).([]model.Ride), args.Get(1).(int64), args.Error(2)
}

func (m *MockRideRepo) CreateRating(ctx context.Context, rating *model.Rating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockRideRepo) FindRatingByRideAndRater(ctx context.Context, rideID, raterID uuid.UUID) (*model.Rating, error) {
	args := m.Called(ctx, rideID, raterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rating), args.Error(1)
}

type MockGeoCache struct {
	mock.Mock
}

func (m *MockGeoCache) UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lng float64) error {
	args := m.Called(ctx, driverID, lat, lng)
	return args.Error(0)
}

func (m *MockGeoCache) RemoveDriver(ctx context.Context, driverID uuid.UUID) error {
	args := m.Called(ctx, driverID)
	return args.Error(0)
}

func (m *MockGeoCache) SetDriverStatus(ctx context.Context, driverID uuid.UUID, status model.DriverStatus) error {
	args := m.Called(ctx, driverID, status)
	return args.Error(0)
}

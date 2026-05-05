package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/wardayadev/ub-mager-api/internal/model"
	"github.com/wardayadev/ub-mager-api/internal/service/testutil"
)

func newTestRideService() (*RideService, *testutil.MockRideRepo, *testutil.MockDriverRepo, *testutil.MockUserRepo) {
	rideRepo := new(testutil.MockRideRepo)
	driverRepo := new(testutil.MockDriverRepo)
	userRepo := new(testutil.MockUserRepo)
	svc := NewRideService(rideRepo, driverRepo, userRepo)
	return svc, rideRepo, driverRepo, userRepo
}

func TestEstimate_Success(t *testing.T) {
	svc, _, _, _ := newTestRideService()
	ctx := context.Background()

	input := EstimateInput{
		Pickup:      LocationInput{Lat: -7.9526, Lng: 112.6146},
		Dropoff:     LocationInput{Lat: -7.9666, Lng: 112.6326},
		VehicleType: "motorcycle",
	}

	breakdown, distanceM, durationS, err := svc.Estimate(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, breakdown)
	assert.Greater(t, distanceM, 0.0)
	assert.Greater(t, durationS, 0)
	assert.Greater(t, breakdown.TotalEstimate, 0.0)
}

func TestRequestRide_Success(t *testing.T) {
	svc, rideRepo, _, _ := newTestRideService()
	ctx := context.Background()
	passengerID := uuid.New()

	rideRepo.On("FindActiveByPassenger", ctx, passengerID).Return(nil, gorm.ErrRecordNotFound)
	rideRepo.On("Create", ctx, mock.AnythingOfType("*model.Ride")).Return(nil)

	input := RequestRideInput{
		Pickup:        LocationWithAddress{Lat: -7.9526, Lng: 112.6146, Address: "FILKOM UB"},
		Dropoff:       LocationWithAddress{Lat: -7.9666, Lng: 112.6326, Address: "Mall Dinoyo"},
		VehicleType:   "motorcycle",
		PaymentMethod: "cash",
	}

	ride, err := svc.RequestRide(ctx, passengerID, input)

	assert.NoError(t, err)
	assert.NotNil(t, ride)
	assert.Equal(t, model.RideStatusSearching, ride.Status)
	assert.Equal(t, passengerID, ride.PassengerID)
	assert.Equal(t, model.PaymentCash, ride.PaymentMethod)
	rideRepo.AssertExpectations(t)
}

func TestRequestRide_ActiveRideExists(t *testing.T) {
	svc, rideRepo, _, _ := newTestRideService()
	ctx := context.Background()
	passengerID := uuid.New()

	existingRide := &model.Ride{ID: uuid.New(), Status: model.RideStatusSearching}
	rideRepo.On("FindActiveByPassenger", ctx, passengerID).Return(existingRide, nil)

	input := RequestRideInput{
		Pickup:        LocationWithAddress{Lat: -7.9526, Lng: 112.6146, Address: "FILKOM"},
		Dropoff:       LocationWithAddress{Lat: -7.9666, Lng: 112.6326, Address: "Mall"},
		VehicleType:   "motorcycle",
		PaymentMethod: "cash",
	}

	ride, err := svc.RequestRide(ctx, passengerID, input)

	assert.ErrorIs(t, err, ErrActiveRideExists)
	assert.Nil(t, ride)
}

func TestAcceptRide_Success(t *testing.T) {
	svc, rideRepo, driverRepo, _ := newTestRideService()
	ctx := context.Background()
	rideID := uuid.New()
	userID := uuid.New()
	driverProfileID := uuid.New()

	ride := &model.Ride{ID: rideID, Status: model.RideStatusSearching}
	rideRepo.On("FindByID", ctx, rideID).Return(ride, nil)

	profile := &model.DriverProfile{ID: driverProfileID, UserID: userID}
	driverRepo.On("FindByUserID", ctx, userID).Return(profile, nil)
	rideRepo.On("UpdateStatus", ctx, rideID, model.RideStatusMatched, mock.Anything).Return(nil)
	driverRepo.On("IncrementAcceptanceStats", ctx, driverProfileID, true).Return(nil)

	err := svc.AcceptRide(ctx, rideID, userID)

	assert.NoError(t, err)
	rideRepo.AssertExpectations(t)
	driverRepo.AssertExpectations(t)
}

func TestAcceptRide_RideNotSearching(t *testing.T) {
	svc, rideRepo, _, _ := newTestRideService()
	ctx := context.Background()
	rideID := uuid.New()
	userID := uuid.New()

	ride := &model.Ride{ID: rideID, Status: model.RideStatusMatched}
	rideRepo.On("FindByID", ctx, rideID).Return(ride, nil)

	err := svc.AcceptRide(ctx, rideID, userID)

	assert.ErrorIs(t, err, ErrRideNoLongerAvailable)
}

func TestCancelRide_Success(t *testing.T) {
	svc, rideRepo, _, _ := newTestRideService()
	ctx := context.Background()
	rideID := uuid.New()
	passengerID := uuid.New()

	ride := &model.Ride{ID: rideID, PassengerID: passengerID, Status: model.RideStatusSearching}
	rideRepo.On("FindByID", ctx, rideID).Return(ride, nil)
	rideRepo.On("UpdateStatus", ctx, rideID, model.RideStatusCancelled, mock.Anything).Return(nil)

	err := svc.CancelRide(ctx, rideID, passengerID, "Changed my mind")

	assert.NoError(t, err)
	rideRepo.AssertExpectations(t)
}

func TestCancelRide_Unauthorized(t *testing.T) {
	svc, rideRepo, _, _ := newTestRideService()
	ctx := context.Background()
	rideID := uuid.New()
	passengerID := uuid.New()
	otherUserID := uuid.New()

	ride := &model.Ride{ID: rideID, PassengerID: passengerID, Status: model.RideStatusSearching}
	rideRepo.On("FindByID", ctx, rideID).Return(ride, nil)

	err := svc.CancelRide(ctx, rideID, otherUserID, "reason")

	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestCancelRide_NotCancellable(t *testing.T) {
	svc, rideRepo, _, _ := newTestRideService()
	ctx := context.Background()
	rideID := uuid.New()
	passengerID := uuid.New()

	ride := &model.Ride{ID: rideID, PassengerID: passengerID, Status: model.RideStatusCompleted}
	rideRepo.On("FindByID", ctx, rideID).Return(ride, nil)

	err := svc.CancelRide(ctx, rideID, passengerID, "reason")

	assert.ErrorIs(t, err, ErrRideNotCancellable)
}

func TestGetHistory_Defaults(t *testing.T) {
	svc, rideRepo, _, _ := newTestRideService()
	ctx := context.Background()
	userID := uuid.New()

	rideRepo.On("FindHistory", ctx, userID, "PASSENGER", 1, 20).Return([]model.Ride{}, int64(0), nil)

	rides, total, err := svc.GetHistory(ctx, userID, "PASSENGER", 0, 0)

	assert.NoError(t, err)
	assert.Empty(t, rides)
	assert.Equal(t, int64(0), total)
}

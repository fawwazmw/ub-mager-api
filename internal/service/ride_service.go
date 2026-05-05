package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/wardayadev/ub-mager-api/internal/model"
)

var (
	ErrRideNotFound          = errors.New("ride not found")
	ErrActiveRideExists      = errors.New("you already have an active ride")
	ErrRideNotCancellable    = errors.New("ride cannot be cancelled in current state")
	ErrAlreadyRated          = errors.New("you have already rated this ride")
	ErrRideNotCompleted      = errors.New("ride must be completed before rating")
	ErrUnauthorized          = errors.New("unauthorized to perform this action")
	ErrRideNoLongerAvailable = errors.New("ride is no longer available")
	ErrInvalidTransition     = errors.New("invalid status transition")
	ErrNoDriverToRate        = errors.New("no driver to rate")
)

var validTransitions = map[model.RideStatus][]model.RideStatus{
	model.RideStatusMatched:         {model.RideStatusDriverEnRoute},
	model.RideStatusDriverEnRoute:   {model.RideStatusArrivedAtPickup},
	model.RideStatusArrivedAtPickup: {model.RideStatusInProgress},
	model.RideStatusInProgress:      {model.RideStatusCompleted},
}

type RideService struct {
	rideRepo   RideRepo
	driverRepo DriverRepo
	userRepo   UserRepo
}

func NewRideService(
	rideRepo RideRepo,
	driverRepo DriverRepo,
	userRepo UserRepo,
) *RideService {
	return &RideService{
		rideRepo:   rideRepo,
		driverRepo: driverRepo,
		userRepo:   userRepo,
	}
}

type EstimateInput struct {
	Pickup      LocationInput `json:"pickup" binding:"required"`
	Dropoff     LocationInput `json:"dropoff" binding:"required"`
	VehicleType string        `json:"vehicle_type" binding:"required,oneof=motorcycle car car_xl"`
}

type LocationWithAddress struct {
	Lat     float64 `json:"lat" binding:"required"`
	Lng     float64 `json:"lng" binding:"required"`
	Address string  `json:"address" binding:"required"`
}

type RequestRideInput struct {
	Pickup        LocationWithAddress `json:"pickup" binding:"required"`
	Dropoff       LocationWithAddress `json:"dropoff" binding:"required"`
	VehicleType   string              `json:"vehicle_type" binding:"required,oneof=motorcycle car car_xl"`
	PaymentMethod string              `json:"payment_method" binding:"required,oneof=cash ewallet"`
	PickupZone    string              `json:"pickup_zone"`
	DropoffZone   string              `json:"dropoff_zone"`
	Notes         string              `json:"notes" binding:"max=200"`
}

type RateInput struct {
	Score   int    `json:"score" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"max=500"`
}

func estimateRoute(pickupLat, pickupLng, dropoffLat, dropoffLng float64) (distanceM float64, durationS int) {
	distanceM = haversineDistance(pickupLat, pickupLng, dropoffLat, dropoffLng)
	avgSpeedMps := 25.0 * 1000.0 / 3600.0
	durationS = int(distanceM / avgSpeedMps)
	return
}

func (s *RideService) Estimate(ctx context.Context, input EstimateInput) (*FareBreakdown, float64, int, error) {
	distanceM, durationS := estimateRoute(input.Pickup.Lat, input.Pickup.Lng, input.Dropoff.Lat, input.Dropoff.Lng)
	vt := stringToVehicleType(input.VehicleType)
	breakdown := CalculateFare(vt, distanceM, durationS, 1.0)

	return &breakdown, distanceM, durationS, nil
}

func (s *RideService) RequestRide(ctx context.Context, passengerID uuid.UUID, input RequestRideInput) (*model.Ride, error) {
	existing, err := s.rideRepo.FindActiveByPassenger(ctx, passengerID)
	if err == nil && existing != nil {
		return nil, ErrActiveRideExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	distanceM, durationS := estimateRoute(input.Pickup.Lat, input.Pickup.Lng, input.Dropoff.Lat, input.Dropoff.Lng)

	vt := stringToVehicleType(input.VehicleType)
	fare := CalculateFare(vt, distanceM, durationS, 1.0)

	pm := model.PaymentCash
	if input.PaymentMethod == "ewallet" {
		pm = model.PaymentEwallet
	}

	ride := &model.Ride{
		ID:                 uuid.New(),
		PassengerID:        passengerID,
		Status:             model.RideStatusSearching,
		VehicleType:        vt,
		PickupLat:          input.Pickup.Lat,
		PickupLng:          input.Pickup.Lng,
		PickupAddress:      input.Pickup.Address,
		PickupZone:         model.CampusZone(input.PickupZone),
		DropoffLat:         input.Dropoff.Lat,
		DropoffLng:         input.Dropoff.Lng,
		DropoffAddress:     input.Dropoff.Address,
		DropoffZone:        model.CampusZone(input.DropoffZone),
		EstimatedDistanceM: distanceM,
		EstimatedDurationS: durationS,
		BaseFare:           fare.TotalEstimate,
		SurgeMultiplier:    1.0,
		TotalFare:          fare.TotalEstimate,
		PaymentMethod:      pm,
		Notes:              input.Notes,
		RequestedAt:        time.Now(),
	}

	if err := s.rideRepo.Create(ctx, ride); err != nil {
		return nil, err
	}

	return ride, nil
}

func (s *RideService) GetRide(ctx context.Context, rideID uuid.UUID) (*model.Ride, error) {
	ride, err := s.rideRepo.FindByID(ctx, rideID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRideNotFound
		}
		return nil, err
	}
	return ride, nil
}

func (s *RideService) GetActiveRide(ctx context.Context, passengerID uuid.UUID) (*model.Ride, error) {
	ride, err := s.rideRepo.FindActiveByPassenger(ctx, passengerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRideNotFound
		}
		return nil, err
	}
	return ride, nil
}

func (s *RideService) CancelRide(ctx context.Context, rideID, userID uuid.UUID, reason string) error {
	ride, err := s.rideRepo.FindByID(ctx, rideID)
	if err != nil {
		return ErrRideNotFound
	}

	if ride.PassengerID != userID && (ride.DriverID == nil || *ride.DriverID != userID) {
		return ErrUnauthorized
	}

	switch ride.Status {
	case model.RideStatusSearching, model.RideStatusMatched, model.RideStatusDriverEnRoute, model.RideStatusArrivedAtPickup:
	default:
		return ErrRideNotCancellable
	}

	now := time.Now()
	return s.rideRepo.UpdateStatus(ctx, rideID, model.RideStatusCancelled, map[string]any{
		"cancelled_at":        now,
		"cancellation_reason": reason,
	})
}

func (s *RideService) AcceptRide(ctx context.Context, rideID, userID uuid.UUID) error {
	ride, err := s.rideRepo.FindByID(ctx, rideID)
	if err != nil {
		return ErrRideNotFound
	}
	if ride.Status != model.RideStatusSearching {
		return ErrRideNoLongerAvailable
	}

	profile, err := s.driverRepo.FindByUserID(ctx, userID)
	if err != nil {
		return ErrDriverNotFound
	}

	now := time.Now()
	err = s.rideRepo.UpdateStatus(ctx, rideID, model.RideStatusMatched, map[string]any{
		"driver_id":  profile.ID,
		"matched_at": now,
	})
	if err != nil {
		return err
	}

	_ = s.driverRepo.IncrementAcceptanceStats(ctx, profile.ID, true)
	return nil
}

func (s *RideService) UpdateRideStatus(ctx context.Context, rideID uuid.UUID, userID uuid.UUID, newStatus model.RideStatus) error {
	ride, err := s.rideRepo.FindByID(ctx, rideID)
	if err != nil {
		return ErrRideNotFound
	}

	profile, err := s.driverRepo.FindByUserID(ctx, userID)
	if err != nil {
		return ErrUnauthorized
	}

	if ride.DriverID == nil || *ride.DriverID != profile.ID {
		return ErrUnauthorized
	}

	allowed, ok := validTransitions[ride.Status]
	if !ok {
		return ErrInvalidTransition
	}

	valid := false
	for _, st := range allowed {
		if st == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return ErrInvalidTransition
	}

	now := time.Now()
	updates := map[string]any{}

	switch newStatus {
	case model.RideStatusDriverEnRoute:
	case model.RideStatusArrivedAtPickup:
		updates["driver_arrived_at"] = now
	case model.RideStatusInProgress:
		updates["picked_up_at"] = now
	case model.RideStatusCompleted:
		updates["completed_at"] = now
		if ride.PickedUpAt != nil {
			updates["actual_duration_s"] = int(now.Sub(*ride.PickedUpAt).Seconds())
		}
	}

	return s.rideRepo.UpdateStatus(ctx, rideID, newStatus, updates)
}

func (s *RideService) RateRide(ctx context.Context, rideID, raterID uuid.UUID, input RateInput) error {
	ride, err := s.rideRepo.FindByID(ctx, rideID)
	if err != nil {
		return ErrRideNotFound
	}

	if ride.Status != model.RideStatusCompleted {
		return ErrRideNotCompleted
	}

	existing, err := s.rideRepo.FindRatingByRideAndRater(ctx, rideID, raterID)
	if err == nil && existing != nil {
		return ErrAlreadyRated
	}

	var rateeID uuid.UUID
	if ride.PassengerID == raterID {
		if ride.DriverID == nil {
			return ErrNoDriverToRate
		}
		driver, err := s.driverRepo.FindByID(ctx, *ride.DriverID)
		if err != nil {
			return err
		}
		rateeID = driver.UserID
	} else {
		rateeID = ride.PassengerID
	}

	rating := &model.Rating{
		ID:      uuid.New(),
		RideID:  rideID,
		RaterID: raterID,
		RateeID: rateeID,
		Score:   input.Score,
		Comment: input.Comment,
	}

	return s.rideRepo.CreateRating(ctx, rating)
}

func (s *RideService) GetHistory(ctx context.Context, userID uuid.UUID, role string, page, perPage int) ([]model.Ride, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.rideRepo.FindHistory(ctx, userID, role, page, perPage)
}

func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

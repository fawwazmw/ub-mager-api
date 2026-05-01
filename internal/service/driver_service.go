package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/wardayadev/ub-mager-api/internal/cache"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"github.com/wardayadev/ub-mager-api/internal/repository"
)

var (
	ErrDriverNotFound       = errors.New("driver profile not found")
	ErrDriverAlreadyExists  = errors.New("driver profile already exists")
	ErrDriverNotVerified    = errors.New("driver not verified")
	ErrNotADriver           = errors.New("user is not a driver")
)

type DriverService struct {
	driverRepo  *repository.DriverRepository
	userRepo    *repository.UserRepository
	geoCache    *cache.DriverGeoCache
}

func NewDriverService(
	driverRepo *repository.DriverRepository,
	userRepo *repository.UserRepository,
	geoCache *cache.DriverGeoCache,
) *DriverService {
	return &DriverService{
		driverRepo: driverRepo,
		userRepo:   userRepo,
		geoCache:   geoCache,
	}
}

type RegisterDriverInput struct {
	LicenseNumber string       `json:"license_number" binding:"required"`
	LicenseExpiry string       `json:"license_expiry" binding:"required"`
	Vehicle       VehicleInput `json:"vehicle" binding:"required"`
}

type VehicleInput struct {
	Type        string `json:"type" binding:"required,oneof=motorcycle car car_xl"`
	PlateNumber string `json:"plate_number" binding:"required"`
	Brand       string `json:"brand" binding:"required,max=100"`
	Model       string `json:"model" binding:"required,max=100"`
	Year        int    `json:"year" binding:"required,min=2010"`
	Color       string `json:"color" binding:"required,max=50"`
}

type DriverProfileResponse struct {
	DriverID       uuid.UUID        `json:"driver_id"`
	UserID         uuid.UUID        `json:"user_id"`
	FullName       string           `json:"full_name"`
	Phone          string           `json:"phone"`
	AvatarURL      *string          `json:"avatar_url"`
	LicenseNumber  string           `json:"license_number"`
	IsVerified     bool             `json:"is_verified"`
	IsOnline       bool             `json:"is_online"`
	Rating         float64          `json:"rating"`
	TotalRides     int              `json:"total_rides"`
	Vehicle        VehicleResponse  `json:"vehicle"`
	Location       *LocationResponse `json:"current_location,omitempty"`
	AcceptanceRate float64          `json:"acceptance_rate"`
	CreatedAt      time.Time        `json:"created_at"`
}

type VehicleResponse struct {
	Type        string `json:"type"`
	PlateNumber string `json:"plate_number"`
	Brand       string `json:"brand"`
	Model       string `json:"model"`
	Year        int    `json:"year"`
	Color       string `json:"color"`
}

type LocationResponse struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (s *DriverService) Register(ctx context.Context, userID uuid.UUID, input RegisterDriverInput) (*model.DriverProfile, error) {
	// Check if driver profile already exists
	existing, err := s.driverRepo.FindByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, ErrDriverAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Parse license expiry
	expiry, err := time.Parse("2006-01-02", input.LicenseExpiry)
	if err != nil {
		return nil, errors.New("invalid license_expiry format, use YYYY-MM-DD")
	}

	vehicleType := model.VehicleType(stringToVehicleType(input.Vehicle.Type))

	profile := &model.DriverProfile{
		ID:             uuid.New(),
		UserID:         userID,
		VehicleType:    vehicleType,
		VehicleBrand:   input.Vehicle.Brand,
		VehicleModel:   input.Vehicle.Model,
		VehicleYear:    input.Vehicle.Year,
		VehicleColor:   input.Vehicle.Color,
		LicensePlate:   input.Vehicle.PlateNumber,
		LicenseNumber:  input.LicenseNumber,
		LicenseExpiry:  expiry,
		IsOnline:       false,
		IsAvailable:    true,
		RatingAvg:      5.00,
		TotalTrips:     0,
		AcceptanceRate: 100.00,
		IsVerified:     false,
	}

	if err := s.driverRepo.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *DriverService) GetProfile(ctx context.Context, userID uuid.UUID) (*DriverProfileResponse, error) {
	profile, err := s.driverRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDriverNotFound
		}
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := &DriverProfileResponse{
		DriverID:       profile.ID,
		UserID:         profile.UserID,
		FullName:       user.FullName,
		Phone:          user.Phone,
		AvatarURL:      user.AvatarURL,
		LicenseNumber:  profile.LicenseNumber,
		IsVerified:     profile.IsVerified,
		IsOnline:       profile.IsOnline,
		Rating:         profile.RatingAvg,
		TotalRides:     profile.TotalTrips,
		AcceptanceRate: profile.AcceptanceRate,
		Vehicle: VehicleResponse{
			Type:        string(profile.VehicleType),
			PlateNumber: profile.LicensePlate,
			Brand:       profile.VehicleBrand,
			Model:       profile.VehicleModel,
			Year:        profile.VehicleYear,
			Color:       profile.VehicleColor,
		},
		CreatedAt: profile.CreatedAt,
	}

	if profile.Latitude != nil && profile.Longitude != nil {
		resp.Location = &LocationResponse{
			Lat: *profile.Latitude,
			Lng: *profile.Longitude,
		}
	}

	return resp, nil
}

type ToggleStatusInput struct {
	IsOnline bool             `json:"is_online"`
	Location *LocationInput   `json:"location"`
}

type LocationInput struct {
	Lat float64 `json:"lat" binding:"required"`
	Lng float64 `json:"lng" binding:"required"`
}

type UpdateLocationInput struct {
	Lat       float64  `json:"lat" binding:"required"`
	Lng       float64  `json:"lng" binding:"required"`
	Speed     float64  `json:"speed"`
	Heading   float64  `json:"heading"`
	Accuracy  float64  `json:"accuracy"`
}

func (s *DriverService) ToggleStatus(ctx context.Context, userID uuid.UUID, input ToggleStatusInput) (map[string]interface{}, error) {
	profile, err := s.driverRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDriverNotFound
		}
		return nil, err
	}

	if !profile.IsVerified {
		return nil, ErrDriverNotVerified
	}

	if input.IsOnline && input.Location == nil {
		return nil, errors.New("location is required when going online")
	}

	// Update status
	if err := s.driverRepo.SetOnlineStatus(ctx, profile.ID, input.IsOnline); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"is_online": input.IsOnline,
	}

	if input.IsOnline && input.Location != nil {
		// Update location in DB and Redis
		if err := s.driverRepo.UpdateLocation(ctx, profile.ID, input.Location.Lat, input.Location.Lng, 0, 0); err != nil {
			return nil, err
		}
		if err := s.geoCache.UpdateLocation(ctx, profile.ID, input.Location.Lat, input.Location.Lng); err != nil {
			return nil, err
		}
		s.geoCache.SetDriverStatus(ctx, profile.ID, "ONLINE_AVAILABLE")

		result["location"] = LocationResponse{Lat: input.Location.Lat, Lng: input.Location.Lng}
		result["went_online_at"] = time.Now().Format(time.RFC3339)
	} else {
		// Remove from geo index when going offline
		s.geoCache.RemoveDriver(ctx, profile.ID)
		s.geoCache.SetDriverStatus(ctx, profile.ID, "OFFLINE")
	}

	return result, nil
}

func (s *DriverService) UpdateLocation(ctx context.Context, userID uuid.UUID, input UpdateLocationInput) error {
	profile, err := s.driverRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDriverNotFound
		}
		return err
	}

	if !profile.IsOnline {
		return errors.New("driver must be online to update location")
	}

	// Update in PostgreSQL
	if err := s.driverRepo.UpdateLocation(ctx, profile.ID, input.Lat, input.Lng, input.Heading, input.Speed); err != nil {
		return err
	}

	// Update in Redis GeoSet
	if err := s.geoCache.UpdateLocation(ctx, profile.ID, input.Lat, input.Lng); err != nil {
		return err
	}

	return nil
}

func (s *DriverService) FindNearbyDrivers(ctx context.Context, lat, lng, radiusKm float64, vehicleType string, limit int) ([]DriverProfileResponse, error) {
	vt := model.VehicleType("")
	if vehicleType != "" {
		vt = model.VehicleType(stringToVehicleType(vehicleType))
	}

	drivers, err := s.driverRepo.FindNearbyDrivers(ctx, lat, lng, radiusKm, vt, limit)
	if err != nil {
		return nil, err
	}

	var results []DriverProfileResponse
	for _, d := range drivers {
		resp := DriverProfileResponse{
			DriverID:   d.ID,
			UserID:     d.UserID,
			IsOnline:   d.IsOnline,
			Rating:     d.RatingAvg,
			TotalRides: d.TotalTrips,
			Vehicle: VehicleResponse{
				Type:        string(d.VehicleType),
				PlateNumber: d.LicensePlate,
				Brand:       d.VehicleBrand,
				Model:       d.VehicleModel,
				Year:        d.VehicleYear,
				Color:       d.VehicleColor,
			},
		}
		if d.Latitude != nil && d.Longitude != nil {
			resp.Location = &LocationResponse{Lat: *d.Latitude, Lng: *d.Longitude}
		}
		results = append(results, resp)
	}

	return results, nil
}

func stringToVehicleType(s string) string {
	switch s {
	case "motorcycle":
		return "MOTORCYCLE"
	case "car":
		return "CAR"
	case "car_xl":
		return "CAR_XL"
	default:
		return "MOTORCYCLE"
	}
}

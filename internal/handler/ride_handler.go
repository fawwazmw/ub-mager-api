package handler

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"github.com/wardayadev/ub-mager-api/internal/service"
)

type RideHandler struct {
	rideService *service.RideService
}

func NewRideHandler(rideService *service.RideService) *RideHandler {
	return &RideHandler{rideService: rideService}
}

type EstimateResponse struct {
	Pickup        service.LocationInput  `json:"pickup"`
	Dropoff       service.LocationInput  `json:"dropoff"`
	VehicleType   string                 `json:"vehicle_type"`
	DistanceKm    float64                `json:"distance_km"`
	DurationMin   int                    `json:"duration_min"`
	FareBreakdown *service.FareBreakdown `json:"fare_breakdown"`
}

type LocationDetail struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
}

type RequestRideResponse struct {
	RideID        uuid.UUID           `json:"ride_id"`
	Status        model.RideStatus    `json:"status"`
	Pickup        LocationDetail      `json:"pickup"`
	Dropoff       LocationDetail      `json:"dropoff"`
	VehicleType   model.VehicleType   `json:"vehicle_type"`
	PaymentMethod model.PaymentMethod `json:"payment_method"`
	FareEstimate  float64             `json:"fare_estimate"`
	CreatedAt     time.Time           `json:"created_at"`
	Message       string              `json:"message"`
}

func (h *RideHandler) Estimate(c *gin.Context) {
	var input service.EstimateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if !ValidCoordinate(input.Pickup.Lat, input.Pickup.Lng) || !ValidCoordinate(input.Dropoff.Lat, input.Dropoff.Lng) {
		Error(c, http.StatusBadRequest, "INVALID_COORDINATES", "Coordinates must be valid lat/lng values")
		return
	}

	breakdown, distanceM, durationS, err := h.rideService.Estimate(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to estimate fare")
		return
	}

	Success(c, http.StatusOK, EstimateResponse{
		Pickup:        input.Pickup,
		Dropoff:       input.Dropoff,
		VehicleType:   input.VehicleType,
		DistanceKm:    math.Round(distanceM/100) / 10,
		DurationMin:   durationS / 60,
		FareBreakdown: breakdown,
	})
}

func (h *RideHandler) RequestRide(c *gin.Context) {
	userID, _ := GetUserID(c)

	var input service.RequestRideInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if !ValidCoordinate(input.Pickup.Lat, input.Pickup.Lng) || !ValidCoordinate(input.Dropoff.Lat, input.Dropoff.Lng) {
		Error(c, http.StatusBadRequest, "INVALID_COORDINATES", "Coordinates must be valid lat/lng values")
		return
	}

	ride, err := h.rideService.RequestRide(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, service.ErrActiveRideExists) {
			Error(c, http.StatusConflict, "ACTIVE_RIDE_EXISTS", "You already have an active ride")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create ride")
		return
	}

	Success(c, http.StatusCreated, RequestRideResponse{
		RideID:        ride.ID,
		Status:        ride.Status,
		Pickup:        LocationDetail{Lat: ride.PickupLat, Lng: ride.PickupLng, Address: ride.PickupAddress},
		Dropoff:       LocationDetail{Lat: ride.DropoffLat, Lng: ride.DropoffLng, Address: ride.DropoffAddress},
		VehicleType:   ride.VehicleType,
		PaymentMethod: ride.PaymentMethod,
		FareEstimate:  ride.TotalFare,
		CreatedAt:     ride.CreatedAt,
		Message:       "Searching for nearby drivers...",
	})
}

func (h *RideHandler) GetRide(c *gin.Context) {
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	ride, err := h.rideService.GetRide(c.Request.Context(), rideID)
	if err != nil {
		if errors.Is(err, service.ErrRideNotFound) {
			Error(c, http.StatusNotFound, "RIDE_NOT_FOUND", "Ride not found")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get ride")
		return
	}

	Success(c, http.StatusOK, ride)
}

func (h *RideHandler) GetActiveRide(c *gin.Context) {
	userID, _ := GetUserID(c)

	ride, err := h.rideService.GetActiveRide(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrRideNotFound) {
			Error(c, http.StatusNotFound, "NO_ACTIVE_RIDE", "No active ride found")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get active ride")
		return
	}

	Success(c, http.StatusOK, ride)
}

func (h *RideHandler) CancelRide(c *gin.Context) {
	userID, _ := GetUserID(c)
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	var input struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&input)

	err = h.rideService.CancelRide(c.Request.Context(), rideID, userID, input.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRideNotFound):
			Error(c, http.StatusNotFound, "RIDE_NOT_FOUND", "Ride not found")
		case errors.Is(err, service.ErrRideNotCancellable):
			Error(c, http.StatusBadRequest, "NOT_CANCELLABLE", "Ride cannot be cancelled in current state")
		case errors.Is(err, service.ErrUnauthorized):
			Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized to cancel this ride")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel ride")
		}
		return
	}

	SuccessMessage(c, "Ride cancelled successfully")
}

func (h *RideHandler) RateRide(c *gin.Context) {
	userID, _ := GetUserID(c)
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	var input service.RateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	err = h.rideService.RateRide(c.Request.Context(), rideID, userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRideNotFound):
			Error(c, http.StatusNotFound, "RIDE_NOT_FOUND", "Ride not found")
		case errors.Is(err, service.ErrAlreadyRated):
			Error(c, http.StatusConflict, "ALREADY_RATED", "You have already rated this ride")
		case errors.Is(err, service.ErrRideNotCompleted):
			Error(c, http.StatusBadRequest, "NOT_COMPLETED", "Ride must be completed before rating")
		case errors.Is(err, service.ErrNoDriverToRate):
			Error(c, http.StatusBadRequest, "NO_DRIVER", "No driver assigned to rate")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to rate ride")
		}
		return
	}

	SuccessMessage(c, "Rating submitted")
}

func (h *RideHandler) GetHistory(c *gin.Context) {
	userID, _ := GetUserID(c)
	userRole := GetUserRole(c)

	page, perPage := ParsePagination(c)

	rides, total, err := h.rideService.GetHistory(c.Request.Context(), userID, userRole, page, perPage)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get history")
		return
	}

	PaginatedSuccess(c, rides, page, perPage, total)
}

func (h *RideHandler) AcceptRide(c *gin.Context) {
	userID, _ := GetUserID(c)
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	err = h.rideService.AcceptRide(c.Request.Context(), rideID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRideNotFound):
			Error(c, http.StatusNotFound, "RIDE_NOT_FOUND", "Ride not found")
		case errors.Is(err, service.ErrRideNoLongerAvailable):
			Error(c, http.StatusConflict, "RIDE_UNAVAILABLE", "Ride is no longer available")
		case errors.Is(err, service.ErrDriverNotFound):
			Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver profile not found")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to accept ride")
		}
		return
	}

	SuccessMessage(c, "Ride accepted")
}

func (h *RideHandler) DriverUpdateStatus(c *gin.Context) {
	userID, _ := GetUserID(c)
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	newStatus := model.RideStatus(input.Status)
	err = h.rideService.UpdateRideStatus(c.Request.Context(), rideID, userID, newStatus)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRideNotFound):
			Error(c, http.StatusNotFound, "RIDE_NOT_FOUND", "Ride not found")
		case errors.Is(err, service.ErrUnauthorized):
			Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized")
		case errors.Is(err, service.ErrInvalidTransition):
			Error(c, http.StatusBadRequest, "INVALID_TRANSITION", "Invalid status transition")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update ride status")
		}
		return
	}

	SuccessMessage(c, "Status updated to "+input.Status)
}

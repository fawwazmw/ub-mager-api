package handler

import (
	"errors"
	"math"
	"net/http"
	"strconv"

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

func (h *RideHandler) Estimate(c *gin.Context) {
	var input service.EstimateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	breakdown, distanceM, durationS, err := h.rideService.Estimate(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to estimate fare")
		return
	}

	Success(c, http.StatusOK, gin.H{
		"pickup":       input.Pickup,
		"dropoff":      input.Dropoff,
		"vehicle_type": input.VehicleType,
		"distance_km":  math.Round(distanceM/100) / 10,
		"duration_min": durationS / 60,
		"fare_breakdown": breakdown,
	})
}

func (h *RideHandler) RequestRide(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input service.RequestRideInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	ride, err := h.rideService.RequestRide(c.Request.Context(), userID.(uuid.UUID), input)
	if err != nil {
		if errors.Is(err, service.ErrActiveRideExists) {
			Error(c, http.StatusConflict, "ACTIVE_RIDE_EXISTS", "You already have an active ride")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create ride")
		return
	}

	Success(c, http.StatusCreated, gin.H{
		"ride_id":        ride.ID,
		"status":         ride.Status,
		"pickup":         gin.H{"lat": ride.PickupLat, "lng": ride.PickupLng, "address": ride.PickupAddress},
		"dropoff":        gin.H{"lat": ride.DropoffLat, "lng": ride.DropoffLng, "address": ride.DropoffAddress},
		"vehicle_type":   ride.VehicleType,
		"payment_method": ride.PaymentMethod,
		"fare_estimate":  ride.TotalFare,
		"created_at":     ride.CreatedAt,
		"message":        "Searching for nearby drivers...",
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
	userID, _ := c.Get("user_id")

	ride, err := h.rideService.GetActiveRide(c.Request.Context(), userID.(uuid.UUID))
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
	userID, _ := c.Get("user_id")
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	var input struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&input)

	err = h.rideService.CancelRide(c.Request.Context(), rideID, userID.(uuid.UUID), input.Reason)
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

	Success(c, http.StatusOK, gin.H{"message": "Ride cancelled successfully"})
}

func (h *RideHandler) RateRide(c *gin.Context) {
	userID, _ := c.Get("user_id")
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

	err = h.rideService.RateRide(c.Request.Context(), rideID, userID.(uuid.UUID), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAlreadyRated):
			Error(c, http.StatusConflict, "ALREADY_RATED", "You have already rated this ride")
		case errors.Is(err, service.ErrRideNotCompleted):
			Error(c, http.StatusBadRequest, "NOT_COMPLETED", "Ride must be completed before rating")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to rate ride")
		}
		return
	}

	Success(c, http.StatusCreated, gin.H{"message": "Rating submitted"})
}

func (h *RideHandler) GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	rides, total, err := h.rideService.GetHistory(c.Request.Context(), userID.(uuid.UUID), userRole.(string), page, perPage)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get history")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	SuccessWithMeta(c, http.StatusOK, rides, &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// Driver-side ride actions

func (h *RideHandler) AcceptRide(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	// We need the driver profile ID, not user ID
	// For now, pass userID — the service will need to look up the driver profile
	err = h.rideService.AcceptRide(c.Request.Context(), rideID, userID.(uuid.UUID))
	if err != nil {
		Error(c, http.StatusBadRequest, "ACCEPT_FAILED", err.Error())
		return
	}

	Success(c, http.StatusOK, gin.H{"message": "Ride accepted"})
}

func (h *RideHandler) DriverUpdateStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
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
	err = h.rideService.UpdateRideStatus(c.Request.Context(), rideID, userID.(uuid.UUID), newStatus)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized")
		default:
			Error(c, http.StatusBadRequest, "STATUS_UPDATE_FAILED", err.Error())
		}
		return
	}

	Success(c, http.StatusOK, gin.H{"status": input.Status, "message": "Status updated"})
}

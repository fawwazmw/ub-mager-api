package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/service"
)

type DriverHandler struct {
	driverService *service.DriverService
}

func NewDriverHandler(driverService *service.DriverService) *DriverHandler {
	return &DriverHandler{driverService: driverService}
}

func (h *DriverHandler) Register(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input service.RegisterDriverInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	profile, err := h.driverService.Register(c.Request.Context(), userID.(uuid.UUID), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDriverAlreadyExists):
			Error(c, http.StatusConflict, "DRIVER_EXISTS", "Driver profile already exists")
		default:
			Error(c, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		}
		return
	}

	Success(c, http.StatusCreated, gin.H{
		"driver_id": profile.ID,
		"status":    "pending_verification",
		"message":   "Registration submitted. You will be notified once verified.",
	})
}

func (h *DriverHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	profile, err := h.driverService.GetProfile(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, service.ErrDriverNotFound) {
			Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver profile not found")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get profile")
		return
	}

	Success(c, http.StatusOK, profile)
}

func (h *DriverHandler) ToggleStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input service.ToggleStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	result, err := h.driverService.ToggleStatus(c.Request.Context(), userID.(uuid.UUID), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDriverNotFound):
			Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver profile not found")
		case errors.Is(err, service.ErrDriverNotVerified):
			Error(c, http.StatusForbidden, "NOT_VERIFIED", "Driver must be verified to go online")
		default:
			Error(c, http.StatusBadRequest, "STATUS_UPDATE_FAILED", err.Error())
		}
		return
	}

	Success(c, http.StatusOK, result)
}

func (h *DriverHandler) UpdateLocation(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input service.UpdateLocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	err := h.driverService.UpdateLocation(c.Request.Context(), userID.(uuid.UUID), input)
	if err != nil {
		if errors.Is(err, service.ErrDriverNotFound) {
			Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver profile not found")
			return
		}
		Error(c, http.StatusBadRequest, "LOCATION_UPDATE_FAILED", err.Error())
		return
	}

	Success(c, http.StatusOK, gin.H{"acknowledged": true})
}

func (h *DriverHandler) GetNearbyDrivers(c *gin.Context) {
	// This is used internally by the matching service, but also exposed for admin
	lat := c.DefaultQuery("lat", "0")
	lng := c.DefaultQuery("lng", "0")
	radius := c.DefaultQuery("radius", "3")
	vehicleType := c.DefaultQuery("vehicle_type", "")

	var latF, lngF, radiusF float64
	fmt.Sscanf(lat, "%f", &latF)
	fmt.Sscanf(lng, "%f", &lngF)
	fmt.Sscanf(radius, "%f", &radiusF)

	if radiusF == 0 {
		radiusF = 3
	}

	drivers, err := h.driverService.FindNearbyDrivers(c.Request.Context(), latF, lngF, radiusF, vehicleType, 20)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to find drivers")
		return
	}

	Success(c, http.StatusOK, gin.H{
		"drivers": drivers,
		"count":   len(drivers),
	})
}

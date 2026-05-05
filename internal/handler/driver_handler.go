package handler

import (
	"errors"
	"net/http"
	"strconv"

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

type RegisterDriverResponse struct {
	DriverID uuid.UUID `json:"driver_id"`
	Status   string    `json:"status"`
	Message  string    `json:"message"`
}

type NearbyDriversResponse struct {
	Drivers []service.DriverProfileResponse `json:"drivers"`
	Count   int                             `json:"count"`
}

func (h *DriverHandler) Register(c *gin.Context) {
	userID, _ := GetUserID(c)

	var input service.RegisterDriverInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	profile, err := h.driverService.Register(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDriverAlreadyExists):
			Error(c, http.StatusConflict, "DRIVER_EXISTS", "Driver profile already exists")
		case errors.Is(err, service.ErrInvalidLicenseExpiry):
			Error(c, http.StatusBadRequest, "INVALID_LICENSE_EXPIRY", "Invalid license expiry format, use YYYY-MM-DD")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register driver")
		}
		return
	}

	Success(c, http.StatusCreated, RegisterDriverResponse{
		DriverID: profile.ID,
		Status:   "pending_verification",
		Message:  "Registration submitted. You will be notified once verified.",
	})
}

func (h *DriverHandler) GetProfile(c *gin.Context) {
	userID, _ := GetUserID(c)

	profile, err := h.driverService.GetProfile(c.Request.Context(), userID)
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
	userID, _ := GetUserID(c)

	var input service.ToggleStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	result, err := h.driverService.ToggleStatus(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDriverNotFound):
			Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver profile not found")
		case errors.Is(err, service.ErrDriverNotVerified):
			Error(c, http.StatusForbidden, "NOT_VERIFIED", "Driver must be verified to go online")
		case errors.Is(err, service.ErrLocationRequired):
			Error(c, http.StatusBadRequest, "LOCATION_REQUIRED", "Location is required when going online")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update driver status")
		}
		return
	}

	Success(c, http.StatusOK, result)
}

func (h *DriverHandler) UpdateLocation(c *gin.Context) {
	userID, _ := GetUserID(c)

	var input service.UpdateLocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if !ValidCoordinate(input.Lat, input.Lng) {
		Error(c, http.StatusBadRequest, "INVALID_COORDINATES", "Coordinates must be valid lat/lng values")
		return
	}

	err := h.driverService.UpdateLocation(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDriverNotFound):
			Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver profile not found")
		case errors.Is(err, service.ErrDriverMustBeOnline):
			Error(c, http.StatusConflict, "DRIVER_OFFLINE", "Driver must be online to update location")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update location")
		}
		return
	}

	SuccessMessage(c, "Location updated")
}

func (h *DriverHandler) GetNearbyDrivers(c *gin.Context) {
	latF, _ := strconv.ParseFloat(c.DefaultQuery("lat", "0"), 64)
	lngF, _ := strconv.ParseFloat(c.DefaultQuery("lng", "0"), 64)
	radiusF, _ := strconv.ParseFloat(c.DefaultQuery("radius", "3"), 64)
	vehicleType := c.DefaultQuery("vehicle_type", "")

	if latF < -90 || latF > 90 || lngF < -180 || lngF > 180 || (latF == 0 && lngF == 0) {
		Error(c, http.StatusBadRequest, "INVALID_COORDINATES", "Valid lat and lng query parameters are required")
		return
	}

	if radiusF <= 0 || radiusF > 50 {
		radiusF = 3
	}

	drivers, err := h.driverService.FindNearbyDrivers(c.Request.Context(), latF, lngF, radiusF, vehicleType, 20)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to find drivers")
		return
	}

	Success(c, http.StatusOK, NearbyDriversResponse{
		Drivers: drivers,
		Count:   len(drivers),
	})
}

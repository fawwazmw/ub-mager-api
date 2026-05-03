package handler

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wardayadev/ub-mager-api/internal/repository"
)

type AdminHandler struct {
	analyticsRepo *repository.AnalyticsRepository
}

func NewAdminHandler(analyticsRepo *repository.AnalyticsRepository) *AdminHandler {
	return &AdminHandler{analyticsRepo: analyticsRepo}
}

// GetDashboardStats returns overview stats for the admin dashboard
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.analyticsRepo.GetDashboardStats(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get stats")
		return
	}
	Success(c, http.StatusOK, stats)
}

// GetRevenueStats returns revenue analytics
func (h *AdminHandler) GetRevenueStats(c *gin.Context) {
	period := c.DefaultQuery("period", "today")

	stats, err := h.analyticsRepo.GetRevenueStats(c.Request.Context(), period)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get revenue stats")
		return
	}
	Success(c, http.StatusOK, stats)
}

// GetDailyRevenue returns daily revenue breakdown for charts
func (h *AdminHandler) GetDailyRevenue(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	data, err := h.analyticsRepo.GetDailyRevenue(c.Request.Context(), days)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get daily revenue")
		return
	}
	Success(c, http.StatusOK, data)
}

// GetRideStats returns ride analytics
func (h *AdminHandler) GetRideStats(c *gin.Context) {
	period := c.DefaultQuery("period", "today")

	stats, err := h.analyticsRepo.GetRideStats(c.Request.Context(), period)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get ride stats")
		return
	}
	Success(c, http.StatusOK, stats)
}

// ListDrivers returns paginated driver list for admin management
func (h *AdminHandler) ListDrivers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	status := c.DefaultQuery("status", "")
	search := c.DefaultQuery("search", "")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	drivers, total, err := h.analyticsRepo.ListDrivers(c.Request.Context(), page, perPage, status, search)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list drivers")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	SuccessWithMeta(c, http.StatusOK, drivers, &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// VerifyDriver marks a driver as verified
func (h *AdminHandler) VerifyDriver(c *gin.Context) {
	driverID := c.Param("id")
	if driverID == "" {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Driver ID is required")
		return
	}

	err := h.analyticsRepo.VerifyDriver(c.Request.Context(), driverID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify driver")
		return
	}

	Success(c, http.StatusOK, gin.H{"message": "Driver verified successfully"})
}

// GetRideCountsByStatus returns ride counts grouped by status
func (h *AdminHandler) GetRideCountsByStatus(c *gin.Context) {
	counts, err := h.analyticsRepo.GetRideCountsByStatus(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get counts")
		return
	}
	Success(c, http.StatusOK, counts)
}

// CancelRide allows admin to cancel any ride
func (h *AdminHandler) CancelRide(c *gin.Context) {
	rideID := c.Param("id")
	var input struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&input)

	reason := input.Reason
	if reason == "" {
		reason = "Cancelled by admin"
	}

	err := h.analyticsRepo.AdminCancelRide(c.Request.Context(), rideID, reason)
	if err != nil {
		Error(c, http.StatusBadRequest, "CANCEL_FAILED", err.Error())
		return
	}

	Success(c, http.StatusOK, gin.H{"message": "Ride cancelled"})
}

// GetDriverDetail returns full driver profile for admin
func (h *AdminHandler) GetDriverDetail(c *gin.Context) {
	driverID := c.Param("id")
	detail, err := h.analyticsRepo.GetDriverDetail(c.Request.Context(), driverID)
	if err != nil || detail == nil {
		Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver not found")
		return
	}
	Success(c, http.StatusOK, detail)
}

// ToggleDriverOnline forces a driver online/offline from admin
func (h *AdminHandler) ToggleDriverOnline(c *gin.Context) {
	driverID := c.Param("id")
	var input struct {
		IsOnline bool `json:"is_online"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	err := h.analyticsRepo.ToggleDriverOnline(c.Request.Context(), driverID, input.IsOnline)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update driver status")
		return
	}

	status := "offline"
	if input.IsOnline {
		status = "online"
	}
	Success(c, http.StatusOK, gin.H{"message": "Driver set to " + status})
}

// GetDriverLeaderboard returns top drivers by revenue
func (h *AdminHandler) GetDriverLeaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	drivers, err := h.analyticsRepo.GetDriverLeaderboard(c.Request.Context(), limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get leaderboard")
		return
	}
	Success(c, http.StatusOK, drivers)
}

// GetRideDetail returns full ride details for admin
func (h *AdminHandler) GetRideDetail(c *gin.Context) {
	rideID := c.Param("id")
	if rideID == "" {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Ride ID is required")
		return
	}

	ride, err := h.analyticsRepo.GetRideDetail(c.Request.Context(), rideID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get ride detail")
		return
	}
	if ride == nil {
		Error(c, http.StatusNotFound, "RIDE_NOT_FOUND", "Ride not found")
		return
	}

	Success(c, http.StatusOK, ride)
}

// GetPeakHours returns ride distribution by hour of day
func (h *AdminHandler) GetPeakHours(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	data, err := h.analyticsRepo.GetPeakHours(c.Request.Context(), days)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get peak hours")
		return
	}
	Success(c, http.StatusOK, data)
}

// BulkCancelStuckRides cancels all rides stuck in SEARCHING for 30+ minutes
func (h *AdminHandler) BulkCancelStuckRides(c *gin.Context) {
	affected, err := h.analyticsRepo.BulkCancelStuckRides(c.Request.Context(), "Auto-cancelled: no driver found within 30 minutes")
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel stuck rides")
		return
	}
	Success(c, http.StatusOK, map[string]interface{}{
		"cancelled": affected,
		"message":   "Stuck rides cancelled",
	})
}

// GetDriverRides returns recent rides for a specific driver
func (h *AdminHandler) GetDriverRides(c *gin.Context) {
	driverID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	rides, err := h.analyticsRepo.GetDriverRides(c.Request.Context(), driverID, limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get driver rides")
		return
	}
	Success(c, http.StatusOK, rides)
}

// GetRecentActivity returns recent platform activity for the admin feed
func (h *AdminHandler) GetRecentActivity(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	activities, err := h.analyticsRepo.GetRecentActivity(c.Request.Context(), limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get activity")
		return
	}
	Success(c, http.StatusOK, activities)
}

// ListRides returns paginated ride list for admin (ALL rides, not user-scoped)
func (h *AdminHandler) ListRides(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	status := c.DefaultQuery("status", "")
	search := c.DefaultQuery("search", "")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	rides, total, err := h.analyticsRepo.ListRides(c.Request.Context(), page, perPage, status, search)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list rides")
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

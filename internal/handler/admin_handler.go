package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wardayadev/ub-mager-api/internal/repository"
)

type AdminHandler struct {
	analyticsRepo AnalyticsRepo
}

func NewAdminHandler(analyticsRepo AnalyticsRepo) *AdminHandler {
	return &AdminHandler{analyticsRepo: analyticsRepo}
}

type BulkCancelResponse struct {
	Cancelled int64  `json:"cancelled"`
	Message   string `json:"message"`
}

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.analyticsRepo.GetDashboardStats(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get stats")
		return
	}
	Success(c, http.StatusOK, stats)
}

func (h *AdminHandler) GetRevenueStats(c *gin.Context) {
	period := c.DefaultQuery("period", "today")

	stats, err := h.analyticsRepo.GetRevenueStats(c.Request.Context(), period)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get revenue stats")
		return
	}
	Success(c, http.StatusOK, stats)
}

func (h *AdminHandler) GetDailyRevenue(c *gin.Context) {
	days := ParseIntQuery(c, "days", 7, 1, 90)

	data, err := h.analyticsRepo.GetDailyRevenue(c.Request.Context(), days)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get daily revenue")
		return
	}
	Success(c, http.StatusOK, data)
}

func (h *AdminHandler) GetRideStats(c *gin.Context) {
	period := c.DefaultQuery("period", "today")

	stats, err := h.analyticsRepo.GetRideStats(c.Request.Context(), period)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get ride stats")
		return
	}
	Success(c, http.StatusOK, stats)
}

func (h *AdminHandler) ListDrivers(c *gin.Context) {
	page, perPage := ParsePagination(c)
	status := c.DefaultQuery("status", "")
	search := c.DefaultQuery("search", "")

	drivers, total, err := h.analyticsRepo.ListDrivers(c.Request.Context(), page, perPage, status, search)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list drivers")
		return
	}

	PaginatedSuccess(c, drivers, page, perPage, total)
}

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

	SuccessMessage(c, "Driver verified successfully")
}

func (h *AdminHandler) GetRideCountsByStatus(c *gin.Context) {
	counts, err := h.analyticsRepo.GetRideCountsByStatus(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get counts")
		return
	}
	Success(c, http.StatusOK, counts)
}

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
		if errors.Is(err, repository.ErrRideNotFoundOrTerminal) {
			Error(c, http.StatusNotFound, "RIDE_NOT_FOUND", "Ride not found or already completed/cancelled")
		} else {
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel ride")
		}
		return
	}

	SuccessMessage(c, "Ride cancelled")
}

func (h *AdminHandler) GetDriverDetail(c *gin.Context) {
	driverID := c.Param("id")
	detail, err := h.analyticsRepo.GetDriverDetail(c.Request.Context(), driverID)
	if err != nil || detail == nil {
		Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver not found")
		return
	}
	Success(c, http.StatusOK, detail)
}

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
	SuccessMessage(c, "Driver set to "+status)
}

func (h *AdminHandler) GetDriverLeaderboard(c *gin.Context) {
	limit := ParseIntQuery(c, "limit", 10, 1, 100)

	drivers, err := h.analyticsRepo.GetDriverLeaderboard(c.Request.Context(), limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get leaderboard")
		return
	}
	Success(c, http.StatusOK, drivers)
}

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

func (h *AdminHandler) GetPeakHours(c *gin.Context) {
	days := ParseIntQuery(c, "days", 7, 1, 90)
	data, err := h.analyticsRepo.GetPeakHours(c.Request.Context(), days)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get peak hours")
		return
	}
	Success(c, http.StatusOK, data)
}

func (h *AdminHandler) BulkCancelStuckRides(c *gin.Context) {
	affected, err := h.analyticsRepo.BulkCancelStuckRides(c.Request.Context(), "Auto-cancelled: no driver found within 30 minutes")
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel stuck rides")
		return
	}
	Success(c, http.StatusOK, BulkCancelResponse{
		Cancelled: affected,
		Message:   "Stuck rides cancelled",
	})
}

func (h *AdminHandler) GetDriverRides(c *gin.Context) {
	driverID := c.Param("id")
	limit := ParseIntQuery(c, "limit", 5, 1, 50)

	rides, err := h.analyticsRepo.GetDriverRides(c.Request.Context(), driverID, limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get driver rides")
		return
	}
	Success(c, http.StatusOK, rides)
}

func (h *AdminHandler) GetRecentActivity(c *gin.Context) {
	limit := ParseIntQuery(c, "limit", 20, 1, 100)

	activities, err := h.analyticsRepo.GetRecentActivity(c.Request.Context(), limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get activity")
		return
	}
	Success(c, http.StatusOK, activities)
}

func (h *AdminHandler) ListRides(c *gin.Context) {
	page, perPage := ParsePagination(c)
	status := c.DefaultQuery("status", "")
	search := c.DefaultQuery("search", "")

	rides, total, err := h.analyticsRepo.ListRides(c.Request.Context(), page, perPage, status, search)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list rides")
		return
	}

	PaginatedSuccess(c, rides, page, perPage, total)
}

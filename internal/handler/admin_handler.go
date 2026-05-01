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

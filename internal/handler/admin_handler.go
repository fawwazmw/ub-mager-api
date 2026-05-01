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
	status := c.DefaultQuery("status", "") // online, verified, pending, or empty for all

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	drivers, total, err := h.analyticsRepo.ListDrivers(c.Request.Context(), page, perPage, status)
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

// ListRides returns paginated ride list for admin
func (h *AdminHandler) ListRides(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// Reuse ride repo through analytics repo
	// For now, simple query
	Success(c, http.StatusOK, gin.H{
		"message": "Use GET /rides/history with admin token for ride listing",
	})
}

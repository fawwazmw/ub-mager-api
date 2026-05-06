package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wardayadev/ub-mager-api/internal/repository"
	"github.com/wardayadev/ub-mager-api/internal/service"
	"gorm.io/gorm"
)

type EarningsHandler struct {
	earningsRepo *repository.EarningsRepository
	driverRepo   service.DriverRepo
}

func NewEarningsHandler(earningsRepo *repository.EarningsRepository, driverRepo service.DriverRepo) *EarningsHandler {
	return &EarningsHandler{earningsRepo: earningsRepo, driverRepo: driverRepo}
}

func (h *EarningsHandler) GetSummary(c *gin.Context) {
	userID, _ := GetUserID(c)
	period := c.DefaultQuery("period", "today")

	profile, err := h.driverRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver profile not found")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get profile")
		return
	}

	summary, err := h.earningsRepo.GetSummary(c.Request.Context(), profile.ID, period)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get earnings")
		return
	}

	Success(c, http.StatusOK, summary)
}

func (h *EarningsHandler) GetDaily(c *gin.Context) {
	userID, _ := GetUserID(c)
	days := ParseIntQuery(c, "days", 7, 1, 90)

	profile, err := h.driverRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver profile not found")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get profile")
		return
	}

	daily, err := h.earningsRepo.GetDaily(c.Request.Context(), profile.ID, days)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get daily earnings")
		return
	}

	Success(c, http.StatusOK, daily)
}

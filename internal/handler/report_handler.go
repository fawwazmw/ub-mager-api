package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"github.com/wardayadev/ub-mager-api/internal/repository"
)

type ReportHandler struct {
	reportRepo *repository.ReportRepository
}

func NewReportHandler(reportRepo *repository.ReportRepository) *ReportHandler {
	return &ReportHandler{reportRepo: reportRepo}
}

type SubmitReportInput struct {
	ReportedUserID string `json:"reported_user_id" binding:"required,uuid"`
	RideID         string `json:"ride_id" binding:"omitempty,uuid"`
	Category       string `json:"category" binding:"required,oneof=RUDE_BEHAVIOR SAFETY_CONCERN FRAUD SPAM OTHER"`
	Description    string `json:"description" binding:"required,min=10,max=500"`
}

func (h *ReportHandler) Submit(c *gin.Context) {
	userID, _ := GetUserID(c)

	var input SubmitReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	reportedID, _ := uuid.Parse(input.ReportedUserID)
	if reportedID == userID {
		Error(c, http.StatusBadRequest, "SELF_REPORT", "Cannot report yourself")
		return
	}

	report := &model.Report{
		ID:             uuid.New(),
		ReporterID:     userID,
		ReportedUserID: reportedID,
		Category:       model.ReportCategory(input.Category),
		Description:    input.Description,
		Status:         model.ReportStatusPending,
	}

	if input.RideID != "" {
		rideID, _ := uuid.Parse(input.RideID)
		report.RideID = &rideID
	}

	if err := h.reportRepo.Create(c.Request.Context(), report); err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to submit report")
		return
	}

	SuccessMessage(c, "Report submitted successfully")
}

func (h *ReportHandler) List(c *gin.Context) {
	page, perPage := ParsePagination(c)
	status := c.DefaultQuery("status", "")

	reports, total, err := h.reportRepo.List(c.Request.Context(), page, perPage, status)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list reports")
		return
	}

	PaginatedSuccess(c, reports, page, perPage, total)
}

type ResolveReportInput struct {
	Status    string `json:"status" binding:"required,oneof=REVIEWED RESOLVED"`
	AdminNote string `json:"admin_note" binding:"max=500"`
}

func (h *ReportHandler) Resolve(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid report ID")
		return
	}

	var input ResolveReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.reportRepo.Resolve(c.Request.Context(), reportID, input.AdminNote, model.ReportStatus(input.Status)); err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to resolve report")
		return
	}

	SuccessMessage(c, "Report updated")
}

func (h *ReportHandler) CountPending(c *gin.Context) {
	count, err := h.reportRepo.CountPending(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to count reports")
		return
	}

	Success(c, http.StatusOK, map[string]int64{"pending": count})
}

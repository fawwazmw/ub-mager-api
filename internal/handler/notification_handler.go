package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/repository"
)

type NotificationHandler struct {
	notifRepo *repository.NotificationRepository
}

func NewNotificationHandler(notifRepo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{notifRepo: notifRepo}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, _ := GetUserID(c)
	page, perPage := ParsePagination(c)

	notifs, total, err := h.notifRepo.FindByUser(c.Request.Context(), userID, page, perPage)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get notifications")
		return
	}

	PaginatedSuccess(c, notifs, page, perPage, total)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, _ := GetUserID(c)
	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid notification ID")
		return
	}

	if err := h.notifRepo.MarkRead(c.Request.Context(), userID, notifID); err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark as read")
		return
	}

	SuccessMessage(c, "Marked as read")
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, _ := GetUserID(c)

	if err := h.notifRepo.MarkAllRead(c.Request.Context(), userID); err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark all as read")
		return
	}

	SuccessMessage(c, "All notifications marked as read")
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID, _ := GetUserID(c)

	count, err := h.notifRepo.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get count")
		return
	}

	Success(c, http.StatusOK, map[string]int64{"unread": count})
}

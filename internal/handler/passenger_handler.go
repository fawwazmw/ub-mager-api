package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/service"
)

type UserHandler struct {
	authService *service.AuthService
}

func NewUserHandler(authService *service.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.authService.GetUserByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	Success(c, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input service.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID.(uuid.UUID), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			Error(c, http.StatusConflict, "EMAIL_EXISTS", "Email already in use")
		case errors.Is(err, service.ErrUserNotFound):
			Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update profile")
		}
		return
	}

	Success(c, http.StatusOK, user)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input service.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), userID.(uuid.UUID), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Current password is incorrect")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to change password")
		return
	}

	Success(c, http.StatusOK, gin.H{"message": "Password changed successfully"})
}

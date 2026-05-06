package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/wardayadev/ub-mager-api/internal/model"
)

type PasswordResetHandler struct {
	db *gorm.DB
}

func NewPasswordResetHandler(db *gorm.DB) *PasswordResetHandler {
	return &PasswordResetHandler{db: db}
}

type RequestResetInput struct {
	Phone string `json:"phone" binding:"required"`
}

func (h *PasswordResetHandler) RequestReset(c *gin.Context) {
	var input RequestResetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	var user model.User
	if err := h.db.WithContext(c.Request.Context()).Where("phone = ?", input.Phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			SuccessMessage(c, "If the account exists, a reset token has been generated")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process request")
		return
	}

	token := generateResetToken()
	reset := &model.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	h.db.WithContext(c.Request.Context()).
		Where("user_id = ? AND used_at IS NULL", user.ID).
		Delete(&model.PasswordReset{})

	if err := h.db.WithContext(c.Request.Context()).Create(reset).Error; err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create reset token")
		return
	}

	Success(c, http.StatusOK, map[string]string{
		"message":     "Reset token generated (valid for 15 minutes)",
		"reset_token": token,
	})
}

type ResetPasswordInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	var reset model.PasswordReset
	err := h.db.WithContext(c.Request.Context()).
		Where("token = ? AND used_at IS NULL AND expires_at > ?", input.Token, time.Now()).
		First(&reset).Error

	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_TOKEN", "Reset token is invalid or expired")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process password")
		return
	}

	now := time.Now()
	h.db.WithContext(c.Request.Context()).
		Model(&model.User{}).
		Where("id = ?", reset.UserID).
		Update("password_hash", string(hash))

	h.db.WithContext(c.Request.Context()).
		Model(&reset).
		Update("used_at", now)

	SuccessMessage(c, "Password reset successfully")
}

func (h *PasswordResetHandler) AdminResetForUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	token := generateResetToken()
	reset := &model.PasswordReset{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	h.db.WithContext(c.Request.Context()).
		Where("user_id = ? AND used_at IS NULL", userID).
		Delete(&model.PasswordReset{})

	if err := h.db.WithContext(c.Request.Context()).Create(reset).Error; err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate reset token")
		return
	}

	Success(c, http.StatusOK, map[string]string{
		"reset_token": token,
		"expires_in":  "30 minutes",
	})
}

func generateResetToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wardayadev/ub-mager-api/internal/service"
)

type AuthHandler struct {
	authService     *service.AuthService
	secureCookie    bool
	refreshCookieTTL int
}

func NewAuthHandler(authService *service.AuthService, isProduction bool, refreshTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		secureCookie:    isProduction,
		refreshCookieTTL: int(refreshTTL.Seconds()),
	}
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", token, maxAge, "/", "", h.secureCookie, true)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input service.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, refreshToken, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPhoneAlreadyExists):
			Error(c, http.StatusConflict, "PHONE_EXISTS", "Phone number already registered")
		case errors.Is(err, service.ErrEmailAlreadyExists):
			Error(c, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user")
		}
		return
	}

	h.setRefreshCookie(c, refreshToken, h.refreshCookieTTL)

	Success(c, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, refreshToken, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Phone number or password is incorrect")
			return
		}
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to login")
		return
	}

	h.setRefreshCookie(c, refreshToken, h.refreshCookieTTL)

	Success(c, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "No refresh token provided")
		return
	}

	resp, newRefreshToken, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		Error(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "Refresh token is invalid or expired")
		return
	}

	h.setRefreshCookie(c, newRefreshToken, h.refreshCookieTTL)

	Success(c, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.setRefreshCookie(c, "", -1)

	SuccessMessage(c, "Successfully logged out")
}

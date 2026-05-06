package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wardayadev/ub-mager-api/internal/service"

	jwtpkg "github.com/wardayadev/ub-mager-api/internal/pkg/jwt"
	"github.com/wardayadev/ub-mager-api/internal/service/testutil"
)

func setupAuthRouter() (*gin.Engine, *testutil.MockUserRepo) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	userRepo := new(testutil.MockUserRepo)
	jwtService := jwtpkg.NewJWTService("test-secret-key-32chars-minimum!", 15, 10080)
	authService := service.NewAuthService(userRepo, jwtService)
	authHandler := NewAuthHandler(authService, false, 7*24*time.Hour)

	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	return r, userRepo
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	r, _ := setupAuthRouter()

	body := `{"phone": ""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
}

func TestRegisterHandler_InvalidEmail(t *testing.T) {
	r, _ := setupAuthRouter()

	body := `{"phone":"+6281234567890","email":"test@gmail.com","password":"password123","full_name":"Test User","role":"passenger"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "INVALID_EMAIL_DOMAIN", resp.Error.Code)
}

func TestLoginHandler_InvalidBody(t *testing.T) {
	r, _ := setupAuthRouter()

	body := `{"phone": ""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
}

package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTaskCreate_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	taskHandler := NewTaskHandler(nil, nil)
	r.POST("/tasks", func(c *gin.Context) {
		c.Set("user_id", nil)
		taskHandler.Create(c)
	})

	body := `{"title": ""}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
}

func TestTaskCreate_InvalidCoordinates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	taskHandler := NewTaskHandler(nil, nil)
	r.POST("/tasks", func(c *gin.Context) {
		c.Set("user_id", nil)
		taskHandler.Create(c)
	})

	body := `{
		"category": "JASTIP_MAKANAN",
		"title": "Test task title here",
		"description": "This is a test description for the task",
		"fee": 5000,
		"pickup_lat": 999,
		"pickup_lng": 112.6,
		"pickup_address": "Test",
		"delivery_lat": -7.9,
		"delivery_lng": 112.6,
		"delivery_address": "Test2"
	}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "INVALID_COORDINATES", resp.Error.Code)
}

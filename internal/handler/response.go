package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Response struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Meta    *Meta      `json:"meta,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, Response{Success: true, Data: data})
}

func SuccessWithMeta(c *gin.Context, status int, data any, meta *Meta) {
	c.JSON(status, Response{Success: true, Data: data, Meta: meta})
}

func PaginatedSuccess(c *gin.Context, data any, page, perPage int, total int64) {
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	SuccessWithMeta(c, http.StatusOK, data, &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

type messageData struct {
	Message string `json:"message"`
}

func SuccessMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    messageData{Message: message},
	})
}

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message},
	})
	c.Abort()
}

func ValidationError(c *gin.Context, details []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: details,
		},
	})
	c.Abort()
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

func GetUserRole(c *gin.Context) string {
	val, _ := c.Get("user_role")
	role, _ := val.(string)
	return role
}

func ParsePagination(c *gin.Context) (page, perPage int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return
}

func ParseIntQuery(c *gin.Context, key string, defaultVal, min, max int) int {
	val, _ := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(defaultVal)))
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func ValidCoordinate(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

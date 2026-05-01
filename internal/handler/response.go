package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
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

func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{Success: true, Data: data})
}

func SuccessWithMeta(c *gin.Context, status int, data interface{}, meta *Meta) {
	c.JSON(status, Response{Success: true, Data: data, Meta: meta})
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

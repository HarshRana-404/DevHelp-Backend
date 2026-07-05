// Package utils provides shared helper utilities used across DevHelp handlers and services.
package utils

import (
	"net/http"

	"devhelp/internal/services/dto"

	"github.com/gin-gonic/gin"
)

// SuccessResponse writes a 200 OK JSON envelope to the Gin context.
func SuccessResponse(c *gin.Context, data interface{}) {
	requestID, _ := c.Get("request_id")
	id, _ := requestID.(string)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		RequestID: id,
		Data:      data,
	})
}

// ErrorResponse writes an error JSON envelope with the provided HTTP status code.
func ErrorResponse(c *gin.Context, status int, code, message, details string) {
	requestID, _ := c.Get("request_id")
	id, _ := requestID.(string)

	c.JSON(status, dto.APIResponse{
		Success:   false,
		RequestID: id,
		Error: &dto.APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// BadRequest is a shorthand for 400 Bad Request error responses.
func BadRequest(c *gin.Context, message, details string) {
	ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

// InternalError is a shorthand for 500 Internal Server Error responses.
func InternalError(c *gin.Context, details string) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred", details)
}

// TooManyRequests is a shorthand for 429 Too Many Requests responses.
func TooManyRequests(c *gin.Context) {
	ErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMITED", "Rate limit exceeded. Try again later.", "100 requests per minute allowed per IP")
}

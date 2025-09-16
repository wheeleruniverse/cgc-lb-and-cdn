package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Data    interface{}       `json:"data"`
	Message string            `json:"message,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// RespondWithError sends an error response
func RespondWithError(c *gin.Context, statusCode int, message, code string, details map[string]string) {
	c.JSON(statusCode, ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}

// RespondWithSuccess sends a success response
func RespondWithSuccess(c *gin.Context, data interface{}, message string, meta map[string]string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data:    data,
		Message: message,
		Meta:    meta,
	})
}

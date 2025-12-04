package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standardized API response structure
type APIResponse struct {
	Status string         `json:"status"`
	Data   interface{}    `json:"data,omitempty"`
	Error  *ErrorResponse `json:"error,omitempty"`
}

// ErrorResponse represents the error structure in API responses
type ErrorResponse struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// SuccessResponse sends a successful API response
func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Status: "success",
		Data:   data,
		Error:  nil,
	})
}

// SendErrorResponse sends an error API response
func SendErrorResponse(c *gin.Context, statusCode int, errorCode string, errorMessage string) {
	c.JSON(statusCode, APIResponse{
		Status: "error",
		Data:   nil,
		Error: &ErrorResponse{
			ErrorCode:    errorCode,
			ErrorMessage: errorMessage,
		},
	})
}

// BadRequestResponse sends a 400 Bad Request error response
func BadRequestResponse(c *gin.Context, errorCode string, errorMessage string) {
	SendErrorResponse(c, http.StatusBadRequest, errorCode, errorMessage)
}

// UnauthorizedResponse sends a 401 Unauthorized error response
func UnauthorizedResponse(c *gin.Context, errorCode string, errorMessage string) {
	SendErrorResponse(c, http.StatusUnauthorized, errorCode, errorMessage)
}

// ForbiddenResponse sends a 403 Forbidden error response
func ForbiddenResponse(c *gin.Context, errorCode string, errorMessage string) {
	SendErrorResponse(c, http.StatusForbidden, errorCode, errorMessage)
}

// NotFoundResponse sends a 404 Not Found error response
func NotFoundResponse(c *gin.Context, errorCode string, errorMessage string) {
	SendErrorResponse(c, http.StatusNotFound, errorCode, errorMessage)
}

// InternalServerErrorResponse sends a 500 Internal Server Error response
func InternalServerErrorResponse(c *gin.Context, errorCode string, errorMessage string) {
	SendErrorResponse(c, http.StatusInternalServerError, errorCode, errorMessage)
}

// ValidationErrorResponse sends a 400 Bad Request for validation errors
func ValidationErrorResponse(c *gin.Context, errorMessage string) {
	BadRequestResponse(c, "VALIDATION_ERROR", errorMessage)
}

// getErrorCodeFromMessage maps error messages to standardized error codes
func getErrorCodeFromMessage(errorMessage string) string {
	errorMessageLower := strings.ToLower(errorMessage)

	switch {
	case strings.Contains(errorMessageLower, "access denied"):
		if strings.Contains(errorMessageLower, "category") {
			return "CATEGORY_ACCESS_DENIED"
		}
		if strings.Contains(errorMessageLower, "currency") {
			return "CURRENCY_ACCESS_DENIED"
		}
		return "ACCESS_DENIED"
	case strings.Contains(errorMessageLower, "not found"):
		if strings.Contains(errorMessageLower, "transaction") {
			return "TRANSACTION_NOT_FOUND"
		}
		if strings.Contains(errorMessageLower, "category") {
			return "CATEGORY_NOT_FOUND"
		}
		if strings.Contains(errorMessageLower, "currency") {
			return "CURRENCY_NOT_FOUND"
		}
		if strings.Contains(errorMessageLower, "budget") {
			return "BUDGET_NOT_FOUND"
		}
		return "RESOURCE_NOT_FOUND"
	case strings.Contains(errorMessageLower, "invalid credentials"):
		return "INVALID_CREDENTIALS"
	case strings.Contains(errorMessageLower, "invalid email"):
		return "INVALID_EMAIL"
	case strings.Contains(errorMessageLower, "invalid token"):
		return "INVALID_TOKEN"
	case strings.Contains(errorMessageLower, "transaction type mismatch"):
		return "TRANSACTION_TYPE_MISMATCH"
	case strings.Contains(errorMessageLower, "invalid"):
		return "INVALID_REQUEST"
	default:
		return "UNKNOWN_ERROR"
	}
}

// HandleError handles errors and sends appropriate error response
func HandleError(c *gin.Context, err error, defaultStatusCode int) {
	errorMessage := err.Error()
	errorCode := getErrorCodeFromMessage(errorMessage)

	// Determine status code based on error code
	statusCode := defaultStatusCode
	switch errorCode {
	case "INVALID_CREDENTIALS", "INVALID_TOKEN":
		statusCode = http.StatusUnauthorized
	case "ACCESS_DENIED", "CATEGORY_ACCESS_DENIED", "CURRENCY_ACCESS_DENIED":
		statusCode = http.StatusForbidden
	case "TRANSACTION_NOT_FOUND", "CATEGORY_NOT_FOUND", "CURRENCY_NOT_FOUND", "BUDGET_NOT_FOUND", "RESOURCE_NOT_FOUND":
		statusCode = http.StatusNotFound
	case "VALIDATION_ERROR", "INVALID_REQUEST", "INVALID_EMAIL", "TRANSACTION_TYPE_MISMATCH":
		statusCode = http.StatusBadRequest
	default:
		statusCode = defaultStatusCode
	}

	SendErrorResponse(c, statusCode, errorCode, errorMessage)
}

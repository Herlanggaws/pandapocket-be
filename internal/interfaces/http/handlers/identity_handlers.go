package handlers

import (
	"net/http"
	"panda-pocket/internal/application/identity"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// IdentityHandlers handles identity-related HTTP requests
type IdentityHandlers struct {
	registerUserUseCase   *identity.RegisterUserUseCase
	loginUserUseCase      *identity.LoginUserUseCase
	getUsersUseCase       *identity.GetUsersUseCase
	forgotPasswordUseCase *identity.ForgotPasswordUseCase
}

// NewIdentityHandlers creates a new identity handlers instance
func NewIdentityHandlers(
	registerUserUseCase *identity.RegisterUserUseCase,
	loginUserUseCase *identity.LoginUserUseCase,
	getUsersUseCase *identity.GetUsersUseCase,
	forgotPasswordUseCase *identity.ForgotPasswordUseCase,
) *IdentityHandlers {
	return &IdentityHandlers{
		registerUserUseCase:   registerUserUseCase,
		loginUserUseCase:      loginUserUseCase,
		getUsersUseCase:       getUsersUseCase,
		forgotPasswordUseCase: forgotPasswordUseCase,
	}
}

// formatValidationError converts technical validation errors to user-friendly messages
func formatValidationError(err error) string {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, e := range validationErrs {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				messages = append(messages, field+" is required")
			case "email":
				messages = append(messages, "Please provide a valid email address")
			case "min":
				messages = append(messages, field+" must be at least "+e.Param()+" characters")
			case "max":
				messages = append(messages, field+" must not exceed "+e.Param()+" characters")
			default:
				messages = append(messages, field+" is invalid")
			}
		}
		if len(messages) > 0 {
			return strings.Join(messages, ", ")
		}
	}
	return "Invalid request. Please check your input and try again"
}

// Register handles user registration
func (h *IdentityHandlers) Register(c *gin.Context) {
	var req identity.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, formatValidationError(err))
		return
	}

	response, err := h.registerUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusCreated, gin.H{
		"token": response.Token,
		"user": gin.H{
			"id":    response.UserID,
			"email": response.Email,
		},
	})
}

// Login handles user login
func (h *IdentityHandlers) Login(c *gin.Context) {
	var req identity.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, formatValidationError(err))
		return
	}

	response, err := h.loginUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		HandleError(c, err, http.StatusUnauthorized)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"token": response.Token,
		"user": gin.H{
			"id":    response.UserID,
			"email": response.Email,
		},
	})
}

// GetUsers handles getting all users
func (h *IdentityHandlers) GetUsers(c *gin.Context) {
	response, err := h.getUsersUseCase.Execute(c.Request.Context())
	if err != nil {
		HandleError(c, err, http.StatusInternalServerError)
		return
	}

	SuccessResponse(c, http.StatusOK, response)
}

// Logout handles user logout
func (h *IdentityHandlers) Logout(c *gin.Context) {
	// In a real application, you might want to blacklist the token
	// For this implementation, we'll just return a success response
	SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// ForgotPassword handles forgot password requests
func (h *IdentityHandlers) ForgotPassword(c *gin.Context) {
	var req identity.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, formatValidationError(err))
		return
	}

	response, err := h.forgotPasswordUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "email not found" {
			HandleError(c, err, http.StatusBadRequest) // Or Not Found, but keeping consistent with "IF email was found"
			return
		}
		HandleError(c, err, http.StatusInternalServerError)
		return
	}

	SuccessResponse(c, http.StatusOK, response)
}

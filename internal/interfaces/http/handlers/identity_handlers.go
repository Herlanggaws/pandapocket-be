package handlers

import (
	"net/http"
	"panda-pocket/internal/application/identity"

	"github.com/gin-gonic/gin"
)

// IdentityHandlers handles identity-related HTTP requests
type IdentityHandlers struct {
	registerUserUseCase *identity.RegisterUserUseCase
	loginUserUseCase    *identity.LoginUserUseCase
}

// NewIdentityHandlers creates a new identity handlers instance
func NewIdentityHandlers(
	registerUserUseCase *identity.RegisterUserUseCase,
	loginUserUseCase *identity.LoginUserUseCase,
) *IdentityHandlers {
	return &IdentityHandlers{
		registerUserUseCase: registerUserUseCase,
		loginUserUseCase:    loginUserUseCase,
	}
}

// Register handles user registration
func (h *IdentityHandlers) Register(c *gin.Context) {
	var req identity.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.registerUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   response.Token,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.loginUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   response.Token,
		"user": gin.H{
			"id":    response.UserID,
			"email": response.Email,
		},
	})
}

// Logout handles user logout
func (h *IdentityHandlers) Logout(c *gin.Context) {
	// In a real application, you might want to blacklist the token
	// For this implementation, we'll just return a success response
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

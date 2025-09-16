package middleware

import (
	"net/http"
	"panda-pocket/internal/application/identity"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	tokenService identity.TokenService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenService identity.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// RequireAuth is the middleware function that validates JWT tokens
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := m.tokenService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

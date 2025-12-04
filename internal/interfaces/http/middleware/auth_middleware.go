package middleware

import (
	"panda-pocket/internal/application/identity"
	"panda-pocket/internal/interfaces/http/handlers"

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
			handlers.UnauthorizedResponse(c, "AUTHORIZATION_HEADER_REQUIRED", "Authorization header required")
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := m.tokenService.ValidateToken(tokenString)
		if err != nil {
			handlers.UnauthorizedResponse(c, "INVALID_TOKEN", "Invalid token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole is a middleware that checks if the user has the required role
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			handlers.UnauthorizedResponse(c, "USER_ROLE_NOT_FOUND", "User role not found")
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			handlers.InternalServerErrorResponse(c, "INVALID_ROLE_TYPE", "Invalid role type")
			c.Abort()
			return
		}

		// Check if user has required role or higher
		if !hasRequiredRole(role, requiredRole) {
			handlers.ForbiddenResponse(c, "INSUFFICIENT_PERMISSIONS", "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasRequiredRole checks if the user's role meets the required role
// Role hierarchy: super_admin > admin > user
func hasRequiredRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		"user":        1,
		"admin":       2,
		"super_admin": 3,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

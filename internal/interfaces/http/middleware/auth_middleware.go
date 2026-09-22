package middleware

import (
	"context"
	"panda-pocket/internal/application/identity"
	domainIdentity "panda-pocket/internal/domain/identity"
	"panda-pocket/internal/interfaces/http/handlers"

	"github.com/gin-gonic/gin"
)

type activeUserChecker interface {
	ExistsActive(ctx context.Context, id domainIdentity.UserID) (bool, error)
}

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	tokenService identity.TokenService
	userChecker  activeUserChecker
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenService identity.TokenService, userChecker activeUserChecker) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		userChecker:  userChecker,
	}
}

// RequireAuth is the middleware function that validates JWT tokens
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.authenticate(c, true) {
			return
		}
		c.Next()
	}
}

// OptionalAuth attaches user claims when a valid Bearer token is present.
// No Authorization header → anonymous (public catalog).
// Authorization present but invalid/expired/deleted → 401 (same as RequireAuth),
// so logged-in clients still clear session instead of silently getting a public view.
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.Next()
			return
		}
		if !m.authenticate(c, true) {
			return
		}
		c.Next()
	}
}

// authenticate validates Authorization when present. When requireAuth is true,
// missing/invalid tokens abort with Unauthorized. Returns whether user context was set.
func (m *AuthMiddleware) authenticate(c *gin.Context, requireAuth bool) bool {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		if requireAuth {
			handlers.UnauthorizedResponse(c, "AUTHORIZATION_HEADER_REQUIRED", "Authorization header required")
			c.Abort()
		}
		return false
	}

	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims, err := m.tokenService.ValidateToken(tokenString)
	if err != nil {
		m.tokenService.CleanupExpiredToken(c.Request.Context(), tokenString)
		if requireAuth {
			handlers.UnauthorizedResponse(c, "INVALID_TOKEN", "Invalid token")
			c.Abort()
		}
		return false
	}

	if claims.UserID <= 0 {
		if requireAuth {
			handlers.UnauthorizedResponse(c, "INVALID_TOKEN", "Invalid token")
			c.Abort()
		}
		return false
	}

	if m.userChecker != nil {
		active, err := m.userChecker.ExistsActive(c.Request.Context(), domainIdentity.NewUserID(claims.UserID))
		if err != nil {
			if requireAuth {
				handlers.InternalServerErrorResponse(c, "USER_LOOKUP_FAILED", "Failed to verify user")
				c.Abort()
			}
			return false
		}
		if !active {
			if requireAuth {
				handlers.UnauthorizedResponse(c, "ACCOUNT_DELETED", "Account has been deleted")
				c.Abort()
			}
			return false
		}
	}

	c.Set("user_id", claims.UserID)
	c.Set("email", claims.Email)
	c.Set("role", claims.Role)
	return true
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

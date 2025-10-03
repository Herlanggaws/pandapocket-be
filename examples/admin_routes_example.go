package main

import (
	"net/http"
	"panda-pocket/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

// Example of how to set up admin routes with role-based authentication
func setupAdminRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// Admin-only routes group
	adminRoutes := router.Group("/admin")
	adminRoutes.Use(authMiddleware.RequireAuth())        // Require authentication
	adminRoutes.Use(authMiddleware.RequireRole("admin")) // Require admin role or higher

	// Super admin only routes
	superAdminRoutes := router.Group("/admin/super")
	superAdminRoutes.Use(authMiddleware.RequireAuth())              // Require authentication
	superAdminRoutes.Use(authMiddleware.RequireRole("super_admin")) // Require super admin role

	// Admin routes
	adminRoutes.GET("/users", getAdminUsers)
	adminRoutes.GET("/analytics", getAdminAnalytics)
	adminRoutes.POST("/users/:id/role", updateUserRole)

	// Super admin routes
	superAdminRoutes.DELETE("/users/:id", deleteUser)
	superAdminRoutes.POST("/admin", createAdmin)
}

// Example admin endpoints
func getAdminUsers(c *gin.Context) {
	// This endpoint is only accessible to admin and super_admin users
	userID := c.GetInt("user_id")
	userRole := c.GetString("role")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Admin users endpoint",
		"admin_id":   userID,
		"admin_role": userRole,
	})
}

func getAdminAnalytics(c *gin.Context) {
	// This endpoint is only accessible to admin and super_admin users
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin analytics endpoint",
		"data":    "Analytics data here",
	})
}

func updateUserRole(c *gin.Context) {
	// This endpoint is only accessible to admin and super_admin users
	userID := c.Param("id")
	newRole := c.PostForm("role")

	c.JSON(http.StatusOK, gin.H{
		"message":  "User role updated",
		"user_id":  userID,
		"new_role": newRole,
	})
}

func deleteUser(c *gin.Context) {
	// This endpoint is only accessible to super_admin users
	userID := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted",
		"user_id": userID,
	})
}

func createAdmin(c *gin.Context) {
	// This endpoint is only accessible to super_admin users
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin user created",
	})
}

package handlers

import (
	"net/http"
	"time"

	"panda-pocket/internal/interfaces/http/versioning"

	"github.com/gin-gonic/gin"
)

// DeprecationHandler handles deprecation warnings and notifications
type DeprecationHandler struct {
	versionManager *versioning.VersionManager
}

// NewDeprecationHandler creates a new deprecation handler instance
func NewDeprecationHandler(vm *versioning.VersionManager) *DeprecationHandler {
	return &DeprecationHandler{
		versionManager: vm,
	}
}

// HandleDeprecatedVersion handles requests to deprecated API versions
func (dh *DeprecationHandler) HandleDeprecatedVersion(c *gin.Context) {
	version := c.GetString("api_version")

	if deprecationInfo, exists := dh.versionManager.GetDeprecationInfo(version); exists {
		// Check if version has reached sunset date
		if dh.versionManager.CheckSunsetDate(version) {
			c.JSON(http.StatusGone, gin.H{
				"error":          "API version no longer available",
				"version":        version,
				"sunset_date":    deprecationInfo.SunsetDate.Format("2006-01-02"),
				"latest_version": dh.versionManager.GetCurrentVersion(),
				"upgrade_url":    deprecationInfo.UpgradeURL,
				"message":        "This API version has reached its end of life. Please upgrade to the latest version.",
			})
			return
		}

		// Add deprecation warning to response
		c.Header("X-API-Deprecated", "true")
		c.Header("X-API-Sunset-Date", deprecationInfo.SunsetDate.Format("2006-01-02"))
		c.Header("X-API-Upgrade-URL", deprecationInfo.UpgradeURL)
		c.Header("X-API-Latest", dh.versionManager.GetCurrentVersion())

		// Continue with the request but add warning to response
		c.Next()

		// Add deprecation warning to response body if it's a successful response
		if c.Writer.Status() < 400 {
			dh.addDeprecationWarningToResponse(c, version, deprecationInfo)
		}
	} else {
		c.Next()
	}
}

// addDeprecationWarningToResponse adds deprecation warning to the response body
func (dh *DeprecationHandler) addDeprecationWarningToResponse(c *gin.Context, version string, info versioning.DeprecationInfo) {
	// This would typically be handled by a response interceptor
	// For now, we'll add it as a header that clients can check
	c.Header("X-API-Deprecation-Warning", info.WarningMessage)
	c.Header("X-API-Deprecation-Details", "This API version is deprecated and will be removed on "+info.SunsetDate.Format("2006-01-02"))
}

// GetDeprecationInfo returns deprecation information for a version
func (dh *DeprecationHandler) GetDeprecationInfo(c *gin.Context) {
	version := c.Param("version")

	if version == "" {
		version = c.GetString("api_version")
	}

	if version == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Version parameter is required",
		})
		return
	}

	if deprecationInfo, exists := dh.versionManager.GetDeprecationInfo(version); exists {
		c.JSON(http.StatusOK, gin.H{
			"version":         version,
			"deprecated":      true,
			"sunset_date":     deprecationInfo.SunsetDate.Format("2006-01-02"),
			"warning_message": deprecationInfo.WarningMessage,
			"upgrade_url":     deprecationInfo.UpgradeURL,
			"latest_version":  dh.versionManager.GetCurrentVersion(),
			"is_sunset":       dh.versionManager.CheckSunsetDate(version),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"version":    version,
			"deprecated": false,
			"status":     dh.versionManager.GetVersionStatus(version),
		})
	}
}

// GetVersionStatus returns the status of a version
func (dh *DeprecationHandler) GetVersionStatus(c *gin.Context) {
	version := c.Param("version")

	if version == "" {
		version = c.GetString("api_version")
	}

	if version == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Version parameter is required",
		})
		return
	}

	status := dh.versionManager.GetVersionStatus(version)
	lifecycle := dh.versionManager.GetVersionLifecycle(version)

	c.JSON(http.StatusOK, gin.H{
		"version":            version,
		"status":             status,
		"lifecycle":          lifecycle,
		"current_version":    dh.versionManager.GetCurrentVersion(),
		"supported_versions": dh.versionManager.GetSupportedVersions(),
	})
}

// GetVersionMatrix returns a matrix of all versions and their status
func (dh *DeprecationHandler) GetVersionMatrix(c *gin.Context) {
	matrix := dh.versionManager.GetVersionMatrix()
	c.JSON(http.StatusOK, matrix)
}

// GetMigrationPath returns the recommended migration path for a version
func (dh *DeprecationHandler) GetMigrationPath(c *gin.Context) {
	fromVersion := c.Param("version")

	if fromVersion == "" {
		fromVersion = c.GetString("api_version")
	}

	if fromVersion == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Version parameter is required",
		})
		return
	}

	migrationPath := dh.versionManager.GetMigrationPath(fromVersion)

	c.JSON(http.StatusOK, gin.H{
		"from_version":    fromVersion,
		"migration_path":  migrationPath,
		"current_version": dh.versionManager.GetCurrentVersion(),
		"features":        dh.versionManager.GetVersionFeatures(fromVersion),
	})
}

// CompareVersions compares two versions
func (dh *DeprecationHandler) CompareVersions(c *gin.Context) {
	version1 := c.Query("version1")
	version2 := c.Query("version2")

	if version1 == "" || version2 == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Both version1 and version2 parameters are required",
		})
		return
	}

	comparison := dh.versionManager.CompareVersions(version1, version2)
	c.JSON(http.StatusOK, comparison)
}

// GetVersionFeatures returns the features available in a version
func (dh *DeprecationHandler) GetVersionFeatures(c *gin.Context) {
	version := c.Param("version")

	if version == "" {
		version = c.GetString("api_version")
	}

	if version == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Version parameter is required",
		})
		return
	}

	features := dh.versionManager.GetVersionFeatures(version)

	c.JSON(http.StatusOK, gin.H{
		"version":  version,
		"features": features,
		"status":   dh.versionManager.GetVersionStatus(version),
	})
}

// ValidateVersionTransition validates if a version transition is valid
func (dh *DeprecationHandler) ValidateVersionTransition(c *gin.Context) {
	fromVersion := c.Query("from")
	toVersion := c.Query("to")

	if fromVersion == "" || toVersion == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Both from and to parameters are required",
		})
		return
	}

	err := dh.versionManager.ValidateVersionTransition(fromVersion, toVersion)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        err.Error(),
			"from_version": fromVersion,
			"to_version":   toVersion,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":        true,
		"from_version": fromVersion,
		"to_version":   toVersion,
		"message":      "Version transition is valid",
	})
}

// GetDeprecationTimeline returns the deprecation timeline for all versions
func (dh *DeprecationHandler) GetDeprecationTimeline(c *gin.Context) {
	timeline := make(map[string]interface{})

	// Get all versions
	allVersions := make(map[string]bool)
	for _, v := range dh.versionManager.GetSupportedVersions() {
		allVersions[v] = true
	}
	for _, v := range dh.versionManager.GetDeprecatedVersions() {
		allVersions[v] = true
	}

	// Build timeline
	for version := range allVersions {
		lifecycle := dh.versionManager.GetVersionLifecycle(version)
		timeline[version] = lifecycle
	}

	c.JSON(http.StatusOK, gin.H{
		"timeline":        timeline,
		"current_date":    time.Now().Format("2006-01-02"),
		"current_version": dh.versionManager.GetCurrentVersion(),
	})
}

// GetUpgradeRecommendations returns upgrade recommendations for a version
func (dh *DeprecationHandler) GetUpgradeRecommendations(c *gin.Context) {
	version := c.Param("version")

	if version == "" {
		version = c.GetString("api_version")
	}

	if version == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Version parameter is required",
		})
		return
	}

	currentVersion := dh.versionManager.GetCurrentVersion()

	recommendations := gin.H{
		"current_version":   currentVersion,
		"requested_version": version,
		"is_deprecated":     dh.versionManager.IsVersionDeprecated(version),
		"is_sunset":         dh.versionManager.CheckSunsetDate(version),
		"migration_path":    dh.versionManager.GetMigrationPath(version),
		"features":          dh.versionManager.GetVersionFeatures(version),
		"new_features":      dh.versionManager.GetVersionFeatures(currentVersion),
	}

	// Add specific recommendations based on version status
	if dh.versionManager.IsVersionDeprecated(version) {
		if info, exists := dh.versionManager.GetDeprecationInfo(version); exists {
			recommendations["sunset_date"] = info.SunsetDate.Format("2006-01-02")
			recommendations["warning_message"] = info.WarningMessage
			recommendations["upgrade_url"] = info.UpgradeURL
		}
	}

	c.JSON(http.StatusOK, recommendations)
}

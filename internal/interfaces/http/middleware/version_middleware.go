package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// VersionInfo represents information about an API version
type VersionInfo struct {
	Version      string
	IsSupported  bool
	IsDeprecated bool
	SunsetDate   string
	UpgradeURL   string
}

// VersionMiddleware handles API version extraction and validation
type VersionMiddleware struct {
	supportedVersions map[string]VersionInfo
}

// NewVersionMiddleware creates a new version middleware instance
func NewVersionMiddleware() *VersionMiddleware {
	return &VersionMiddleware{
		supportedVersions: map[string]VersionInfo{
			"v100": {
				Version:      "v100",
				IsSupported:  true,
				IsDeprecated: false,
				SunsetDate:   "",
				UpgradeURL:   "",
			},
		},
	}
}

// ExtractVersion extracts and validates API version from URL path
func (vm *VersionMiddleware) ExtractVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Extract version from path like /api/v120/transactions
		parts := strings.Split(path, "/")
		if len(parts) >= 3 && strings.HasPrefix(parts[2], "v") {
			version := parts[2]

			if versionInfo, exists := vm.supportedVersions[version]; exists {
				c.Set("api_version", version)
				c.Set("version_info", versionInfo)

				// Add version headers
				c.Header("X-API-Version", version)
				if versionInfo.IsDeprecated {
					c.Header("X-API-Deprecated", "true")
					c.Header("X-API-Sunset-Date", versionInfo.SunsetDate)
					c.Header("X-API-Upgrade-URL", versionInfo.UpgradeURL)
				}

				c.Next()
				return
			}
		}

		// No version specified - redirect to latest
		c.Header("X-API-Version", "v100")
		c.Header("X-API-Latest", "v100")
		c.Next()
	}
}

// ValidateVersion validates if the requested version is supported
func (vm *VersionMiddleware) ValidateVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetString("api_version")

		if version == "" {
			// No version specified, use latest
			c.Set("api_version", "v100")
			c.Next()
			return
		}

		versionInfo, exists := vm.supportedVersions[version]
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":              "Unsupported API version",
				"supported_versions": []string{"v100"},
				"latest_version":     "v100",
			})
			c.Abort()
			return
		}

		if !versionInfo.IsSupported {
			c.JSON(http.StatusGone, gin.H{
				"error":          "API version no longer supported",
				"version":        version,
				"latest_version": "v100",
				"upgrade_url":    "https://docs.pandapocket.com/upgrade",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AddDeprecationWarning adds deprecation warnings to responses
func (vm *VersionMiddleware) AddDeprecationWarning() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetString("api_version")

		if versionInfo, exists := vm.supportedVersions[version]; exists && versionInfo.IsDeprecated {
			// Add deprecation warning to response headers
			c.Header("X-API-Deprecated", "true")
			c.Header("X-API-Sunset-Date", versionInfo.SunsetDate)
			c.Header("X-API-Upgrade-URL", versionInfo.UpgradeURL)

			// Add deprecation warning to response body
			c.Next()

			// Check if response is successful and add deprecation warning
			if c.Writer.Status() < 400 {
				// This will be handled by a response interceptor
				vm.addDeprecationWarningToResponse(c)
			}
		} else {
			c.Next()
		}
	}
}

// addDeprecationWarningToResponse adds deprecation warning to response body
func (vm *VersionMiddleware) addDeprecationWarningToResponse(c *gin.Context) {
	version := c.GetString("api_version")
	versionInfo := vm.supportedVersions[version]

	// Get the current response
	response := c.Writer.Header().Get("Content-Type")
	if strings.Contains(response, "application/json") {
		// Add deprecation warning to JSON response
		c.Header("X-API-Deprecation-Warning", fmt.Sprintf("API version %s is deprecated and will be removed on %s. Please upgrade to v100.", version, versionInfo.SunsetDate))
	}
}

// GetSupportedVersions returns list of supported versions
func (vm *VersionMiddleware) GetSupportedVersions() []string {
	var versions []string
	for version, info := range vm.supportedVersions {
		if info.IsSupported {
			versions = append(versions, version)
		}
	}
	return versions
}

// IsVersionDeprecated checks if a version is deprecated
func (vm *VersionMiddleware) IsVersionDeprecated(version string) bool {
	if info, exists := vm.supportedVersions[version]; exists {
		return info.IsDeprecated
	}
	return false
}

// GetVersionInfo returns version information
func (vm *VersionMiddleware) GetVersionInfo(version string) (VersionInfo, bool) {
	info, exists := vm.supportedVersions[version]
	return info, exists
}

// UpdateVersionStatus updates the status of a version
func (vm *VersionMiddleware) UpdateVersionStatus(version string, isDeprecated bool, sunsetDate string) {
	if info, exists := vm.supportedVersions[version]; exists {
		info.IsDeprecated = isDeprecated
		info.SunsetDate = sunsetDate
		vm.supportedVersions[version] = info
	}
}

// AddVersion adds a new version to the supported versions
func (vm *VersionMiddleware) AddVersion(version string, isDeprecated bool, sunsetDate string) {
	vm.supportedVersions[version] = VersionInfo{
		Version:      version,
		IsSupported:  true,
		IsDeprecated: isDeprecated,
		SunsetDate:   sunsetDate,
		UpgradeURL:   "https://docs.pandapocket.com/upgrade",
	}
}

// RemoveVersion removes a version from supported versions
func (vm *VersionMiddleware) RemoveVersion(version string) {
	if info, exists := vm.supportedVersions[version]; exists {
		info.IsSupported = false
		vm.supportedVersions[version] = info
	}
}

// GetCurrentVersion returns the current/latest version
func (vm *VersionMiddleware) GetCurrentVersion() string {
	return "v100"
}

// GetDeprecatedVersions returns list of deprecated versions
func (vm *VersionMiddleware) GetDeprecatedVersions() []string {
	var versions []string
	for version, info := range vm.supportedVersions {
		if info.IsDeprecated {
			versions = append(versions, version)
		}
	}
	return versions
}

// CheckSunsetDate checks if a version has reached its sunset date
func (vm *VersionMiddleware) CheckSunsetDate(version string) bool {
	if info, exists := vm.supportedVersions[version]; exists && info.SunsetDate != "" {
		sunsetDate, err := time.Parse("2006-01-02", info.SunsetDate)
		if err != nil {
			return false
		}
		return time.Now().After(sunsetDate)
	}
	return false
}

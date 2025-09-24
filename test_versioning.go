package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"panda-pocket/internal/application"
	"panda-pocket/internal/infrastructure/database"
	"panda-pocket/internal/interfaces/http/handlers"
	"panda-pocket/internal/interfaces/http/middleware"
	"panda-pocket/internal/interfaces/http/versioning"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIVersioning(t *testing.T) {
	// Setup test database
	db, err := database.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create application
	app := application.NewApp(db)

	// Setup routes
	router := app.SetupRoutes()

	// Test cases
	tests := []struct {
		name            string
		method          string
		path            string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:           "v120 transactions endpoint",
			method:         "GET",
			path:           "/api/v120/transactions",
			expectedStatus: http.StatusUnauthorized, // No auth token
			expectedHeaders: map[string]string{
				"X-API-Version": "v120",
			},
		},
		{
			name:           "v100 transactions endpoint (deprecated)",
			method:         "GET",
			path:           "/api/v100/transactions",
			expectedStatus: http.StatusUnauthorized, // No auth token
			expectedHeaders: map[string]string{
				"X-API-Version":     "v100",
				"X-API-Deprecated":  "true",
				"X-API-Sunset-Date": "2024-06-01",
				"X-API-Upgrade-URL": "https://docs.pandapocket.com/upgrade",
			},
		},
		{
			name:           "legacy transactions endpoint",
			method:         "GET",
			path:           "/api/transactions",
			expectedStatus: http.StatusUnauthorized, // No auth token
			expectedHeaders: map[string]string{
				"X-API-Version": "v120",
				"X-API-Latest":  "v120",
			},
		},
		{
			name:           "version info endpoint",
			method:         "GET",
			path:           "/api/version/info/v100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "version matrix endpoint",
			method:         "GET",
			path:           "/api/version/matrix",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(tt.method, tt.path, nil)
			assert.NoError(t, err)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			// Check headers
			for header, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(header)
				assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", header)
			}

			// Print response for debugging
			if w.Code != tt.expectedStatus {
				fmt.Printf("Response body: %s\n", w.Body.String())
			}
		})
	}
}

func TestVersionMiddleware(t *testing.T) {
	// Test version extraction
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add version middleware
	versionMiddleware := middleware.NewVersionMiddleware()
	router.Use(versionMiddleware.ExtractVersion())

	// Add test route
	router.GET("/api/v120/test", func(c *gin.Context) {
		version := c.GetString("api_version")
		c.JSON(http.StatusOK, gin.H{"version": version})
	})

	// Test request
	req, _ := http.NewRequest("GET", "/api/v120/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "v120", w.Header().Get("X-API-Version"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "v120", response["version"])
}

func TestDeprecationHandler(t *testing.T) {
	// Test deprecation handling
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add version middleware
	versionMiddleware := middleware.NewVersionMiddleware()
	router.Use(versionMiddleware.ExtractVersion())

	// Add deprecation handler
	versionManager := versioning.NewVersionManager()
	deprecationHandler := handlers.NewDeprecationHandler(versionManager)
	router.Use(deprecationHandler.HandleDeprecatedVersion)

	// Add test route
	router.GET("/api/v100/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test request
	req, _ := http.NewRequest("GET", "/api/v100/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "true", w.Header().Get("X-API-Deprecated"))
	assert.Equal(t, "2024-06-01", w.Header().Get("X-API-Sunset-Date"))
	assert.Equal(t, "https://docs.pandapocket.com/upgrade", w.Header().Get("X-API-Upgrade-URL"))
}

func TestVersionManager(t *testing.T) {
	vm := versioning.NewVersionManager()

	// Test supported versions
	assert.True(t, vm.IsVersionSupported("v120"))
	assert.True(t, vm.IsVersionSupported("v110"))
	assert.True(t, vm.IsVersionSupported("v100"))
	assert.False(t, vm.IsVersionSupported("v090"))

	// Test deprecated versions
	assert.True(t, vm.IsVersionDeprecated("v100"))
	assert.False(t, vm.IsVersionDeprecated("v120"))

	// Test current version
	assert.Equal(t, "v120", vm.GetCurrentVersion())

	// Test supported versions list
	supportedVersions := vm.GetSupportedVersions()
	assert.Contains(t, supportedVersions, "v120")
	assert.Contains(t, supportedVersions, "v110")
	assert.Contains(t, supportedVersions, "v100")
}

func TestVersionComparison(t *testing.T) {
	vm := versioning.NewVersionManager()

	// Test version comparison
	comparison := vm.CompareVersions("v100", "v120")
	assert.NotNil(t, comparison)
	assert.Equal(t, "v100", comparison["version1"])
	assert.Equal(t, "v120", comparison["version2"])

	// Test version features
	features := vm.GetVersionFeatures("v120")
	assert.NotNil(t, features)
	assert.True(t, features["analytics"].(bool))
	assert.True(t, features["advanced_filtering"].(bool))
	assert.True(t, features["bulk_operations"].(bool))
	assert.True(t, features["export_functionality"].(bool))
}

func TestMigrationPath(t *testing.T) {
	vm := versioning.NewVersionManager()

	// Test migration path
	migrationPath := vm.GetMigrationPath("v100")
	assert.NotNil(t, migrationPath)
	assert.Contains(t, migrationPath, "v120")

	// Test version transition validation
	err := vm.ValidateVersionTransition("v100", "v120")
	assert.NoError(t, err)

	err = vm.ValidateVersionTransition("v100", "v090")
	assert.Error(t, err)
}

// Benchmark tests
func BenchmarkVersionMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	versionMiddleware := middleware.NewVersionMiddleware()
	router.Use(versionMiddleware.ExtractVersion())

	router.GET("/api/v120/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "v120"})
	})

	req, _ := http.NewRequest("GET", "/api/v120/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, httptest.NewRecorder())
	}
}

func BenchmarkVersionManager(b *testing.B) {
	vm := versioning.NewVersionManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.IsVersionSupported("v120")
		vm.IsVersionDeprecated("v100")
		vm.GetCurrentVersion()
	}
}

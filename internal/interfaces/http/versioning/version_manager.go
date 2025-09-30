package versioning

import (
	"fmt"
	"time"
)

// VersionManager manages API versions and their lifecycle
type VersionManager struct {
	currentVersion     string
	supportedVersions  []string
	deprecatedVersions map[string]DeprecationInfo
}

// DeprecationInfo contains information about deprecated versions
type DeprecationInfo struct {
	Version        string
	SunsetDate     time.Time
	WarningMessage string
	UpgradeURL     string
}

// NewVersionManager creates a new version manager instance
func NewVersionManager() *VersionManager {
	return &VersionManager{
		currentVersion:     "v100",
		supportedVersions:  []string{"v100"},
		deprecatedVersions: map[string]DeprecationInfo{},
	}
}

// IsVersionSupported checks if a version is supported
func (vm *VersionManager) IsVersionSupported(version string) bool {
	for _, v := range vm.supportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

// IsVersionDeprecated checks if a version is deprecated
func (vm *VersionManager) IsVersionDeprecated(version string) bool {
	_, exists := vm.deprecatedVersions[version]
	return exists
}

// GetDeprecationInfo returns deprecation information for a version
func (vm *VersionManager) GetDeprecationInfo(version string) (DeprecationInfo, bool) {
	info, exists := vm.deprecatedVersions[version]
	return info, exists
}

// GetCurrentVersion returns the current/latest version
func (vm *VersionManager) GetCurrentVersion() string {
	return vm.currentVersion
}

// GetSupportedVersions returns list of supported versions
func (vm *VersionManager) GetSupportedVersions() []string {
	return vm.supportedVersions
}

// GetDeprecatedVersions returns list of deprecated versions
func (vm *VersionManager) GetDeprecatedVersions() []string {
	var versions []string
	for version := range vm.deprecatedVersions {
		versions = append(versions, version)
	}
	return versions
}

// AddDeprecatedVersion adds a version to the deprecated list
func (vm *VersionManager) AddDeprecatedVersion(version string, sunsetDate time.Time, warningMessage string) {
	vm.deprecatedVersions[version] = DeprecationInfo{
		Version:        version,
		SunsetDate:     sunsetDate,
		WarningMessage: warningMessage,
		UpgradeURL:     "https://docs.pandapocket.com/upgrade",
	}
}

// RemoveDeprecatedVersion removes a version from the deprecated list
func (vm *VersionManager) RemoveDeprecatedVersion(version string) {
	delete(vm.deprecatedVersions, version)
}

// UpdateCurrentVersion updates the current version
func (vm *VersionManager) UpdateCurrentVersion(version string) {
	vm.currentVersion = version
}

// AddSupportedVersion adds a version to the supported list
func (vm *VersionManager) AddSupportedVersion(version string) {
	// Check if version already exists
	for _, v := range vm.supportedVersions {
		if v == version {
			return
		}
	}
	vm.supportedVersions = append(vm.supportedVersions, version)
}

// RemoveSupportedVersion removes a version from the supported list
func (vm *VersionManager) RemoveSupportedVersion(version string) {
	for i, v := range vm.supportedVersions {
		if v == version {
			vm.supportedVersions = append(vm.supportedVersions[:i], vm.supportedVersions[i+1:]...)
			break
		}
	}
}

// CheckSunsetDate checks if a version has reached its sunset date
func (vm *VersionManager) CheckSunsetDate(version string) bool {
	if info, exists := vm.deprecatedVersions[version]; exists {
		return time.Now().After(info.SunsetDate)
	}
	return false
}

// GetVersionStatus returns the status of a version
func (vm *VersionManager) GetVersionStatus(version string) string {
	if vm.IsVersionSupported(version) {
		if vm.IsVersionDeprecated(version) {
			return "deprecated"
		}
		return "supported"
	}
	return "unsupported"
}

// GetVersionLifecycle returns the lifecycle information for a version
func (vm *VersionManager) GetVersionLifecycle(version string) map[string]interface{} {
	status := vm.GetVersionStatus(version)

	lifecycle := map[string]interface{}{
		"version": version,
		"status":  status,
		"current": version == vm.currentVersion,
	}

	if vm.IsVersionDeprecated(version) {
		if info, exists := vm.GetDeprecationInfo(version); exists {
			lifecycle["sunset_date"] = info.SunsetDate.Format("2006-01-02")
			lifecycle["warning_message"] = info.WarningMessage
			lifecycle["upgrade_url"] = info.UpgradeURL
			lifecycle["is_sunset"] = vm.CheckSunsetDate(version)
		}
	}

	return lifecycle
}

// GetVersionMatrix returns a matrix of all versions and their status
func (vm *VersionManager) GetVersionMatrix() map[string]interface{} {
	matrix := map[string]interface{}{
		"current_version":     vm.currentVersion,
		"supported_versions":  vm.supportedVersions,
		"deprecated_versions": vm.GetDeprecatedVersions(),
		"versions":            make(map[string]interface{}),
	}

	// Add all versions to the matrix
	allVersions := make(map[string]bool)
	for _, v := range vm.supportedVersions {
		allVersions[v] = true
	}
	for v := range vm.deprecatedVersions {
		allVersions[v] = true
	}

	for version := range allVersions {
		matrix["versions"].(map[string]interface{})[version] = vm.GetVersionLifecycle(version)
	}

	return matrix
}

// ValidateVersionTransition validates if a version transition is valid
func (vm *VersionManager) ValidateVersionTransition(fromVersion, toVersion string) error {
	// Check if both versions are supported
	if !vm.IsVersionSupported(fromVersion) {
		return fmt.Errorf("source version %s is not supported", fromVersion)
	}

	if !vm.IsVersionSupported(toVersion) {
		return fmt.Errorf("target version %s is not supported", toVersion)
	}

	// Check if target version is not deprecated
	if vm.IsVersionDeprecated(toVersion) {
		return fmt.Errorf("target version %s is deprecated", toVersion)
	}

	return nil
}

// GetMigrationPath returns the recommended migration path for a version
func (vm *VersionManager) GetMigrationPath(fromVersion string) []string {
	if fromVersion == vm.currentVersion {
		return []string{vm.currentVersion}
	}

	// Simple migration path: go directly to current version
	return []string{vm.currentVersion}
}

// GetVersionFeatures returns the features available in a version
func (vm *VersionManager) GetVersionFeatures(version string) map[string]interface{} {
	features := map[string]interface{}{
		"basic_transactions": true,
		"categories":         true,
		"currencies":         true,
		"budgets":            true,
	}

	// Add version-specific features
	switch version {
	case "v100":
		features["analytics"] = true
		features["advanced_filtering"] = true
		features["bulk_operations"] = true
		features["export_functionality"] = true
	}

	return features
}

// CompareVersions compares two versions and returns the differences
func (vm *VersionManager) CompareVersions(version1, version2 string) map[string]interface{} {
	features1 := vm.GetVersionFeatures(version1)
	features2 := vm.GetVersionFeatures(version2)

	comparison := map[string]interface{}{
		"version1":    version1,
		"version2":    version2,
		"features1":   features1,
		"features2":   features2,
		"differences": make(map[string]interface{}),
	}

	// Find differences
	allFeatures := make(map[string]bool)
	for feature := range features1 {
		allFeatures[feature] = true
	}
	for feature := range features2 {
		allFeatures[feature] = true
	}

	differences := make(map[string]interface{})
	for feature := range allFeatures {
		has1 := features1[feature].(bool)
		has2 := features2[feature].(bool)

		if has1 != has2 {
			differences[feature] = map[string]bool{
				version1: has1,
				version2: has2,
			}
		}
	}

	comparison["differences"] = differences
	return comparison
}

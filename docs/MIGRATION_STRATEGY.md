# API Versioning Migration Strategy

This document outlines the comprehensive migration strategy for implementing API versioning in the PandaPocket backend, ensuring smooth transitions between versions while maintaining backward compatibility.

## Table of Contents

- [Migration Overview](#migration-overview)
- [Version Lifecycle Management](#version-lifecycle-management)
- [Migration Phases](#migration-phases)
- [Client Migration Strategy](#client-migration-strategy)
- [Database Migration Considerations](#database-migration-considerations)
- [Testing Strategy](#testing-strategy)
- [Rollback Procedures](#rollback-procedures)
- [Monitoring and Metrics](#monitoring-and-metrics)
- [Communication Plan](#communication-plan)

## Migration Overview

### Current State
- **API Version**: Unversioned (legacy)
- **Endpoints**: `/api/transactions`, `/api/categories`, etc.
- **Clients**: Unknown number of clients using current API

### Target State
- **API Versioning**: `/api/v100/transactions`, `/api/v110/transactions`, `/api/v120/transactions`
- **Backward Compatibility**: Maintain last 3 versions
- **Client Awareness**: Clear deprecation warnings and upgrade paths

## Version Lifecycle Management

### Version Numbering Scheme

```
v100 = Version 1.0.0 (Initial versioned release)
v110 = Version 1.1.0 (Minor updates, backward compatible)
v120 = Version 1.2.0 (Latest version with new features)
```

### Version Lifecycle Stages

#### 1. **Development Phase**
- New features developed in latest version
- Previous versions maintained for bug fixes only
- No breaking changes in supported versions

#### 2. **Release Phase**
- New version deployed alongside existing versions
- Gradual client migration encouraged
- Deprecation warnings for older versions

#### 3. **Deprecation Phase**
- Strong warnings for deprecated versions
- Sunset date announced
- Migration assistance provided

#### 4. **End of Life Phase**
- Deprecated versions return errors
- Clients must upgrade to supported versions

## Migration Phases

### Phase 1: Infrastructure Setup (Week 1-2)

#### 1.1 Version Middleware Implementation
```go
// internal/interfaces/http/middleware/version_middleware.go
package middleware

import (
    "fmt"
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
)

type VersionMiddleware struct {
    supportedVersions map[string]VersionInfo
}

type VersionInfo struct {
    Version     string
    IsSupported bool
    IsDeprecated bool
    SunsetDate  string
    UpgradeURL  string
}

func NewVersionMiddleware() *VersionMiddleware {
    return &VersionMiddleware{
        supportedVersions: map[string]VersionInfo{
            "v120": {
                Version:     "v120",
                IsSupported: true,
                IsDeprecated: false,
                SunsetDate:  "",
                UpgradeURL:  "",
            },
            "v110": {
                Version:     "v110",
                IsSupported: true,
                IsDeprecated: false,
                SunsetDate:  "",
                UpgradeURL:  "",
            },
            "v100": {
                Version:     "v100",
                IsSupported: true,
                IsDeprecated: true,
                SunsetDate:  "2024-06-01",
                UpgradeURL:  "https://docs.pandapocket.com/upgrade",
            },
        },
    }
}

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
        c.Header("X-API-Version", "v120")
        c.Header("X-API-Latest", "v120")
        c.Next()
    }
}
```

#### 1.2 Version Manager Service
```go
// internal/interfaces/http/versioning/version_manager.go
package versioning

import (
    "time"
    "fmt"
)

type VersionManager struct {
    currentVersion string
    supportedVersions []string
    deprecatedVersions map[string]DeprecationInfo
}

type DeprecationInfo struct {
    Version     string
    SunsetDate  time.Time
    WarningMessage string
    UpgradeURL  string
}

func NewVersionManager() *VersionManager {
    return &VersionManager{
        currentVersion: "v120",
        supportedVersions: []string{"v120", "v110", "v100"},
        deprecatedVersions: map[string]DeprecationInfo{
            "v100": {
                Version: "v100",
                SunsetDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
                WarningMessage: "API version v100 is deprecated and will be removed on 2024-06-01. Please upgrade to v120.",
                UpgradeURL: "https://docs.pandapocket.com/upgrade",
            },
        },
    }
}

func (vm *VersionManager) IsVersionSupported(version string) bool {
    for _, v := range vm.supportedVersions {
        if v == version {
            return true
        }
    }
    return false
}

func (vm *VersionManager) IsVersionDeprecated(version string) bool {
    _, exists := vm.deprecatedVersions[version]
    return exists
}

func (vm *VersionManager) GetDeprecationInfo(version string) (DeprecationInfo, bool) {
    info, exists := vm.deprecatedVersions[version]
    return info, exists
}
```

### Phase 2: Route Restructuring (Week 3-4)

#### 2.1 Versioned Route Structure
```go
// internal/application/app.go - Updated SetupRoutes method
func (app *App) SetupRoutes() *gin.Engine {
    r := gin.Default()
    
    // CORS configuration
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{
        "http://localhost:3000",
        "http://localhost:3001",
        "http://localhost:3002",
        "http://localhost:3003",
        "http://localhost:3004",
    }
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Version"}
    config.AllowCredentials = false
    r.Use(cors.New(config))
    
    // Version middleware
    versionMiddleware := middleware.NewVersionMiddleware()
    r.Use(versionMiddleware.ExtractVersion())
    
    // Legacy routes (for backward compatibility)
    legacy := r.Group("/api")
    {
        // Auth routes
        auth := legacy.Group("/auth")
        {
            auth.POST("/register", app.IdentityHandlers.Register)
            auth.POST("/login", app.IdentityHandlers.Login)
            auth.POST("/logout", app.IdentityHandlers.Logout)
        }
        
        // Protected routes
        protected := legacy.Group("")
        protected.Use(app.AuthMiddleware.RequireAuth())
        {
            // Legacy endpoints with deprecation warnings
            protected.GET("/transactions", app.FinanceHandlers.GetAllTransactions)
            protected.POST("/expenses", app.FinanceHandlers.CreateExpense)
            protected.POST("/incomes", app.FinanceHandlers.CreateIncome)
            // ... other legacy routes
        }
    }
    
    // Versioned routes
    versioned := r.Group("/api")
    {
        // v120 routes (latest)
        v120 := versioned.Group("/v120")
        {
            // Auth routes
            auth := v120.Group("/auth")
            {
                auth.POST("/register", app.IdentityHandlers.Register)
                auth.POST("/login", app.IdentityHandlers.Login)
                auth.POST("/logout", app.IdentityHandlers.Logout)
            }
            
            // Protected routes
            protected := v120.Group("")
            protected.Use(app.AuthMiddleware.RequireAuth())
            {
                // Latest features
                protected.GET("/transactions", app.FinanceHandlers.GetAllTransactions)
                protected.POST("/transactions", app.FinanceHandlers.CreateTransaction)
                protected.PUT("/transactions/:id", app.FinanceHandlers.UpdateTransaction)
                protected.DELETE("/transactions/:id", app.FinanceHandlers.DeleteTransaction)
                // ... other v120 routes
            }
        }
        
        // v110 routes (previous version)
        v110 := versioned.Group("/v110")
        {
            // Similar structure but with v110-specific handlers
            // ... v110 routes
        }
        
        // v100 routes (legacy)
        v100 := versioned.Group("/v100")
        {
            // Legacy structure with deprecation warnings
            // ... v100 routes
        }
    }
    
    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })
    
    return r
}
```

### Phase 3: Client Migration (Week 5-8)

#### 3.1 Gradual Migration Strategy

**Week 5-6: Soft Launch**
- Deploy versioned API alongside legacy
- Add deprecation warnings to legacy endpoints
- Monitor usage patterns

**Week 7-8: Active Migration**
- Send migration notifications to known clients
- Provide migration guides and examples
- Offer support for migration process

#### 3.2 Client Notification System
```go
// internal/interfaces/http/middleware/deprecation_middleware.go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func DeprecationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        version := c.GetString("api_version")
        
        // Check if version is deprecated
        if version == "v100" {
            c.Header("X-API-Deprecated", "true")
            c.Header("X-API-Sunset-Date", "2024-06-01")
            c.Header("X-API-Upgrade-URL", "https://docs.pandapocket.com/upgrade")
            
            // Add deprecation warning to response
            c.Next()
            
            // Add deprecation info to response body
            if c.Writer.Status() < 400 {
                // This would be handled in a response interceptor
            }
        } else {
            c.Next()
        }
    }
}
```

### Phase 4: Legacy Deprecation (Week 9-12)

#### 4.1 Deprecation Timeline

**Week 9-10: Warning Phase**
- Strong deprecation warnings
- Migration assistance provided
- Usage analytics collected

**Week 11-12: End of Life**
- Legacy endpoints return errors
- Clients must use versioned endpoints
- Legacy code removed

#### 4.2 Deprecation Response Handler
```go
// internal/interfaces/http/handlers/deprecation_handler.go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type DeprecationHandler struct {
    versionManager *versioning.VersionManager
}

func NewDeprecationHandler(vm *versioning.VersionManager) *DeprecationHandler {
    return &DeprecationHandler{
        versionManager: vm,
    }
}

func (dh *DeprecationHandler) HandleDeprecatedVersion(c *gin.Context) {
    version := c.GetString("api_version")
    
    if deprecationInfo, exists := dh.versionManager.GetDeprecationInfo(version); exists {
        c.JSON(http.StatusOK, gin.H{
            "warning": "API version deprecated",
            "current_version": version,
            "latest_version": "v120",
            "sunset_date": deprecationInfo.SunsetDate.Format("2006-01-02"),
            "upgrade_url": deprecationInfo.UpgradeURL,
            "message": deprecationInfo.WarningMessage,
            "data": nil, // Actual response data would go here
        })
        return
    }
    
    c.Next()
}
```

## Client Migration Strategy

### 1. **Immediate Actions (Week 1-2)**
- Update client libraries to support versioned endpoints
- Implement version detection and fallback logic
- Add version headers to requests

### 2. **Gradual Migration (Week 3-6)**
- Migrate to latest version (v120) for new features
- Maintain compatibility with previous versions
- Test migration with staging environment

### 3. **Full Migration (Week 7-8)**
- Complete migration to versioned endpoints
- Remove legacy endpoint usage
- Update documentation and examples

### 4. **Client Code Examples**

#### JavaScript/TypeScript Client
```typescript
class PandaPocketClient {
    private baseURL: string;
    private version: string;
    
    constructor(baseURL: string, version: string = 'v120') {
        this.baseURL = baseURL;
        this.version = version;
    }
    
    async getTransactions(): Promise<Transaction[]> {
        const response = await fetch(`${this.baseURL}/api/${this.version}/transactions`, {
            headers: {
                'Authorization': `Bearer ${this.token}`,
                'X-API-Version': this.version
            }
        });
        
        if (response.headers.get('X-API-Deprecated') === 'true') {
            console.warn('API version is deprecated:', {
                current: this.version,
                latest: response.headers.get('X-API-Latest'),
                sunset: response.headers.get('X-API-Sunset-Date'),
                upgrade: response.headers.get('X-API-Upgrade-URL')
            });
        }
        
        return response.json();
    }
}
```

#### Go Client
```go
type PandaPocketClient struct {
    BaseURL string
    Version string
    Token   string
}

func NewPandaPocketClient(baseURL, version, token string) *PandaPocketClient {
    return &PandaPocketClient{
        BaseURL: baseURL,
        Version: version,
        Token:   token,
    }
}

func (c *PandaPocketClient) GetTransactions() ([]Transaction, error) {
    url := fmt.Sprintf("%s/api/%s/transactions", c.BaseURL, c.Version)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("X-API-Version", c.Version)
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Check for deprecation warnings
    if resp.Header.Get("X-API-Deprecated") == "true" {
        log.Printf("API version %s is deprecated. Upgrade to %s",
            c.Version,
            resp.Header.Get("X-API-Latest"))
    }
    
    var transactions []Transaction
    err = json.NewDecoder(resp.Body).Decode(&transactions)
    return transactions, err
}
```

## Database Migration Considerations

### 1. **Schema Changes**
- New features may require database schema changes
- Maintain backward compatibility for existing data
- Use database migrations for schema updates

### 2. **Data Migration**
- Ensure data consistency across versions
- Handle data format changes gracefully
- Provide data migration scripts if needed

### 3. **Migration Scripts**
```sql
-- Example: Adding new column for v120
ALTER TABLE transactions ADD COLUMN metadata JSONB;

-- Example: Creating index for performance
CREATE INDEX idx_transactions_metadata ON transactions USING GIN (metadata);

-- Example: Data migration for new features
UPDATE transactions 
SET metadata = '{"version": "v120"}'::jsonb 
WHERE metadata IS NULL;
```

## Testing Strategy

### 1. **Unit Testing**
- Test version middleware functionality
- Test deprecation handling
- Test version-specific handlers

### 2. **Integration Testing**
- Test versioned endpoints
- Test backward compatibility
- Test deprecation warnings

### 3. **Load Testing**
- Test performance with multiple versions
- Test version switching under load
- Test deprecation handling under load

### 4. **Client Testing**
- Test client migration scenarios
- Test fallback mechanisms
- Test error handling

## Rollback Procedures

### 1. **Immediate Rollback (0-1 hour)**
- Revert to previous deployment
- Restore legacy endpoints
- Notify clients of temporary issues

### 2. **Partial Rollback (1-4 hours)**
- Disable problematic version
- Maintain working versions
- Investigate and fix issues

### 3. **Full Rollback (4-24 hours)**
- Complete system restoration
- Data consistency checks
- Client communication

## Monitoring and Metrics

### 1. **Version Usage Metrics**
- Track API version usage
- Monitor deprecation warnings
- Measure migration progress

### 2. **Performance Metrics**
- Response times by version
- Error rates by version
- Client satisfaction metrics

### 3. **Migration Metrics**
- Client migration progress
- Deprecation warning effectiveness
- Support ticket volume

## Communication Plan

### 1. **Pre-Migration (Week 1-2)**
- Announce versioning strategy
- Provide migration timeline
- Share documentation and examples

### 2. **During Migration (Week 3-8)**
- Regular progress updates
- Migration assistance
- Issue resolution support

### 3. **Post-Migration (Week 9-12)**
- Migration completion announcement
- Performance improvements summary
- Future roadmap communication

### 4. **Communication Channels**
- Developer documentation
- Email notifications
- Support tickets
- Community forums

## Success Criteria

### 1. **Technical Success**
- All clients migrated to versioned endpoints
- No breaking changes in supported versions
- Smooth deprecation of legacy endpoints

### 2. **Business Success**
- Improved API stability
- Better client experience
- Reduced support burden

### 3. **Operational Success**
- Clear version lifecycle
- Effective monitoring
- Smooth future migrations

## Risk Mitigation

### 1. **Technical Risks**
- **Risk**: Breaking changes in supported versions
- **Mitigation**: Comprehensive testing and gradual rollout

### 2. **Business Risks**
- **Risk**: Client migration resistance
- **Mitigation**: Clear communication and support

### 3. **Operational Risks**
- **Risk**: Increased complexity
- **Mitigation**: Proper documentation and training

## Conclusion

This migration strategy provides a comprehensive approach to implementing API versioning in the PandaPocket backend. The phased approach ensures minimal disruption to existing clients while providing a clear path for future API evolution.

Key success factors:
- Gradual migration approach
- Strong client communication
- Comprehensive testing
- Clear rollback procedures
- Effective monitoring

The strategy balances technical requirements with business needs, ensuring a smooth transition to versioned APIs while maintaining backward compatibility and client satisfaction.

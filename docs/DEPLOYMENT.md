# Deployment Guide

This guide covers deployment strategies, environment configuration, and production considerations for the PandaPocket backend application.

## Table of Contents

- [Deployment Overview](#deployment-overview)
- [Development Guide](#development-guide)
- [Environment Configuration](#environment-configuration)
- [Database Setup](#database-setup)
- [Docker Deployment](#docker-deployment)
- [Cloud Deployment](#cloud-deployment)
- [Production Considerations](#production-considerations)
- [Monitoring and Logging](#monitoring-and-logging)
- [Security Configuration](#security-configuration)
- [Backup and Recovery](#backup-and-recovery)
- [Troubleshooting](#troubleshooting)

## Deployment Overview

PandaPocket supports multiple deployment strategies:

- **Docker Containers**: Recommended for most deployments
- **Binary Deployment**: Direct binary execution
- **Cloud Platforms**: AWS, GCP, Azure, Heroku
- **Traditional Servers**: VPS, dedicated servers

### System Requirements

#### Minimum Requirements
- **CPU**: 1 vCPU
- **RAM**: 512 MB
- **Storage**: 1 GB
- **OS**: Linux (Ubuntu 20.04+, CentOS 8+), macOS, Windows

#### Recommended Requirements
- **CPU**: 2+ vCPUs
- **RAM**: 2+ GB
- **Storage**: 10+ GB SSD
- **OS**: Linux (Ubuntu 22.04 LTS)

## Development Guide

This section provides comprehensive guidance for developers working on the PandaPocket backend, including local development setup, API versioning, testing strategies, and best practices.

### Development Environment Setup

#### Prerequisites

**Required Software**:
- Go 1.23+ (latest stable version)
- PostgreSQL 14+ or Docker
- Git
- Make (optional, for build automation)

**Recommended Tools**:
- VS Code with Go extension
- Postman or Insomnia for API testing
- pgAdmin for database management
- Docker Desktop for containerized development

#### Local Development Setup

**1. Clone and Setup Repository**:
```bash
# Clone the repository
git clone <repository-url>
cd PandaPocket/backend

# Install dependencies
go mod download

# Copy environment configuration
cp .env.example .env
```

**2. Environment Configuration**:
```bash
# .env file for development
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=panda_pocket_dev
DB_PASSWORD=dev_password
DB_NAME=panda_pocket_dev
DB_SSL_MODE=disable

JWT_SECRET=dev-jwt-secret-key-change-in-production
JWT_EXPIRY=24h

PORT=8080
GIN_MODE=debug
HOST=localhost

CORS_ORIGINS=http://localhost:3000,http://localhost:3001

LOG_LEVEL=debug
LOG_FORMAT=text

BCRYPT_COST=4  # Lower cost for development
```

**3. Database Setup**:
```bash
# Using Docker (recommended)
docker run --name panda-pocket-postgres \
  -e POSTGRES_DB=panda_pocket_dev \
  -e POSTGRES_USER=panda_pocket_dev \
  -e POSTGRES_PASSWORD=dev_password \
  -p 5432:5432 \
  -d postgres:15-alpine

# Or using local PostgreSQL
createdb panda_pocket_dev
```

**4. Run Database Migrations**:
```bash
# Run the currency setup script
go run scripts/add_currencies.go

# Or run the SQL script directly
psql -h localhost -U panda_pocket_dev -d panda_pocket_dev -f scripts/add_currencies.sql
```

**5. Start Development Server**:
```bash
# Run the application
go run main.go

# Or with hot reload (requires air)
air
```

### API Versioning Development

#### Version Structure

The API follows a versioning strategy where each version is maintained independently:

```
/api/v100/transactions  # Version 1.0.0 (Legacy)
/api/v110/transactions  # Version 1.1.0 (Previous)
/api/v120/transactions  # Version 1.2.0 (Latest)
```

#### Adding New Features

**1. Create Version-Specific Handlers**:
```go
// internal/interfaces/http/handlers/v120/finance_handlers.go
package v120

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type FinanceHandlersV120 struct {
    // Include all necessary use cases
    createTransactionUseCase  *finance.CreateTransactionUseCase
    getTransactionsUseCase    *finance.GetTransactionsUseCase
    // ... other use cases
}

// New features specific to v120
func (h *FinanceHandlersV120) GetTransactionsWithAnalytics(c *gin.Context) {
    // Implementation with new analytics features
    userID := c.GetInt("user_id")
    
    // New v120-specific logic
    response, err := h.getTransactionsUseCase.Execute(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
        return
    }
    
    // Enhanced response with analytics
    c.JSON(http.StatusOK, gin.H{
        "transactions": response.Transactions,
        "analytics": gin.H{
            "total_amount": response.TotalAmount,
            "category_breakdown": response.CategoryBreakdown,
            "monthly_trends": response.MonthlyTrends,
        },
    })
}
```

**2. Update Route Configuration**:
```go
// internal/application/app.go
func (app *App) SetupRoutes() *gin.Engine {
    r := gin.Default()
    
    // Version middleware
    versionMiddleware := middleware.NewVersionMiddleware()
    r.Use(versionMiddleware.ExtractVersion())
    
    // Versioned routes
    versioned := r.Group("/api")
    {
        // v120 routes (latest)
        v120 := versioned.Group("/v120")
        {
            v120Handlers := handlers.NewFinanceHandlersV120(
                app.createTransactionUseCase,
                app.getTransactionsUseCase,
                // ... other use cases
            )
            
            protected := v120.Group("")
            protected.Use(app.AuthMiddleware.RequireAuth())
            {
                protected.GET("/transactions", v120Handlers.GetTransactionsWithAnalytics)
                protected.POST("/transactions", v120Handlers.CreateTransaction)
                // ... other v120 routes
            }
        }
        
        // v110 routes (previous version)
        v110 := versioned.Group("/v110")
        {
            // v110-specific handlers
        }
        
        // v100 routes (legacy)
        v100 := versioned.Group("/v100")
        {
            // v100-specific handlers
        }
    }
    
    return r
}
```

#### Backward Compatibility

**1. Maintain Previous Versions**:
```go
// internal/interfaces/http/handlers/v100/finance_handlers.go
package v100

// Legacy handlers maintain original functionality
func (h *FinanceHandlersV100) GetTransactions(c *gin.Context) {
    // Original implementation without new features
    userID := c.GetInt("user_id")
    
    response, err := h.getTransactionsUseCase.Execute(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
        return
    }
    
    // Simple response without analytics
    c.JSON(http.StatusOK, response.Transactions)
}
```

**2. Deprecation Handling**:
```go
// internal/interfaces/http/middleware/deprecation_middleware.go
func DeprecationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        version := c.GetString("api_version")
        
        if version == "v100" {
            c.Header("X-API-Deprecated", "true")
            c.Header("X-API-Sunset-Date", "2024-06-01")
            c.Header("X-API-Upgrade-URL", "https://docs.pandapocket.com/upgrade")
        }
        
        c.Next()
    }
}
```

### Testing Strategy

#### Unit Testing

**1. Handler Testing**:
```go
// internal/interfaces/http/handlers/v120/finance_handlers_test.go
package v120

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestGetTransactionsWithAnalytics(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    // Mock use case
    mockUseCase := &MockGetTransactionsUseCase{}
    handler := &FinanceHandlersV120{
        getTransactionsUseCase: mockUseCase,
    }
    
    // Setup route
    router.GET("/transactions", handler.GetTransactionsWithAnalytics)
    
    // Test request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/transactions", nil)
    req.Header.Set("user_id", "1")
    
    router.ServeHTTP(w, req)
    
    // Assertions
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.Contains(t, response, "transactions")
    assert.Contains(t, response, "analytics")
}
```

**2. Use Case Testing**:
```go
// internal/application/finance/get_transactions_use_case_test.go
package finance

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestGetTransactionsUseCase_Execute(t *testing.T) {
    // Setup
    mockRepo := &MockTransactionRepository{}
    useCase := NewGetTransactionsUseCase(mockRepo)
    
    // Mock expectations
    mockRepo.On("GetByUserID", mock.Anything, 1).Return([]Transaction{}, nil)
    
    // Execute
    result, err := useCase.Execute(context.Background(), 1)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
```

#### Integration Testing

**1. API Integration Tests**:
```go
// tests/integration/api_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestAPIVersioning(t *testing.T) {
    // Setup test server
    gin.SetMode(gin.TestMode)
    router := setupTestRouter()
    
    tests := []struct {
        name     string
        version  string
        endpoint string
        expected int
    }{
        {
            name:     "v120 transactions",
            version:  "v120",
            endpoint: "/api/v120/transactions",
            expected: http.StatusOK,
        },
        {
            name:     "v100 transactions (deprecated)",
            version:  "v100",
            endpoint: "/api/v100/transactions",
            expected: http.StatusOK,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := httptest.NewRecorder()
            req, _ := http.NewRequest("GET", tt.endpoint, nil)
            req.Header.Set("Authorization", "Bearer test-token")
            
            router.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expected, w.Code)
            
            if tt.version == "v100" {
                assert.Equal(t, "true", w.Header().Get("X-API-Deprecated"))
            }
        })
    }
}
```

**2. Database Integration Tests**:
```go
// tests/integration/database_test.go
package integration

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func TestDatabaseOperations(t *testing.T) {
    // Setup in-memory database
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    assert.NoError(t, err)
    
    // Run migrations
    err = db.AutoMigrate(&Transaction{}, &Category{}, &User{})
    assert.NoError(t, err)
    
    // Test operations
    transaction := &Transaction{
        UserID:     1,
        CategoryID: 1,
        Amount:     100.0,
        Type:       "expense",
    }
    
    result := db.Create(transaction)
    assert.NoError(t, result.Error)
    assert.NotZero(t, transaction.ID)
}
```

#### End-to-End Testing

**1. API End-to-End Tests**:
```go
// tests/e2e/api_e2e_test.go
package e2e

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCompleteUserFlow(t *testing.T) {
    baseURL := "http://localhost:8080"
    
    // 1. Register user
    registerData := map[string]string{
        "email":    "test@example.com",
        "password": "password123",
    }
    
    resp, err := http.Post(baseURL+"/api/auth/register", "application/json", 
        bytes.NewBuffer(mustMarshal(registerData)))
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    
    // 2. Login
    loginData := map[string]string{
        "email":    "test@example.com",
        "password": "password123",
    }
    
    resp, err = http.Post(baseURL+"/api/auth/login", "application/json",
        bytes.NewBuffer(mustMarshal(loginData)))
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // 3. Create transaction
    transactionData := map[string]interface{}{
        "category_id": 1,
        "amount":      100.0,
        "description": "Test transaction",
        "date":        "2024-01-01",
    }
    
    req, _ := http.NewRequest("POST", baseURL+"/api/v120/transactions", 
        bytes.NewBuffer(mustMarshal(transactionData)))
    req.Header.Set("Authorization", "Bearer "+extractToken(resp))
    
    client := &http.Client{}
    resp, err = client.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
```

### Development Best Practices

#### Code Organization

**1. Directory Structure**:
```
internal/
├── application/
│   ├── finance/
│   │   ├── create_transaction_use_case.go
│   │   ├── get_transactions_use_case.go
│   │   └── ...
│   └── identity/
├── domain/
│   ├── finance/
│   │   ├── transaction.go
│   │   ├── repository.go
│   │   └── service.go
│   └── identity/
├── infrastructure/
│   └── database/
├── interfaces/
│   └── http/
│       ├── handlers/
│       │   ├── v100/
│       │   ├── v110/
│       │   └── v120/
│       └── middleware/
└── versioning/
```

**2. Naming Conventions**:
- Use descriptive names for functions and variables
- Follow Go naming conventions (PascalCase for exported, camelCase for private)
- Use meaningful package names
- Include version suffixes for version-specific code

#### Error Handling

**1. Structured Error Responses**:
```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (h *FinanceHandlers) CreateTransaction(c *gin.Context) {
    // ... validation logic
    
    if err != nil {
        c.JSON(http.StatusBadRequest, APIError{
            Code:    "VALIDATION_ERROR",
            Message: "Invalid request data",
            Details: err.Error(),
        })
        return
    }
    
    // ... business logic
}
```

**2. Error Logging**:
```go
import "log/slog"

func (h *FinanceHandlers) CreateTransaction(c *gin.Context) {
    // ... business logic
    
    if err != nil {
        slog.Error("Failed to create transaction",
            "user_id", userID,
            "error", err.Error(),
            "request_id", c.GetString("request_id"),
        )
        
        c.JSON(http.StatusInternalServerError, APIError{
            Code:    "INTERNAL_ERROR",
            Message: "Failed to create transaction",
        })
        return
    }
}
```

#### Performance Optimization

**1. Database Query Optimization**:
```go
// Use specific field selection
func (r *GormTransactionRepository) GetByUserID(ctx context.Context, userID int) ([]Transaction, error) {
    var transactions []Transaction
    
    err := r.db.Select("id, user_id, category_id, amount, description, date, type").
        Where("user_id = ?", userID).
        Order("date DESC").
        Find(&transactions).Error
        
    return transactions, err
}
```

**2. Caching Strategy**:
```go
// internal/infrastructure/cache/redis_cache.go
package cache

import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    
    return json.Unmarshal([]byte(val), dest)
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return r.client.Set(ctx, key, data, expiration).Err()
}
```

#### Security Best Practices

**1. Input Validation**:
```go
import "github.com/go-playground/validator/v10"

type CreateTransactionRequest struct {
    CategoryID  int     `json:"category_id" binding:"required,min=1"`
    Amount      float64 `json:"amount" binding:"required,gt=0"`
    Description string  `json:"description" binding:"required,min=1,max=255"`
    Date        string  `json:"date" binding:"required,datetime=2006-01-02"`
}

func (h *FinanceHandlers) CreateTransaction(c *gin.Context) {
    var req CreateTransactionRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Additional validation
    if err := validateTransactionRequest(req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // ... business logic
}
```

**2. Rate Limiting**:
```go
import "golang.org/x/time/rate"

func RateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(time.Minute), 100)
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Development Workflow

#### Git Workflow

**1. Feature Branch Strategy**:
```bash
# Create feature branch
git checkout -b feature/api-versioning

# Make changes
git add .
git commit -m "feat: implement API versioning middleware"

# Push and create PR
git push origin feature/api-versioning
```

**2. Commit Message Convention**:
```
feat: add new feature
fix: bug fix
docs: documentation changes
style: code formatting
refactor: code refactoring
test: add tests
chore: maintenance tasks
```

#### Code Review Process

**1. Pull Request Template**:
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

**2. Review Checklist**:
- Code follows Go best practices
- Proper error handling
- Adequate test coverage
- Documentation updated
- No security vulnerabilities
- Performance considerations

### Debugging and Troubleshooting

#### Local Debugging

**1. Enable Debug Mode**:
```bash
export GIN_MODE=debug
export LOG_LEVEL=debug
go run main.go
```

**2. Database Debugging**:
```go
// Enable SQL logging
db.Logger = logger.Default.LogMode(logger.Info)

// Add debug middleware
func DebugMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        log.Printf("Request: %s %s - Status: %d - Duration: %v",
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            time.Since(start),
        )
    }
}
```

#### Common Issues and Solutions

**1. Database Connection Issues**:
```bash
# Check database status
docker ps | grep postgres

# Check connection
psql -h localhost -U panda_pocket_dev -d panda_pocket_dev -c "SELECT 1;"

# Reset database
docker-compose down
docker-compose up -d
```

**2. Port Conflicts**:
```bash
# Check port usage
lsof -i :8080

# Kill process using port
kill -9 $(lsof -t -i:8080)
```

**3. Environment Variable Issues**:
```bash
# Check environment variables
env | grep DB_

# Load from .env file
source .env
```

### Performance Monitoring

#### Application Metrics

**1. Response Time Monitoring**:
```go
func ResponseTimeMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        
        // Log slow requests
        if duration > 1*time.Second {
            slog.Warn("Slow request detected",
                "method", c.Request.Method,
                "path", c.Request.URL.Path,
                "duration", duration,
            )
        }
    }
}
```

**2. Database Query Monitoring**:
```go
// Enable query logging
db.Logger = logger.Default.LogMode(logger.Info)

// Add query timing
func (r *GormTransactionRepository) GetByUserID(ctx context.Context, userID int) ([]Transaction, error) {
    start := time.Now()
    defer func() {
        slog.Info("Database query completed",
            "query", "GetByUserID",
            "duration", time.Since(start),
        )
    }()
    
    // ... query logic
}
```

## Environment Configuration

### Environment Variables

Create a `.env` file or set environment variables for production:

```bash
# Database Configuration
DB_TYPE=postgres
DB_HOST=your-db-host.com
DB_PORT=5432
DB_USER=panda_pocket_user
DB_PASSWORD=secure_password_here
DB_NAME=panda_pocket_prod
DB_SSL_MODE=require

# JWT Configuration
JWT_SECRET=your-super-secure-jwt-secret-key-here
JWT_EXPIRY=24h

# Server Configuration
PORT=8080
GIN_MODE=release
HOST=0.0.0.0

# CORS Configuration
CORS_ORIGINS=https://your-frontend-domain.com,https://admin.your-domain.com

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Security
BCRYPT_COST=12
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

### Configuration Validation

The application validates required environment variables on startup:

```go
func validateConfig() error {
    required := []string{
        "DB_TYPE",
        "JWT_SECRET",
    }
    
    for _, key := range required {
        if os.Getenv(key) == "" {
            return fmt.Errorf("required environment variable %s is not set", key)
        }
    }
    
    return nil
}
```

## Database Setup

### PostgreSQL Production Setup

#### 1. Install PostgreSQL

**Ubuntu/Debian**:
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**CentOS/RHEL**:
```bash
sudo yum install postgresql-server postgresql-contrib
sudo postgresql-setup initdb
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### 2. Create Database and User

```bash
# Switch to postgres user
sudo -u postgres psql

-- Create database
CREATE DATABASE panda_pocket_prod;

-- Create user
CREATE USER panda_pocket_user WITH PASSWORD 'secure_password_here';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE panda_pocket_prod TO panda_pocket_user;

-- Connect to the database and grant schema privileges
\c panda_pocket_prod
GRANT ALL ON SCHEMA public TO panda_pocket_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO panda_pocket_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO panda_pocket_user;

-- Exit
\q
```

#### 3. Configure PostgreSQL

Edit `/etc/postgresql/14/main/postgresql.conf`:

```conf
# Connection settings
listen_addresses = 'localhost'
port = 5432
max_connections = 100

# Memory settings
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB

# Logging
log_destination = 'stderr'
logging_collector = on
log_directory = '/var/log/postgresql'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_statement = 'mod'
log_min_duration_statement = 1000

# Security
ssl = on
```

Edit `/etc/postgresql/14/main/pg_hba.conf`:

```conf
# Local connections
local   all             panda_pocket_user                    md5
host    all             panda_pocket_user    127.0.0.1/32   md5
host    all             panda_pocket_user    ::1/128        md5

# SSL connections
hostssl all             panda_pocket_user    0.0.0.0/0       md5
```

#### 4. Restart PostgreSQL

```bash
sudo systemctl restart postgresql
```

### Database Migration

#### Manual Migration
```bash
# Set environment variables
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=panda_pocket_user
export DB_PASSWORD=secure_password_here
export DB_NAME=panda_pocket_prod

# Run migration
go run cmd/migrate/main.go
```

#### Automated Migration
```bash
# Create migration script
cat > migrate.sh << 'EOF'
#!/bin/bash
set -e

echo "Starting database migration..."

# Wait for database to be ready
until pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do
    echo "Waiting for database..."
    sleep 2
done

# Run migration
go run cmd/migrate/main.go

echo "Migration completed successfully"
EOF

chmod +x migrate.sh
./migrate.sh
```

## Docker Deployment

### Dockerfile

Create a production-ready Dockerfile:

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Create non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
```

### Docker Compose

Create `docker-compose.yml` for production:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_TYPE=postgres
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=panda_pocket_user
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=panda_pocket_prod
      - JWT_SECRET=${JWT_SECRET}
      - GIN_MODE=release
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - panda-pocket-network

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=panda_pocket_prod
      - POSTGRES_USER=panda_pocket_user
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U panda_pocket_user -d panda_pocket_prod"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - panda-pocket-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    restart: unless-stopped
    networks:
      - panda-pocket-network

volumes:
  postgres_data:

networks:
  panda-pocket-network:
    driver: bridge
```

### Build and Deploy

```bash
# Build the image
docker build -t panda-pocket:latest .

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f app

# Scale the application
docker-compose up -d --scale app=3
```

## Cloud Deployment

### AWS Deployment

#### 1. EC2 Deployment

**Launch EC2 Instance**:
```bash
# Create security group
aws ec2 create-security-group \
    --group-name panda-pocket-sg \
    --description "Security group for PandaPocket"

# Add rules
aws ec2 authorize-security-group-ingress \
    --group-name panda-pocket-sg \
    --protocol tcp \
    --port 22 \
    --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
    --group-name panda-pocket-sg \
    --protocol tcp \
    --port 80 \
    --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
    --group-name panda-pocket-sg \
    --protocol tcp \
    --port 443 \
    --cidr 0.0.0.0/0
```

**Deploy Application**:
```bash
# Connect to instance
ssh -i your-key.pem ubuntu@your-instance-ip

# Install Docker
sudo apt update
sudo apt install docker.io docker-compose
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker ubuntu

# Clone and deploy
git clone <your-repo>
cd PandaPocket/backend
docker-compose up -d
```

#### 2. ECS Deployment

**Create ECS Task Definition**:
```json
{
  "family": "panda-pocket",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "panda-pocket",
      "image": "your-account.dkr.ecr.region.amazonaws.com/panda-pocket:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DB_TYPE",
          "value": "postgres"
        },
        {
          "name": "GIN_MODE",
          "value": "release"
        }
      ],
      "secrets": [
        {
          "name": "JWT_SECRET",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:panda-pocket/jwt-secret"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/panda-pocket",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

### Google Cloud Platform

#### 1. Cloud Run Deployment

```bash
# Build and push to GCR
gcloud builds submit --tag gcr.io/PROJECT-ID/panda-pocket

# Deploy to Cloud Run
gcloud run deploy panda-pocket \
    --image gcr.io/PROJECT-ID/panda-pocket \
    --platform managed \
    --region us-central1 \
    --allow-unauthenticated \
    --set-env-vars DB_TYPE=postgres,DB_HOST=your-db-host,GIN_MODE=release
```

#### 2. GKE Deployment

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: panda-pocket
spec:
  replicas: 3
  selector:
    matchLabels:
      app: panda-pocket
  template:
    metadata:
      labels:
        app: panda-pocket
    spec:
      containers:
      - name: panda-pocket
        image: gcr.io/PROJECT-ID/panda-pocket:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_TYPE
          value: "postgres"
        - name: GIN_MODE
          value: "release"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: panda-pocket-secrets
              key: jwt-secret
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: panda-pocket-service
spec:
  selector:
    app: panda-pocket
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### Heroku Deployment

#### 1. Heroku Setup

```bash
# Install Heroku CLI
# Create Procfile
echo "web: ./panda-pocket" > Procfile

# Create app
heroku create panda-pocket-api

# Set environment variables
heroku config:set DB_TYPE=postgres
heroku config:set JWT_SECRET=your-secret-key
heroku config:set GIN_MODE=release

# Add PostgreSQL addon
heroku addons:create heroku-postgresql:hobby-dev

# Deploy
git push heroku main
```

#### 2. Heroku Configuration

```bash
# Scale the application
heroku ps:scale web=2

# View logs
heroku logs --tail

# Run database migration
heroku run go run cmd/migrate/main.go
```

## Production Considerations

### Performance Optimization

#### 1. Application Configuration

```go
// main.go - Production optimizations
func main() {
    // Set GIN mode for production
    gin.SetMode(gin.ReleaseMode)
    
    // Initialize database with connection pooling
    db, err := database.InitDB()
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer db.Close()
    
    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    // Create application
    app := application.NewApp(db)
    
    // Setup routes with middleware
    router := app.SetupRoutes()
    
    // Add production middleware
    router.Use(gin.Recovery())
    router.Use(rateLimitMiddleware())
    router.Use(securityHeadersMiddleware())
    
    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    router.Run(":" + port)
}
```

#### 2. Database Optimization

```sql
-- Create indexes for better performance
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
CREATE INDEX idx_transactions_type ON transactions(type);

-- Analyze tables for query optimization
ANALYZE transactions;
ANALYZE categories;
ANALYZE users;
```

### Load Balancing

#### Nginx Configuration

```nginx
# nginx.conf
upstream panda_pocket_backend {
    server app1:8080;
    server app2:8080;
    server app3:8080;
}

server {
    listen 80;
    server_name your-domain.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL configuration
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    
    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";
    
    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;
    
    location / {
        proxy_pass http://panda_pocket_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # Health check endpoint
    location /health {
        proxy_pass http://panda_pocket_backend/health;
        access_log off;
    }
}
```

## Monitoring and Logging

### Application Monitoring

#### 1. Health Checks

```go
// health.go
func healthCheck(c *gin.Context) {
    // Check database connectivity
    if err := checkDatabaseHealth(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unhealthy",
            "error":  err.Error(),
        })
        return
    }
    
    // Check external services
    if err := checkExternalServices(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "degraded",
            "error":  err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":    "healthy",
        "timestamp": time.Now().UTC(),
        "version":   "1.0.0",
    })
}

func checkDatabaseHealth() error {
    // Test database connection
    db := getDB()
    return db.Ping()
}
```

#### 2. Metrics Collection

```go
// metrics.go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}

func metricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
    }
}
```

### Logging Configuration

#### 1. Structured Logging

```go
// logger.go
import "log/slog"

func setupLogger() *slog.Logger {
    level := os.Getenv("LOG_LEVEL")
    if level == "" {
        level = "info"
    }
    
    var logLevel slog.Level
    switch level {
    case "debug":
        logLevel = slog.LevelDebug
    case "info":
        logLevel = slog.LevelInfo
    case "warn":
        logLevel = slog.LevelWarn
    case "error":
        logLevel = slog.LevelError
    default:
        logLevel = slog.LevelInfo
    }
    
    format := os.Getenv("LOG_FORMAT")
    if format == "json" {
        return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: logLevel,
        }))
    }
    
    return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: logLevel,
    }))
}
```

#### 2. Log Aggregation

**ELK Stack Configuration**:

```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.8.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  logstash:
    image: docker.elastic.co/logstash/logstash:8.8.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    ports:
      - "5044:5044"
    depends_on:
      - elasticsearch

  kibana:
    image: docker.elastic.co/kibana/kibana:8.8.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch

volumes:
  elasticsearch_data:
```

## Security Configuration

### 1. SSL/TLS Configuration

#### Let's Encrypt with Certbot

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain SSL certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

#### Self-Signed Certificate (Development)

```bash
# Generate private key
openssl genrsa -out key.pem 2048

# Generate certificate
openssl req -new -x509 -key key.pem -out cert.pem -days 365 \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
```

### 2. Security Headers

```go
// security.go
func securityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Content-Security-Policy", "default-src 'self'")
        
        c.Next()
    }
}
```

### 3. Rate Limiting

```go
// ratelimit.go
import "golang.org/x/time/rate"

func rateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(time.Minute), 100)
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## Backup and Recovery

### 1. Database Backup

#### Automated Backup Script

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backups"
DB_NAME="panda_pocket_prod"
DB_USER="panda_pocket_user"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/panda_pocket_$DATE.sql"

# Create backup directory
mkdir -p $BACKUP_DIR

# Create database backup
pg_dump -h localhost -U $DB_USER -d $DB_NAME > $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Upload to S3 (optional)
aws s3 cp $BACKUP_FILE.gz s3://your-backup-bucket/database/

# Clean old backups (keep last 30 days)
find $BACKUP_DIR -name "panda_pocket_*.sql.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_FILE.gz"
```

#### Cron Job

```bash
# Add to crontab
crontab -e

# Daily backup at 2 AM
0 2 * * * /path/to/backup.sh
```

### 2. Application Backup

```bash
#!/bin/bash
# app-backup.sh

APP_DIR="/opt/panda-pocket"
BACKUP_DIR="/backups/app"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup
tar -czf $BACKUP_DIR/panda-pocket_$DATE.tar.gz -C $APP_DIR .

# Upload to S3
aws s3 cp $BACKUP_DIR/panda-pocket_$DATE.tar.gz s3://your-backup-bucket/app/

echo "Application backup completed"
```

### 3. Recovery Procedures

#### Database Recovery

```bash
# Restore from backup
gunzip -c panda_pocket_20240115_020000.sql.gz | psql -h localhost -U panda_pocket_user -d panda_pocket_prod

# Verify restoration
psql -h localhost -U panda_pocket_user -d panda_pocket_prod -c "SELECT COUNT(*) FROM transactions;"
```

#### Application Recovery

```bash
# Stop application
sudo systemctl stop panda-pocket

# Restore from backup
tar -xzf panda-pocket_20240115_020000.tar.gz -C /opt/panda-pocket/

# Start application
sudo systemctl start panda-pocket
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Issues

```bash
# Check database status
sudo systemctl status postgresql

# Check database logs
sudo tail -f /var/log/postgresql/postgresql-14-main.log

# Test connection
psql -h localhost -U panda_pocket_user -d panda_pocket_prod -c "SELECT 1;"
```

#### 2. Application Startup Issues

```bash
# Check application logs
journalctl -u panda-pocket -f

# Check environment variables
systemctl show-environment

# Test configuration
./panda-pocket --config-test
```

#### 3. Performance Issues

```bash
# Check system resources
htop
iostat -x 1
free -h

# Check database performance
psql -h localhost -U panda_pocket_user -d panda_pocket_prod -c "
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"
```

### Debugging Tools

#### 1. Application Debugging

```go
// Enable debug mode
export GIN_MODE=debug
export LOG_LEVEL=debug

// Add debug middleware
func debugMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        log.Printf("Request: %s %s - Status: %d - Duration: %v",
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            time.Since(start),
        )
    }
}
```

#### 2. Database Debugging

```sql
-- Enable query logging
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_min_duration_statement = 0;
SELECT pg_reload_conf();

-- Monitor active connections
SELECT * FROM pg_stat_activity WHERE state = 'active';

-- Check slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
WHERE mean_time > 1000 
ORDER BY mean_time DESC;
```

### Health Check Endpoints

```go
// health.go
func setupHealthChecks(router *gin.Engine) {
    router.GET("/health", healthCheck)
    router.GET("/health/ready", readinessCheck)
    router.GET("/health/live", livenessCheck)
}

func readinessCheck(c *gin.Context) {
    // Check if application is ready to serve requests
    if !isDatabaseReady() {
        c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func livenessCheck(c *gin.Context) {
    // Check if application is alive
    c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
```

---

This deployment guide provides comprehensive instructions for deploying PandaPocket in various environments. For additional support or questions, please refer to the project documentation or create an issue in the repository.



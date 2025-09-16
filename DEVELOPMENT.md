# Development Guide

This guide provides comprehensive information for developers working on the PandaPocket backend project.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Development Workflow](#development-workflow)
- [Testing Guidelines](#testing-guidelines)
- [Database Development](#database-development)
- [API Development](#api-development)
- [Debugging](#debugging)
- [Performance Considerations](#performance-considerations)
- [Contributing Guidelines](#contributing-guidelines)

## Development Environment Setup

### Prerequisites

- **Go 1.23.0+**: [Download and install Go](https://golang.org/dl/)
- **Git**: For version control
- **PostgreSQL**: For database (primary database)
- **Docker** (optional): For containerized development

### IDE/Editor Setup

#### Recommended IDEs
- **VS Code** with Go extension
- **GoLand** (JetBrains)
- **Vim/Neovim** with Go plugins

#### VS Code Extensions
```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "bradlc.vscode-tailwindcss",
    "ms-vscode.vscode-typescript-next"
  ]
}
```

### Environment Configuration

#### 1. Clone the Repository
```bash
git clone <repository-url>
cd PandaPocket/backend
```

#### 2. Install Dependencies
```bash
go mod download
```

#### 3. Environment Variables
Create a `.env` file for local development:

```bash
# Database Configuration
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=panda_pocket

# JWT Configuration
JWT_SECRET=your-secret-key-here
JWT_EXPIRY=24h

# Server Configuration
PORT=8080
GIN_MODE=debug
```

#### 4. Database Setup

**PostgreSQL (Primary Database)**:
```bash
# Using Docker Compose (Recommended)
docker compose up -d postgres

# Or using Docker directly
docker run --name panda-pocket-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=panda_pocket \
  -p 5432:5432 \
  -d postgres:15

# Or using local PostgreSQL installation
createdb panda_pocket
```

#### 5. Run the Application
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## Project Structure

```
panda-pocket/backend/
├── cmd/                          # Application entry points
│   └── migrate/                  # Database migration tool
│       └── main.go
├── internal/                     # Private application code
│   ├── application/              # Application layer (Use Cases)
│   │   ├── app.go               # Main application setup
│   │   ├── finance/             # Financial use cases
│   │   │   ├── create_budget_use_case.go
│   │   │   ├── create_category_use_case.go
│   │   │   ├── create_transaction_use_case.go
│   │   │   ├── get_analytics_use_case.go
│   │   │   ├── get_budgets_use_case.go
│   │   │   ├── get_categories_use_case.go
│   │   │   └── get_transactions_use_case.go
│   │   └── identity/            # Identity use cases
│   │       ├── login_user_use_case.go
│   │       ├── register_user_use_case.go
│   │       └── token_service.go
│   ├── domain/                   # Domain layer (Business Logic)
│   │   ├── finance/             # Financial domain
│   │   │   ├── budget.go
│   │   │   ├── category.go
│   │   │   ├── currency.go
│   │   │   ├── currency_service.go
│   │   │   ├── recurring_transaction.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── transaction.go
│   │   └── identity/            # Identity domain
│   │       ├── repository.go
│   │       ├── service.go
│   │       └── user.go
│   ├── infrastructure/           # Infrastructure layer
│   │   └── database/            # Database implementations
│   │       ├── init.go
│   │       ├── postgres_budget_repository.go
│   │       ├── sqlite_category_repository.go
│   │       ├── sqlite_currency_repository.go
│   │       ├── sqlite_transaction_repository.go
│   │       └── sqlite_user_repository.go
│   └── interfaces/              # Interface layer
│       └── http/                # HTTP interface
│           ├── handlers/        # HTTP handlers
│           │   ├── finance_handlers.go
│           │   └── identity_handlers.go
│           └── middleware/      # HTTP middleware
│               └── auth_middleware.go
├── scripts/                     # Utility scripts
│   └── setup-postgres.sh
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── main.go                      # Application entry point
└── *.md                         # Documentation files
```

## Coding Standards

### Go Code Style

#### 1. Formatting
- Use `gofmt` for code formatting
- Use `goimports` for import organization
- Follow Go's official style guide

```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .
```

#### 2. Naming Conventions
- **Packages**: lowercase, single word (e.g., `finance`, `identity`)
- **Types**: PascalCase (e.g., `Transaction`, `UserID`)
- **Functions**: PascalCase for public, camelCase for private
- **Variables**: camelCase (e.g., `userID`, `transactionAmount`)
- **Constants**: PascalCase or UPPER_CASE

#### 3. Error Handling
```go
// Good: Explicit error handling
func CreateUser(email string) (*User, error) {
    if email == "" {
        return nil, errors.New("email cannot be empty")
    }
    // ... implementation
}

// Bad: Ignoring errors
func CreateUser(email string) *User {
    // ... implementation without error handling
}
```

#### 4. Documentation
- Document all public functions and types
- Use Go doc format
- Include examples for complex functions

```go
// User represents a user in the identity domain
type User struct {
    id        UserID
    email     Email
    password  PasswordHash
    createdAt time.Time
}

// NewUser creates a new user entity with the provided parameters
func NewUser(id UserID, email Email, password PasswordHash) *User {
    return &User{
        id:        id,
        email:     email,
        password:  password,
        createdAt: time.Now(),
    }
}
```

### Domain-Driven Design Guidelines

#### 1. Domain Layer
- Keep business logic in the domain layer
- Use value objects for immutable concepts
- Use entities for objects with identity
- Keep domain services stateless

```go
// Value Object Example
type Money struct {
    amount   float64
    currency CurrencyID
}

func NewMoney(amount float64, currency CurrencyID) (Money, error) {
    if amount < 0 {
        return Money{}, errors.New("amount cannot be negative")
    }
    return Money{amount: amount, currency: currency}, nil
}
```

#### 2. Application Layer
- Use cases should be single-purpose
- Coordinate between domain and infrastructure
- Handle transaction boundaries

```go
// Use Case Example
type CreateTransactionUseCase struct {
    transactionService TransactionService
    currencyService    CurrencyService
}

func (uc *CreateTransactionUseCase) Execute(req CreateTransactionRequest) (*Transaction, error) {
    // Validate request
    // Create domain objects
    // Execute business logic
    // Return result
}
```

#### 3. Infrastructure Layer
- Implement domain interfaces
- Handle external concerns (database, APIs)
- Keep infrastructure details isolated

```go
// Repository Implementation Example
type SQLiteTransactionRepository struct {
    db *sql.DB
}

func (r *SQLiteTransactionRepository) Create(transaction *Transaction) error {
    // Database-specific implementation
}
```

## Development Workflow

### 1. Feature Development

#### Branch Strategy
```bash
# Create feature branch
git checkout -b feature/add-budget-management

# Make changes and commit
git add .
git commit -m "feat: add budget creation use case"

# Push and create PR
git push origin feature/add-budget-management
```

#### Commit Message Convention
```
type(scope): description

feat: add new feature
fix: bug fix
docs: documentation changes
style: formatting changes
refactor: code refactoring
test: add or update tests
chore: maintenance tasks
```

### 2. Code Review Process

#### Before Submitting PR
- [ ] Code follows Go style guidelines
- [ ] All tests pass
- [ ] Documentation is updated
- [ ] No linting errors
- [ ] Feature is properly tested

#### Review Checklist
- [ ] Business logic is in domain layer
- [ ] Error handling is appropriate
- [ ] Code is readable and maintainable
- [ ] Performance considerations are addressed
- [ ] Security implications are considered

### 3. Testing Workflow

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/domain/finance

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

## Testing Guidelines

### 1. Unit Tests

#### Domain Layer Tests
```go
func TestNewMoney(t *testing.T) {
    tests := []struct {
        name      string
        amount    float64
        currency  CurrencyID
        wantError bool
    }{
        {
            name:      "valid amount",
            amount:    100.50,
            currency:  NewCurrencyID(1),
            wantError: false,
        },
        {
            name:      "negative amount",
            amount:    -10.0,
            currency:  NewCurrencyID(1),
            wantError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            money, err := NewMoney(tt.amount, tt.currency)
            if tt.wantError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.amount, money.Amount())
            }
        })
    }
}
```

#### Use Case Tests
```go
func TestCreateTransactionUseCase_Execute(t *testing.T) {
    // Setup mocks
    mockTransactionService := &MockTransactionService{}
    mockCurrencyService := &MockCurrencyService{}
    
    useCase := NewCreateTransactionUseCase(mockTransactionService, mockCurrencyService)
    
    // Test execution
    req := CreateTransactionRequest{
        UserID:      1,
        CategoryID:  1,
        Amount:      100.0,
        Description: "Test transaction",
        Date:        time.Now(),
        Type:        TransactionTypeExpense,
    }
    
    result, err := useCase.Execute(req)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, req.Amount, result.Amount().Amount())
}
```

### 2. Integration Tests

#### Repository Tests
```go
func TestSQLiteTransactionRepository_Create(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := NewSQLiteTransactionRepository(db)
    
    // Create test transaction
    transaction := createTestTransaction()
    
    // Test repository method
    err := repo.Create(transaction)
    
    // Assertions
    assert.NoError(t, err)
    
    // Verify data was stored
    retrieved, err := repo.GetByID(transaction.ID())
    assert.NoError(t, err)
    assert.Equal(t, transaction.Amount().Amount(), retrieved.Amount().Amount())
}
```

### 3. API Tests

#### HTTP Handler Tests
```go
func TestFinanceHandlers_CreateExpense(t *testing.T) {
    // Setup test server
    router := setupTestRouter(t)
    
    // Create test request
    reqBody := CreateExpenseRequest{
        CategoryID:  1,
        Amount:      50.0,
        Description: "Test expense",
        Date:        "2024-01-15",
    }
    
    req, _ := http.NewRequest("POST", "/api/expenses", 
        strings.NewReader(marshalJSON(t, reqBody)))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+testToken)
    
    // Execute request
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assertions
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response CreateExpenseResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, reqBody.Amount, response.Expense.Amount)
}
```

## Database Development

### 1. Schema Changes

#### Adding New Tables
```sql
-- Create migration file: migrations/001_add_budgets.sql
CREATE TABLE budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

#### Modifying Existing Tables
```sql
-- Create migration file: migrations/002_add_budget_id_to_transactions.sql
ALTER TABLE transactions ADD COLUMN budget_id INTEGER;
ALTER TABLE transactions ADD FOREIGN KEY (budget_id) REFERENCES budgets(id);
```

### 2. Repository Development

#### Interface Definition
```go
// In domain layer
type BudgetRepository interface {
    Save(ctx context.Context, budget *Budget) error
    FindByID(ctx context.Context, id BudgetID) (*Budget, error)
    FindByUserID(ctx context.Context, userID UserID) ([]*Budget, error)
    FindByUserIDAndCategory(ctx context.Context, userID UserID, categoryID CategoryID) ([]*Budget, error)
    FindActiveByUserID(ctx context.Context, userID UserID) ([]*Budget, error)
    Delete(ctx context.Context, id BudgetID) error
}
```

#### Implementation
```go
// In infrastructure layer
type PostgresBudgetRepository struct {
    db *sql.DB
}

func NewPostgresBudgetRepository(db *sql.DB) *PostgresBudgetRepository {
    return &PostgresBudgetRepository{db: db}
}

func (r *PostgresBudgetRepository) Save(ctx context.Context, budget *Budget) error {
    if budget.ID().Value() == 0 {
        // Insert new budget
        query := `INSERT INTO budgets (user_id, category_id, amount, period, start_date, end_date, created_at) 
                  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
        
        var id int
        err := r.db.QueryRowContext(ctx, query,
            budget.UserID().Value(),
            budget.CategoryID().Value(),
            budget.Amount().Amount(),
            string(budget.Period()),
            budget.StartDate(),
            budget.EndDate(),
            budget.CreatedAt(),
        ).Scan(&id)
        
        if err != nil {
            return err
        }
        
        // Note: In a real implementation, you'd want to handle the ID assignment properly
        _ = id
    } else {
        // Update existing budget
        query := `UPDATE budgets SET amount = $1, period = $2, start_date = $3, end_date = $4 
                  WHERE id = $5`
        _, err := r.db.ExecContext(ctx, query,
            budget.Amount().Amount(),
            string(budget.Period()),
            budget.StartDate(),
            budget.EndDate(),
            budget.ID().Value(),
        )
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## API Development

### 1. New Features Added

#### Analytics Endpoint
The analytics endpoint provides financial insights and balance calculations:

```go
// GET /api/analytics?period=monthly
type GetAnalyticsResponse struct {
    TotalIncome      float64 `json:"total_income"`
    TotalSpent       float64 `json:"total_spent"`
    NetAmount        float64 `json:"net_amount"`
    Period           string  `json:"period"`
    TransactionCount int     `json:"transaction_count"`
}
```

#### Budget Management
Budget functionality allows users to set spending limits:

```go
// POST /api/budgets
type CreateBudgetRequest struct {
    CategoryID int     `json:"category_id" binding:"required"`
    Amount     float64 `json:"amount" binding:"required,gt=0"`
    Period     string  `json:"period" binding:"required,oneof=weekly monthly yearly"`
    StartDate  string  `json:"start_date" binding:"required"`
}
```

### 2. Handler Development

#### Request/Response DTOs
```go
type CreateBudgetRequest struct {
    CategoryID  int    `json:"category_id" binding:"required"`
    Amount      float64 `json:"amount" binding:"required,min=0"`
    PeriodStart string `json:"period_start" binding:"required"`
    PeriodEnd   string `json:"period_end" binding:"required"`
}

type CreateBudgetResponse struct {
    Message string `json:"message"`
    Budget  BudgetDTO `json:"budget"`
}

type BudgetDTO struct {
    ID          int     `json:"id"`
    UserID      int     `json:"user_id"`
    CategoryID  int     `json:"category_id"`
    Amount      float64 `json:"amount"`
    PeriodStart string  `json:"period_start"`
    PeriodEnd   string  `json:"period_end"`
    CreatedAt   string  `json:"created_at"`
}
```

#### Handler Implementation
```go
func (h *FinanceHandlers) CreateBudget(c *gin.Context) {
    var req CreateBudgetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Get user from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    
    // Execute use case
    budget, err := h.createBudgetUseCase.Execute(CreateBudgetUseCaseRequest{
        UserID:      userID.(int),
        CategoryID:  req.CategoryID,
        Amount:      req.Amount,
        PeriodStart: req.PeriodStart,
        PeriodEnd:   req.PeriodEnd,
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // Convert to DTO
    budgetDTO := BudgetDTO{
        ID:          budget.ID().Value(),
        UserID:      budget.UserID().Value(),
        CategoryID:  budget.CategoryID().Value(),
        Amount:      budget.Amount().Amount(),
        PeriodStart: budget.PeriodStart().Format("2006-01-02"),
        PeriodEnd:   budget.PeriodEnd().Format("2006-01-02"),
        CreatedAt:   budget.CreatedAt().Format(time.RFC3339),
    }
    
    c.JSON(http.StatusCreated, CreateBudgetResponse{
        Message: "Budget created successfully",
        Budget:  budgetDTO,
    })
}
```

### 2. Middleware Development

#### Custom Middleware
```go
func LoggingMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
            param.ClientIP,
            param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
            param.Method,
            param.Path,
            param.Request.Proto,
            param.StatusCode,
            param.Latency,
            param.Request.UserAgent(),
            param.ErrorMessage,
        )
    })
}
```

## Debugging

### 1. Logging

#### Structured Logging
```go
import "log/slog"

func (h *FinanceHandlers) CreateExpense(c *gin.Context) {
    logger := slog.With(
        "handler", "CreateExpense",
        "user_id", c.GetString("userID"),
        "request_id", c.GetString("requestID"),
    )
    
    logger.Info("Creating new expense")
    
    // ... handler logic
    
    if err != nil {
        logger.Error("Failed to create expense", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }
    
    logger.Info("Expense created successfully", "expense_id", expense.ID().Value())
}
```

### 2. Debugging Tools

#### Delve Debugger
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug main.go

# Set breakpoints
(dlv) break internal/application/finance/create_transaction_use_case.go:25
(dlv) continue
```

#### VS Code Debugging
Create `.vscode/launch.json`:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/main.go"
        }
    ]
}
```

## Performance Considerations

### 1. Database Optimization

#### Query Optimization
```go
// Good: Use prepared statements with PostgreSQL
func (r *PostgresTransactionRepository) GetByUserID(ctx context.Context, userID UserID) ([]*Transaction, error) {
    query := `SELECT id, user_id, category_id, currency_id, amount, description, date, type, created_at 
              FROM transactions WHERE user_id = $1 ORDER BY date DESC`
    
    rows, err := r.db.QueryContext(ctx, query, userID.Value())
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    // ... process rows
}

// Bad: String concatenation (SQL injection risk)
func (r *PostgresTransactionRepository) GetByUserID(ctx context.Context, userID UserID) ([]*Transaction, error) {
    query := "SELECT * FROM transactions WHERE user_id = " + strconv.Itoa(userID.Value())
    // ... vulnerable to SQL injection
}
```

#### Connection Pooling
```go
func InitDB() (*sql.DB, error) {
    db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost/panda_pocket?sslmode=disable")
    if err != nil {
        return nil, err
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return db, nil
}
```

### 2. Memory Management

#### Avoiding Memory Leaks
```go
// Good: Proper resource cleanup
func (r *PostgresTransactionRepository) GetByUserID(ctx context.Context, userID UserID) ([]*Transaction, error) {
    rows, err := r.db.QueryContext(ctx, query, userID.Value())
    if err != nil {
        return nil, err
    }
    defer rows.Close() // Always close rows
    
    var transactions []*Transaction
    for rows.Next() {
        // ... process row
    }
    
    return transactions, rows.Err()
}
```

## Contributing Guidelines

### 1. Pull Request Process

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** following the coding standards
4. **Add tests** for new functionality
5. **Update documentation** if needed
6. **Run tests** and ensure they pass
7. **Commit your changes**: `git commit -m 'feat: add amazing feature'`
8. **Push to your branch**: `git push origin feature/amazing-feature`
9. **Create a Pull Request**

### 2. Code Review Process

- All code must be reviewed before merging
- Address review feedback promptly
- Ensure CI/CD checks pass
- Update documentation as needed

### 3. Issue Reporting

When reporting issues, include:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, etc.)
- Relevant logs or error messages

### 4. Feature Requests

When requesting features:
- Describe the use case
- Explain the expected behavior
- Consider implementation complexity
- Discuss with maintainers before starting work

---

This development guide should help you contribute effectively to the PandaPocket project. For questions or clarifications, please open an issue or reach out to the maintainers.


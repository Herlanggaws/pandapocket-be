# PandaPocket Architecture Documentation

## Overview

PandaPocket is built using **Domain-Driven Design (DDD)** principles with a clean architecture approach. This document outlines the architectural decisions, design patterns, and structure of the application.

## Architecture Principles

### 1. Domain-Driven Design (DDD)
- **Domain-First Approach**: Business logic is the core of the application
- **Ubiquitous Language**: Domain concepts are clearly defined and consistently used
- **Bounded Contexts**: Clear separation between different business domains
- **Value Objects**: Immutable objects that represent concepts without identity

### 2. Clean Architecture
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Separation of Concerns**: Each layer has a single responsibility
- **Testability**: Business logic can be tested independently of infrastructure
- **Independence**: Framework, database, and external concerns are isolated

## Layer Structure

```
┌─────────────────────────────────────────────────────────────┐
│                    Interface Layer                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │   HTTP Handlers │  │   Middleware    │  │   Routes    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                  Application Layer                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │   Use Cases     │  │   DTOs          │  │   Services  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │   Entities      │  │  Value Objects  │  │   Services  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │   Repositories  │  │   Database      │  │   External  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Domain Layer

The domain layer contains the core business logic and is independent of any external concerns.

### Bounded Contexts

#### 1. Identity Context
**Purpose**: Manages user authentication and authorization

**Entities**:
- `User`: Represents a user account with email and password

**Value Objects**:
- `UserID`: Unique identifier for users
- `Email`: Email address with validation
- `PasswordHash`: Hashed password representation

**Services**:
- `UserService`: Handles user-related business operations

#### 2. Finance Context
**Purpose**: Manages financial transactions, categories, and currencies

**Entities**:
- `Transaction`: Represents a financial transaction (expense or income)
- `Category`: Represents a transaction category
- `Currency`: Represents a supported currency
- `Budget`: Represents spending limits (planned)
- `RecurringTransaction`: Represents recurring financial transactions (planned)

**Value Objects**:
- `TransactionID`, `CategoryID`, `CurrencyID`: Unique identifiers
- `Money`: Monetary amount with currency
- `TransactionType`: Enum for expense/income types

**Services**:
- `TransactionService`: Handles transaction business logic
- `CategoryService`: Manages category operations
- `CurrencyService`: Handles currency-related operations

### Domain Rules

#### Transaction Rules
- Transactions must have a positive amount
- Transactions must belong to a valid category
- Transactions must have a valid currency
- Transaction dates cannot be in the future (business rule)

#### User Rules
- Email addresses must be unique
- Passwords must be hashed before storage
- Users can only access their own data

#### Category Rules
- Categories must have a name and type (expense/income)
- Categories must have a color for UI representation
- Default categories are system-managed

## Application Layer

The application layer orchestrates use cases and coordinates between the domain and infrastructure layers.

### Use Cases

#### Identity Use Cases
- `RegisterUserUseCase`: Creates new user accounts
- `LoginUserUseCase`: Authenticates users and generates tokens
- `TokenService`: Manages JWT token operations

#### Finance Use Cases
- `CreateTransactionUseCase`: Creates new financial transactions
- `GetTransactionsUseCase`: Retrieves user transactions
- `CreateCategoryUseCase`: Creates new categories
- `GetCategoriesUseCase`: Retrieves available categories

### Application Services
- `App`: Main application service that wires all dependencies
- Dependency injection container for all layers

## Infrastructure Layer

The infrastructure layer handles external concerns like database persistence and external services.

### Repository Pattern
All repositories implement interfaces defined in the domain layer:

```go
type UserRepository interface {
    Create(user *User) error
    GetByID(id UserID) (*User, error)
    GetByEmail(email Email) (*User, error)
    Update(user *User) error
    Delete(id UserID) error
}
```

### Database Implementations
- `SQLiteUserRepository`: SQLite implementation of user repository
- `SQLiteTransactionRepository`: SQLite implementation of transaction repository
- `SQLiteCategoryRepository`: SQLite implementation of category repository
- `SQLiteCurrencyRepository`: SQLite implementation of currency repository

### Database Abstraction
The application supports multiple database backends:
- **SQLite**: Default for development and simple deployments
- **PostgreSQL**: For production and scalable deployments

## Interface Layer

The interface layer handles external communication and adapts external requests to the application layer.

### HTTP Handlers
- `IdentityHandlers`: Handles authentication endpoints
- `FinanceHandlers`: Handles financial operation endpoints

### Middleware
- `AuthMiddleware`: JWT token validation and user context injection
- CORS middleware for cross-origin requests

### Request/Response DTOs
- Input validation and sanitization
- Response formatting and error handling
- Content-Type negotiation

## Design Patterns

### 1. Repository Pattern
- Abstracts data access logic
- Enables easy testing with mocks
- Supports multiple data sources

### 2. Use Case Pattern
- Encapsulates business workflows
- Coordinates between domain and infrastructure
- Provides clear application boundaries

### 3. Value Object Pattern
- Immutable objects representing concepts
- No identity, equality by value
- Examples: Money, Email, IDs

### 4. Entity Pattern
- Objects with identity and lifecycle
- Business logic encapsulation
- Examples: User, Transaction, Category

### 5. Service Pattern
- Stateless operations that don't belong to entities
- Domain services for complex business logic
- Application services for orchestration

## Data Flow

### Request Flow
1. **HTTP Request** → Interface Layer (Handlers)
2. **Validation** → Input DTOs and validation
3. **Use Case** → Application Layer orchestration
4. **Domain Service** → Business logic execution
5. **Repository** → Infrastructure Layer persistence
6. **Response** → Interface Layer formatting

### Authentication Flow
1. **Login Request** → Identity Handler
2. **User Validation** → User Service
3. **Token Generation** → Token Service
4. **JWT Response** → Client
5. **Protected Requests** → Auth Middleware validation

## Error Handling

### Error Types
- **Domain Errors**: Business rule violations
- **Application Errors**: Use case failures
- **Infrastructure Errors**: Database and external service failures
- **Interface Errors**: HTTP and validation errors

### Error Propagation
- Domain errors bubble up through layers
- Application layer translates domain errors to application errors
- Interface layer formats errors for HTTP responses

## Security Considerations

### Authentication
- JWT tokens with expiration
- Password hashing with bcrypt
- Token validation middleware

### Authorization
- User context injection
- Resource ownership validation
- Protected route middleware

### Data Protection
- Input validation and sanitization
- SQL injection prevention
- CORS configuration

## Testing Strategy

### Unit Tests
- Domain entities and value objects
- Domain services and business logic
- Use cases and application services

### Integration Tests
- Repository implementations
- HTTP handlers and middleware
- End-to-end API testing

### Test Doubles
- Repository interfaces for mocking
- Service interfaces for dependency injection
- HTTP client mocking for external services

## Performance Considerations

### Database Optimization
- Indexed queries for common operations
- Connection pooling for PostgreSQL
- Query optimization and N+1 prevention

### Caching Strategy
- Category and currency caching (planned)
- User session caching (planned)
- Response caching for static data (planned)

### Scalability
- Stateless application design
- Horizontal scaling support
- Database connection management

## Future Architecture Considerations

### Microservices Migration
- Service boundaries based on bounded contexts
- API Gateway for service coordination
- Event-driven communication

### Event Sourcing
- Transaction event streams
- Audit trail and history
- Event replay capabilities

### CQRS (Command Query Responsibility Segregation)
- Separate read and write models
- Optimized query performance
- Event-driven updates

## Monitoring and Observability

### Logging
- Structured logging with context
- Request/response logging
- Error tracking and alerting

### Metrics
- Application performance metrics
- Business metrics (transactions, users)
- Infrastructure metrics

### Health Checks
- Database connectivity
- External service availability
- Application status endpoints

---

This architecture provides a solid foundation for the PandaPocket application, ensuring maintainability, testability, and scalability while following industry best practices and DDD principles.


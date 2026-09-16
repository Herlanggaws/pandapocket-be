# PandaPocket Backend

A personal finance management API built with Go using Domain-Driven Design (DDD) architecture. PandaPocket allows users to track expenses, incomes, and manage financial categories with a clean, maintainable codebase.

## 🚀 Features

### ✅ Currently Implemented
- **User Authentication**: JWT-based authentication with registration and login
- **Category Management**: Create and retrieve expense/income categories
- **Transaction Management**: Create, retrieve, and delete expenses and incomes
- **Multi-Currency Support**: Support for multiple currencies
- **Clean Architecture**: Domain-Driven Design with proper separation of concerns

### 🔄 Planned Features
- Budget management and tracking
- Recurring transactions
- Financial analytics and reports
- User preferences and settings
- Admin API endpoints
- Notification system

## 🏗️ Architecture

PandaPocket follows Domain-Driven Design (DDD) principles with a clean architecture:

```
├── cmd/                    # Application entry points
├── internal/
│   ├── application/        # Application layer (Use Cases)
│   ├── domain/            # Domain layer (Entities, Value Objects, Services)
│   ├── infrastructure/    # Infrastructure layer (Repositories, Database)
│   └── interfaces/        # Interface layer (HTTP handlers, Middleware)
├── scripts/               # Setup and utility scripts
└── main.go               # Main application entry point
```

### Layer Responsibilities

- **Domain Layer**: Contains business logic, entities, value objects, and domain services
- **Application Layer**: Orchestrates use cases and coordinates between domain and infrastructure
- **Infrastructure Layer**: Handles data persistence, external services, and technical concerns
- **Interface Layer**: Manages HTTP requests, authentication, and API responses

## 🛠️ Technology Stack

- **Language**: Go 1.23.0
- **Web Framework**: Gin
- **Database**: SQLite (default) / PostgreSQL
- **Authentication**: JWT tokens
- **Architecture**: Domain-Driven Design (DDD)
- **Dependencies**: See [go.mod](go.mod) for complete list

## 📋 Prerequisites

- Go 1.23.0 or higher
- SQLite3 (for default database)
- PostgreSQL (optional, for production)

## 🚀 Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd PandaPocket/backend
go mod download
```

### 2. Run with SQLite (Default)

```bash
go run main.go
```

The application will:
- Create a SQLite database (`panda_pocket.db`)
- Initialize default categories and currencies
- Start the server on `http://localhost:8080`

### 3. Run with PostgreSQL

```bash
# Setup PostgreSQL environment
source scripts/setup-postgres.sh

# Start PostgreSQL (using Docker)
docker-compose up -d postgres

# Run the application
go run main.go
```

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_TYPE` | `sqlite` | Database type (`sqlite` or `postgres`) |
| `DB_HOST` | `localhost` | Database host (PostgreSQL only) |
| `DB_PORT` | `5432` | Database port (PostgreSQL only) |
| `DB_USER` | `postgres` | Database user (PostgreSQL only) |
| `DB_PASSWORD` | `postgres` | Database password (PostgreSQL only) |
| `DB_NAME` | `panda_pocket` | Database name (PostgreSQL only) |

### Database Setup

#### SQLite (Default)
No additional setup required. The database file will be created automatically.

#### PostgreSQL
1. Install PostgreSQL or use Docker:
   ```bash
   docker run --name panda-pocket-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=panda_pocket -p 5432:5432 -d postgres:13
   ```

2. Set environment variables:
   ```bash
   export DB_TYPE=postgres
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=panda_pocket
   ```

## 📚 API Documentation

The API provides endpoints under `/api` (see [API_DOCUMENTATION.md](API_DOCUMENTATION.md) for the full contract), including:

- **Authentication**: `/api/auth/*`
- **Categories / Expenses / Incomes / Transactions / Budgets / Currencies / Analytics**
- **Users**: `/api/users`
- **Admin dashboard**: `/api/dashboard/stats`
- **Health Check**: `/health`

**Keep docs in sync:** any change to routes, request/response shapes, auth, or error codes must update [API_DOCUMENTATION.md](API_DOCUMENTATION.md) in the same change. Cursor rule: `.cursor/rules/api-documentation.mdc`.

For detailed API documentation, see [API_DOCUMENTATION.md](API_DOCUMENTATION.md).

### Example Usage

1. **Register a new user**:
   ```bash
   curl -X POST http://localhost:8080/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email": "user@example.com", "password": "password123"}'
   ```

2. **Login**:
   ```bash
   curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email": "user@example.com", "password": "password123"}'
   ```

3. **Create an expense** (with token from login):
   ```bash
   curl -X POST http://localhost:8080/api/expenses \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <your-token>" \
     -d '{"category_id": 1, "amount": 25.50, "description": "Lunch", "date": "2024-01-15"}'
   ```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/domain/finance
```

### API Testing

Test reports are available:
- [API_TEST_REPORT_1.md](API_TEST_REPORT_1.md)
- [API_TEST_REPORT_FINAL.md](API_TEST_REPORT_FINAL.md)

## 🏗️ Development

### Project Structure

```
internal/
├── application/           # Use cases and application services
│   ├── finance/          # Financial use cases
│   └── identity/         # User identity use cases
├── domain/               # Business logic and entities
│   ├── finance/          # Financial domain models
│   └── identity/         # User identity domain models
├── infrastructure/       # External concerns
│   └── database/         # Database implementations
└── interfaces/           # External interfaces
    └── http/             # HTTP handlers and middleware
```

### Adding New Features

1. **Domain Layer**: Define entities, value objects, and domain services
2. **Application Layer**: Create use cases that orchestrate domain operations
3. **Infrastructure Layer**: Implement repositories and external services
4. **Interface Layer**: Add HTTP handlers and middleware

For detailed development guidelines, see [DEVELOPMENT.md](DEVELOPMENT.md).

## 🚀 Deployment

### Production Deployment

1. **Build the application**:
   ```bash
   go build -o panda-pocket main.go
   ```

2. **Set production environment variables**:
   ```bash
   export DB_TYPE=postgres
   export DB_HOST=your-db-host
   export DB_USER=your-db-user
   export DB_PASSWORD=your-db-password
   export DB_NAME=panda_pocket
   ```

3. **Run the application**:
   ```bash
   ./panda-pocket
   ```

For detailed deployment instructions, see [DEPLOYMENT.md](DEPLOYMENT.md).

## 📊 Database Schema

### Tables

- **users**: User accounts and authentication
- **categories**: Expense and income categories
- **currencies**: Supported currencies
- **transactions**: Financial transactions (expenses and incomes)

### Default Data

The application automatically creates:
- Default expense categories (Food, Transportation, Entertainment, etc.)
- Default income categories (Salary, Freelance, Investment, etc.)
- Common currencies (USD, EUR, IDR, etc.)

## 🔒 Security

- JWT-based authentication
- Password hashing with bcrypt
- CORS configuration for development
- Input validation and sanitization
- SQL injection prevention through parameterized queries

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed contribution guidelines.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Check the [API_DOCUMENTATION.md](API_DOCUMENTATION.md) for API details
- Review the [DEVELOPMENT.md](DEVELOPMENT.md) for development guidelines

## 📈 Roadmap

- [ ] Budget management system
- [ ] Recurring transactions
- [ ] Financial analytics and reporting
- [ ] Mobile app integration
- [ ] Multi-user support
- [ ] Advanced security features
- [ ] Performance optimizations
- [ ] Comprehensive test coverage

---

**PandaPocket** - Your personal finance companion 🐼💰



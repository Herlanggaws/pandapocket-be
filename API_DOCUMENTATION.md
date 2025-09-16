# PandaPocket API Documentation

## Overview

PandaPocket is a personal finance management API built with Domain-Driven Design (DDD) architecture. The API allows users to track expenses, incomes, and categories. The application follows clean architecture principles with proper separation of concerns.

**Base URL:** `http://localhost:8080`  
**API Version:** v1 (DDD Refactored)  
**Content-Type:** `application/json`  
**Architecture:** Domain-Driven Design (DDD)

## Current Implementation Status

### ✅ Implemented Endpoints
- **Authentication**: Register, Login, Logout
- **Categories**: Get all categories, Create category
- **Expenses**: Get all expenses, Create expense, Delete expense
- **Incomes**: Get all incomes, Create income, Delete income
- **Health Check**: Server status

### 🔄 Not Yet Implemented (Future Versions)
- **Categories**: Get by ID, Update, Delete individual categories
- **Currencies**: All currency management endpoints
- **Budgets**: All budget management endpoints
- **Recurring Transactions**: All recurring transaction endpoints
- **Analytics**: Spending analytics and reports
- **User Preferences**: User settings and preferences
- **Notifications**: User notification system
- **Admin API**: Administrative endpoints for back-office management

### 📊 Architecture Overview
The application follows Domain-Driven Design principles with the following structure:
- **Domain Layer**: Entities, Value Objects, Domain Services
- **Application Layer**: Use Cases, Application Services
- **Infrastructure Layer**: Repository implementations, Database
- **Interface Layer**: HTTP handlers, Middleware

## Authentication

The API uses token-based authentication. Include the authorization token in the request header:

```
Authorization: Bearer <your-token>
```

### Getting Started

1. Register a new user account
2. Login to receive an authentication token
3. Use the token for all subsequent API calls

## CORS Configuration

The API allows requests from the following origins:
- `http://localhost:3000`
- `http://localhost:3001`
- `http://localhost:3002`
- `http://localhost:3003`
- `http://localhost:3004` (Back office)

## Health Check

### GET /health

Check if the API is running.

**Response:**
```json
{
  "status": "ok"
}
```

---

## User Authentication

### POST /api/auth/register

Register a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "message": "User registered successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

### POST /api/auth/login

Login to get an authentication token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com"
  }
}
```

### POST /api/auth/logout

Logout and invalidate the current token.

**Response:**
```json
{
  "message": "Logout successful"
}
```

---

## Categories

### GET /api/categories

Get all categories available to the user (default + user-created).

**Query Parameters:**
- `type` (optional): Filter by category type (`expense` or `income`)

**Response:**
```json
[
  {
    "id": 1,
    "name": "Food",
    "color": "#EF4444",
    "type": "expense"
  },
  {
    "id": 9,
    "name": "Salary",
    "color": "#10B981",
    "type": "income"
  }
]
```

### POST /api/categories

Create a new category.

**Request Body:**
```json
{
  "name": "Custom Category",
  "color": "#3B82F6",
  "type": "expense"
}
```

**Response:**
```json
{
  "message": "Category created successfully",
  "category": {
    "id": 0,
    "name": "Custom Category",
    "color": "#3B82F6",
    "type": "expense"
  }
}
```


---

## Expenses

### GET /api/expenses

Get all expenses for the authenticated user.

**Response:**
```json
[
  {
    "id": 13,
    "user_id": 7,
    "category_id": 1,
    "currency_id": 6,
    "amount": 50,
    "description": "Final Test Expense",
    "date": "2024-01-17",
    "type": "expense",
    "created_at": "2025-09-10T11:02:29+07:00"
  }
]
```

### POST /api/expenses

Create a new expense.

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 25.50,
  "description": "Lunch at restaurant",
  "date": "2024-01-15"
}
```

**Response:**
```json
{
  "message": "Expense created successfully",
  "expense": {
    "id": 0,
    "user_id": 7,
    "category_id": 1,
    "currency_id": 6,
    "amount": 25.5,
    "description": "Lunch at restaurant",
    "date": "2024-01-15",
    "type": "expense",
    "created_at": "2025-09-10T10:56:56+07:00"
  }
}
```

### DELETE /api/expenses/:id

Delete an expense.

**Response:**
```json
{
  "message": "Expense deleted successfully"
}
```

---

## Incomes

### GET /api/incomes

Get all incomes for the authenticated user.

**Response:**
```json
[
  {
    "id": 8,
    "user_id": 7,
    "category_id": 9,
    "currency_id": 6,
    "amount": 2000,
    "description": "Final Test Income",
    "date": "2024-01-17",
    "type": "income",
    "created_at": "2025-09-10T11:02:50+07:00"
  }
]
```

### POST /api/incomes

Create a new income.

**Request Body:**
```json
{
  "category_id": 5,
  "amount": 3000.00,
  "description": "Monthly salary",
  "date": "2024-01-01"
}
```

**Response:**
```json
{
  "message": "Income created successfully",
  "income": {
    "id": 0,
    "user_id": 7,
    "category_id": 9,
    "currency_id": 6,
    "amount": 3000,
    "description": "Monthly salary",
    "date": "2024-01-01",
    "type": "income",
    "created_at": "2025-09-10T10:57:04+07:00"
  }
}
```

### DELETE /api/incomes/:id

Delete an income.

**Response:**
```json
{
  "message": "Income deleted successfully"
}
```


---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "Invalid request data"
}
```

### 401 Unauthorized
```json
{
  "error": "Authorization header required"
}
```

### 403 Forbidden
```json
{
  "error": "Access denied"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```

---

## Data Types

### Category Types
- `expense`: For expense categories
- `income`: For income categories

### Transaction Types
- `expense`: For expense transactions
- `income`: For income transactions

---

## Rate Limiting

Currently, no rate limiting is implemented. In production, consider implementing rate limiting to prevent abuse.

---

## Database Support

The API supports both SQLite (default) and PostgreSQL databases. The database type can be configured using the `DB_TYPE` environment variable.

---

## Development Notes

- The API uses Domain-Driven Design (DDD) architecture
- Built with Go and Gin framework for HTTP routing
- Authentication is implemented using JWT tokens
- CORS is configured to allow localhost development
- All timestamps are in UTC format
- Date formats should be in `YYYY-MM-DD` format for input
- Amounts are stored as floating-point numbers
- SQLite database is used by default
- Clean architecture with proper separation of concerns

---

## Version History

- **v1.0.0**: Initial release with basic CRUD operations
- **v1.1.0**: Added budgets and recurring transactions
- **v1.2.0**: Added analytics and user preferences
- **v1.3.0**: Added admin API endpoints
- **v2.0.0**: **DDD Refactored** - Complete architectural overhaul with Domain-Driven Design

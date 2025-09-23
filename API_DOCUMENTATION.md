# PandaPocket API Documentation

## Overview

PandaPocket is a personal finance management API built with Domain-Driven Design (DDD) architecture. The API allows users to track expenses, incomes, categories, budgets, and currencies with comprehensive analytics.

**Base URL:** `http://localhost:8080`  
**API Version:** v1 (DDD Refactored)  
**Content-Type:** `application/json`  
**Architecture:** Domain-Driven Design (DDD)

## Current Implementation Status

### ✅ Implemented Endpoints
- **Authentication**: Register, Login, Logout
- **Categories**: Full CRUD operations (Get, Create, Update, Delete)
- **Expenses**: Full CRUD operations (Get, Create, Update, Delete)
- **Incomes**: Full CRUD operations (Get, Create, Update, Delete)
- **Transactions**: Get all transactions with advanced filtering
- **Budgets**: Full CRUD operations (Get, Create, Update, Delete)
- **Currencies**: Full CRUD operations (Get, Create, Update, Delete)
- **Analytics**: Spending analytics and reports
- **Health Check**: Server status

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

### PUT /api/categories/:id

Update an existing category.

**Request Body:**
```json
{
  "name": "Updated Category",
  "color": "#10B981",
  "type": "income"
}
```

**Response:**
```json
{
  "message": "Category updated successfully",
  "category": {
    "id": 1,
    "name": "Updated Category",
    "color": "#10B981",
    "type": "income"
  }
}
```

### DELETE /api/categories/:id

Delete a category.

**Response:**
```json
{
  "message": "Category deleted successfully"
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

### PUT /api/expenses/:id

Update an existing expense.

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 150.00,
  "description": "Updated lunch at restaurant",
  "date": "2024-01-15"
}
```

**Response:**
```json
{
  "message": "Expense updated successfully",
  "expense": {
    "id": 1,
    "user_id": 7,
    "category_id": 1,
    "currency_id": 6,
    "amount": 150,
    "description": "Updated lunch at restaurant",
    "date": "2024-01-15",
    "type": "expense"
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

### PUT /api/incomes/:id

Update an existing income.

**Request Body:**
```json
{
  "category_id": 9,
  "amount": 4000.00,
  "description": "Updated monthly salary",
  "date": "2024-01-01"
}
```

**Response:**
```json
{
  "message": "Income updated successfully",
  "income": {
    "id": 1,
    "user_id": 7,
    "category_id": 9,
    "currency_id": 6,
    "amount": 4000,
    "description": "Updated monthly salary",
    "date": "2024-01-01",
    "type": "income"
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

## Transactions

### GET /api/transactions

Get all transactions (both income and expense) for the authenticated user with advanced filtering capabilities.

**Query Parameters:**
- `type` (optional): Filter by transaction type (`expense` or `income`)
- `category_ids` (optional): Filter by category IDs (comma-separated, e.g., `1,2,3`)
- `start_date` (optional): Filter transactions from this date (YYYY-MM-DD format)
- `end_date` (optional): Filter transactions until this date (YYYY-MM-DD format)
- `page` (optional): Page number for pagination (1-based, default: 1)
- `limit` (optional): Number of items per page (default: 20, max: 100)

**Examples:**
- Get all transactions: `GET /api/transactions`
- Get only expenses: `GET /api/transactions?type=expense`
- Get transactions from specific date range: `GET /api/transactions?start_date=2024-01-01&end_date=2024-12-31`
- Get transactions from specific categories: `GET /api/transactions?category_ids=1,2,3`
- Combined filters: `GET /api/transactions?type=expense&start_date=2024-01-01&end_date=2024-12-31&category_ids=1,2`
- Paginated results: `GET /api/transactions?page=2&limit=10`
- Paginated with filters: `GET /api/transactions?type=expense&page=1&limit=5`

**Response:**
```json
{
  "transactions": [
    {
      "id": 4,
      "user_id": 1,
      "category_id": 1,
      "currency_id": 1,
      "amount": 50,
      "description": "Test expense",
      "date": "2024-01-15",
      "type": "expense",
      "created_at": "2025-09-24T04:35:44+07:00"
    },
    {
      "id": 5,
      "user_id": 1,
      "category_id": 9,
      "currency_id": 1,
      "amount": 1000,
      "description": "Test salary",
      "date": "2024-01-01",
      "type": "income",
      "created_at": "2025-09-24T04:35:44+07:00"
    }
  ],
  "total": 2,
  "page": 1,
  "limit": 20,
  "total_pages": 1,
  "filters": {
    "type": "expense",
    "start_date": "2024-01-01",
    "end_date": "2024-12-31",
    "page": 1,
    "limit": 20
  }
}
```

**Response Fields:**
- `transactions`: Array of transaction objects
- `total`: Total number of transactions matching the filters (across all pages)
- `page`: Current page number (1-based)
- `limit`: Number of items per page
- `total_pages`: Total number of pages available
- `filters`: Object showing the applied filters for transparency

**Transaction Object Fields:**
- `id`: Unique transaction identifier
- `user_id`: ID of the user who owns the transaction
- `category_id`: ID of the category this transaction belongs to
- `currency_id`: ID of the currency used for this transaction
- `amount`: Transaction amount
- `description`: Transaction description
- `date`: Transaction date (YYYY-MM-DD format)
- `type`: Transaction type (`expense` or `income`)
- `created_at`: Timestamp when the transaction was created

---

## Budgets

### GET /api/budgets

Get all budgets for the authenticated user.

**Response:**
```json
[
  {
    "id": 1,
    "user_id": 7,
    "category_id": 1,
    "currency_id": 6,
    "amount": 500,
    "period": "monthly",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "created_at": "2025-09-10T11:02:29+07:00"
  }
]
```

### POST /api/budgets

Create a new budget.

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 500.00,
  "period": "monthly",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
```

**Response:**
```json
{
  "message": "Budget created successfully",
  "budget": {
    "id": 0,
    "user_id": 7,
    "category_id": 1,
    "currency_id": 6,
    "amount": 500,
    "period": "monthly",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "created_at": "2025-09-10T10:56:56+07:00"
  }
}
```

### PUT /api/budgets/:id

Update an existing budget.

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 750.00,
  "period": "monthly",
  "start_date": "2024-01-01"
}
```

**Response:**
```json
{
  "message": "Budget updated successfully"
}
```

### DELETE /api/budgets/:id

Delete a budget.

**Response:**
```json
{
  "message": "Budget deleted successfully"
}
```

---

## Currencies

### GET /api/currencies

Get all currencies available in the system.

**Response:**
```json
[
  {
    "id": 1,
    "code": "USD",
    "name": "US Dollar",
    "symbol": "$",
    "is_default": true
  },
  {
    "id": 2,
    "code": "EUR",
    "name": "Euro",
    "symbol": "€",
    "is_default": false
  }
]
```

### POST /api/currencies

Create a new currency.

**Request Body:**
```json
{
  "code": "GBP",
  "name": "British Pound",
  "symbol": "£",
  "is_default": false
}
```

**Response:**
```json
{
  "message": "Currency created successfully",
  "currency": {
    "id": 0,
    "code": "GBP",
    "name": "British Pound",
    "symbol": "£",
    "is_default": false
  }
}
```

### PUT /api/currencies/:id

Update an existing currency.

**Request Body:**
```json
{
  "code": "GBP",
  "name": "British Pound Sterling",
  "symbol": "£",
  "is_default": false
}
```

**Response:**
```json
{
  "message": "Currency updated successfully",
  "currency": {
    "id": 1,
    "code": "GBP",
    "name": "British Pound Sterling",
    "symbol": "£",
    "is_default": false
  }
}
```

### DELETE /api/currencies/:id

Delete a currency.

**Response:**
```json
{
  "message": "Currency deleted successfully"
}
```

### PUT /api/currencies/:id/set-default

Set a currency as the user's default currency.

**Response:**
```json
{
  "message": "Default currency set successfully"
}
```

### GET /api/currencies/default

Get the user's default currency.

**Response:**
```json
{
  "id": 3,
  "code": "GBP",
  "name": "British Pound",
  "symbol": "£",
  "is_default": true,
  "created_at": "2025-09-23T16:20:51.976667+07:00"
}
```

---

## Analytics

### GET /api/analytics

Get spending analytics and reports.

**Query Parameters:**
- `period` (optional): Time period for analytics (`daily`, `weekly`, `monthly`, `yearly`)
- `start_date` (optional): Start date for custom period (YYYY-MM-DD)
- `end_date` (optional): End date for custom period (YYYY-MM-DD)

**Response:**
```json
{
  "total_expenses": 1250.50,
  "total_incomes": 3000.00,
  "net_balance": 1749.50,
  "expenses_by_category": [
    {
      "category_id": 1,
      "category_name": "Food",
      "amount": 450.00,
      "percentage": 36.0
    },
    {
      "category_id": 2,
      "category_name": "Transport",
      "amount": 200.50,
      "percentage": 16.0
    }
  ],
  "monthly_trends": [
    {
      "month": "2024-01",
      "expenses": 1250.50,
      "incomes": 3000.00
    }
  ]
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

### Budget Periods
- `daily`: Daily budget
- `weekly`: Weekly budget
- `monthly`: Monthly budget
- `yearly`: Yearly budget

---

## Development Notes

- The API uses Domain-Driven Design (DDD) architecture
- Built with Go and Gin framework for HTTP routing
- Authentication is implemented using JWT tokens
- CORS is configured to allow localhost development
- All timestamps are in UTC format
- Date formats should be in `YYYY-MM-DD` format for input
- Amounts are stored as floating-point numbers
- PostgreSQL database is used by default
- Clean architecture with proper separation of concerns

---

## Version History

- **v2.1.0**: **Enhanced Transaction API** - Advanced filtering and pagination for transaction retrieval
  - New unified transactions endpoint with filtering capabilities
  - Support for filtering by transaction type, categories, and date ranges
  - Combined filter support for complex queries
  - Pagination support with configurable page size (default: 20, max: 100)
  - Improved API response structure with pagination metadata and filter transparency

- **v2.0.0**: **DDD Refactored** - Complete architectural overhaul with Domain-Driven Design
  - Full CRUD operations for Categories, Currencies
  - Budget management system
  - Analytics and reporting
  - Comprehensive transaction tracking

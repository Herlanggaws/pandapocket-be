# PandaPocket API Documentation

## Overview

PandaPocket is a personal finance management API built with Domain-Driven Design (DDD) architecture. The API allows users to track expenses, incomes, categories, budgets, and currencies with comprehensive analytics.

**Base URL:** `http://localhost:8080/api`  
**Content-Type:** `application/json`  
**Architecture:** Domain-Driven Design (DDD)

## Current Implementation Status

### Implemented Endpoints

- **Authentication**: Register, Login, Logout, Refresh, Forgot/Reset Password, Change Password
- **Categories**: Full CRUD operations
- **Expenses**: Full CRUD operations
- **Incomes**: Full CRUD operations
- **Transactions**: Get all transactions with advanced filtering and pagination
- **Budgets**: Full CRUD operations
- **Currencies**: Full CRUD operations
- **Analytics**: Spending analytics and reports
- **Dashboard**: Admin-only dashboard statistics
- **Health Check**: Server status (`GET /health`)

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

CORS currently allows all origins (`*`). Allowed request headers: `Origin`, `Content-Type`, `Accept`, `Authorization`.

## Endpoint Index

| Method | Path | Auth | Notes |
|--------|------|------|-------|
| GET | `/health` | No | Health check (outside `/api`) |
| POST | `/api/auth/register` | No | |
| POST | `/api/auth/login` | No | |
| POST | `/api/auth/refresh` | No | Refresh access token |
| POST | `/api/auth/logout` | No | Requires `refresh_token` body |
| POST | `/api/auth/forgot` | No | Forgot password |
| POST | `/api/auth/reset-password` | No | Reset with token from email |
| POST | `/api/auth/change-password` | Yes | Authenticated password change |
| GET | `/api/users` | Yes | List users |
| GET | `/api/dashboard/stats` | Yes (admin) | Admin dashboard statistics |
| GET/POST | `/api/categories` | Yes | |
| PUT/DELETE | `/api/categories/:id` | Yes | |
| GET/POST | `/api/expenses` | Yes | |
| PUT/DELETE | `/api/expenses/:id` | Yes | |
| GET/POST | `/api/incomes` | Yes | |
| PUT/DELETE | `/api/incomes/:id` | Yes | |
| GET | `/api/transactions` | Yes | Filtered/paginated list |
| GET/POST | `/api/budgets` | Yes | |
| PUT/DELETE | `/api/budgets/:id` | Yes | |
| GET/POST | `/api/currencies` | Yes | |
| GET | `/api/currencies/default` | Yes | |
| PUT | `/api/currencies/:id/set-default` | Yes | |
| PUT/DELETE | `/api/currencies/:id` | Yes | |
| GET | `/api/analytics` | Yes | |

## Health Check

### GET /health

Check if the API is running.

**Response:**
```json
{
  "status": "success",
  "data": {
    "status": "ok"
  },
  "error": null
}
```

---

## Standardized Response Structure

All API endpoints follow a standardized response structure:

**Success Response:**
```json
{
  "status": "success",
  "data": {
    // Response data here
  },
  "error": null
}
```

**Error Response:**
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "ERROR_CODE",
    "error_message": "Human-readable error message"
  }
}
```

### Response Fields

- `status` (string): Either `"success"` or `"error"`
- `data` (object/array/null): The response data for successful requests, `null` for errors
- `error` (object/null): Error details for failed requests, `null` for successful requests
  - `error_code` (string): Machine-readable error code (e.g., `VALIDATION_ERROR`, `ACCESS_DENIED`)
  - `error_message` (string): Human-readable error message

### Common Error Codes

- `VALIDATION_ERROR`: Request validation failed
- `INVALID_CREDENTIALS`: Invalid email or password
- `INVALID_TOKEN`: Invalid or expired authentication token
- `AUTHORIZATION_HEADER_REQUIRED`: Missing Authorization header
- `ACCESS_DENIED`: User doesn't have permission to access the resource
- `CATEGORY_ACCESS_DENIED`: User doesn't have access to the category
- `CURRENCY_ACCESS_DENIED`: User doesn't have access to the currency
- `TRANSACTION_NOT_FOUND`: Transaction not found
- `CATEGORY_NOT_FOUND`: Category not found
- `CURRENCY_NOT_FOUND`: Currency not found
- `BUDGET_NOT_FOUND`: Budget not found
- `TRANSACTION_TYPE_MISMATCH`: Transaction type doesn't match the endpoint
- `INVALID_CATEGORY_ID`: Invalid category ID format
- `INVALID_CURRENCY_ID`: Invalid currency ID format
- `FETCH_EXPENSES_ERROR`: Failed to fetch expenses
- `FETCH_INCOMES_ERROR`: Failed to fetch incomes
- `FETCH_TRANSACTIONS_ERROR`: Failed to fetch transactions
- `FETCH_CATEGORIES_ERROR`: Failed to fetch categories
- `FETCH_BUDGETS_ERROR`: Failed to fetch budgets
- `FETCH_CURRENCIES_ERROR`: Failed to fetch currencies
- `FETCH_ANALYTICS_ERROR`: Failed to fetch analytics
- `FETCH_DASHBOARD_STATS_ERROR`: Failed to fetch dashboard statistics

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
  "status": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "...",
    "user": {
      "id": 1,
      "email": "user@example.com"
    }
  },
  "error": null
}
```

### POST /api/auth/login

Login to get access and refresh tokens.

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
  "status": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "...",
    "user": {
      "id": 1,
      "email": "user@example.com"
    }
  },
  "error": null
}
```

### POST /api/auth/refresh

Exchange a refresh token for a new access token and refresh token pair.

**Request Body:**
```json
{
  "refresh_token": "..."
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "..."
  },
  "error": null
}
```

### POST /api/auth/logout

Revoke a refresh token.

**Request Body:**
```json
{
  "refresh_token": "..."
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Logout successful"
  },
  "error": null
}
```

### POST /api/auth/forgot

Request a password reset email/link for the given address.

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "...",
    "token": "...",
    "reset_link": "..."
  },
  "error": null
}
```

### POST /api/auth/reset-password

Reset password using the token from the forgot-password flow.

**Request Body:**
```json
{
  "token": "...",
  "new_password": "newpassword123",
  "confirm_new_password": "newpassword123"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "..."
  },
  "error": null
}
```

### POST /api/auth/change-password

Change password for the authenticated user. Requires `Authorization: Bearer <token>`.

**Request Body:**
```json
{
  "old_password": "password123",
  "new_password": "newpassword123",
  "confirm_new_password": "newpassword123"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Password changed successfully"
  },
  "error": null
}
```

---

## Users

### GET /api/users

List users. Requires authentication.

**Response:**
```json
{
  "status": "success",
  "data": {
    "users": [
      {
        "id": 1,
        "email": "user@example.com",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "error": null
}
```

---

## Dashboard (Admin)

### GET /api/dashboard/stats

Admin-only dashboard statistics. Requires authentication and `admin` role.

**Response:**
```json
{
  "status": "success",
  "data": {
    "total_users": 100,
    "active_users": 80,
    "total_budgets": 50,
    "total_transactions": 1000,
    "total_expenses": 25000.5,
    "total_income": 40000.0,
    "budgets_created_this_week": 5,
    "budgets_created_this_month": 20
  },
  "error": null
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
    "type": "expense",
    "is_default": true
  },
  {
    "id": 9,
    "name": "Salary",
    "color": "#10B981",
    "type": "income",
    "is_default": true
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

Get all expense transactions for the authenticated user.

**Response:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "category_id": 1,
    "category": {
      "id": 1,
      "name": "Food",
      "color": "#EF4444",
      "type": "expense"
    },
    "amount": 50.0,
    "description": "Lunch at restaurant",
    "date": "2024-01-15",
    "created_at": "2024-01-15T10:00:00Z"
  }
]
```

### POST /api/expenses

Create a new expense transaction.

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 50.0,
  "description": "Lunch at restaurant",
  "date": "2024-01-15"
}
```

**Response:**
```json
{
  "message": "Expense created successfully",
  "expense": {
    "id": 1,
    "user_id": 1,
    "category_id": 1,
    "amount": 50.0,
    "description": "Lunch at restaurant",
    "date": "2024-01-15",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

### PUT /api/expenses/:id

Update an existing expense transaction.

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 75.0,
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
    "user_id": 1,
    "category_id": 1,
    "currency_id": 1,
    "amount": 75.0,
    "description": "Updated lunch at restaurant",
    "date": "2024-01-15",
    "type": "expense"
  }
}
```

**Error Responses:**

**400 Bad Request - Access Denied:**
The "access denied" error can occur in the following scenarios:

1. **Transaction Ownership**: The transaction with the given ID does not belong to the authenticated user.
   ```json
   {
     "error": "access denied"
   }
   ```

2. **Transaction Type Mismatch**: The transaction ID exists but is of a different type (e.g., trying to update an expense but the ID points to an income, or vice versa).
   ```json
   {
     "error": "transaction type mismatch"
   }
   ```
   **Note**: This can happen if an expense and income share the same ID in their respective tables. The system now validates that the transaction type matches the endpoint being used.

3. **Category Access**: The `category_id` provided is not a default category and does not belong to the authenticated user.
   ```json
   {
     "error": "access denied to category"
   }
   ```
   **Solution**: Ensure you're using either:
   - A default category (available to all users)
   - A category that you created (belongs to your user account)

4. **Currency Access**: The currency being used is not a default currency and does not belong to the authenticated user.
   ```json
   {
     "error": "access denied to currency"
   }
   ```
   **Note**: Currently, the handler uses currency ID `1` (default USD). If this currency doesn't exist or isn't accessible, you'll get this error.

**400 Bad Request - Transaction Not Found:**
```json
{
  "error": "transaction not found"
}
```

**400 Bad Request - Category Not Found:**
```json
{
  "error": "category not found"
}
```

**400 Bad Request - Currency Not Found:**
```json
{
  "error": "currency not found"
}
```

### DELETE /api/expenses/:id

Delete an expense transaction.

**Response:**
```json
{
  "message": "Expense deleted successfully"
}
```

---

## Incomes

### GET /api/incomes

Get all income transactions for the authenticated user.

**Response:**
```json
{
  "status": "success",
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "category_id": 9,
      "category": {
        "id": 9,
        "name": "Salary",
        "color": "#10B981",
        "type": "income"
      },
      "amount": 3000.0,
      "description": "Monthly salary",
      "date": "2024-01-01",
      "created_at": "2024-01-01T09:00:00Z"
    }
  ],
  "error": null
}
```

### POST /api/incomes

Create a new income transaction.

**Request Body:**
```json
{
  "category_id": 9,
  "amount": 3000.0,
  "description": "Monthly salary",
  "date": "2024-01-01"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "income": {
      "id": 1,
      "user_id": 1,
      "category_id": 9,
      "amount": 3000.0,
      "description": "Monthly salary",
      "date": "2024-01-01",
      "created_at": "2024-01-01T09:00:00Z"
    }
  },
  "error": null
}
```

### PUT /api/incomes/:id

Update an existing income transaction.

**Request Body:**
```json
{
  "category_id": 9,
  "amount": 3500.0,
  "description": "Updated monthly salary",
  "date": "2024-01-01"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "income": {
      "id": 1,
      "user_id": 1,
      "category_id": 9,
      "currency_id": 1,
      "amount": 3500.0,
      "description": "Updated monthly salary",
      "date": "2024-01-01",
      "type": "income"
    }
  },
  "error": null
}
```

**Error Responses:**

**400 Bad Request - Access Denied:**
The "access denied" error can occur in the following scenarios:

1. **Transaction Ownership**: The transaction with the given ID does not belong to the authenticated user.
   ```json
   {
     "status": "error",
     "data": null,
     "error": {
       "error_code": "ACCESS_DENIED",
       "error_message": "access denied"
     }
   }
   ```

2. **Transaction Type Mismatch**: The transaction ID exists but is of a different type (e.g., trying to update an income but the ID points to an expense, or vice versa).
   ```json
   {
     "status": "error",
     "data": null,
     "error": {
       "error_code": "TRANSACTION_TYPE_MISMATCH",
       "error_message": "transaction type mismatch"
     }
   }
   ```
   **Note**: This can happen if an expense and income share the same ID in their respective tables. The system now validates that the transaction type matches the endpoint being used.

3. **Category Access**: The `category_id` provided is not a default category and does not belong to the authenticated user.
   ```json
   {
     "status": "error",
     "data": null,
     "error": {
       "error_code": "CATEGORY_ACCESS_DENIED",
       "error_message": "access denied to category"
     }
   }
   ```
   **Solution**: Ensure you're using either:
   - A default category (available to all users)
   - A category that you created (belongs to your user account)

4. **Currency Access**: The currency being used is not a default currency and does not belong to the authenticated user.
   ```json
   {
     "status": "error",
     "data": null,
     "error": {
       "error_code": "CURRENCY_ACCESS_DENIED",
       "error_message": "access denied to currency"
     }
   }
   ```
   **Note**: Currently, the handler uses currency ID `1` (default USD). If this currency doesn't exist or isn't accessible, you'll get this error.

**400 Bad Request - Transaction Not Found:**
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "TRANSACTION_NOT_FOUND",
    "error_message": "transaction not found"
  }
}
```

**400 Bad Request - Category Not Found:**
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "CATEGORY_NOT_FOUND",
    "error_message": "category not found"
  }
}
```

**400 Bad Request - Currency Not Found:**
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "CURRENCY_NOT_FOUND",
    "error_message": "currency not found"
  }
}
```

### DELETE /api/incomes/:id

Delete an income transaction.

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Income deleted successfully"
  },
  "error": null
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
      "created_at": "2025-09-24T04:35:44+07:00",
      "category": {
        "id": 1,
        "name": "Food",
        "color": "#EF4444",
        "type": "expense"
      }
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
      "created_at": "2025-09-24T04:35:44+07:00",
      "category": {
        "id": 9,
        "name": "Salary",
        "color": "#10B981",
        "type": "income"
      }
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
  },
  "error": null
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
{
  "status": "success",
  "data": [
    {
      "id": 8,
      "user_id": 1,
      "amount": 500,
      "period": "monthly",
      "start_date": "2024-01-01",
      "end_date": "2024-02-01",
      "created_at": "2025-09-30T09:51:35+07:00",
      "category": {
        "id": 1,
        "name": "Food",
        "color": "#EF4444",
        "type": "expense"
      }
    }
  ],
  "error": null
}
```

**Response Fields:**
- `id` (integer): Budget ID
- `user_id` (integer): User ID who owns the budget
- `amount` (number): Budget amount
- `period` (string): Budget period (weekly, monthly, yearly)
- `start_date` (string): Budget start date (YYYY-MM-DD)
- `end_date` (string): Budget end date (YYYY-MM-DD)
- `created_at` (string): Budget creation timestamp (ISO 8601)
- `category` (object, optional): Category information
  - `id` (integer): Category ID
  - `name` (string): Category name
  - `color` (string): Category color (hex code)
  - `type` (string): Category type (expense or income)

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
  "status": "success",
  "data": {
    "amount": 500,
    "period": "monthly",
    "start_date": "2024-01-01",
    "end_date": "2024-02-01",
    "category": {
      "id": 1,
      "name": "Food",
      "color": "#EF4444",
      "type": "expense"
    }
  },
  "error": null
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
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
```

**Response:**
```json
{
  "amount": 750,
  "period": "monthly",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "category": {
    "id": 1,
    "name": "Food",
    "color": "#EF4444",
    "type": "expense"
  }
}
```

### DELETE /api/budgets/:id

Delete a budget.

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Budget deleted successfully"
  },
  "error": null
}
```

---

## Currencies

### GET /api/currencies

Get all currencies available in the system.

**Response:**
```json
{
  "status": "success",
  "data": [
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
  ],
  "error": null
}
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
  "status": "success",
  "data": {
    "currency": {
      "id": 0,
      "code": "GBP",
      "name": "British Pound",
      "symbol": "£",
      "is_default": false
    }
  },
  "error": null
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
  "status": "success",
  "data": {
    "currency": {
      "id": 1,
      "code": "GBP",
      "name": "British Pound Sterling",
      "symbol": "£",
      "is_default": false
    }
  },
  "error": null
}
```

### DELETE /api/currencies/:id

Delete a currency.

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Currency deleted successfully"
  },
  "error": null
}
```

### PUT /api/currencies/:id/set-default

Set a currency as the user's default currency.

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Default currency set successfully"
  },
  "error": null
}
```

### GET /api/currencies/default

Get the user's default currency.

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 3,
    "code": "GBP",
    "name": "British Pound",
    "symbol": "£",
    "is_default": true,
    "created_at": "2025-09-23T16:20:51.976667+07:00"
  },
  "error": null
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
  "status": "success",
  "data": {
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
  },
  "error": null
}
```

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "VALIDATION_ERROR",
    "error_message": "Invalid request data"
  }
}
```

### 401 Unauthorized
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "AUTHORIZATION_HEADER_REQUIRED",
    "error_message": "Authorization header required"
  }
}
```

### 403 Forbidden
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "ACCESS_DENIED",
    "error_message": "Access denied"
  }
}
```

### 404 Not Found
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "RESOURCE_NOT_FOUND",
    "error_message": "Resource not found"
  }
}
```

### 500 Internal Server Error
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "INTERNAL_SERVER_ERROR",
    "error_message": "Internal server error"
  }
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

## Documentation Maintenance

Keep this file in sync with the running API. When routes, request/response shapes, auth rules, or error codes change in the backend, update `API_DOCUMENTATION.md` in the same change. Source of truth for mounted routes is `internal/application/app.go`.

## Development Notes

- The API uses Domain-Driven Design (DDD) architecture
- Built with Go and Gin framework for HTTP routing
- Authentication uses bearer access tokens plus refresh tokens
- CORS allows all origins (`*`)
- All timestamps are in UTC format
- Date formats should be in `YYYY-MM-DD` format for input
- Amounts are stored as floating-point numbers
- PostgreSQL database is used by default
- Clean architecture with proper separation of concerns

---

## Version History

- **v2.5.0**: **Removed API Versioning** - Endpoints live under unversioned `/api/*` paths
  - Dropped `/api/v100` URL prefix and version middleware
  - Removed unused version lifecycle / deprecation handlers
  - Documented refresh, forgot/reset/change password, users, and admin dashboard endpoints

- **v2.4.0**: **Dashboard Statistics API** - Added comprehensive dashboard statistics for back office
  - New admin-only dashboard statistics endpoint (`GET /api/dashboard/stats`)
  - Role-based access control ensuring only admin users can access dashboard data
  - User registration automatically assigns "user" role by default

- **v2.3.0**: **Enhanced API Documentation** - Updated documentation to reflect current implementation
  - Added detailed documentation for Expenses and Incomes endpoints
  - Updated Transactions endpoint with enhanced response structure including category details

- **v2.2.0**: **API Versioning System** (superseded by v2.5.0)
  - Historical note: multi-version URL prefixes were introduced and later removed

- **v2.1.0**: **Enhanced Transaction API** - Advanced filtering and pagination for transaction retrieval
  - Unified transactions endpoint with filtering and pagination

- **v2.0.0**: **DDD Refactored** - Complete architectural overhaul with Domain-Driven Design
  - Full CRUD for Categories, Currencies, Budgets, and transaction tracking

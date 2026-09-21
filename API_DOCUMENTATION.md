# PandaPocket API Documentation

## Overview

PandaPocket (product brand: **Berbudget**) is a personal finance management API built with Domain-Driven Design (DDD) architecture. The API allows users to track expenses, incomes, categories, budgets, recurring rules, and preferences with analytics and in-app notifications.

**Base URL:** `http://localhost:8080/api`  
**Content-Type:** `application/json`  
**Architecture:** Domain-Driven Design (DDD)

## Current Implementation Status

### Implemented Endpoints

- **Authentication**: Register, Login, Logout, Refresh, Forgot/Reset Password, Change Password, Delete Account
- **Categories**: Full CRUD operations
- **Expenses**: Full CRUD operations
- **Incomes**: Full CRUD operations
- **Transactions**: Get all transactions with advanced filtering and pagination
- **Budgets**: Full CRUD operations
- **Currencies**: Full CRUD operations
- **Preferences**: GET/PUT user preferences and onboarding
- **Account reset**: Challenge + wipe user financial data (login kept)
- **Notifications**: In-app list, mark read, delete
- **Recurring Transactions**: Create/list/delete; due items enqueue as pending for confirm/reject
- **Analytics**: Totals plus spending by category and period
- **Dashboard**: Admin-only dashboard statistics
- **Users**: Admin-only user list
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
| DELETE | `/api/auth/account` | Yes | Soft-delete account (password required); purge after 14 days |
| GET | `/api/users` | Yes (admin) | List users |
| GET | `/api/dashboard/stats` | Yes (admin) | Admin dashboard statistics |
| GET/POST | `/api/categories` | Yes | |
| PUT/DELETE | `/api/categories/:id` | Yes | |
| GET/POST | `/api/expenses` | Yes | |
| PUT/DELETE | `/api/expenses/:id` | Yes | |
| GET/POST | `/api/incomes` | Yes | |
| PUT/DELETE | `/api/incomes/:id` | Yes | |
| GET | `/api/transactions` | Yes | Filtered/paginated list |
| GET | `/api/export/transactions` | Yes (Pro) | Download CSV or PDF (max 5000 rows) |
| GET/POST | `/api/budgets` | Yes | `limit_type` fixed\|percent |
| PUT/DELETE | `/api/budgets/:id` | Yes | |
| GET/POST | `/api/currencies` | Yes | |
| GET | `/api/currencies/default` | Yes | |
| PUT | `/api/currencies/:id/set-default` | Yes | |
| PUT/DELETE | `/api/currencies/:id` | Yes | |
| GET | `/api/analytics` | Yes | |
| GET | `/api/health-score` | Yes | Financial health score 0–100 (+ monthly snapshot upsert) |
| GET | `/api/health-score/history` | Yes | Monthly score history (`limit`, default 12) |
| GET/POST | `/api/goals` | Yes | Savings goals with deadline |
| GET/PUT/DELETE | `/api/goals/:id` | Yes | |
| GET/POST | `/api/assets` | Yes | Non-wallet assets |
| PUT | `/api/assets/:id` | Yes | |
| POST | `/api/assets/:id/archive` | Yes | |
| POST | `/api/assets/:id/unarchive` | Yes | |
| GET/POST | `/api/liabilities` | Yes | Liabilities / debts |
| PUT | `/api/liabilities/:id` | Yes | |
| POST | `/api/liabilities/:id/archive` | Yes | |
| POST | `/api/liabilities/:id/unarchive` | Yes | |
| GET/POST | `/api/liabilities/:id/payments` | Yes | Payment history / record payment |
| GET | `/api/net-worth/summary` | Yes | Liquid + assets − liabilities (primary currency) |
| GET/PUT | `/api/preferences` | Yes | User preferences & onboarding |
| GET | `/api/me/subscription` | Yes | Current billing subscription + `is_pro` |
| POST | `/api/account/reset/challenge` | Yes | Issue one-time confirmation string for data reset |
| POST | `/api/account/reset` | Yes | Wipe user financial data after typing confirmation |
| POST | `/api/onboarding/complete` | Yes | Finish onboarding; seed pending income/expense + budget/(debt) |
| GET | `/api/notifications` | Yes | In-app notifications |
| PUT | `/api/notifications/:id/read` | Yes | Mark notification read |
| DELETE | `/api/notifications/:id` | Yes | Delete notification |
| POST | `/api/feedback` | Yes | Submit product feedback |
| POST | `/api/tickets` | Yes (Pro) | Create support ticket |
| GET | `/api/tickets` | Yes | List own support tickets |
| GET | `/api/tickets/:id` | Yes | Get own support ticket |
| POST | `/api/tickets/:id/reopen` | Yes | Reopen own done ticket |
| GET | `/api/admin/tickets` | Admin | List all support tickets |
| GET | `/api/admin/tickets/:id` | Admin | Get support ticket |
| PATCH | `/api/admin/tickets/:id/status` | Admin | Update ticket status |
| GET/POST | `/api/recurring-transactions` | Yes | Recurring rules (GET also enqueues due items as pending) |
| DELETE | `/api/recurring-transactions/:id` | Yes | |
| GET | `/api/pending-transactions` | Yes | List open pending recurring occurrences (also enqueues dues) |
| POST | `/api/pending-transactions/:id/confirm` | Yes | Confirm pending → create real expense/income |
| POST | `/api/pending-transactions/:id/reject` | Yes | Reject pending (no transaction) |
| GET/POST | `/api/wallets` | Yes | List/create wallets (`include_archived` query on GET) |
| GET | `/api/wallets/summary` | Yes | Liquid net worth (primary currency) |
| GET/PUT | `/api/wallets/:id` | Yes | Get/update wallet |
| POST | `/api/wallets/:id/default` | Yes | Set default wallet |
| POST | `/api/wallets/:id/archive` | Yes | Archive wallet |
| POST | `/api/wallets/:id/unarchive` | Yes | Unarchive wallet |
| GET | `/api/wallets/:id/balance` | Yes | Ledger balance breakdown |
| GET/POST | `/api/transfers` | Yes | List/create same-currency transfers |

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
- `BUDGET_OVERLAP`: Overlapping budget already exists for this category
- `INVALID_BUDGET_CATEGORY`: Budget category must be expense type
- `CATEGORY_HAS_BUDGETS`: Category cannot be deleted while budgets reference it
- `INVALID_DATE_RANGE`: End date must be on or after start date
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

Register a new user account. Also creates a billing subscription with a **14-day Pro trial** (`status=trialing`, `plan=free`, `trial_ends_at=now+14d`). Existing backfilled accounts do not receive a trial.

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

**Response:** Always returns a generic success message (does not reveal whether the email exists). The reset token is sent only via email.

```json
{
  "status": "success",
  "data": {
    "message": "If an account exists for that email, a password reset link has been sent."
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

### DELETE /api/auth/account

Soft-delete the authenticated user's account. Requires `Authorization: Bearer <token>` and the current password. Login is blocked immediately; all sessions are revoked. Personal data is retained for up to 14 days, then hard-purged by a background job. The original email can be registered again right after soft-delete.

**Request Body:**
```json
{
  "password": "password123"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Account scheduled for deletion",
    "scheduled_purge_at": "2026-10-02T10:00:00Z"
  },
  "error": null
}
```

**Errors:** `401` for invalid password; `400` for missing password or other client errors.

---

## Users

### GET /api/users

List users. Requires authentication and `admin` role (or higher).

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
- `wallet_id` (optional): Filter by wallet ID
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

---

### GET /api/export/transactions

Export transactions for the authenticated user as a **CSV** or **PDF** file download (Pro-gated via entitlement checker).

**Auth:** Bearer token required.  
**Premium:** Returns `403` with `PREMIUM_REQUIRED` (+ `feature`/`limit`/`used`) when the user is not Pro per subscription `IsPro()`.

**Query Parameters:**
- `format` (required): `csv` or `pdf`
- `type` (optional): `expense` or `income`
- `category_ids` (optional): comma-separated category IDs
- `wallet_id` (optional): wallet ID
- `start_date` / `end_date` (optional): `YYYY-MM-DD` — same defaults as `GET /api/transactions` (last 30 days if both empty)

**Limits:** Maximum **5000** rows. Over limit → `400` `EXPORT_TOO_LARGE`.

**Success response:** Raw file body (not JSON envelope)
- CSV: `Content-Type: text/csv; charset=utf-8`, `Content-Disposition: attachment; filename="berbudget-transactions-{start}-{end}.csv"`
- PDF: `Content-Type: application/pdf`, same disposition pattern with `.pdf`

**CSV columns:** `date,type,category,description,amount,currency_id,wallet_id`

**Examples:**
- `GET /api/export/transactions?format=csv`
- `GET /api/export/transactions?format=pdf&start_date=2026-01-01&end_date=2026-01-31&type=expense`

**Error examples:**
```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "PREMIUM_REQUIRED",
    "error_message": "export requires Pro",
    "feature": "export",
    "limit": 0,
    "used": 0
  }
}
```

---

## Transactions (list response)

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
    "total_incomes": 1000,
    "total_expenses": 50,
    "currency_id": 1,
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
- `transactions`: Array of transaction objects (all currencies; each row has its own `currency_id`)
- `total`: Total number of transactions matching the filters (across all pages)
- `page`: Current page number (1-based)
- `limit`: Number of items per page
- `total_pages`: Total number of pages available
- `total_incomes` / `total_expenses`: Sums for the **current page**, **primary currency only** (no FX); other-currency rows on the page are listed but excluded from these totals
- `currency_id`: User primary currency used for the page totals
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

Budgets are limited to **expense** categories. Date windows are inclusive. Create auto-calculates `end_date` from `period` + `start_date` (weekly = 7 days, monthly/yearly = calendar period minus one day). Overlapping budgets for the same user and category are rejected. Spending reports only sum expenses in the budget's currency.

**Limit types:**
- `fixed` (default): cap is `amount`
- `percent`: cap is `percent`% of **income** in the budget date range (same currency). Stored `amount` is `0`; use `effective_amount` for the resolved limit. If period income is `0`, `effective_amount` is `0` (exceeded only if spent > 0).

Report fields (`percentage_used`, `remaining`, `is_on_track`) are computed against `effective_amount`. Budget alerts (≥80%) also use `effective_amount`.

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
      "limit_type": "fixed",
      "effective_amount": 500,
      "period": "monthly",
      "start_date": "2024-01-01",
      "end_date": "2024-01-31",
      "created_at": "2025-09-30T09:51:35+07:00",
      "category": {
        "id": 1,
        "name": "Food",
        "color": "#EF4444",
        "type": "expense"
      },
      "report": {
        "is_on_track": true,
        "total_spent": 120.5,
        "remaining": 379.5,
        "percentage_used": 24.1
      }
    }
  ],
  "error": null
}
```

**Response Fields:**
- `id` (integer): Budget ID
- `user_id` (integer): User ID who owns the budget
- `amount` (number): Stored fixed amount (`0` when `limit_type` is `percent`)
- `limit_type` (string): `fixed` or `percent`
- `percent` (number, optional): 1–100 when `limit_type` is `percent`
- `effective_amount` (number): Resolved spending cap for the period
- `period` (string): Budget period (`weekly`, `monthly`, `yearly`)
- `start_date` (string): Budget start date (YYYY-MM-DD)
- `end_date` (string): Budget end date (YYYY-MM-DD), inclusive
- `created_at` (string): Budget creation timestamp (ISO 8601)
- `category` (object, optional): Category information
  - `id` (integer): Category ID
  - `name` (string): Category name
  - `color` (string): Category color (hex code)
  - `type` (string): Category type (`expense`)
- `report` (object, optional): Spending vs budget for the window
  - `is_on_track` (boolean): `true` when spent ≤ effective amount
  - `total_spent` (number): Expense total in the same currency
  - `remaining` (number): Effective amount minus total spent
  - `percentage_used` (number): Percent of effective amount used

### POST /api/budgets

Create a new budget. `end_date` is computed server-side from `period` and `start_date` (client-sent `end_date` is ignored if present).

**Request Body (fixed):**
```json
{
  "category_id": 1,
  "limit_type": "fixed",
  "amount": 500.00,
  "period": "monthly",
  "start_date": "2024-01-01"
}
```

**Request Body (percent of income):**
```json
{
  "category_id": 1,
  "limit_type": "percent",
  "percent": 30,
  "period": "monthly",
  "start_date": "2024-01-01"
}
```

- `limit_type` (optional): defaults to `fixed`
- `amount` (required when `fixed`): must be > 0
- `percent` (required when `percent`): 1–100

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 8,
    "user_id": 1,
    "amount": 500,
    "limit_type": "fixed",
    "effective_amount": 500,
    "period": "monthly",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "created_at": "2025-09-30T09:51:35+07:00",
    "category": {
      "id": 1,
      "name": "Food",
      "color": "#EF4444",
      "type": "expense"
    },
    "report": {
      "is_on_track": true,
      "total_spent": 0,
      "remaining": 500,
      "percentage_used": 0
    }
  },
  "error": null
}
```

### PUT /api/budgets/:id

Update an existing budget. `end_date` must be on or after `start_date`. Supports the same `limit_type` / `amount` / `percent` rules as create.

**Request Body:**
```json
{
  "category_id": 1,
  "limit_type": "fixed",
  "amount": 750.00,
  "period": "monthly",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
```

**Response:** same shape as GET/POST budget items (includes `limit_type`, `percent`, `effective_amount`, `report`).

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

Get spending analytics and reports for the authenticated user.

**Query Parameters:**
- `period` (optional): `weekly`, `monthly` (default), or `yearly`
- `start_date` / `end_date` (optional, Pro): `YYYY-MM-DD` custom range (both required together; response `period` becomes `custom`)
- `wallet_id` (optional): filter transactions to one wallet

**Entitlement:** Free may use `weekly` / `monthly` only. `yearly` or custom dates return **403** `PREMIUM_REQUIRED` (`feature`: `insights`).

**Currency:** Totals and breakdowns include **primary currency only** (no FX). Other-currency transactions are counted in `excluded_transaction_count` and omitted from sums.

**Response:**
```json
{
  "status": "success",
  "data": {
    "total_income": 3000.00,
    "total_spent": 1250.50,
    "net_amount": 1749.50,
    "period": "monthly",
    "currency_id": 1,
    "transaction_count": 42,
    "excluded_transaction_count": 3,
    "spending_by_category": [
      {
        "category_id": 1,
        "category_name": "Food",
        "category_color": "#22c55e",
        "amount": 450.00,
        "percentage": 36.0
      }
    ],
    "spending_by_period": [
      {
        "period": "Week 1",
        "amount": 320.00,
        "date": "2024-01-03"
      }
    ]
  },
  "error": null
}
```

---

## Financial Health Score

### GET /api/health-score

Server-computed score for the current calendar month (aligned with dashboard analytics).

**Formula (v0):**
- Budget adherence (50%): average of `max(0, 100 - min(percentage_used, 150))` across budgets overlapping this month; no budgets ⇒ 50
- Cashflow (30%): if monthly `total_income > 0` → `clamp(100 * (1 - total_spent/total_income), 0, 100)`; else 50
- Coverage (20%): ≥1 overlapping budget ⇒ 100; else 40

`score = round(0.5*A + 0.3*C + 0.2*Cov)`

On `GET /api/health-score`, the current month score is **upserted** into `health_score_snapshots` (`user_id` + `year_month`).

**Response:**
```json
{
  "status": "success",
  "data": {
    "score": 72,
    "components": {
      "budget_adherence": 80,
      "cashflow": 65,
      "coverage": 70
    },
    "period": "monthly",
    "year_month": "2026-09",
    "as_of": "2026-09-17T15:00:00Z"
  },
  "error": null
}
```

### GET /api/health-score/history

Newest-first monthly snapshots. Query `limit` (default 12, max 24).

```json
{
  "status": "success",
  "data": {
    "history": [
      {
        "year_month": "2026-09",
        "score": 72,
        "budget_adherence": 80,
        "cashflow": 65,
        "coverage": 70,
        "computed_at": "2026-09-17T15:00:00Z"
      }
    ]
  },
  "error": null
}
```

---

## Goals

Savings goals with a required deadline. Progress is **manual** (`current_amount`) by default, or **wallet-linked** when `wallet_id` is set (effective progress = full wallet balance). Auto-`completed` when effective amount ≥ `target_amount`.

- Manual create: currency = user primary.
- Linked create: currency = wallet currency; `current_amount` in body is ignored.
- Linked GET responses refresh `current_amount` from wallet balance; may persist `completed` + `goal_completed` notification.
- Unlink (`wallet_id: null` on PUT): freezes last wallet balance into `current_amount`.
- Archive wallet: auto-unlinks goals on that wallet (freeze amount).
- Deadline alerts (in-app `goal_deadline`): D−7 and D−1 for active incomplete goals when `goal_deadline_alerts` pref is on (hourly job).

### GET /api/goals

Query: `include_archived=true` to include archived goals.

Response goal fields include: `wallet_id`, `wallet_name` (when linked), `progress_source` (`manual` | `wallet`), effective `current_amount` / `progress_percent`.

### POST /api/goals

```json
{
  "name": "Emergency fund",
  "target_amount": 10000000,
  "current_amount": 1500000,
  "target_date": "2026-12-31",
  "wallet_id": 3
}
```

`wallet_id` optional. When set, wallet must belong to the user, not archived.

### GET /api/goals/:id

### PUT /api/goals/:id

```json
{
  "name": "Emergency fund",
  "target_amount": 10000000,
  "current_amount": 2000000,
  "target_date": "2026-12-31",
  "status": "active",
  "wallet_id": 3
}
```

`status`: `active` | `completed` | `archived`  
`wallet_id`: set to link; `null` to unlink (clients should always send this field on edit). Currency must match when linking an existing goal.

### DELETE /api/goals/:id

---

## Assets & Liabilities

Manual balance-sheet positions (no market feeds). Wallets remain liquid assets.

### Assets

`type`: `property` | `vehicle` | `investment` | `other`

- `GET/POST /api/assets` (`include_archived` on GET)
- `PUT /api/assets/:id`
- `POST /api/assets/:id/archive`
- `POST /api/assets/:id/unarchive`

### Liabilities / Debts

`type`: `loan` | `credit_card` | `mortgage` | `other`

Create/update body fields:
- `name`, `type`, `currency_id`, `current_balance` (required as before)
- `original_principal` (optional; defaults to `current_balance` on create when omitted)
- `interest_rate_apr` (optional annual percent)
- `minimum_payment` (optional installment / minimum payment)
- `next_due_date` (`YYYY-MM-DD`, optional)
- `notes`, `as_of_date`

Response extras:
- `payoff_progress_percent` when `original_principal` is set
- `estimated_months_remaining` when `minimum_payment > 0` and balance remains

- `GET/POST /api/liabilities`
- `PUT /api/liabilities/:id`
- `POST /api/liabilities/:id/archive`
- `POST /api/liabilities/:id/unarchive`
- `GET /api/liabilities/:id/payments`
- `POST /api/liabilities/:id/payments`

#### POST /api/liabilities/:id/payments

Records a payment and reduces `current_balance`. Optionally creates an expense on the default (or specified) wallet using the Debt category when available.

```json
{
  "amount": 500000,
  "paid_at": "2026-09-18",
  "note": "September installment",
  "create_expense": true,
  "category_id": null,
  "wallet_id": null
}
```

```json
{
  "status": "success",
  "data": {
    "liability": { "id": 1, "current_balance": 9500000 },
    "payment": { "id": 1, "amount": 500000, "paid_at": "2026-09-18", "expense_id": 42 }
  },
  "error": null
}
```

### GET /api/net-worth/summary

Primary-currency only (no FX). Other-currency wallets/assets/liabilities are counted in `excluded_*` fields.

```json
{
  "status": "success",
  "data": {
    "currency_id": 1,
    "liquid_net_worth": 1250000,
    "assets_total": 500000000,
    "liabilities_total": 200000000,
    "net_worth": 301250000,
    "excluded_asset_count": 0,
    "excluded_liability_count": 0,
    "excluded_wallet_count": 1
  },
  "error": null
}
```

`net_worth = liquid_net_worth + assets_total − liabilities_total`

---

## Preferences

### POST /api/onboarding/complete

Authenticated. Marks onboarding complete, updates preferences, and seeds baseline setup. Monthly income/expense become **pending** items (via monthly recurring rules) that the user must confirm or reject — they are not posted to the ledger until confirmed.

**Body:**
```json
{
  "primary_currency_id": 1,
  "goal": "budget",
  "topics": ["Food & Dining", "Transport"],
  "monthly_income": 10000000,
  "monthly_expense": 7000000,
  "debt_balance": 5000000,
  "debt_name": "Personal loan",
  "debt_type": "loan"
}
```

- First completion (when amounts > 0): creates monthly recurring income/expense rules, enqueues open `pending_transactions` for today, and a monthly fixed budget from expense
- Confirming a pending posts a real income/expense; rejecting skips that occurrence (recurring schedule still continues next month)
- Re-running after `onboarding_completed` is already true updates preferences only (no duplicate seed); overlapping budgets are ignored so retries/redos do not fail
- When `goal` is `debt` and `debt_balance` > 0, creates one liability on first completion

**Response** includes `onboarding` map and optional `health_score`.

### GET /api/preferences

Returns the authenticated user's preferences. Creates defaults on first access.

### PUT /api/preferences

Partial update. Accepts any of:
- `primary_currency_id`
- `email_notifications`, `budget_alerts`, `recurring_reminders`, `goal_deadline_alerts`
- `language` (`id` | `en`, default `id`)
- Onboarding fields: `onboarding_completed`, `goal`, `topics`, `cadence`, `start_path`

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Preferences updated successfully",
    "preferences": {
      "id": 1,
      "user_id": 1,
      "primary_currency_id": 1,
      "email_notifications": true,
      "budget_alerts": true,
      "recurring_reminders": true,
      "goal_deadline_alerts": true,
      "language": "id",
      "onboarding": {},
      "onboarding_completed": true
    }
  },
  "error": null
}
```

### GET /api/me/subscription

Returns the authenticated user's current subscription. Creates a Free row (`plan=free`, `status=expired`, no trial) if missing. If `status=trialing` and `trial_ends_at` has passed without paid access, normalizes to `status=expired` (keeps `trial_ends_at` so trial cannot restart).

**Example after register (active trial):**
```json
{
  "status": "success",
  "data": {
    "subscription": {
      "plan": "free",
      "status": "trialing",
      "billing_interval": null,
      "trial_ends_at": "2026-10-02T12:00:00Z",
      "current_period_end": null,
      "grace_ends_at": null,
      "cancel_at_period_end": false,
      "is_pro": true
    }
  },
  "error": null
}
```

`is_pro` is derived from DB state: active period, past_due within grace, or active trial (`trial_ends_at` in the future). Create gates and Pro-only endpoints use the same `SubscriptionChecker` / `IsPro()`.

---

## Freemium create gates (billing PR2)

Free users are limited on **create** writes. GET list/read stays open (including data created while Pro).

| Action | Free | Pro |
| --- | --- | --- |
| Create expense/income | max **50** / calendar month (UTC) | Unlimited |
| Create custom category | max **10** (non-default) | Unlimited |
| Create budget | max **3** active | Unlimited |
| Create recurring | blocked | Unlimited |
| Create support ticket | blocked | Allowed |
| Export transactions | blocked | Allowed |

Over limit / Pro-only → **403** with:

```json
{
  "status": "error",
  "data": null,
  "error": {
    "error_code": "PREMIUM_REQUIRED",
    "error_message": "budgets limit reached on Free plan (3/3). Upgrade to Pro.",
    "feature": "budgets",
    "limit": 3,
    "used": 3
  }
}
```

`feature` values: `transactions`, `categories`, `budgets`, `recurring`, `tickets`, `export`.

---

## Account data reset

Resets all financial data for the authenticated user while **keeping** the login (email/password). Session stays valid. After reset, the client should send the user through onboarding again.

Does **not** delete the `users` row or system-wide categories/currencies (`user_id IS NULL`).

### POST /api/account/reset/challenge

Authenticated. Issues a random confirmation string (TTL ~5 minutes). Previous challenges for the user are replaced.

**Response:**
```json
{
  "status": "success",
  "data": {
    "confirmation_text": "AFkLJdl9879x",
    "expires_at": "2026-09-18T03:00:00Z"
  },
  "error": null
}
```

### POST /api/account/reset

Authenticated. Requires exact match of `confirmation_text` from the active challenge (case-sensitive). On success, hard-deletes user-owned rows: pending transactions, liability payments, recurring, transfers, expenses, incomes, budgets, goals, assets, liabilities, wallets, health snapshots, notifications, user-owned categories/currencies, preferences, password-reset tokens, and the challenge itself.

**Body:**
```json
{
  "confirmation_text": "AFkLJdl9879x"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "message": "Account data reset successfully"
  },
  "error": null
}
```

**Errors (400):** missing/expired challenge, confirmation mismatch, or too many failed attempts (max 5).

---

## Notifications

### GET /api/notifications

List in-app notifications for the authenticated user (newest first).

### PUT /api/notifications/:id/read

Mark a notification as read.

### DELETE /api/notifications/:id

Delete a notification.

---

## Feedback

### POST /api/feedback

Submit product feedback for the authenticated user. Stored in `user_feedbacks` (no email/backoffice in v1).

**Request body:**

```json
{
  "category": "suggestion",
  "message": "Would love CSV export for expenses."
}
```

| Field | Type | Rules |
|-------|------|-------|
| `category` | string | Required. One of `bug`, `suggestion`, `other` |
| `message` | string | Required. 1–2000 characters |

**Response `201`:**

```json
{
  "status": "success",
  "data": {
    "feedback": {
      "id": 1,
      "category": "suggestion",
      "message": "Would love CSV export for expenses.",
      "created_at": "2026-09-18T10:00:00Z"
    }
  }
}
```

---

## Support Tickets

Support tickets are Pro-only for **create**. Feedback (`POST /api/feedback`) remains a separate product-input channel for all authenticated users.

Create returns `403` + `PREMIUM_REQUIRED` when subscription `IsPro()` is false.

### POST /api/tickets

Create a support ticket. Status starts as `open`.

**Request body:**

```json
{
  "subject": "Cannot sync wallet balance",
  "body": "After adding a transfer, balance stays stale until refresh.",
  "category": "technical",
  "priority": "medium"
}
```

| Field | Type | Rules |
|-------|------|-------|
| `subject` | string | Required. 1–200 characters |
| `body` | string | Required. 1–2000 characters |
| `category` | string | Required. One of `technical`, `payment`, `other` |
| `priority` | string | Optional. One of `low`, `medium`, `high` (default `medium`) |

**Response `201`:**

```json
{
  "status": "success",
  "data": {
    "ticket": {
      "id": 1,
      "subject": "Cannot sync wallet balance",
      "body": "After adding a transfer, balance stays stale until refresh.",
      "category": "technical",
      "priority": "medium",
      "status": "open",
      "created_at": "2026-09-18T10:00:00Z",
      "updated_at": "2026-09-18T10:00:00Z"
    }
  }
}
```

**Error `403` `PREMIUM_REQUIRED`:** caller is not Pro (`feature=tickets`).

### GET /api/tickets

List the authenticated user's tickets. Optional query: `status`, `category`.

**Response `200`:** `{ "status": "success", "data": { "tickets": [ ... ] } }`

### GET /api/tickets/:id

Get one of the authenticated user's tickets.

### POST /api/tickets/:id/reopen

Reopen a ticket that is `done` → `open`. Sends a status-change email to the ticket owner. Returns `400` if the ticket is not `done`.

### GET /api/admin/tickets

Admin-only. List all tickets (includes `user_id`, `user_email`). Optional query: `status`, `category`, `user_id`.

### GET /api/admin/tickets/:id

Admin-only. Ticket detail with `user_id` and `user_email`.

### PATCH /api/admin/tickets/:id/status

Admin-only. Update status and email the ticket owner.

**Request body:**

```json
{
  "status": "in_progress"
}
```

| Field | Type | Rules |
|-------|------|-------|
| `status` | string | Required. One of `open`, `in_progress`, `done` |

---

## Recurring Transactions

### GET /api/recurring-transactions

List recurring rules. Also enqueues any **due** active rules as **pending transactions** (does not create expenses/incomes) and advances `next_due_date` using the schedule. May create `recurring_reminder` notifications when enabled in preferences.

Each item includes schedule fields (`weekday`, `day_of_month`, `month_of_year` as applicable), `schedule_label`, `wallet_id`, and `currency_id` (inherited from the wallet).

### POST /api/recurring-transactions

Creates a recurring rule. Currency is taken from the resolved wallet (`wallet_id` optional; defaults to the user’s default wallet). `next_due_date` is computed from the schedule (first occurrence on/after today), unless an optional `next_due_date` seed date is provided as the search start.

**Weekly** — requires `weekday` (`0`=Sunday … `6`=Saturday):

```json
{
  "type": "expense",
  "category_id": 1,
  "amount": 50000,
  "description": "Gym",
  "frequency": "weekly",
  "weekday": 1
}
```

**Monthly** — requires `day_of_month` (`1`–`31`). If the month has fewer days, the due date is **clamped to the last day** of that month (e.g. day 31 → Feb 28/29, Apr 30):

```json
{
  "type": "expense",
  "category_id": 1,
  "amount": 150000,
  "description": "Rent",
  "frequency": "monthly",
  "day_of_month": 31
}
```

**Yearly** — requires `month_of_year` (`1`–`12`) and `day_of_month`. Feb 29 clamps to Feb 28 in non-leap years:

```json
{
  "type": "income",
  "category_id": 2,
  "amount": 1000000,
  "description": "Annual bonus",
  "frequency": "yearly",
  "month_of_year": 3,
  "day_of_month": 15
}
```

Currency is inherited from the wallet (not from primary preference alone).

### DELETE /api/recurring-transactions/:id

Delete a recurring rule owned by the authenticated user.

---

### GET /api/pending-transactions

List open pending recurring occurrences for the authenticated user. Also runs due enqueue (same as GET recurring) so visiting Dashboard/Transactions can surface new pendings without opening Recurring.

**Response** — array of:
- `id`, `user_id`, `wallet_id`, `currency_id`, `recurring_transaction_id`, `due_date`, `amount`, `description`, `type`, `category_id`, `status` (`pending`), `category`, `created_at`

### POST /api/pending-transactions/:id/confirm

Confirm a pending occurrence owned by the user: creates the corresponding expense/income on `due_date`, sets status to `confirmed`.

### POST /api/pending-transactions/:id/reject

Reject a pending occurrence owned by the user: sets status to `rejected` without creating a transaction. The recurring schedule was already advanced when the pending was enqueued.

---

## Wallets

### GET /api/wallets

List wallets for the authenticated user. Each item includes computed `balance`.

**Query Parameters:**
- `include_archived` (optional): `true` / `1` to include archived wallets

### POST /api/wallets

```json
{
  "name": "BCA",
  "type": "bank",
  "currency_id": 1,
  "opening_balance": 100000,
  "is_default": false
}
```

`type`: `cash` | `bank` | `e_wallet`

Free plan: max **1** non-archived wallet. Additional creates return **403** `PREMIUM_REQUIRED` (`feature`: `wallets`, `limit`: 1). Pro (trial / paid / grace) is unlimited.

### GET /api/wallets/summary

Liquid net worth in the user's **primary** currency: sum of balances for non-archived wallets with matching `currency_id`. Other-currency active wallets are counted in `excluded_wallet_count` (not converted; no FX).

Register this route **before** `/wallets/:id`.

```json
{
  "status": "success",
  "data": {
    "currency_id": 1,
    "liquid_net_worth": 1250000,
    "wallet_count": 3,
    "excluded_wallet_count": 1
  },
  "error": null
}
```

### GET /api/wallets/:id

### PUT /api/wallets/:id

Update `name`, `type`, and/or `opening_balance`. Currency is fixed after create.

### POST /api/wallets/:id/default

### POST /api/wallets/:id/archive

Cannot archive the default wallet or the last active wallet.

### POST /api/wallets/:id/unarchive

### GET /api/wallets/:id/balance

```json
{
  "wallet_id": 1,
  "opening_balance": 0,
  "total_income": 500000,
  "total_expense": 200000,
  "transfers_in": 0,
  "transfers_out": 50000,
  "balance": 250000
}
```

Balance = `opening_balance + total_income − total_expense + transfers_in − transfers_out`.

## Transfers

Same-currency moves between wallets. Transfers are **not** income/expense and do not affect budgets/analytics cashflow.

### GET /api/transfers

**Query Parameters:** `wallet_id`, `start_date`, `end_date`

### POST /api/transfers

```json
{
  "from_wallet_id": 1,
  "to_wallet_id": 2,
  "amount": 50000,
  "description": "Top up",
  "date": "2026-09-17"
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
- `weekly`: Weekly budget (inclusive 7-day window)
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

- **v2.23.0**: **Goals wallet link + deadline alerts (C7)**
  - Optional `wallet_id` on goals; linked progress mirrors wallet balance
  - Response: `wallet_id`, `wallet_name`, `progress_source`
  - Pref `goal_deadline_alerts`; in-app `goal_deadline` (D−7/D−1) and `goal_completed`
  - Archive wallet unlinks linked goals (freeze amount)

- **v2.22.0**: **Primary-currency aggregates + currency_id on recurring/pending**
  - `GET /api/analytics` — sums/breakdowns use primary currency only (no FX); response adds `currency_id`, `excluded_transaction_count`
  - `GET /api/transactions` — `total_incomes` / `total_expenses` are primary-currency only for the current page; response adds `currency_id`
  - Recurring + pending list responses include `currency_id` (from wallet); create recurring inherits wallet currency

- **v2.21.0**: **Pro differentiation gates (P1/P2)**
  - `POST /api/wallets` — Free max 1 non-archived wallet; **403** `PREMIUM_REQUIRED` (`feature`: `wallets`)
  - `GET /api/analytics` — Free: `weekly`/`monthly`; Pro: `yearly` + optional `start_date`/`end_date`; Free advanced range → **403** `PREMIUM_REQUIRED` (`feature`: `insights`)

- **v2.20.0**: **Trial on register (billing PR3)**
  - `POST /api/auth/register` creates `status=trialing`, `trial_ends_at=now+14d` (plan stays `free` until paid)
  - `GET /api/me/subscription` normalizes ended trials to `expired` without clearing `trial_ends_at`
  - Existing / backfilled users remain Free without trial; trial once per account
- **v2.19.0**: **Entitlement gates (billing PR2)**
  - `SubscriptionChecker` replaces interim `BILLING_ENTITLEMENTS_ENABLED` bypass
  - Free create limits: transactions 50/mo, custom categories 10, active budgets 3, recurring blocked
  - Ticket create + export use real subscription `IsPro()`
  - `403 PREMIUM_REQUIRED` includes optional `feature` / `limit` / `used`
  - Onboarding seed bypasses gates via trusted context (Free users can still complete onboarding)
- **v2.18.0**: **Subscription schema (billing PR1)**
  - Tables `subscriptions` + `billing_webhook_events` (webhook handler later)
  - Register creates subscription row; existing users backfilled Free
  - `GET /api/me/subscription` returns plan/status/period/trial fields + `is_pro`
  - Domain `IsPro(now)`
- **v2.17.0**: **Transaction export (CSV/PDF)**
  - `GET /api/export/transactions?format=csv|pdf` with list filters; Pro-gated (`PREMIUM_REQUIRED`)
  - Shared `domain/entitlement` package (`ErrPremiumRequired`)
- **v2.16.0**: **Support tickets**
  - User: `POST/GET /api/tickets`, `GET /api/tickets/:id`, `POST /api/tickets/:id/reopen`
  - Admin: `GET /api/admin/tickets`, `GET /api/admin/tickets/:id`, `PATCH /api/admin/tickets/:id/status`
  - Categories `technical` \| `payment` \| `other`; priority `low` \| `medium` \| `high`; status `open` \| `in_progress` \| `done`
  - Create is Pro-gated (`PREMIUM_REQUIRED`)
  - Email notification on admin status change and user reopen (SMTP / mock)

- **v2.15.0**: **Preferences language**
  - `GET/PUT /api/preferences` includes `language` (`id` \| `en`, default `id`)

- **v2.14.0**: **User feedback**
  - `POST /api/feedback` stores authenticated feedback (`bug` \| `suggestion` \| `other` + message) in `user_feedbacks`

- **v2.13.0**: **Onboarding seeds pending transactions**
  - `POST /api/onboarding/complete` no longer posts income/expense to the ledger immediately
  - Monthly income/expense create recurring rules + open pending items awaiting confirm/reject
  - Budget and optional debt liability seeding unchanged

- **v2.12.0**: **Delete account**
  - `DELETE /api/auth/account` soft-deletes the authenticated account after password confirmation
  - Sessions revoked immediately; login/refresh blocked for deleted users
  - Hard purge of user-owned data after 14 days via hourly background job

- **v2.11.0**: **Account data reset**
  - `POST /api/account/reset/challenge` issues a one-time confirmation string (TTL 5m)
  - `POST /api/account/reset` wipes user financial data after exact confirmation match; login and session kept

- **v2.10.1**: **Idempotent onboarding complete**
  - `POST /api/onboarding/complete` skips re-seeding when already completed and ignores overlapping budget on retry/redo

- **v2.10.0**: **Debt tracker + onboarding complete**
  - Liabilities support `original_principal`, `interest_rate_apr`, `minimum_payment`, `next_due_date`
  - `GET/POST /api/liabilities/:id/payments` with optional expense creation (Debt category)
  - Response includes payoff progress and estimated months remaining
  - `POST /api/onboarding/complete` seeds income/expense/budget/(debt) and returns health score

- **v2.9.0**: **Goals, full net worth, health history**
  - Savings goals CRUD with deadline and manual progress
  - Assets + liabilities CRUD; `GET /api/net-worth/summary` (liquid + assets − liabilities, no FX)
  - Health score upserts monthly snapshots; `GET /api/health-score/history`

- **v2.8.0**: **Budget %, health score, liquid net worth**
  - Budgets support `limit_type` `fixed` | `percent`; responses include `effective_amount`
  - Percent budgets use share of same-currency income in the budget date range
  - `GET /api/health-score` — monthly Financial Health Score (0–100)
  - `GET /api/wallets/summary` — liquid net worth in primary currency (no FX)

- **v2.7.0**: **Multi-wallet**
  - New `wallets` and `transfers` resources (types: `cash` | `bank` | `e_wallet`; default, archive, opening balance)
  - Expenses, incomes, recurring, and pending stamp `wallet_id`; currency inherited from wallet
  - `GET /api/transactions` and `GET /api/analytics` accept optional `wallet_id`
  - Startup backfill creates a default Cash wallet per user and attaches existing rows
  - Transfers are same-currency only and excluded from budget/analytics cashflow

- **v2.6.1**: **Recurring schedule rules**
  - Weekly requires `weekday`; monthly `day_of_month` (clamp to last day); yearly `month_of_year` + `day_of_month`
  - Due advancement is schedule-aware; list responses include `schedule_label`

- **v2.6.0**: **Berbudget backlog ship**
  - `GET /api/users` is admin-only; forgot-password no longer returns token/reset_link in JSON
  - Preferences GET/PUT, in-app notifications, recurring transactions (due → pending confirm/reject)
  - Analytics returns `spending_by_category` and `spending_by_period`
  - Budget alerts create in-app notifications when spending approaches/exceeds limits

- **v2.5.1**: **Budgets correctness**
  - Create/update responses include `id`, `created_at`, and `report`
  - Budgets persist `currency_id`; reports sum same-currency expenses only
  - Reject overlapping budgets and non-expense categories
  - Create auto-calculates inclusive `end_date` from period

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

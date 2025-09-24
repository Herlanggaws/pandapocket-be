# PandaPocket API Documentation

## Overview

PandaPocket is a personal finance management API built with Domain-Driven Design (DDD) architecture. The API allows users to track expenses, incomes, categories, budgets, and currencies with comprehensive analytics.

**Base URL:** `http://localhost:8080`  
**API Versioning:** Versioned endpoints only (v100)  
**Content-Type:** `application/json`  
**Architecture:** Domain-Driven Design (DDD)

## API Versioning

The PandaPocket API supports multiple versions to ensure backward compatibility and smooth evolution:

### Supported Versions

| Version | Status | Features | Sunset Date |
|---------|--------|----------|-------------|
| **v100** | ✅ Current | Core features, transactions, categories, budgets, currencies, analytics | - |

### Version Endpoints

- **Current Version (v100):** `/api/v100/transactions`
- **All endpoints require versioning** - Version must be specified in URL

### Version Headers

All API responses include version information:

```
X-API-Version: v110
X-API-Latest: v110 (for legacy routes)
```

### Migration Guide

For future version management:

1. **Check version status:** `GET /api/version/info/{version}`
2. **Get version features:** `GET /api/version/features/{version}`
3. **Get version matrix:** `GET /api/version/matrix`

## Current Implementation Status

### ✅ Implemented Endpoints

#### Version 1.1.0 (v110) - Current
- **Authentication**: Register, Login, Logout
- **Transactions**: Enhanced CRUD with analytics and advanced filtering
- **Categories**: Full CRUD operations with enhanced validation
- **Budgets**: Full CRUD operations with enhanced validation
- **Currencies**: Full CRUD operations with enhanced validation
- **Analytics**: Advanced analytics with detailed insights
- **Version Management**: Complete version lifecycle management

#### Versioned Routes (v100 Only)
- **Authentication**: Register, Login, Logout (versioned)
- **Categories**: Full CRUD operations (Get, Create, Update, Delete)
- **Expenses**: Core operations (Get, Create, Delete)
- **Incomes**: Core operations (Get, Create, Delete)
- **Transactions**: Get all transactions with advanced filtering and pagination
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

## Version-Specific Endpoints

### Version 1.0.0 (v100) - Current Features

#### Authentication
- **POST** `/api/v100/auth/register` - Register new user
- **POST** `/api/v100/auth/login` - Login user
- **POST** `/api/v100/auth/logout` - Logout user

#### Categories
- **GET** `/api/v100/categories` - Get categories
- **POST** `/api/v100/categories` - Create category
- **PUT** `/api/v100/categories/{id}` - Update category
- **DELETE** `/api/v100/categories/{id}` - Delete category

#### Expenses
- **GET** `/api/v100/expenses` - Get expenses
- **POST** `/api/v100/expenses` - Create expense
- **DELETE** `/api/v100/expenses/{id}` - Delete expense

#### Incomes
- **GET** `/api/v100/incomes` - Get incomes
- **POST** `/api/v100/incomes` - Create income
- **DELETE** `/api/v100/incomes/{id}` - Delete income

#### Transactions
- **GET** `/api/v100/transactions` - Get all transactions with filtering

#### Budgets
- **GET** `/api/v100/budgets` - Get budgets
- **POST** `/api/v100/budgets` - Create budget
- **PUT** `/api/v100/budgets/{id}` - Update budget
- **DELETE** `/api/v100/budgets/{id}` - Delete budget

#### Currencies
- **GET** `/api/v100/currencies` - Get currencies
- **POST** `/api/v100/currencies` - Create currency
- **PUT** `/api/v100/currencies/{id}` - Update currency
- **DELETE** `/api/v100/currencies/{id}` - Delete currency
- **GET** `/api/v100/currencies/default` - Get default currency
- **PUT** `/api/v100/currencies/{id}/set-default` - Set default currency

#### Analytics
- **GET** `/api/v100/analytics` - Get spending analytics


---

## Version-Specific API Endpoints

### Version 1.1.0 (v110) - Current Features

#### Enhanced Transactions

##### GET /api/v110/transactions

Get transactions with advanced analytics and enhanced filtering.

**Query Parameters:**
- `type` (optional): Filter by transaction type (`expense` or `income`)
- `category_ids` (optional): Filter by category IDs (comma-separated)
- `start_date` (optional): Filter transactions from this date (YYYY-MM-DD)
- `end_date` (optional): Filter transactions until this date (YYYY-MM-DD)
- `page` (optional): Page number for pagination (default: 1)
- `limit` (optional): Number of items per page (default: 20, max: 100)

**Response:**
```json
{
  "transactions": [
    {
      "id": 1,
      "user_id": 1,
      "category_id": 1,
      "currency_id": 1,
      "amount": 50.0,
      "description": "Test transaction",
      "date": "2024-01-15",
      "type": "expense",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1
  },
  "analytics": {
    "version": "v110",
    "features": ["analytics", "advanced_filtering", "pagination"]
  }
}
```

##### POST /api/v110/transactions

Create a transaction with enhanced validation.

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 100.0,
  "description": "Enhanced transaction",
  "date": "2024-01-15",
  "type": "expense"
}
```

**Response:**
```json
{
  "message": "Transaction created successfully",
  "transaction": {
    "id": 1,
    "user_id": 1,
    "category_id": 1,
    "currency_id": 1,
    "amount": 100.0,
    "description": "Enhanced transaction",
    "date": "2024-01-15",
    "type": "expense"
  },
  "analytics": {
    "version": "v110",
    "features": ["enhanced_validation", "analytics"]
  }
}
```

##### GET /api/v110/transactions/analytics

Get detailed transaction analytics (v110 specific feature).

**Query Parameters:**
- `period` (optional): Time period (`daily`, `weekly`, `monthly`, `yearly`)

**Response:**
```json
{
  "transactions": [...],
  "analytics": {
    "period": "monthly",
    "data": {
      "total_expenses": 1250.50,
      "total_incomes": 3000.00,
      "net_balance": 1749.50
    },
    "version": "v110",
    "features": ["detailed_analytics", "period_analysis", "trend_analysis"]
  },
  "pagination": {
    "total": 10
  }
}
```

#### Enhanced Categories

##### GET /api/v120/categories

Get categories with analytics.

**Response:**
```json
{
  "categories": [
    {
      "id": 1,
      "name": "Food",
      "color": "#EF4444",
      "type": "expense"
    }
  ],
  "analytics": {
    "version": "v110",
    "features": ["analytics", "category_insights"]
  }
}
```

##### POST /api/v120/categories

Create category with enhanced validation.

**Request Body:**
```json
{
  "name": "Enhanced Category",
  "color": "#3B82F6",
  "type": "expense"
}
```

**Response:**
```json
{
  "message": "Category created successfully",
  "category": {
    "id": 1,
    "name": "Enhanced Category",
    "color": "#3B82F6",
    "type": "expense"
  },
  "analytics": {
    "version": "v110",
    "features": ["enhanced_validation", "analytics"]
  }
}
```

#### Enhanced Budgets

##### GET /api/v120/budgets

Get budgets with analytics.

**Response:**
```json
{
  "budgets": [
    {
      "id": 1,
      "user_id": 1,
      "category_id": 1,
      "amount": 500.0,
      "period": "monthly",
      "start_date": "2024-01-01",
      "end_date": "2024-01-31"
    }
  ],
  "analytics": {
    "version": "v110",
    "features": ["analytics", "budget_insights"]
  }
}
```

#### Enhanced Currencies

##### GET /api/v120/currencies

Get currencies with analytics.

**Response:**
```json
{
  "currencies": [
    {
      "id": 1,
      "code": "USD",
      "name": "US Dollar",
      "symbol": "$",
      "is_default": true
    }
  ],
  "analytics": {
    "version": "v110",
    "features": ["analytics", "currency_insights"]
  }
}
```

#### Enhanced Analytics

##### GET /api/v120/analytics

Get advanced analytics with detailed insights.

**Query Parameters:**
- `period` (optional): Time period for analytics

**Response:**
```json
{
  "analytics": {
    "total_expenses": 1250.50,
    "total_incomes": 3000.00,
    "net_balance": 1749.50,
    "expenses_by_category": [...],
    "monthly_trends": [...]
  },
  "version": "v120",
  "features": ["detailed_analytics", "period_analysis", "trend_analysis", "export_functionality"]
}
```


---

## Version Headers and Client Usage

### Response Headers

All API responses include version information in headers:

```
X-API-Version: v110
X-API-Deprecated: true (for deprecated versions)
X-API-Sunset-Date: 2024-06-01 (for deprecated versions)
X-API-Upgrade-URL: https://docs.pandapocket.com/upgrade
X-API-Latest: v110 (for legacy routes)
```

### Client Implementation Examples

#### JavaScript/TypeScript Client

```typescript
class PandaPocketClient {
    private baseURL: string;
    private version: string;
    private token: string;
    
    constructor(baseURL: string, version: string = 'v110', token: string = '') {
        this.baseURL = baseURL;
        this.version = version;
        this.token = token;
    }
    
    async getTransactions(): Promise<any> {
        const response = await fetch(`${this.baseURL}/api/${this.version}/transactions`, {
            headers: {
                'Authorization': `Bearer ${this.token}`,
                'X-API-Version': this.version,
                'Content-Type': 'application/json'
            }
        });
        
        // Check for deprecation warnings
        if (response.headers.get('X-API-Deprecated') === 'true') {
            console.warn('⚠️ API version is deprecated:', {
                version: this.version,
                sunsetDate: response.headers.get('X-API-Sunset-Date'),
                upgradeURL: response.headers.get('X-API-Upgrade-URL')
            });
        }
        
        return response.json();
    }
    
    async createTransaction(transaction: any): Promise<any> {
        const response = await fetch(`${this.baseURL}/api/${this.version}/transactions`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.token}`,
                'X-API-Version': this.version,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(transaction)
        });
        
        return response.json();
    }
}

// Usage examples
const client = new PandaPocketClient('http://localhost:8080', 'v110', 'your-token');

// Current version (v110)
const transactions = await client.getTransactions();
```

#### Go Client

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type PandaPocketClient struct {
    BaseURL string
    Version string
    Token   string
    Client  *http.Client
}

func NewPandaPocketClient(baseURL, version, token string) *PandaPocketClient {
    return &PandaPocketClient{
        BaseURL: baseURL,
        Version: version,
        Token:   token,
        Client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *PandaPocketClient) GetTransactions() (map[string]interface{}, error) {
    url := fmt.Sprintf("%s/api/%s/transactions", c.BaseURL, c.Version)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("X-API-Version", c.Version)
    
    resp, err := c.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Check for deprecation warnings
    if resp.Header.Get("X-API-Deprecated") == "true" {
        fmt.Printf("⚠️ Warning: API version %s is deprecated!\n", c.Version)
        fmt.Printf("   Sunset date: %s\n", resp.Header.Get("X-API-Sunset-Date"))
        fmt.Printf("   Upgrade URL: %s\n", resp.Header.Get("X-API-Upgrade-URL"))
    }
    
    var response map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&response)
    return response, err
}

func (c *PandaPocketClient) CreateTransaction(transaction map[string]interface{}) (map[string]interface{}, error) {
    url := fmt.Sprintf("%s/api/%s/transactions", c.BaseURL, c.Version)
    
    jsonData, err := json.Marshal(transaction)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("X-API-Version", c.Version)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var response map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&response)
    return response, err
}

// Usage example
func main() {
    // Current version (v110)
    client := NewPandaPocketClient("http://localhost:8080", "v110", "your-token")
    transactions, err := client.GetTransactions()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Transactions: %+v\n", transactions)
}
```

#### Python Client

```python
import requests
import json
from typing import Dict, Any, Optional

class PandaPocketClient:
    def __init__(self, base_url: str, version: str = 'v110', token: str = ''):
        self.base_url = base_url
        self.version = version
        self.token = token
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {token}',
            'X-API-Version': version,
            'Content-Type': 'application/json'
        })
    
    def get_transactions(self) -> Dict[str, Any]:
        url = f"{self.base_url}/api/{self.version}/transactions"
        response = self.session.get(url)
        
        # Check for deprecation warnings
        if response.headers.get('X-API-Deprecated') == 'true':
            print(f"⚠️ Warning: API version {self.version} is deprecated!")
            print(f"   Sunset date: {response.headers.get('X-API-Sunset-Date')}")
            print(f"   Upgrade URL: {response.headers.get('X-API-Upgrade-URL')}")
        
        response.raise_for_status()
        return response.json()
    
    def create_transaction(self, transaction: Dict[str, Any]) -> Dict[str, Any]:
        url = f"{self.base_url}/api/{self.version}/transactions"
        response = self.session.post(url, json=transaction)
        response.raise_for_status()
        return response.json()

# Usage examples
# Current version (v110)
client = PandaPocketClient('http://localhost:8080', 'v110', 'your-token')
transactions = client.get_transactions()
```

### Migration Strategies

#### 1. Gradual Migration

```typescript
// Start with current version
const client = new PandaPocketClient('http://localhost:8080', 'v110', token);

// Check if version is supported
const versionInfo = await fetch('/api/version/info/v110').then(r => r.json());
if (versionInfo.deprecated) {
    console.warn('Version is deprecated, consider upgrading');
}
```

#### 2. Version Detection

```typescript
async function detectBestVersion(baseURL: string): Promise<string> {
    try {
        // Try current version first
        const response = await fetch(`${baseURL}/api/v110/transactions`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            return 'v110';
        }
    } catch (error) {
        console.log('v110 not available, using legacy routes...');
    }
    
    // Fallback to legacy routes
    return 'legacy';
}
```

#### 3. Feature Detection

```typescript
async function checkFeatures(baseURL: string, version: string): Promise<string[]> {
    const response = await fetch(`${baseURL}/api/version/features/${version}`);
    const data = await response.json();
    return Object.keys(data.features).filter(feature => data.features[feature]);
}

// Check if analytics are available
const features = await checkFeatures('http://localhost:8080', 'v110');
if (features.includes('analytics')) {
    // Use enhanced analytics endpoint
    const analytics = await fetch('/api/v110/transactions/analytics');
} else {
    // Use basic analytics endpoint
    const analytics = await fetch('/api/analytics');
}
```

---

## User Authentication

### POST /api/v100/auth/register

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
    "email": "user@example.com",
    "name": "User"
  }
}
```

### POST /api/v100/auth/login

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
    "email": "user@example.com",
    "name": "User"
  }
}
```

### POST /api/v100/auth/logout

Logout and invalidate the current token.

**Response:**
```json
{
  "message": "Logout successful"
}
```

---

## Categories

### GET /api/v100/categories

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

### POST /api/v100/categories

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

### PUT /api/v100/categories/:id

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

### DELETE /api/v100/categories/:id

Delete a category.

**Response:**
```json
{
  "message": "Category deleted successfully"
}
```


---

## Expenses

### GET /api/v100/expenses

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

### POST /api/v100/expenses

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

### DELETE /api/v100/expenses/:id

Delete an expense transaction.

**Response:**
```json
{
  "message": "Expense deleted successfully"
}
```

---

## Incomes

### GET /api/v100/incomes

Get all income transactions for the authenticated user.

**Response:**
```json
[
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
]
```

### POST /api/v100/incomes

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
  "message": "Income created successfully",
  "income": {
    "id": 1,
    "user_id": 1,
    "category_id": 9,
    "amount": 3000.0,
    "description": "Monthly salary",
    "date": "2024-01-01",
    "created_at": "2024-01-01T09:00:00Z"
  }
}
```

### DELETE /api/v100/incomes/:id

Delete an income transaction.

**Response:**
```json
{
  "message": "Income deleted successfully"
}
```

---

## Transactions

### GET /api/v100/transactions

Get all transactions (both income and expense) for the authenticated user with advanced filtering capabilities.

**Query Parameters:**
- `type` (optional): Filter by transaction type (`expense` or `income`)
- `category_ids` (optional): Filter by category IDs (comma-separated, e.g., `1,2,3`)
- `start_date` (optional): Filter transactions from this date (YYYY-MM-DD format)
- `end_date` (optional): Filter transactions until this date (YYYY-MM-DD format)
- `page` (optional): Page number for pagination (1-based, default: 1)
- `limit` (optional): Number of items per page (default: 20, max: 100)

**Examples:**
- Get all transactions: `GET /api/v100/transactions`
- Get only expenses: `GET /api/v100/transactions?type=expense`
- Get transactions from specific date range: `GET /api/v100/transactions?start_date=2024-01-01&end_date=2024-12-31`
- Get transactions from specific categories: `GET /api/v100/transactions?category_ids=1,2,3`
- Combined filters: `GET /api/v100/transactions?type=expense&start_date=2024-01-01&end_date=2024-12-31&category_ids=1,2`
- Paginated results: `GET /api/v100/transactions?page=2&limit=10`
- Paginated with filters: `GET /api/v100/transactions?type=expense&page=1&limit=5`

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

### GET /api/v100/budgets

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

### POST /api/v100/budgets

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

### PUT /api/v100/budgets/:id

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

### DELETE /api/v100/budgets/:id

Delete a budget.

**Response:**
```json
{
  "message": "Budget deleted successfully"
}
```

---

## Currencies

### GET /api/v100/currencies

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

### POST /api/v100/currencies

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

### PUT /api/v100/currencies/:id

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

### DELETE /api/v100/currencies/:id

Delete a currency.

**Response:**
```json
{
  "message": "Currency deleted successfully"
}
```

### PUT /api/v100/currencies/:id/set-default

Set a currency as the user's default currency.

**Response:**
```json
{
  "message": "Default currency set successfully"
}
```

### GET /api/v100/currencies/default

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

### GET /api/v100/analytics

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

- **v2.3.0**: **Enhanced API Documentation** - Updated documentation to reflect current implementation
  - Added detailed documentation for Expenses and Incomes endpoints
  - Updated Transactions endpoint with enhanced response structure including category details
  - Improved pagination and filtering documentation
  - Added comprehensive request/response examples for all endpoints
  - Updated version history and implementation status

- **v2.2.0**: **API Versioning System** - Multi-version API support with backward compatibility
  - Complete API versioning implementation with support for v110
  - Version middleware for automatic version detection and validation
  - Version management endpoints for future migration support
  - Enhanced v110 endpoints with advanced analytics and filtering
  - Client library examples for JavaScript, Go, and Python
  - Comprehensive migration strategies and feature detection

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

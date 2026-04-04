# Wallets API Contract

This document outlines the JSON contracts for the new Wallets feature.

**Base Path:** `/api/v100/wallets`  
**Authentication:** Required (Bearer Token)  

---

## 1. Create a Wallet

Creates a new wallet for the authenticated user.

**Endpoint:** `POST /api/v100/wallets`

### Request Body
```json
{
  "name": "My Primary Wallet",
  "amount": 100.50,
  "is_approved": true
}
```
*Note: `amount` is optional and defaults to `0.0` if not provided.*
*Note: `is_approved` is optional and defaults to `true` if not provided.*

### Success Response (201 Created)
```json
{
  "status": "success",
  "data": {
    "wallet": {
      "id": 1,
      "user_id": 42,
      "name": "My Primary Wallet",
      "amount": 100.50,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  },
  "error": null
}
```

---

## 2. Get All Wallets

Retrieves all wallets belonging to the authenticated user.

**Endpoint:** `GET /api/v100/wallets`

### Query Parameters

| Parameter | Type   | Required | Description                      |
| :-------- | :----- | :------- | :------------------------------- |
| `page`    | int    | No       | Page number (defaults to 1).     |
| `limit`   | int    | No       | Number of items per page.        |
| `search`  | string | No       | Search term for the wallet name. |

### Success Response (200 OK)
```json
{
  "status": "success",
  "data": {
    "wallets": [
      {
        "id": 1,
        "user_id": 42,
        "name": "My Primary Wallet",
        "amount": 100.50,
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:00:00Z"
      },
      {
        "id": 2,
        "user_id": 42,
        "name": "Savings",
        "amount": 5000.00,
        "created_at": "2024-01-02T08:30:00Z",
        "updated_at": "2024-01-02T08:30:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "limit": 20,
    "total_pages": 1
  },
  "error": null
}
```

---

## 3. Update a Wallet

Updates the properties of an existing wallet. Only the `name` property can be modified via this endpoint.

**Endpoint:** `PUT /api/v100/wallets/:id`

### Request Body
```json
{
  "name": "Updated Wallet Name"
}
```

### Success Response (200 OK)
```json
{
  "status": "success",
  "data": {
    "wallet": {
      "id": 1,
      "user_id": 42,
      "name": "Updated Wallet Name",
      "amount": 100.50,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-05T09:15:00Z"
    }
  },
  "error": null
}
```
*Note: A `403 Forbidden` error is returned if a user attempts to update a wallet matching an ID they don't own.*

---

## 4. Delete a Wallet

Deletes an existing wallet belonging to the user.

**Endpoint:** `DELETE /api/v100/wallets/:id`

### Success Response (200 OK)
```json
{
  "status": "success",
  "data": {
    "message": "Wallet deleted successfully"
  },
  "error": null
}
```
*Note: A `403 Forbidden` error is returned if a user attempts to delete a wallet matching an ID they don't own.*

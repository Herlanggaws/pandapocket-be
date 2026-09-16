# API Path Notes

API versioning has been removed. All endpoints use the unversioned `/api` prefix.

Examples:

- `POST /api/auth/login`
- `GET /api/transactions`
- `GET /api/categories`
- `GET /health`

This file previously described a multi-version migration strategy (`/api/v100`, `/api/v110`, etc.). That approach is no longer used.

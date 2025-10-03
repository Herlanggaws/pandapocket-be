# Role-Based Authentication System

This document explains how to use the role-based authentication system implemented in PandaPocket.

## Overview

The system now supports three user roles:
- **`user`** - Regular users (default)
- **`admin`** - Administrative users with elevated permissions
- **`super_admin`** - Super administrators with full system access

## Database Changes

### Migration
Run the migration script to add the role column:

```bash
# Option 1: Using the shell script
./scripts/run_add_user_role.sh

# Option 2: Using Go directly
go run scripts/add_user_role.go

# Option 3: Manual SQL execution
psql -d panda_pocket -f scripts/add_user_role.sql
```

### Database Schema
```sql
ALTER TABLE users 
ADD COLUMN role VARCHAR(20) DEFAULT 'user' NOT NULL;

ALTER TABLE users 
ADD CONSTRAINT check_user_role 
CHECK (role IN ('user', 'admin', 'super_admin'));
```

## Usage Examples

### 1. Basic Authentication (Existing)
```go
// Regular user authentication
router.GET("/profile", authMiddleware.RequireAuth(), getUserProfile)
```

### 2. Admin-Only Routes
```go
// Admin routes (admin or super_admin)
adminRoutes := router.Group("/admin")
adminRoutes.Use(authMiddleware.RequireAuth())
adminRoutes.Use(authMiddleware.RequireRole("admin"))

adminRoutes.GET("/users", getAdminUsers)
adminRoutes.GET("/analytics", getAdminAnalytics)
```

### 3. Super Admin-Only Routes
```go
// Super admin routes (super_admin only)
superAdminRoutes := router.Group("/admin/super")
superAdminRoutes.Use(authMiddleware.RequireAuth())
superAdminRoutes.Use(authMiddleware.RequireRole("super_admin"))

superAdminRoutes.DELETE("/users/:id", deleteUser)
superAdminRoutes.POST("/admin", createAdmin)
```

### 4. Role Hierarchy
The system supports role hierarchy:
- `super_admin` can access admin and user routes
- `admin` can access user routes
- `user` can only access user routes

```go
// This works for admin and super_admin
adminRoutes.Use(authMiddleware.RequireRole("admin"))

// This works only for super_admin
superAdminRoutes.Use(authMiddleware.RequireRole("super_admin"))
```

## API Endpoints

### Authentication
- `POST /auth/register` - Register new user (defaults to 'user' role)
- `POST /auth/login` - Login user (returns JWT with role)

### Admin Endpoints (Require Admin Role)
- `GET /admin/users` - List all users
- `GET /admin/analytics` - Get system analytics
- `POST /admin/users/:id/role` - Update user role

### Super Admin Endpoints (Require Super Admin Role)
- `DELETE /admin/super/users/:id` - Delete user
- `POST /admin/super/admin` - Create new admin

## JWT Token Structure

The JWT token now includes the user's role:

```json
{
  "user_id": 123,
  "email": "user@example.com",
  "role": "admin",
  "exp": 1234567890,
  "iat": 1234567890
}
```

## Middleware Usage

### RequireAuth()
Validates JWT token and sets user context:
- `user_id` - User ID from token
- `email` - User email from token  
- `role` - User role from token

### RequireRole(role)
Checks if user has required role or higher:
- `RequireRole("admin")` - Allows admin and super_admin
- `RequireRole("super_admin")` - Allows only super_admin

## Creating Admin Users

### Method 1: Database Direct
```sql
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
UPDATE users SET role = 'super_admin' WHERE email = 'superadmin@example.com';
```

### Method 2: Application Code
```go
// In your application code
user, err := userService.GetUserByEmail(ctx, email)
if err != nil {
    return err
}

adminRole, _ := identity.NewRole("admin")
user.ChangeRole(adminRole)
userService.Save(ctx, user)
```

## Security Considerations

1. **Role Validation**: Roles are validated both in JWT tokens and database
2. **Hierarchy**: Higher roles inherit permissions of lower roles
3. **Default Role**: New users default to 'user' role
4. **Token Security**: Roles are embedded in JWT tokens for stateless authentication

## Testing

### Test Admin Access
```bash
# Login as admin user
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password"}'

# Use the returned token for admin endpoints
curl -X GET http://localhost:8080/admin/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Test Role Hierarchy
```bash
# Super admin can access admin routes
curl -X GET http://localhost:8080/admin/users \
  -H "Authorization: Bearer SUPER_ADMIN_TOKEN"

# Regular user cannot access admin routes
curl -X GET http://localhost:8080/admin/users \
  -H "Authorization: Bearer USER_TOKEN"
# Returns 403 Forbidden
```

## Troubleshooting

### Common Issues

1. **Migration Not Applied**: Ensure the database migration has been run
2. **Invalid Role**: Check that roles are exactly 'user', 'admin', or 'super_admin'
3. **Token Issues**: Verify JWT tokens include the role claim
4. **Permission Denied**: Check that the user has the required role

### Debugging

Check user role in context:
```go
func debugHandler(c *gin.Context) {
    userID := c.GetInt("user_id")
    email := c.GetString("email")
    role := c.GetString("role")
    
    c.JSON(200, gin.H{
        "user_id": userID,
        "email": email,
        "role": role,
    })
}
```

## Migration Checklist

- [x] Add role column to User model
- [x] Update JWT Claims to include role
- [x] Update token service interface
- [x] Update auth middleware to extract role
- [x] Create role-based authorization middleware
- [x] Update user repository to handle role
- [x] Update domain user model
- [x] Create database migration script
- [x] Update login/register use cases
- [x] Create example admin routes
- [x] Document usage and examples

-- Migration script to add role column to users table
-- This script adds a role column with default value 'user' and check constraint

-- Add role column to users table
ALTER TABLE users 
ADD COLUMN role VARCHAR(20) DEFAULT 'user' NOT NULL;

-- Add check constraint to ensure only valid roles are allowed
ALTER TABLE users 
ADD CONSTRAINT check_user_role 
CHECK (role IN ('user', 'admin', 'super_admin'));

-- Create index on role column for better query performance
CREATE INDEX idx_users_role ON users(role);

-- Update existing users to have 'user' role (this is already set by DEFAULT, but explicit for clarity)
UPDATE users SET role = 'user' WHERE role IS NULL;

-- Optional: Create a super admin user (uncomment and modify as needed)
-- INSERT INTO users (email, password_hash, role) 
-- VALUES ('admin@pandapocket.com', '$2a$10$hashedpasswordhere', 'super_admin');

#!/bin/bash

# Script to run the user role migration
# This script adds the role column to the users table

echo "Starting user role migration..."

# Check if DATABASE_URL is set, otherwise use default
if [ -z "$DATABASE_URL" ]; then
    echo "DATABASE_URL not set, using default connection"
    export DATABASE_URL="postgres://postgres:password@localhost:5432/panda_pocket?sslmode=disable"
fi

# Run the migration
cd "$(dirname "$0")/.."
go run scripts/add_user_role.go

echo "Migration completed!"

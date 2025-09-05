#!/bin/bash

# PostgreSQL Setup Script for Panda Pocket

echo "Setting up PostgreSQL environment variables..."

# Set default PostgreSQL configuration
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=panda_pocket

echo "Environment variables set:"
echo "DB_TYPE=$DB_TYPE"
echo "DB_HOST=$DB_HOST"
echo "DB_PORT=$DB_PORT"
echo "DB_USER=$DB_USER"
echo "DB_NAME=$DB_NAME"

echo ""
echo "To make these permanent, add them to your shell profile (.bashrc, .zshrc, etc.)"
echo "Or create a .env file in the backend directory with these values."
echo ""
echo "Next steps:"
echo "1. Start PostgreSQL: docker-compose up -d postgres"
echo "2. Run migration: go run cmd/migrate/main.go ./panda_pocket.db"
echo "3. Start the application: go run main.go"

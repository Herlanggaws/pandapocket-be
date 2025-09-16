#!/bin/bash

# Start Panda Pocket with PostgreSQL
echo "Starting Panda Pocket with PostgreSQL..."

# Set PostgreSQL environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=herlangga.wicaksono
export DB_PASSWORD=""
export DB_NAME=panda_pocket

echo "Environment variables set:"
echo "DB_HOST=$DB_HOST"
echo "DB_PORT=$DB_PORT"
echo "DB_USER=$DB_USER"
echo "DB_NAME=$DB_NAME"

echo ""
echo "Starting application..."
go run main.go

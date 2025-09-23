#!/bin/bash

# Add currencies script for PandaPocket
# This script provides multiple ways to add currencies to the database

set -e

echo "🚀 PandaPocket Currency Setup Script"
echo "====================================="

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the backend directory"
    exit 1
fi

# Set default database connection parameters
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-herlangga.wicaksono}
DB_PASSWORD=${DB_PASSWORD:-}
DB_NAME=${DB_NAME:-panda_pocket}

echo "📊 Database Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo ""

# Function to run Go script
run_go_script() {
    echo "🔧 Running Go script to add currencies..."
    
    # Set environment variables
    export DB_HOST DB_PORT DB_USER DB_PASSWORD DB_NAME
    
    # Run the Go script
    go run scripts/add_currencies.go
}

# Function to run SQL script
run_sql_script() {
    echo "🔧 Running SQL script to add currencies..."
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        echo "❌ Error: psql command not found. Please install PostgreSQL client tools."
        exit 1
    fi
    
    # Run the SQL script
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f scripts/add_currencies.sql
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [go|sql|auto]"
    echo ""
    echo "Options:"
    echo "  go    - Use Go script (recommended)"
    echo "  sql   - Use SQL script directly"
    echo "  auto  - Try Go first, fallback to SQL"
    echo ""
    echo "Environment variables:"
    echo "  DB_HOST     - Database host (default: localhost)"
    echo "  DB_PORT     - Database port (default: 5432)"
    echo "  DB_USER     - Database user (default: herlangga.wicaksono)"
    echo "  DB_PASSWORD - Database password"
    echo "  DB_NAME     - Database name (default: panda_pocket)"
}

# Main script logic
case "${1:-auto}" in
    "go")
        run_go_script
        ;;
    "sql")
        run_sql_script
        ;;
    "auto")
        echo "🔄 Trying Go script first..."
        if run_go_script; then
            echo "✅ Go script completed successfully"
        else
            echo "⚠️  Go script failed, trying SQL script..."
            run_sql_script
        fi
        ;;
    "help"|"-h"|"--help")
        show_usage
        ;;
    *)
        echo "❌ Invalid option: $1"
        show_usage
        exit 1
        ;;
esac

echo ""
echo "🎉 Currency setup completed!"
echo "You can now use the /api/currencies endpoint to see the currencies."

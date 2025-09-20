#!/bin/bash

# PandaPocket Configuration Display Script
echo "📊 PandaPocket Configuration:"
echo ""

if [ -f ".env" ]; then
    echo "📁 .env file contents:"
    echo "----------------------------------------"
    cat .env
    echo "----------------------------------------"
else
    echo "⚠️  No .env file found"
fi

echo ""
echo "🔧 Current environment variables:"
echo "   DB_HOST: ${DB_HOST:-localhost (default)}"
echo "   DB_PORT: ${DB_PORT:-5432 (default)}"
echo "   DB_USER: ${DB_USER:-herlangga.wicaksono (default)}"
echo "   DB_NAME: ${DB_NAME:-panda_pocket (default)}"
echo "   GIN_MODE: ${GIN_MODE:-debug (default)}"

#!/bin/bash

# PandaPocket Backend Server Startup Script
echo "🚀 Starting PandaPocket Backend Server..."

echo "📊 Reading configuration from .env file..."
if [ -f ".env" ]; then
    echo "   ✅ .env file found - using configuration from file"
else
    echo "   ⚠️  .env file not found - using default configuration"
fi
echo ""

# Start the server (will automatically load .env file)
./panda-pocket

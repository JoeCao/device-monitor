#!/bin/bash

echo "=== Device Monitor Go - Quick Start ==="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Check if Node.js is installed (for frontend)
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

echo "✅ Prerequisites check passed"
echo

# Install dependencies
echo "📦 Installing dependencies..."
make deps
echo

# Build the project
echo "🔨 Building project..."
make build
echo

echo "✅ Build complete!"
echo
echo "To run the application:"
echo "  Development mode: make dev"
echo "  Production mode:  make run"
echo
echo "The application will be available at http://localhost:3000"
echo
echo "API endpoints:"
echo "  - Health check: GET /api/health"
echo "  - Start device: POST /api/webhooks/device/start?deviceName={deviceId}"
echo "  - End device:   POST /api/webhooks/device/end?deviceName={deviceId}"
echo "  - Sessions:     GET /api/sessions"
echo
echo "Don't forget to configure your .env file with IoT platform credentials!"
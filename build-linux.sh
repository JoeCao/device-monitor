#!/bin/bash

echo "Building device-monitor for Linux x86_64 from macOS..."

# Build frontend first
echo "Building frontend..."
cd web && npm run build && cd ..

# Check if user has Docker
if command -v docker &> /dev/null; then
    echo "Using Docker for cross-compilation..."
    make build-linux-docker
elif command -v zig &> /dev/null; then
    echo "Using zig for cross-compilation..."
    make build-linux-zig
elif command -v x86_64-linux-musl-gcc &> /dev/null; then
    echo "Using musl-cross for cross-compilation..."
    make build-linux-musl
else
    echo "No suitable cross-compiler found!"
    echo ""
    echo "Please install one of the following:"
    echo "1. Docker Desktop (recommended)"
    echo "2. zig: brew install zig"
    echo "3. musl-cross: brew install FiloSottile/musl-cross/musl-cross"
    echo ""
    echo "Then run 'make build-linux' again."
    exit 1
fi

# Check if build succeeded
if [ -f "./device-monitor-linux-amd64" ]; then
    echo ""
    echo "Build successful!"
    echo "Binary: ./device-monitor-linux-amd64"
    echo "Architecture: $(file ./device-monitor-linux-amd64)"
else
    echo "Build failed!"
    exit 1
fi
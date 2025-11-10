.PHONY: help dev build run clean deps frontend-deps backend-deps test

# Default target
help:
	@echo "Available commands:"
	@echo "  make deps          - Install all dependencies (frontend + backend)"
	@echo "  make dev           - Run in development mode"
	@echo "  make build         - Build for production"
	@echo "  make run           - Run production build"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make test          - Run tests"

# Install all dependencies
deps: backend-deps frontend-deps

# Install backend dependencies
backend-deps:
	@echo "Installing backend dependencies..."
	go mod download
	go mod tidy

# Install frontend dependencies
frontend-deps:
	@echo "Installing frontend dependencies..."
	cd web && npm install

# Development mode - run both frontend and backend
dev:
	@echo "Starting development servers..."
	@trap 'kill %1' INT; \
	(cd web && npm run dev) & \
	NODE_ENV=development go run main.go

# Build for production
build: frontend-build
	@echo "Building Go binary..."
	CGO_ENABLED=1 go build -ldflags="-s -w" -o device-monitor main.go
	@echo "Build complete! Binary: ./device-monitor"

# Build frontend
frontend-build:
	@echo "Building frontend..."
	cd web && npm run build
	@echo "Frontend build complete!"

# Run production build
run:
	@if [ ! -f ./device-monitor ]; then \
		echo "Binary not found. Run 'make build' first."; \
		exit 1; \
	fi
	NODE_ENV=production ./device-monitor

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f device-monitor
	rm -rf web/dist
	rm -rf database/*.db
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Create database directory
init-db:
	@mkdir -p database

# Docker build (optional)
docker-build:
	docker build -t device-monitor-go .

# Cross-platform builds
# Method 1: Using Docker (recommended for macOS Silicon)
build-linux-docker:
	@echo "Building for Linux x86_64 using Docker..."
	cd web && npm run build
	@echo "Building Linux binary in Docker container..."
	docker build -f Dockerfile.build -o . .
	@echo "Linux x86_64 binary created: ./device-monitor-linux-amd64"

# Method 2: Using installed cross-compiler (requires musl-cross)
build-linux-musl:
	@echo "Building for Linux x86_64 with musl..."
	cd web && npm run build
	@echo "Note: This requires x86_64-linux-musl-gcc installed via:"
	@echo "  brew install FiloSottile/musl-cross/musl-cross"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-musl-gcc go build -ldflags="-s -w -extldflags '-static'" -o device-monitor-linux-amd64 main.go

# Method 3: Using zig as cross-compiler (requires zig)
build-linux-zig:
	@echo "Building for Linux x86_64 using zig..."
	cd web && npm run build
	@echo "Note: This requires zig installed via: brew install zig"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC="zig cc -target x86_64-linux" go build -ldflags="-s -w" -o device-monitor-linux-amd64 main.go

# Simplified alias
build-linux: build-linux-docker

build-darwin:
	@echo "Building for macOS..."
	cd web && npm run build
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o device-monitor-darwin main.go

build-windows:
	@echo "Building for Windows..."
	cd web && npm run build
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o device-monitor.exe main.go
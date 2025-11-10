# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Device Operation Monitor (Go Version) - A Go implementation of the device operation monitoring system with identical functionality to the Node.js version but simpler deployment.

## Key Commands

### Development
```bash
# Install all dependencies
make deps

# Run development mode (frontend + backend)
make dev

# Run backend only
go run main.go

# Run frontend only  
cd web && npm run dev
```

### Production
```bash
# Build production binary with embedded frontend
make build

# Run production server
make run
```

### Testing
```bash
# Test IoT connection
curl http://localhost:3000/api/iot/test-connection

# Test webhook
curl -X POST http://localhost:3000/api/webhooks/device/start?deviceName=test123 \
  -H "Content-Type: application/json" \
  -d '{"power":"on"}'
```

## Architecture Overview

### Single Binary Architecture
- Go backend serves both API and static frontend files
- Frontend built files are embedded into the Go binary using `embed` package
- Development mode proxies to Vite dev server, production serves embedded files

### Backend Structure
- **Entry**: `main.go` - Gin router setup and static file serving
- **API Handlers**: `/api/handlers/` - All HTTP endpoint handlers
- **Models**: `/models/` - Database models and business logic
- **Services**: `/services/` - IoT integration service with OAuth2
- **Database**: SQLite with sqlx for query building

### API Compatibility
This Go version maintains 100% API compatibility with the Node.js version:
- Same endpoints, request/response formats
- Same database schema
- Same IoT platform integration
- Frontend code is directly reused without modification

### Key Differences from Node.js Version

1. **Deployment**: Single binary vs npm install + node
2. **Static Files**: Embedded at compile time vs runtime serving
3. **Dependencies**: Go modules vs npm packages
4. **Error Handling**: Go's explicit error returns vs try/catch

### Development Workflow

1. Backend changes: Edit Go files and restart server
2. Frontend changes: Vite hot-reloads automatically
3. Building: `make build` creates a single deployable binary

### Important Implementation Details

- **Timestamps**: All times stored as UTC, formatted as RFC3339
- **IoT Data**: Not stored locally, fetched real-time from platform
- **Session Status**: "running" or "completed" only
- **Webhook Auth**: No authentication required (same as Node.js version)
- **Hilbert Data**: Kept as string to preserve JSON array format

### Common Issues

- **CGO Required**: SQLite driver requires CGO_ENABLED=1
- **Embed Path**: Must use `//go:embed all:web/dist` syntax
- **Frontend Build**: Must build frontend before Go binary for production
- **CORS**: Already handled by middleware for all origins
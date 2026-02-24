# Quick Start Guide

Get up and running with golang-sample in 5 minutes.

## Prerequisites

- **Go 1.25+** - [Download](https://go.dev/dl/)
- **PostgreSQL 15+** - [Download](https://www.postgresql.org/download/)
- **Docker** (optional) - [Download](https://www.docker.com/products/docker-desktop)
- **Git** - [Download](https://git-scm.com/downloads)

## Installation

### 1. Clone Repository

```bash
git clone https://github.com/haipham22/golang-sample.git
cd golang-sample
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment

Create a `.env` file or set environment variables:

```bash
# Application
APP_ENV=development
APP_DEBUG=true
APP_PORT=8080

# Database
DB_DSN="host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable"

# JWT
API_SECRET="your-jwt-secret-key"
```

### 4. Start PostgreSQL

**Option A: Docker Compose (Recommended)**

```bash
docker-compose up -d
```

**Option B: Local PostgreSQL**

```bash
# Create database
createdb golang_sample

# Start PostgreSQL service
brew services start postgresql  # macOS
sudo service postgresql start   # Linux
```

### 5. Run Application

```bash
# Development mode
go run main.go serverd

# Or build and run
go build -o bin/serverd .
./bin/serverd
```

The API will be available at `http://localhost:8080`

## Verify Installation

### Check Health Status

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": "2026-02-24T10:30:00Z",
  "service": "golang-sample-api",
  "database": "ok"
}
```

### Check Readiness (Kubernetes)

```bash
curl http://localhost:8080/readyz
```

### Check Liveness (Kubernetes)

```bash
curl http://localhost:8080/livez
```

## Quick Test

### Register a User

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePassword123!",
    "full_name": "Test User"
  }'
```

Expected response:
```json
{
  "id": "1",
  "username": "testuser",
  "email": "test@example.com",
  "full_name": "Test User"
}
```

### Login

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "SecurePassword123!"
  }'
```

Expected response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "1",
    "username": "testuser",
    "email": "test@example.com"
  }
}
```

### Access Protected Endpoint

```bash
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## View API Documentation

### Swagger UI (Development Only)

```bash
# Generate Swagger docs
./scripts/generate-swagger.sh

# Access Swagger UI
open http://localhost:8080/document/index.html
```

### Prometheus Metrics

```bash
curl http://localhost:8080/metrics
```

## Docker Quick Start

### Build Image

```bash
docker build -t golang-sample .
```

### Run Container

```bash
docker run -d \
  --name golang-sample \
  -p 8080:8080 \
  -e APP_ENV=production \
  -e DB_DSN="host=host.docker.internal user=postgres password=password dbname=golang_sample port=5432 sslmode=disable" \
  -e API_SECRET="your-secret" \
  golang-sample
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Troubleshooting

### Port Already in Use

```bash
# Change port
export APP_PORT=8081
go run main.go serverd
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check database exists
psql -U postgres -l | grep golang_sample
```

### Module Download Issues

```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
```

## Next Steps

- Read [Development Guide](development.md) for testing and building
- Check [System Architecture](system-architecture.md) for architecture details
- Review [Code Standards](code-standards.md) for contribution guidelines

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/haipham22/golang-sample/issues)
- **Documentation**: [docs/](.)
- **Govern Package**: [github.com/haipham22/govern](https://github.com/haipham22/govern)

# Golang Sample API

[![Build docker image](https://github.com/haipham22/golang-sample/actions/workflows/push.yml/badge.svg)](https://github.com/haipham22/golang-sample/actions/workflows/push.yml)

A production-ready Go API demonstrating clean architecture principles, using Go 1.25+ and Echo framework with govern package integration.

## ✅ Security Status

**UPDATE (2026-02-24):** All critical security vulnerabilities have been **FIXED** through govern package integration. This project is now **production-ready**.

**Fixed Issues:**
- ✅ Password verification bug corrected
- ✅ Password hashing implemented (bcrypt)
- ✅ JWT authentication middleware added
- ✅ Health check endpoints added
- ✅ Graceful shutdown implemented
- ✅ Prometheus metrics integrated

**See:** [Govern Integration Plan](plans/260224-1557-govern-integration/plan.md) | [Code Review Summary](plans/260224-1206-codebase-review/REVIEW_SUMMARY.md)

## Documentation

| Document | Description |
|----------|-------------|
| [Project Overview & PDR](docs/project-overview-pdr.md) | Product requirements, roadmap, and current status |
| [Codebase Summary](docs/codebase-summary.md) | Complete directory structure, API endpoints, and dependencies |
| [Code Standards](docs/code-standards.md) | Naming conventions, style guidelines, and best practices |
| [System Architecture](docs/system-architecture.md) | Clean architecture layers, design patterns, and data flow |

## Quick Start

### Prerequisites
- Go 1.25+
- PostgreSQL 15+
- Docker (optional)

### Installation

```bash
# Clone and install
git clone https://github.com/haipham22/golang-sample.git
cd golang-sample
go mod download

# Set environment variables
export APP_ENV=development
export DB_DSN="host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable"
export API_SECRET="your-jwt-secret"

# Run with Docker Compose (includes PostgreSQL)
docker-compose up -d

# Or run locally
go run main.go serverd
```

### Testing the API

```bash
# Register a new user
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123","full_name":"Test User"}'

# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# Check health
curl http://localhost:8080/health

# Check metrics
curl http://localhost:8080/metrics
```

### Generate Swagger Docs
```bash
go install github.com/swaggo/swag/cmd/swag@latest
./scripts/generate-swagger.sh
```

Access Swagger UI: `http://localhost:8080/document/index.html`

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│  HTTP Routes │ Middlewares │ Handlers (Echo)                │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                       │
│  Controllers │ Validators │ Schemas                         │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                           │
│  Models │ Interfaces │ Services                            │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                      │
│  Storage │ Database (PostgreSQL) │ External Services       │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
golang-sample/
├── cmd/                    # Application entry points
├── internal/api/           # API layer (routes, handlers, storage)
├── pkg/                    # Public packages (models, config, utils)
├── scripts/                # Build and deployment scripts
├── docs/                   # Comprehensive documentation
├── main.go                 # Entry point
├── Dockerfile              # Multi-stage build
└── compose.yml             # Docker Compose setup
```

## API Endpoints

### Authentication
| Method | Endpoint | Description | Status |
|--------|----------|-------------|--------|
| POST | `/api/login` | User login with JWT token | ✅ Working |
| POST | `/api/register` | User registration | ✅ Working |

### Health Checks ✅ NEW
| Method | Endpoint | Description | Status |
|--------|----------|-------------|--------|
| GET | `/health` | Application health status | ✅ Implemented |
| GET | `/readyz` | Readiness probe (Kubernetes) | ✅ Implemented |
| GET | `/livez` | Liveness probe (Kubernetes) | ✅ Implemented |

### Monitoring ✅ NEW
| Method | Endpoint | Description | Status |
|--------|----------|-------------|--------|
| GET | `/metrics` | Prometheus metrics | ✅ Implemented |

### Development
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/document/*` | Swagger UI (dev only) |

## Configuration

```bash
# Application
APP_ENV=development          # development | staging | production
APP_DEBUG=true
APP_PORT=8080

# Database (DSN format)
DB_DSN="host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable"

# JWT
API_SECRET=your-secret-key   # JWT signing secret
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Framework | Echo v4 |
| Database | PostgreSQL + GORM |
| Auth | JWT (golang-jwt/jwt/v5) + Bcrypt |
| DI | Google Wire |
| Logging | Zap |
| Monitoring | Prometheus |
| Config | Viper |
| CLI | Cobra |

## Development

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build binary
go build -o bin/api cmd/api.go

# Build Docker image
docker build -t golang-sample .
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [Code Standards](docs/code-standards.md) for detailed guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

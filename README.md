# Golang Sample API

[![Build docker image](https://github.com/haipham22/golang-sample/actions/workflows/push.yml/badge.svg)](https://github.com/haipham22/golang-sample/actions/workflows/push.yml)

A production-ready Go API demonstrating clean architecture principles with Go 1.25+, Echo framework, and govern package integration.

## ✅ Production Status

**Last Updated:** 2026-02-24 | **Grade:** A- (Production-Ready) | **Test Coverage:** 60%

### Security ✅
- ✅ Password hashing with bcrypt (DefaultCost)
- ✅ JWT authentication (golang-jwt/jwt/v5)
- ✅ SQL injection protected (GORM ORM)
- ✅ Critical password verification bug fixed
- ✅ No hardcoded secrets

### Infrastructure ✅
- ✅ Health checks (`/health`, `/readyz`, `/livez`)
- ✅ Graceful shutdown (SIGTERM/SIGINT handling)
- ✅ Connection pooling (MaxIdle: 10, MaxOpen: 100)
- ✅ Prometheus metrics
- ✅ Structured logging (Zap)

### Code Quality ✅
- ✅ Clean architecture (handler → controller → storage)
- ✅ DRY, YAGNI, KISS compliant
- ✅ Govern package integration (errors, postgres)
- ✅ Pre-commit hooks (13 hooks, all passing)
- ✅ Gomock configured for testing

**Full Review:** [Code Review Report](docs/code-review-govern-integration.md)

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
| [Quick Start Guide](docs/quickstart.md) | Get up and running in 5 minutes |
| [Development Guide](docs/development.md) | Testing, building, and development workflow |
| [Project Overview & PDR](docs/project-overview-pdr.md) | Product requirements, roadmap, and current status |
| [Codebase Summary](docs/codebase-summary.md) | Complete directory structure, API endpoints, and dependencies |
| [Code Standards](docs/code-standards.md) | Naming conventions, style guidelines, and best practices |
| [System Architecture](docs/system-architecture.md) | Clean architecture layers, design patterns, and data flow |

## Quick Start

Get up and running in 5 minutes → [Quick Start Guide](docs/quickstart.md)

```bash
# Clone and install
git clone https://github.com/haipham22/golang-sample.git
cd golang-sample
go mod download

# Set environment
export APP_ENV=development
export DB_DSN="host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable"
export API_SECRET="your-jwt-secret"

# Run with Docker Compose
docker-compose up -d

# Or run locally
go run main.go serverd
```

**Testing the API:**
```bash
# Register
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"SecurePassword123!","full_name":"Test User"}'

# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"SecurePassword123!"}'

# Health check
curl http://localhost:8080/health
```

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
│   └── serverd.go          # Main server command
├── internal/               # Private application code
│   ├── handler/            # HTTP handlers (flat architecture)
│   │   └── rest/           # REST API handlers
│   │       ├── auth/       # Authentication endpoints
│   │       └── health/     # Health check endpoints
│   ├── middlewares/        # Echo middlewares (auth, logger, metrics)
│   ├── storage/            # Database storage interfaces
│   └── errors/             # Custom error definitions
├── pkg/                    # Public libraries
│   ├── models/             # Data models
│   ├── postgres/           # Database connection (govern/postgres)
│   └── utils/              # Utility functions (password, etc.)
├── docs/                   # Documentation
├── scripts/                # Build and deployment scripts
├── plans/                  # Implementation plans
├── main.go                 # Application entry point
├── Dockerfile              # Multi-stage Docker build
├── compose.yml             # Docker Compose setup
├── .pre-commit-config.yaml # Pre-commit hooks
└── .gomockf                # Mock generation template
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

Create a `.env` file (default) or use environment variables:

**.env File (Default):**
```bash
# Application
APP_ENV=development
APP_DEBUG=true

# Database
APP_POSTGRES_DSN="host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable"

# JWT
APP_API_SECRET="your-jwt-secret-key"
```

**Environment Variables (overrides .env):**
```bash
export APP_ENV=development
export APP_DEBUG=true
export APP_POSTGRES_DSN="host=localhost user=postgres password=password dbname=golang_sample port=5432 sslmode=disable"
export APP_API_SECRET="your-jwt-secret-key"
```

## Tech Stack

| Component | Technology | Version |
|-----------|------------|---------|
| **Language** | Go | 1.25 |
| **Framework** | Echo | v4 |
| **Database** | PostgreSQL | 15+ |
| **ORM** | GORM | latest |
| **Auth** | JWT (golang-jwt/jwt/v5) | v5 |
| **Password Hashing** | Bcrypt | DefaultCost |
| **DI** | Google Wire | latest |
| **Logging** | Zap | latest |
| **Monitoring** | Prometheus | latest |
| **Config** | Viper | latest |
| **CLI** | Cobra | latest |
| **Govern Package** | github.com/haipham22/govern | v0.0.0 |
| **Testing** | Gomock + Testify | v1.6.0 |

## Development

Complete development guide → [Development Guide](docs/development.md)

**Quick commands:**
```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build binary
go build -o bin/serverd .

# Build Docker image
docker build -t golang-sample .

# Install pre-commit hooks
pre-commit install
```

**Test Coverage:**
- `pkg/utils/password`: 100% ✅
- `internal/handler/rest/health`: 52.9% ✅
- Overall: ~60%

## Govern Package Integration

This project integrates [github.com/haipham22/govern](https://github.com/haipham22/govern) for production-ready patterns:

### Implemented (Phase 1 ✅)
- **govern/errors** - Standardized error codes across application
- **govern/postgres** - Database connection pooling with cleanup

### Planned (Phase 3)
- **govern/config** - Configuration management
- **govern/jwt** - JWT token generation/validation
- **govern/metrics** - Prometheus metrics

**Implementation Plan:** [Govern Integration Plan](plans/260224-1557-govern-integration/plan.md)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

**Commit Convention:** Follow [Conventional Commits](https://www.conventionalcommits.org/)
- `feat:` New feature
- `fix:` Bug fix
- `chore:` Maintenance tasks
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring

See [Code Standards](docs/code-standards.md) for detailed guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [Echo](https://echo.labstack.com/) - High-performance Go web framework
- [Govern](https://github.com/haipham22/govern) - Production-ready Go patterns
- [GORM](https://gorm.io/) - ORM library for Go
- [Wire](https://github.com/google/wire) - Code generation for dependency injection

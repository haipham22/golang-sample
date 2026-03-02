# Golang Sample API

[![Build docker image](https://github.com/haipham22/golang-sample/actions/workflows/push.yml/badge.svg)](https://github.com/haipham22/golang-sample/actions/workflows/push.yml)
[![Test](https://github.com/haipham22/golang-sample/actions/workflows/test.yml/badge.svg)](https://github.com/haipham22/golang-sample/actions/workflows/test.yml)

A production-ready Go API demonstrating clean architecture principles with Go 1.25+, Echo framework, and govern package stack.

## ✅ Production Status

**Last Updated:** 2026-02-26 | **Grade:** A (Production-Ready) | **Test Coverage:** 83.3%

### Security ✅
- ✅ Password hashing with bcrypt (cost 10)
- ✅ JWT authentication (golang-jwt/jwt/v5)
- ✅ SQL injection protected (GORM ORM)
- ✅ Input validation with TrimStrings middleware
- ✅ No PII in logs (usernames/emails redacted)
- ✅ Generic error messages (no internal details leaked)

### Infrastructure ✅
- ✅ Health checks (`/health`, `/readyz`, `/livez`)
- ✅ Graceful shutdown with govern/graceful.Run()
- ✅ Connection pooling (MaxIdle: 10, MaxOpen: 100)
- ✅ Prometheus metrics integration
- ✅ Structured logging (Zap)
- ✅ Signal handling (SIGINT/SIGTERM)

### Performance ✅
- ✅ Sonic JSON parser (2-3x faster than stdlib)
- ✅ Optimized string trimming (4.7ns, 0 allocations)
- ✅ Precompiled regex patterns
- ✅ Connection pooling configured

### Code Quality ✅
- ✅ Clean architecture (Handler → Controller → Service → Storage)
- ✅ DRY, YAGNI, KISS compliant
- ✅ Govern package stack fully integrated
- ✅ Pre-commit hooks (13 hooks, passing)
- ✅ Mockery for test generation
- ✅ Wire for compile-time DI


**Reviews:** [Govern Integration](docs/code-review-govern-integration.md) | [Code Fixes](docs/code-review-fixes.md)

## Documentation

| Document | Description |
|----------|-------------|
| [Quick Start Guide](docs/quickstart.md) | Get up and running in 5 minutes |
| [Development Guide](docs/development.md) | Testing, building, and development workflow |
| [Project Overview & PDR](docs/project-overview-pdr.md) | Product requirements, roadmap, and current status |
| [Codebase Summary](docs/codebase-summary.md) | Complete directory structure, API endpoints, and dependencies |
| [Code Standards](docs/code-standards.md) | Naming conventions, style guidelines, and best practices |
| [System Architecture](docs/system-architecture.md) | Clean architecture layers, design patterns, and data flow |
| [Govern Runner Migration](docs/govern-runner-migration.md) | Migration to graceful.Run() helper |
| [TrimStrings Middleware](docs/trim-strings-middleware.md) | Automatic string trimming middleware |
| [Code Review Fixes](docs/code-review-fixes.md) | Recent fixes from code review |

## Quick Start

Get up and running in 5 minutes → [Quick Start Guide](docs/quickstart.md)

## Architecture

This project follows **Clean Architecture** principles with clear layer separation:

### Layers

1. **HTTP Handler Layer** (`internal/handler/rest/`)
   - Echo framework integration
   - Request binding and validation
   - Response mapping (JSON)
   - No business logic
   - **Controller Layer** (`internal/handler/rest/controllers/`)
     - Request orchestration
     - Calls service layer
     - Error handling and mapping

2. **Service Layer** (`internal/service/auth/`)
   - Business logic
   - JWT generation/validation
   - Password hashing
   - Domain operations
   - Protocol-agnostic

3. **Domain Model Layer** (`internal/model/`)
   - Pure domain entities
   - Business rules
   - No external dependencies
   - Clean separation from persistence

4. **Storage Layer** (`internal/storage/user/`)
   - Database operations
   - GORM ORM entities
   - Data access interface
   - Domain↔ORM conversion

### Dependency Flow

```text
HTTP Request → Handler → Controller → Service → Domain Model ← Storage
                        ↓              ↓            ↓           ↓
                   Orchestration   Business    Domain     Database
                                    Logic       Logic
```

### Key Design Patterns

- **Dependency Injection** - Wire for compile-time DI
- **Interface Segregation** - Service interfaces in handler layer
- **Adapter Pattern** - Password hasher wraps pkg/utils/password
- **Composition Root** - Wire providers at application startup
- **YAGNI Compliance** - No unnecessary abstraction layers

## Project Structure

```text
golang-sample/
├── cmd/                        # Application entry points
│   ├── serverd.go             # Main server command (uses govern/graceful.Run)
│   └── root.go                # Root command configuration
├── internal/
│   ├── handler/               # HTTP handlers
│   │   └── rest/              # Echo HTTP handlers
│   │       ├── controllers/   # Request controllers
│   │       │   ├── auth/      # Auth endpoints
│   │       │   └── health/    # Health check endpoints
│   │       ├── middlewares/   # HTTP middlewares
│   │       │   ├── cors.go    # CORS configuration
│   │       │   ├── security.go # Security headers
│   │       │   ├── compression.go # Gzip compression
│   │       │   └── ratelimit.go # Rate limiting
│   │       ├── swagger/       # API documentation
│   │       ├── handler.go     # HTTP server setup
│   │       ├── routes.go      # Route registration
│   │       └── wire.go        # Dependency injection
│   ├── service/               # Service layer (business logic)
│   │   └── auth/              # Auth service implementation
│   ├── storage/               # Storage interface
│   │   └── user/              # User storage implementation
│   ├── model/                 # Domain models (pure)
│   ├── orm/                   # ORM models (GORM)
│   ├── schemas/               # DTOs and request/response models
│   └── validator/             # Custom validators
├── pkg/                       # Public libraries
│   ├── config/                # Configuration management
│   └── utils/                 # Utility functions
│       └── password/          # Password hashing (bcrypt)
├── plans/                     # Implementation plans
├── docs/                      # Documentation
└── .github/                   # GitHub workflows
```

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

## Security

### Environment Variables
**CRITICAL:** Never commit production credentials. Always use environment variables or secure secret management.

| Variable | Requirements | Generate |
|----------|--------------|----------|
| `APP_API_SECRET` | 32+ characters, random | `openssl rand -base64 32` |
| `APP_POSTGRES_DSN` password | 16+ characters, mixed case | `openssl rand -base64 24` |

### Pre-Production Checklist
- [ ] JWT secret is 32+ characters
- [ ] Database password is strong (16+ chars, mixed case)
- [ ] Debug mode disabled (`APP_DEBUG=false`)
- [ ] HTTPS enforced in production
- [ ] CORS configured with allowed origins
- [ ] Rate limiting configured on auth endpoints

See [Security Checklist](docs/security-checklist.md) for complete deployment guidelines.

## Tech Stack

| Component            | Technology  | Version | Notes                          |
|----------------------|-------------|---------|--------------------------------|
| **Language**         | Go          | 1.25    | Latest stable                  |
| **Framework**        | Echo        | v4.15+  | High-performance web framework |
| **Database**         | PostgreSQL  | 15+     | Production-grade               |
| **ORM**              | GORM        | v1.31+  | Feature-rich ORM               |
| **Auth**             | JWT         | v5.3+   | golang-jwt/jwt                 |
| **Password Hashing** | Bcrypt      | cost 10 | Secure by default              |
| **JSON Parser**      | Sonic       | v1.15+  | 2-3x faster than stdlib        |
| **DI**               | Google Wire | v0.7+   | Compile-time DI                |
| **Logging**          | Zap         | latest  | Structured logging             |
| **Testing**          | Mockery     | latest  | Mock generation                |
| **Config**           | Viper       | latest  | Configuration management       |
| **CLI**              | Cobra       | latest  | CLI framework                  |
| **Govern Stack**     | govern      | v0.0.0+ | Production patterns            |
| **Metrics**          | Prometheus  | latest  | Observability                  |

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

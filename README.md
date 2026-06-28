# Govern

Production-ready Go packages for building scalable microservices and web applications.

[![Test Govern Library](https://github.com/haipham22/govern/actions/workflows/test.yml/badge.svg)](https://github.com/haipham22/govern/actions/workflows/test.yml)
[![Test Sample App](https://github.com/haipham22/govern/actions/workflows/test-sample.yml/badge.svg)](https://github.com/haipham22/govern/actions/workflows/test-sample.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/haipham22/govern)](https://goreportcard.com/report/github.com/haipham22/govern)

## Overview

Govern is a collection of production-tested Go packages implementing common patterns for robust,
scalable web applications and microservices. Each package is framework-agnostic and can be used
independently or combined with the others.

## Repository Layout

This is a **monorepo** with two independent modules (no `go.work` — the sample imports govern
as an external dependency via a `replace` directive):

```
govern/                              # Module: github.com/haipham22/govern
├── http/                            # HTTP server, Echo integration, middleware
├── database/                        # Postgres, Redis
├── config/                          # Config (YAML, .env, env vars)
├── errors/                          # Standardized error handling
├── log/                             # Structured logging (Zap)
├── graceful/                        # Graceful shutdown
├── retry/                           # Exponential backoff retry
├── cron/                            # Cron scheduler
├── mq/                              # Message queues (asynq)
├── metrics/                         # Prometheus metrics
├── healthcheck/                     # Health check endpoints
├── go.mod                           # Module: github.com/haipham22/govern
├── Makefile                         # Govern library targets
└── examples/
    └── golang-sample/               # Module: github.com/haipham22/golang-sample
        ├── cmd/                     # Cobra entry points
        ├── internal/                # Clean-architecture layers
        ├── pkg/                     # config, postgres, utils
        └── go.mod                   # require govern v0.0.0 + replace => ../../
```

## Packages

### HTTP ([`http/`](http/))
- `http/echo` — Echo framework integration with middleware
- `http/jwt` — JWT authentication middleware
- `http/middleware` — Common HTTP middleware (CORS, security, compression)

### Database ([`database/`](database/))
- `database/postgres` — PostgreSQL integration
- `database/redis` — Redis client integration

### Core Services
- [`config/`](config/) — Configuration management (YAML, .env, environment variables)
- [`errors/`](errors/) — Standardized error handling and packaging
- [`log/`](log/) — Structured logging with Zap
- [`graceful/`](graceful/) — Graceful shutdown handling
- [`retry/`](retry/) — Exponential backoff retry logic

### Background Processing
- [`cron/`](cron/) — Cron scheduler integration
- [`mq/asynq`](mq/) — Asynq task queue integration
- [`metrics/`](metrics/) — Prometheus metrics integration
- [`healthcheck/`](healthcheck/) — Health check endpoints

## Quick Start

```bash
go get github.com/haipham22/govern/http
go get github.com/haipham22/govern/config
go get github.com/haipham22/govern/log
```

## Sample Application

See [`examples/golang-sample/`](examples/golang-sample/) for a complete sample application
demonstrating govern package usage with clean architecture (Handler → Controller → Service → Storage),
JWT auth, and GORM with PostgreSQL.

```bash
cd examples/golang-sample
mise install
make run
```

The sample imports govern as an external dependency. For local development, its `go.mod` uses:

```
require github.com/haipham22/govern v0.0.0
replace github.com/haipham22/govern => ../../
```

## Development

```bash
mise install                  # Install Go + tools
make test                     # Run govern library tests
make build                    # Build govern packages
make lint                     # Run linters
```

Sample app commands run from `examples/golang-sample/`:

```bash
cd examples/golang-sample
make test                     # Sample app tests
make build                    # Build bin/serverd
```

## Documentation

- [Development Guide](docs/development.md) — Testing, building, workflow
- [Code Standards](docs/code-standards.md) — Naming, style, best practices
- [Sample App README](examples/golang-sample/README.md) — Sample app guide

Detailed Go development rules are in [`.claude/rules/`](.claude/rules/).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Follow [Conventional Commits](https://www.conventionalcommits.org/).

## License

MIT License — see [LICENSE](LICENSE).

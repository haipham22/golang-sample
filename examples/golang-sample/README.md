# Golang Sample Application

Sample application demonstrating the [govern](../..) library with clean architecture principles.

**Module**: `github.com/haipham22/golang-sample`
**Govern Library**: `github.com/haipham22/govern` (external import via `replace` directive for local development)

## Quick Start

```bash
# Install dependencies & tools
mise install

# Run the server
make run

# Run tests
make test

# Build binary (outputs bin/serverd)
make build
```

## External Import Configuration

This sample imports govern as an **external dependency**, not via a Go workspace. The
`replace` directive in [go.mod](go.mod) points at the root govern module for local development:

```
require github.com/haipham22/govern v0.0.0
replace github.com/haipham22/govern => ../../
```

In production, remove the `replace` directive and use the published govern version.

## Architecture

This sample demonstrates:

- Clean architecture layers: Handler → Controller → Service → Storage
- Govern package integration: `http`, `database`, `config`, `errors`, `log`, `graceful`
- Wire compile-time dependency injection (to be replaced with manual DI)
- Mockery for test generation
- GORM with PostgreSQL (SQLite for tests)
- JWT authentication with bcrypt password hashing

### Layers

| Layer        | Location                                   | Responsibility                          |
| ------------ | ------------------------------------------ | --------------------------------------- |
| HTTP Handler | `internal/handler/rest/`                   | Echo binding, routing, response mapping |
| Controller   | `internal/handler/rest/controllers/`       | Request orchestration, error mapping    |
| Service      | `internal/service/auth/`                   | Business logic, JWT, password hashing   |
| Storage      | `internal/storage/user/`                   | Data access interface + GORM conversion |
| Model        | `internal/model/`                          | Pure domain entities                    |
| ORM          | `internal/orm/`                            | GORM entities                           |
| Schemas      | `internal/schemas/`                        | Request/response DTOs                   |
| Config       | `pkg/config/`                              | Environment configuration               |

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Key variables:

| Variable            | Description                  |
| ------------------- | ---------------------------- |
| `APP_ENV`           | `development` / `production` |
| `APP_DEBUG`         | Enable debug logging         |
| `APP_POSTGRES_DSN`  | PostgreSQL connection string |
| `APP_API_SECRET`    | JWT signing secret (32+ chars) |

## Development

See repository root [docs/](../../docs/) for the full development guide.

## CI/CD

Sample app tests run via the root workflow
[`.github/workflows/test-sample.yml`](../../.github/workflows/test-sample.yml), which targets
`examples/golang-sample/**` and runs in this directory. The Docker image build targets this
directory via [`.github/workflows/push.yml`](../../.github/workflows/push.yml).

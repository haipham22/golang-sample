---
title: "Phase 05: Documentation Migration"
description: "Migrate documentation from govern repository and update project documentation"
status: completed
priority: P1
effort: 2h
branch: feat/monorepo-migration
tags: [documentation, migration, guides]
created: 2026-06-27
dependsOn: [phase-04-root-configuration-update.md]
---

# Phase 05: Documentation Migration

> **Status sync (2026-06-28):** Completed. Historical examples below may still show older Echo v4 or pre-final doc-generation snippets.

## Overview

Migrate documentation from govern repository and update all project documentation to reflect monorepo structure.

**Priority**: P1 (blocks validation and testing)  
**Duration**: 2 hours  
**Risk**: Low (documentation updates)

**Working Directory**: All operations in this phase are performed at the repository root (`golang-sample/`). Documentation files are updated here.

---

## Context

**Current State**: Generator implemented, basic docs created  
**Target State**: Complete documentation for govern library and monorepo  

**Documentation to Migrate**:
- ../govern/QUICKSTART.md → docs/quickstart.md (already created in Phase 04)
- ../govern/DEVELOPMENT.md → docs/contributing.md (already created in Phase 04)
- ../govern package docs → docs/packages/
- Current golang-sample docs → Merge and reorganize

**Documentation to Update**:
- Sample app README (already created in Phase 03)
- CLAUDE.md (already updated in Phase 04)
- Package documentation files

---

## Requirements

### Functional Requirements
- Migrate govern repository documentation
- Create package documentation for each govern package
- Update all documentation for monorepo structure
- Create comprehensive guides
- Ensure all links work correctly

### Non-Functional Requirements
- Documentation clear and comprehensive
- No broken links or references
- Consistent formatting and style
- Easy to navigate

---

## Architecture

**Data Flow**:
```
../govern/ Docs + Current Docs → Merged & Updated → Complete Monorepo Documentation
```

**Component Interactions**:
- Package docs explain each govern package
- Guides explain how to use govern library
- Sample app docs demonstrate usage

---

## Related Code Files

### Files to Read
- `../govern/QUICKSTART.md` - Govern quick start (if exists)
- `../govern/DEVELOPMENT.md` - Govern development guide (if exists)
- `../govern/*/README.md` - Package documentation (if exists)

### Files to Create

**Govern Library Documentation**:
- `docs/packages/http.md` - HTTP package documentation
- `docs/packages/database.md` - Database package documentation
- `docs/packages/config.md` - Config package documentation
- `docs/packages/errors.md` - Errors package documentation
- `docs/packages/log.md` - Log package documentation
- `docs/packages/graceful.md` - Graceful package documentation
- `docs/packages/retry.md` - Retry package documentation
- `docs/packages/cron.md` - Cron package documentation
- `docs/packages/mq.md` - Message queue package documentation
- `docs/packages/metrics.md` - Metrics package documentation
- `docs/packages/healthcheck.md` - Health check package documentation

**Sample Application Documentation** (following affiliate-tracking approach):
- `examples/golang-sample/docs/SPEC.md` - Specification document (feature specs, acceptance criteria)
- `examples/golang-sample/docs/HLD.md` - High-level design (architecture, data flow, component interactions)
- `examples/golang-sample/docs/ROADMAP.md` - Project roadmap (phases, milestones, progress tracking)

### Files to Update
- `README.md` - Verify govern library docs complete
- `CONTRIBUTING.md` - Verify contribution guide complete
- `CLAUDE.md` - Verify monorepo docs complete

---

## Implementation Steps

### Step 1: Check Govern Repository for Documentation
**Duration**: 15 minutes  

```bash
# Check if govern repository has documentation
cd ../govern

# List markdown files
find . -name "*.md" -type f

# Check for QUICKSTART.md
if [ -f QUICKSTART.md ]; then
    echo "QUICKSTART.md exists"
    cat QUICKSTART.md | head -20
fi

# Check for DEVELOPMENT.md
if [ -f DEVELOPMENT.md ]; then
    echo "DEVELOPMENT.md exists"
    cat DEVELOPMENT.md | head -20
fi

# Check for package READMEs
find . -name "README.md" -type f

# Return to golang-sample
cd ../golang-sample
```

**Acceptance Criteria**:
- Govern repository checked for documentation
- Documentation files identified
- Missing documentation noted

---

### Step 2: Create Package Documentation Structure
**Duration**: 20 minutes  

```bash
# Ensure docs/packages/ directory exists
mkdir -p docs/packages

# Create package documentation template
cat > docs/packages/README.md << 'EOF'
# Govern Package Documentation

Complete documentation for all govern packages.

## Packages

### HTTP Packages
- [http/](http.md) - HTTP server integration
- [http/echo/](http.md#echo) - Echo framework integration
- [http/jwt/](http.md#jwt) - JWT authentication middleware
- [http/middleware/](http.md#middleware) - Common HTTP middleware

### Database Packages
- [database/](database.md) - Database interfaces
- [database/postgres/](database.md#postgres) - PostgreSQL integration
- [database/redis/](database.md#redis) - Redis integration

### Core Services
- [config/](config.md) - Configuration management
- [errors/](errors.md) - Error handling
- [log/](log.md) - Structured logging
- [graceful/](graceful.md) - Graceful shutdown
- [retry/](retry.md) - Retry logic

### Background Processing
- [cron/](cron.md) - Cron scheduler
- [mq/](mq.md) - Message queue interfaces
- [mq/asynq/](mq.md#asynq) - Asynq task queue
- [metrics/](metrics.md) - Prometheus metrics
- [healthcheck/](healthcheck.md) - Health checks

## Usage Patterns

### HTTP Server Setup
\`\`\`go
import "github.com/haipham22/govern/http"
// See http.md for details
\`\`\`

### Configuration Management
\`\`\`go
import "github.com/haipham22/govern/config"
// See config.md for details
\`\`\`

### Database Operations
\`\`\`go
import "github.com/haipham22/govern/database/postgres"
// See database.md for details
\`\`\`
EOF

# Verify package docs structure
cat docs/packages/README.md
```

**Acceptance Criteria**:
- docs/packages/ directory exists
- Package docs index created
- Structure organized by category

---

### Step 3: Create HTTP Package Documentation
**Duration**: 20 minutes  

```bash
cat > docs/packages/http.md << 'EOF'
# HTTP Package Documentation

Govern HTTP packages provide web framework integration and middleware.

## Packages

### http/
Base HTTP package with common interfaces and utilities.

### http/echo
Echo framework integration with production-ready middleware.

\`\`\`go
import "github.com/haipham22/govern/http/echo"

server := echohttp.New(e, echohttp.Config{
    Port: 8080,
    ReadTimeout: 30 * time.Second,
    WriteTimeout: 30 * time.Second,
})
\`\`\`

**Features**:
- Request timeout configuration
- Graceful shutdown
- Health check endpoints
- Middleware chain management

### http/jwt
JWT authentication middleware for Echo.

\`\`\`go
import "github.com/haipham22/govern/http/jwt"

middleware := jwtmiddleware.New(jwtmiddleware.Config{
    SigningKey: []byte("secret"),
})
\`\`\`

**Features**:
- JWT token validation
- Custom claims support
- Token extraction from headers/cookies
- Error handling

### http/middleware
Common HTTP middleware for Echo.

\`\`\`go
import "github.com/haipham22/govern/http/middleware"

// CORS middleware
middleware.CORS()

// Security headers
middleware.Security()

// Compression
middleware.Gzip()
\`\`\`

**Available Middleware**:
- CORS - Cross-origin resource sharing
- Security - Security headers (HSTS, X-Frame-Options, etc.)
- Gzip - Response compression
- RateLimit - Request rate limiting
- RequestID - Unique request ID generation
- Logger - Request logging

## Usage Example

\`\`\`go
package main

import (
    "github.com/haipham22/govern/http/echo"
    "github.com/haipham22/govern/http/jwt"
    "github.com/haipham22/govern/http/middleware"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    
    // Add middleware
    e.Use(middleware.CORS())
    e.Use(middleware.Security())
    e.Use(middleware.Gzip())
    
    // Add JWT auth
    jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Config{
        SigningKey: []byte("secret"),
    })
    e.POST("/login", loginHandler)
    
    // Protected routes
    api := e.Group("/api")
    api.Use(jwtMiddleware.Middleware())
    api.GET("/users", getUsersHandler)
    
    // Start server
    server := echohttp.New(e, echohttp.Config{
        Port: 8080,
    })
    graceful.Run(server)
}
\`\`\`

## Configuration

### Echo Server Configuration
\`\`\`go
type Config struct {
    Port            int
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    ShutdownTimeout time.Duration
}
\`\`\`

### JWT Configuration
\`\`\`go
type Config struct {
    SigningKey      []byte
    TokenLookup     string
    AuthScheme      string
    Claims          jwt.Claims
}
\`\`\`

## Testing

\`\`\`go
func TestHTTPHandler(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest("GET", "/health", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    
    handler := HealthHandler()
    err := handler(c)
    
    assert.NoError(t, err)
    assert.Equal(t, 200, rec.Code)
}
\`\`\`

## See Also

- [Sample Application](../../golang-sample/) - Complete usage example
- [Quick Start](../../docs/quickstart.md) - Get started guide
EOF

# Verify HTTP package docs created
cat docs/packages/http.md
```

**Acceptance Criteria**:
- HTTP package documentation created
- Usage examples included
- Configuration documented
- Testing section included

---

### Step 4: Create Database Package Documentation
**Duration**: 15 minutes  

```bash
cat > docs/packages/database.md << 'EOF'
# Database Package Documentation

Govern database packages provide database integration and ORM support.

## Packages

### database/
Base database package with common interfaces.

### database/postgres
PostgreSQL integration with pgx driver.

\`\`\`go
import "github.com/haipham22/govern/database/postgres"

db, err := postgres.New(postgres.Config{
    DSN: "host=localhost user=postgres password=password dbname=mydb port=5432",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
})
\`\`\`

**Features**:
- Connection pooling
- Query timeouts
- Automatic reconnection
- Health checks

### database/redis
Redis client integration.

\`\`\`go
import "github.com/haipham22/govern/database/redis"

client, err := redis.New(redis.Config{
    Addr: "localhost:6379",
    Password: "",
    DB: 0,
    PoolSize: 10,
})
\`\`\`

**Features**:
- Connection pooling
- Pub/Sub support
- Transaction support
- Health checks

## Usage Example

\`\`\`go
package main

import (
    "github.com/haipham22/govern/database/postgres"
    "context"
    "time"
)

func main() {
    // Initialize PostgreSQL
    db, err := postgres.New(postgres.Config{
        DSN: "host=localhost user=postgres password=password dbname=mydb",
        MaxOpenConns: 25,
        ConnMaxLifetime: time.Hour,
    })
    if err != nil {
        panic(err)
    }
    
    // Use database
    var result int
    err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&result)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Total users: %d\n", result)
}
\`\`\`

## Configuration

### PostgreSQL Configuration
\`\`\`go
type Config struct {
    DSN            string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}
\`\`\`

### Redis Configuration
\`\`\`go
type Config struct {
    Addr         string
    Password     string
    DB           int
    PoolSize     int
    MinIdleConns int
    DialTimeout  time.Duration
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}
\`\`\`

## See Also

- [Sample Application](../../golang-sample/) - Database usage example
- [GORM Documentation](https://gorm.io/docs/) - ORM documentation
EOF

# Verify database package docs created
cat docs/packages/database.md
```

**Acceptance Criteria**:
- Database package documentation created
- PostgreSQL and Redis documented
- Usage examples included

---

### Step 5: Create Core Package Documentation
**Duration**: 30 minutes  

```bash
# Create config package docs
cat > docs/packages/config.md << 'EOF'
# Config Package Documentation

Configuration management with support for YAML, .env, and environment variables.

\`\`\`go
import "github.com/haipham22/govern/config"

type Config struct {
    Port     int    \`yaml:"port" env:"PORT"\`
    Database string \`yaml:"database" env:"DATABASE_URL"\`
}

cfg := &Config{Port: 8080}
config.Load("config.yaml", cfg)
config.LoadFromEnv(cfg)
\`\`\`

**Features**:
- YAML file loading
- Environment variable overrides
- .env file support
- Default values
- Validation

## See Also

- [Sample Application](../../golang-sample/) - Config usage example
EOF

# Create errors package docs
cat > docs/packages/errors.md << 'EOF'
# Errors Package Documentation

Standardized error handling and HTTP error mapping.

\`\`\`go
import "github.com/haipham22/govern/errors"

err := errors.New("user not found", errors.NotFound)
httpErr := errors.ToHTTP(err) // 404 Not Found
\`\`\`

**Features**:
- Error wrapping with context
- HTTP error code mapping
- Error type classification
- Structured error responses

## See Also

- [Sample Application](../../golang-sample/) - Error handling example
EOF

# Create log package docs
cat > docs/packages/log.md << 'EOF'
# Log Package Documentation

Structured logging with Zap integration.

\`\`\`go
import "github.com/haipham22/govern/log"

logger := log.New()
logger.Info("Server started", "port", 8080)
logger.Error("Request failed", "error", err, "path", "/api/users")
\`\`\`

**Features**:
- Structured logging
- Multiple log levels
- Context-aware logging
- Log rotation

## See Also

- [Sample Application](../../golang-sample/) - Logging example
EOF

# Create graceful package docs
cat > docs/packages/graceful.md << 'EOF'
# Graceful Package Documentation

Graceful shutdown handling for HTTP servers.

\`\`\`go
import "github.com/haipham22/govern/graceful"

graceful.Run(server, graceful.DefaultConfig())
\`\`\`

**Features**:
- SIGINT/SIGTERM handling
- Connection draining
- Timeout enforcement
- Shutdown hooks

## See Also

- [Sample Application](../../golang-sample/) - Graceful shutdown example
EOF

# Create retry package docs
cat > docs/packages/retry.md << 'EOF'
# Retry Package Documentation

Exponential backoff retry logic.

\`\`\`go
import "github.com/haipham22/govern/retry"

err := retry.Do(func() error {
    return callAPI()
}, retry.DefaultConfig())
\`\`\`

**Features**:
- Exponential backoff
- Max retry attempts
- Jitter support
- Custom retry policies

## See Also

- [Sample Application](../../golang-sample/) - Retry usage example
EOF

# Verify core package docs created
ls -la docs/packages/
```

**Acceptance Criteria**:
- Core package docs created
- Config, errors, log, graceful, retry documented
- Usage examples included

---

### Step 6: Create Background Processing Package Documentation
**Duration**: 20 minutes  

```bash
# Create cron package docs
cat > docs/packages/cron.md << 'EOF'
# Cron Package Documentation

Cron scheduler for recurring tasks.

\`\`\`go
import "github.com/haipham22/govern/cron"

scheduler := cron.New()
scheduler.AddFunc("@every 1h", func() {
    cleanup()
})
scheduler.Start()
\`\`\`

**Features**:
- Cron expression support
- Concurrent task execution
- Task registration
- Graceful shutdown

## See Also

- [Sample Application](../../golang-sample/) - Cron usage example
EOF

# Create mq package docs
cat > docs/packages/mq.md << 'EOF'
# Message Queue Package Documentation

Message queue interfaces and implementations.

## Packages

### mq/
Base message queue interface.

### mq/asynq
Asynq task queue implementation.

\`\`\`go
import "github.com/haipham22/govern/mq/asynq"

server := asynq.NewServer(asynq.Config{
    RedisAddr: "localhost:6379",
})
server.Run()
\`\`\`

**Features**:
- Task queue with Redis
- Retry mechanisms
- Dead letter queue
- Task monitoring

## See Also

- [Sample Application](../../golang-sample/) - Task queue example
EOF

# Create metrics package docs
cat > docs/packages/metrics.md << 'EOF'
# Metrics Package Documentation

Prometheus metrics integration.

\`\`\`go
import "github.com/haipham22/govern/metrics"

metrics.RecordHTTPRequest("api_users", "GET", time.Since(start))
\`\`\`

**Features**:
- HTTP request metrics
- Custom metric registration
- Prometheus exposition
- Label-based metrics

## See Also

- [Sample Application](../../golang-sample/) - Metrics usage example
EOF

# Create healthcheck package docs
cat > docs/packages/healthcheck.md << 'EOF'
# Healthcheck Package Documentation

Health check endpoints and monitors.

\`\`\`go
import "github.com/haipham22/govern/healthcheck"

healthcheck.Register("db", func() error {
    return db.Ping()
})
\`\`\`

**Features**:
- Health check registration
- HTTP health endpoints
- Liveness/Readiness probes
- Health status aggregation

## See Also

- [Sample Application](../../golang-sample/) - Health check example
EOF

# Verify background processing docs created
ls -la docs/packages/
```

**Acceptance Criteria**:
- Background processing docs created
- Cron, mq, metrics, healthcheck documented
- Usage examples included

---

### Step 7: Update Root Documentation
**Duration**: 15 minutes  

```bash
# Verify root README is complete
cat README.md

# Verify CONTRIBUTING.md is complete
cat CONTRIBUTING.md

# Verify CLAUDE.md is complete
cat CLAUDE.md

# Check for any missing sections
grep -r "TODO\|FIXME\|XXX" README.md CONTRIBUTING.md CLAUDE.md docs/
```

**Acceptance Criteria**:
- Root documentation complete
- No TODO/FIXME placeholders
- All sections filled

---

### Step 8: Verify All Links
**Duration**: 15 minutes  

```bash
# Check for broken links in documentation
# (Manual verification required)

# Verify relative paths work
ls -la docs/
ls -la docs/packages/
ls -la docs/samples/

# Verify README links
grep -o '\[.*\](.*)' README.md | head -20

# Verify package docs links
grep -o '\[.*\](.*)' docs/packages/*.md | head -20
```

**Acceptance Criteria**:
- All relative paths correct
- No broken links identified
- Documentation structure valid

---

### Step 9: Commit Documentation Migration
**Duration**: 10 minutes  

```bash
# Stage all documentation
git add docs/ README.md CONTRIBUTING.md CLAUDE.md

# Review changes
git diff --cached --stat

# Commit documentation
git commit -m "docs: migrate and complete documentation for govern monorepo

Migrate documentation from govern repository and create comprehensive
documentation for all govern packages.

Changes:
- Create docs/packages/ structure
- Document all govern packages (http, database, config, errors, log, graceful, retry, cron, mq, metrics, healthcheck)
- Create package documentation index
- Update root README, CONTRIBUTING.md, CLAUDE.md
- Add usage examples for all packages
- Verify all documentation links

Package Documentation:
- HTTP package: Echo integration, JWT middleware, common middleware
- Database package: PostgreSQL, Redis integration
- Core packages: Config, errors, log, graceful, retry
- Background processing: Cron, message queue, metrics, health checks

Documentation Features:
- Clear usage examples for each package
- Configuration options documented
- Integration patterns explained
- Links to sample application examples

Next: Phase 07 - Validation and testing
"

# Verify commit
git log -1 --stat
```

**Acceptance Criteria**:
- Documentation migration committed
- All package docs included
- Clean commit history

---

### Step 10: Create Sample Application Documentation
**Duration**: 30 minutes

Following the affiliate-tracking documentation approach, create comprehensive documentation for the sample application in `examples/golang-sample/docs/`.

```bash
# Create documentation directory for sample app
mkdir -p examples/golang-sample/docs

# Create SPEC.md (Specification document)
cat > examples/golang-sample/docs/SPEC.md << 'EOF'
# golang-sample Specification

## Overview
Sample application demonstrating clean architecture with govern packages.

## Features
- HTTP REST API (Echo framework)
- gRPC support (planned)
- Job workers (planned)
- Kafka event handlers (planned)
- PostgreSQL integration
- Custom error handling
- Manual dependency injection

## Architecture
See [HLD.md](./HLD.md) for detailed architecture.

## Use Cases
1. User authentication (JWT)
2. Health checks
3. Example CRUD operations

## Roadmap
See [ROADMAP.md](./ROADMAP.md) for implementation phases.
EOF

# Create HLD.md (High-level design)
cat > examples/golang-sample/docs/HLD.md << 'EOF'
# golang-sample High-Level Design

## Architecture Overview
Clean architecture with transport-specific handlers.

## Component Diagram
```
HTTP/gRPC/Job/Kafka Requests
        ↓
    handler/{rest,grpc,job,kafka}/
        ↓
    controller/
        ↓
    service/ (use cases)
        ↓
    storage/ → repository/postgres/
        ↓
    model/ (domain entities)
```

## Data Flow
See main plan.md for detailed dependency flow.

## Error Handling
Custom error types in internal/errors/ replacing govern/errors.
EOF

# Create ROADMAP.md (Project roadmap)
cat > examples/golang-sample/docs/ROADMAP.md << 'EOF'
# golang-sample Roadmap

## Current Status
✅ HTTP REST API with Echo
✅ PostgreSQL integration
✅ Custom error handling
✅ Manual dependency injection

## Planned Features
- [x] gRPC server implementation
- [x] Job worker implementation
- [x] Kafka event handlers
- [x] Additional use cases (CRUD operations)

## Phases
See main plan for implementation phases.
EOF
```

**Acceptance Criteria**:
- `examples/golang-sample/docs/SPEC.md` created
- `examples/golang-sample/docs/HLD.md` created
- `examples/golang-sample/docs/ROADMAP.md` created
- Documentation follows affiliate-tracking approach
- Clear links between documents

---

## Success Criteria

### Phase Completion Criteria
- [x] Govern repository documentation checked
- [x] Package documentation structure created
- [x] All govern packages documented
- [x] Root documentation updated
- [x] All links verified
- [x] Documentation committed

### Quality Criteria
- [x] Documentation comprehensive
- [x] No broken links
- [x] Consistent formatting
- [x] Clear examples

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Missing documentation sections | Low | Low | Follow checklist, verify all sections |
| Broken links | Low | Medium | Verify all paths work |
| Inconsistent formatting | Low | Low | Use consistent markdown style |

---

## Security Considerations

**No Security Impact**: Documentation updates only

---

## Rollback Strategy

**If documentation issues**:
```bash
# Reset to before commit
git reset --hard HEAD~1

# Fix documentation issues
# Re-commit after fixes
```

---

## Todo List

- [x] Check govern repository for documentation
- [x] Create package documentation structure
- [x] Create HTTP package documentation
- [x] Create database package documentation
- [x] Create core package documentation
- [x] Create background processing documentation
- [x] Update root documentation
- [x] Verify all links
- [x] Commit documentation migration

---

## Phase Summary

**Input**: Generator implemented, basic docs created  
**Output**: Complete documentation for govern library and monorepo  
**Duration**: 2 hours  
**Risk Level**: Low (documentation updates)  
**Blocks**: Phase 07 (Validation and Testing)

**Status**: Ready to start  
**Next Action**: Execute Step 1 (Check Govern Repository for Documentation)

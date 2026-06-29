# Docker Rules

**Best practices for Docker in Go development workflows.**

---

## Docker Compose for Development

**docker-compose.yml structure:**
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:16-alpine
    container_name: golang_sample_db
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: golang_sample
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

**Docker Compose commands:**
```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f postgres

# Stop services
docker-compose down

# Stop with volumes
docker-compose down -v

# Rebuild and restart
docker-compose up -d --build
```

---

## Database Container Rules

**Use Docker for local database:**
- ✅ PostgreSQL in container for development
- ✅ Match production version in Docker image
- ✅ Use healthchecks for startup dependencies
- ✅ Persist data with named volumes
- ❌ NEVER commit database data to git
- ❌ NEVER use latest tag (pin version)

**Version pinning:**
```yaml
# GOOD - Pinned version
postgres:
  image: postgres:16-alpine

# BAD - Latest tag
postgres:
  image: postgres:latest  # Unpredictable updates
```

---

## Dockerfile for Go Applications

**Multi-stage build for production:**
```dockerfile
# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

**Multi-stage benefits:**
- ✅ Small final image (no build tools)
- ✅ No source code in production image
- ✅ Faster deployment (smaller image)
- ✅ Separation of build and runtime concerns

---

## Docker Build Commands

**Build and run locally:**
```bash
# Build image
docker build -t golang-sample:latest .

# Run container
docker run -p 8080:8080 golang-sample:latest

# Run with environment variables
docker run -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  golang-sample:latest

# Run in background
docker run -d -p 8080:8080 --name api golang-sample:latest
```

---

## Dockerfile Best Practices

**Use .dockerignore:**
```
# .dockerignore
.git
.github
.vscode
.idea
*_test.go
docs/
.env
coverage.out
*.md
```

**Layer caching:**
```dockerfile
# GOOD - Cache dependencies separately
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# BAD - No dependency caching
COPY . .
RUN go mod download
```

**Minimal images:**
```dockerfile
# GOOD - Minimal base
FROM alpine:latest

# OK - Distroless
FROM gcr.io/distroless/static

# BAD - Large base
FROM ubuntu:latest
```

---

## Development Workflow

**Start development environment:**
```bash
# 1. Start database
docker-compose up -d postgres

# 2. Wait for healthy status
docker-compose ps

# 3. Run migrations (if any)
mise exec -- go run cmd/migrate/main.go

# 4. Run application
mise exec -- go run cmd/api/main.go
```

**Database operations:**
```bash
# Connect to PostgreSQL
docker exec -it golang_sample_db psql -U postgres -d golang_sample

# Run SQL script
docker exec -i golang_sample_db psql -U postgres -d golang_sample < schema.sql

# Backup database
docker exec golang_sample_db pg_dump -U postgres golang_sample > backup.sql

# Restore database
docker exec -i golang_sample_db psql -U postgres golang_sample < backup.sql
```

---

## Healthchecks

**Application healthcheck:**
```yaml
# docker-compose.yml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

**Healthcheck endpoint:**
```go
// handler/health.go
func (h *Handler) HealthCheck(c echo.Context) error {
    return c.JSON(http.StatusOK, map[string]string{
        "status": "healthy",
        "timestamp": time.Now().Format(time.RFC3339),
    })
}
```

---

## Docker Compose Patterns

**Service dependencies:**
```yaml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
```

**Multiple environments:**
```bash
# Development
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Production
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up
```

**Override files:**
```yaml
# docker-compose.dev.yml
services:
  api:
    volumes:
      - .:/app  # Hot reload

# docker-compose.prod.yml
services:
  api:
    restart: always
    environment:
      - GIN_MODE=release
```

---

## Security Best Practices

**✅ DO:**
- Use specific version tags (not `latest`)
- Run as non-root user
- Scan images for vulnerabilities
- Use secrets management for sensitive data
- Enable Content Trust for images

**❌ DON'T:**
- Run as root user
- Embed secrets in Dockerfile
- Use `latest` tag in production
- Expose unnecessary ports
- Leave build tools in production image

**Non-root user:**
```dockerfile
RUN addgroup -g 1000 appuser && \
    adduser -u 1000 -G appuser appuser
USER appuser
```

---

## Image Optimization

**Small image tips:**
```dockerfile
# Use alpine base
FROM alpine:latest

# Multi-stage builds
FROM golang:1.26-alpine AS builder
# ... build steps
FROM alpine:latest
COPY --from=builder /app/main .

# Combine RUN commands
RUN apk add --no-cache git ca-certificates && \
    update-ca-certificates

# Clean up in same layer
RUN go build -o main . && \
    rm -rf /tmp/*
```

---

## CI/CD Integration

**GitHub Actions with Docker:**
```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Build Docker image
        run: docker build -t myapp:${{ github.sha }} .
      
      - name: Run tests in container
        run: docker run myapp:${{ github.sha }} go test ./...
      
      - name: Push to registry
        run: docker push myapp:${{ github.sha }}
```

---

## Common Pitfalls

**❌ Avoid these:**

```yaml
# BAD - Latest tag
image: postgres:latest

# BAD - No healthcheck
services:
  postgres:
    image: postgres:16
    # Missing healthcheck - app may start before DB ready

# BAD - No volume persistence
services:
  postgres:
    image: postgres:16
    # Data lost on container restart

# BAD - Running as root
USER root

# BAD - Large layers
COPY . .
RUN go mod download  # Downloads every time code changes
```

---

## Quick Reference

**Docker Compose:**
```bash
docker-compose up -d           # Start services
docker-compose down             # Stop services
docker-compose logs -f          # Follow logs
docker-compose ps               # Show status
docker-compose exec db psql     # Connect to service
```

**Docker Build:**
```bash
docker build -t app:latest .
docker run -p 8080:8080 app:latest
docker images
docker rmi $(docker images -q)  # Remove all images
```

**Container Management:**
```bash
docker ps                       # Running containers
docker logs <container>         # View logs
docker exec -it <container> sh  # SSH into container
docker stop <container>         # Stop container
docker rm <container>           # Remove container
```

---

## Troubleshooting

**Container won't start:**
```bash
# Check logs
docker-compose logs postgres

# Check container status
docker-compose ps

# Rebuild from scratch
docker-compose down -v
docker-compose up -d --build
```

**Database connection issues:**
```bash
# Verify database is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Test connection
docker exec -it golang_sample_db psql -U postgres -c "SELECT 1"
```

**Port conflicts:**
```yaml
# Use different ports locally
services:
  postgres:
    ports:
      - "15432:5432"  # Map localhost:15432 to container:5432
```

---

## Best Practices Summary

| Practice | Do | Don't |
|----------|-----|-------|
| **Base images** | Pin specific version | Use `latest` |
| **Layer caching** | Copy go.mod first | Copy all at once |
| **Image size** | Multi-stage builds | Include build tools |
| **Security** | Run as non-root | Run as root |
| **Data persistence** | Use named volumes | Lose data on restart |
| **Healthchecks** | Define for services | Skip healthchecks |

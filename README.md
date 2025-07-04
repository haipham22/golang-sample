# Golang Sample API

A clean, scalable Go API built with clean architecture principles, using Go 1.22+ and the standard library's net/http package.

## 🏗️ Clean Architecture Design System

This project follows Clean Architecture principles with clear separation of concerns across multiple layers:

### Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   HTTP Routes   │  │   Middlewares   │  │   Handlers   │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Controllers   │  │   Validators    │  │   Schemas    │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │     Models      │  │   Interfaces    │  │   Services   │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Storage       │  │   Database      │  │   External   │ │
│  │   (Repository)  │  │   (PostgreSQL)  │  │   Services   │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

#### 1. **Presentation Layer** (`internal/api/`)
- **HTTP Routes**: Route definitions and endpoint mapping
- **Handlers**: HTTP request/response handling
- **Middlewares**: Cross-cutting concerns (logging, auth, CORS)
- **Schemas**: Request/response data structures

#### 2. **Application Layer** (`internal/api/routes/`)
- **Controllers**: Business logic orchestration
- **Validators**: Input validation and sanitization
- **Error Handling**: Custom error types and responses

#### 3. **Domain Layer** (`pkg/models/`)
- **Models**: Core business entities
- **Interfaces**: Abstract contracts for dependencies
- **Business Rules**: Domain-specific logic

#### 4. **Infrastructure Layer** (`internal/api/storage/`, `pkg/postgres/`)
- **Storage**: Data access implementations
- **Database**: PostgreSQL connection and configuration
- **External Services**: Third-party integrations

### Dependency Injection

The project uses Google Wire for dependency injection, ensuring:
- **Inversion of Control**: Dependencies flow inward
- **Testability**: Easy mocking and unit testing
- **Loose Coupling**: Components depend on abstractions

```go
// Wire configuration (internal/api/wire.go)
func InitApp(
    isDebugMode bool,
    db string,
    log *zap.SugaredLogger,
) (*Handler, func(), error) {
    panic(wire.Build(
        NewHandler,
        echo.New,
        postgres.NewGormDB,
        wire.NewSet(storage.NewStorage),
        wire.NewSet(auth.NewAuthController),
    ))
}
```

## 📁 Project Structure

```
golang-sample/
├── cmd/                    # Application entry points
│   ├── api.go             # API server command
│   └── root.go            # Root command configuration
├── internal/              # Private application code
│   └── api/               # API layer implementation
│       ├── errors/        # Custom error types
│       ├── handler.go     # HTTP server setup
│       ├── middlewares/   # HTTP middlewares
│       ├── routes/        # Route controllers
│       │   └── auth/      # Authentication routes
│       ├── schemas/       # Request/response schemas
│       ├── storage/       # Data access layer
│       ├── swagger/       # API documentation
│       ├── validator/     # Input validation
│       ├── wire.go        # Dependency injection
│       └── wire_gen.go    # Generated wire code
├── pkg/                   # Public packages
│   ├── config/           # Configuration management
│   ├── models/           # Domain models
│   ├── postgres/         # Database connection
│   └── utils/            # Utility functions
│       ├── password/     # Password utilities
│       └── string/       # String utilities
├── scripts/              # Build and deployment scripts
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Dockerfile           # Container configuration
├── compose.yml          # Docker Compose setup
└── README.md            # Project documentation
```

## 🚀 Getting Started

### Prerequisites

- Go 1.22 or newer
- PostgreSQL database
- Docker (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd golang-sample
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   export APP_ENV=development
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_NAME=golang_sample
   export DB_USER=postgres
   export DB_PASSWORD=password
   ```

4. **Run the application**
   ```bash
   go run main.go api
   ```

### Development

#### Run in development mode
```bash
go run main.go api
```

#### Run tests
```bash
go test ./...
go test -v ./...
```

#### Build the application
```bash
go build -o bin/api cmd/api.go
```

### API Documentation

#### Generate Swagger documentation
```bash
# Install Swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
./scripts/generate-swagger.sh
```

#### Access Swagger UI
- Development: `http://localhost:8080/document/index.html`
- Production: Disabled for security

## 🧪 Testing Strategy

### Test Structure
```
├── internal/
│   └── api/
│       ├── routes/
│       │   └── auth/
│       │       ├── auth_test.go
│       │       └── login_test.go
│       └── storage/
│           └── storage_test.go
└── pkg/
    └── utils/
        └── password/
            └── hash_test.go
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/api/routes/auth

# Run with verbose output
go test -v ./...
```

## 🔧 Configuration

### Environment Variables
```bash
# Application
APP_ENV=development          # Environment (development, staging, production)
APP_DEBUG=true              # Debug mode
APP_PORT=8080               # Server port

# Database
DB_HOST=localhost           # Database host
DB_PORT=5432               # Database port
DB_NAME=golang_sample      # Database name
DB_USER=postgres           # Database user
DB_PASSWORD=password       # Database password
DB_SSL_MODE=disable        # SSL mode

# JWT
JWT_SECRET=your-secret-key # JWT signing secret
JWT_EXPIRY=24h             # JWT expiry time
```

## 🐳 Docker Support

### Build and run with Docker
```bash
# Build the image
docker build -t golang-sample .

# Run the container
docker run -p 8080:8080 golang-sample

# Run with Docker Compose
docker-compose up -d
```

### Docker Compose
```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=development
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=golang_sample
      - DB_USER=postgres
      - DB_PASSWORD=password
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=golang_sample
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## 📊 API Endpoints

### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `GET /auth/profile` - Get user profile (protected)

### Health Check
- `GET /health` - Application health status

## 🔒 Security Features

- **Password Hashing**: Bcrypt for secure password storage
- **JWT Authentication**: Stateless authentication tokens
- **Input Validation**: Comprehensive request validation
- **CORS Protection**: Cross-origin resource sharing configuration
- **Rate Limiting**: Request rate limiting (configurable)
- **Request Logging**: Structured logging for security auditing

## 📈 Performance & Scalability

- **Connection Pooling**: Efficient database connection management
- **Graceful Shutdown**: Proper server shutdown handling
- **Middleware Chain**: Optimized request processing pipeline
- **Dependency Injection**: Efficient resource management
- **Structured Logging**: Performance monitoring and debugging

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style Guidelines

- Follow Go formatting standards (`gofmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Write unit tests for new features
- Follow clean architecture principles

## 📝 TODO

- [ ] Add comprehensive test coverage
- [ ] Implement CI/CD pipeline with GitHub Actions
- [ ] Add API rate limiting middleware
- [ ] Implement caching layer (Redis)
- [ ] Add metrics and monitoring (Prometheus)
- [ ] Implement user roles and permissions
- [ ] Add API versioning strategy
- [ ] Create deployment guides for different environments

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

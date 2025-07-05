# API Architecture Documentation

This document describes the clean architecture implementation for the Golang Sample API.

## 🏗️ Clean Architecture Layers

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

## 📋 Layer Responsibilities

### 1. **Presentation Layer** (`internal/api/`)

The presentation layer handles all HTTP-related concerns and user interface logic.

#### Components:
- **HTTP Routes** (`routes/`): Route definitions and endpoint mapping
- **Handlers** (`handler.go`): HTTP server setup and request/response handling
- **Middlewares** (`middlewares/`): Cross-cutting concerns (logging, auth, CORS)
- **Schemas** (`schemas/`): Request/response data structures and validation

#### Key Files:
- `handler.go` - Main HTTP server configuration
- `routes.go` - Route registration
- `middlewares/logger.go` - Request logging middleware
- `schemas/` - Request/response DTOs

#### Responsibilities:
- ✅ HTTP request/response handling
- ✅ Route management and endpoint mapping
- ✅ Middleware chain execution
- ✅ Input/output data transformation
- ✅ Error response formatting
- ✅ CORS and security headers

### 2. **Application Layer** (`internal/api/routes/`)

The application layer orchestrates business logic and coordinates between the presentation and domain layers.

#### Components:
- **Controllers** (`routes/auth/`): Business logic orchestration
- **Validators** (`validator/`): Input validation and sanitization
- **Error Handling** (`errors/`): Custom error types and responses

#### Key Files:
- `routes/auth/auth.go` - Authentication controller
- `routes/auth/login.go` - Login endpoint logic
- `routes/auth/register.go` - Registration endpoint logic
- `validator/validator.go` - Custom validation logic
- `errors/errors.go` - Custom error definitions

#### Responsibilities:
- ✅ Business logic orchestration
- ✅ Input validation and sanitization
- ✅ Error handling and custom error types
- ✅ Request/response transformation
- ✅ Authentication and authorization logic
- ✅ Transaction management

### 3. **Domain Layer** (`pkg/models/`)

The domain layer contains the core business logic and entities.

#### Components:
- **Models** (`pkg/models/`): Core business entities
- **Interfaces** (`internal/api/storage/`): Abstract contracts for dependencies
- **Business Rules**: Domain-specific logic and validation

#### Key Files:
- `pkg/models/user.go` - User domain model
- `internal/api/storage/storage.go` - Storage interface definition

#### Responsibilities:
- ✅ Core business entities and models
- ✅ Domain-specific business rules
- ✅ Interface definitions for dependencies
- ✅ Value objects and domain services
- ✅ Business validation logic

### 4. **Infrastructure Layer** (`internal/api/storage/`, `pkg/postgres/`)

The infrastructure layer handles external concerns like databases, file systems, and third-party services.

#### Components:
- **Storage** (`internal/api/storage/`): Data access implementations
- **Database** (`pkg/postgres/`): PostgreSQL connection and configuration
- **External Services**: Third-party integrations

#### Key Files:
- `internal/api/storage/storage.go` - Storage interface and implementation
- `internal/api/storage/user.go` - User-specific storage operations
- `pkg/postgres/postgres.go` - Database connection management

#### Responsibilities:
- ✅ Data persistence and retrieval
- ✅ Database connection management
- ✅ External service integrations
- ✅ File system operations
- ✅ Caching implementations
- ✅ Message queue operations

## 🔄 Dependency Injection

The project uses Google Wire for dependency injection, ensuring proper dependency flow and testability.

### Wire Configuration (`wire.go`)

```go
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

### Benefits:
- **Inversion of Control**: Dependencies flow inward toward the domain layer
- **Testability**: Easy mocking and unit testing through interface abstractions
- **Loose Coupling**: Components depend on abstractions, not concrete implementations
- **Maintainability**: Clear dependency graph and easy refactoring

## 🏛️ Architecture Principles

### 1. **Dependency Rule**
Dependencies point inward. The domain layer has no dependencies on outer layers.

### 2. **Interface Segregation**
Each layer defines interfaces for the services it needs from outer layers.

### 3. **Single Responsibility**
Each component has a single, well-defined responsibility.

### 4. **Open/Closed Principle**
Open for extension, closed for modification through interface abstractions.

### 5. **Dependency Inversion**
High-level modules don't depend on low-level modules. Both depend on abstractions.

## 📁 Directory Structure

```
internal/api/
├── errors/              # Custom error types
├── handler.go           # HTTP server setup
├── middlewares/         # HTTP middlewares
│   └── logger.go        # Request logging
├── routes/              # Route controllers
│   └── auth/            # Authentication routes
│       ├── auth.go      # Auth controller
│       ├── login.go     # Login endpoint
│       └── register.go  # Register endpoint
├── schemas/             # Request/response schemas
├── storage/             # Data access layer
│   ├── storage.go       # Storage interface
│   └── user.go          # User storage operations
├── swagger/             # API documentation
├── validator/           # Input validation
├── wire.go              # Dependency injection
└── wire_gen.go          # Generated wire code
```

## 🔧 Implementation Guidelines

### Adding New Features

1. **Domain Layer**: Define models and interfaces
2. **Infrastructure Layer**: Implement storage and external services
3. **Application Layer**: Create controllers and business logic
4. **Presentation Layer**: Add routes and handlers

### Testing Strategy

- **Unit Tests**: Test each layer in isolation
- **Integration Tests**: Test layer interactions
- **End-to-End Tests**: Test complete user workflows

### Error Handling

- Use custom error types for domain-specific errors
- Implement proper error wrapping and context
- Provide meaningful error messages to users

### Validation

- Validate input at the presentation layer
- Use domain validation for business rules
- Implement comprehensive error responses

## 🚀 Best Practices

1. **Keep Dependencies Inward**: Outer layers depend on inner layers
2. **Use Interfaces**: Define contracts for external dependencies
3. **Implement Proper Error Handling**: Use custom error types
4. **Write Tests**: Maintain high test coverage
5. **Follow Go Conventions**: Use idiomatic Go patterns
6. **Document APIs**: Use Swagger for API documentation
7. **Log Appropriately**: Use structured logging
8. **Handle Graceful Shutdown**: Implement proper cleanup

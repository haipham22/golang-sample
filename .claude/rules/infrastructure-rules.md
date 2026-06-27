# Infrastructure Rules

**Rules for Zap logging, environment configuration, and infrastructure setup.**

---

## Zap (Logging)

**Logger setup with error envelope:**
```go
// GOOD - Proper zap logger setup with environment-based config
func NewLogger(cfg LogConfig) (*zap.Logger, error) {
    var config zap.Config
    
    switch cfg.Environment {
    case "production":
        config = zap.NewProductionConfig()
        config.EncoderConfig.TimeKey = "timestamp"
        config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
        config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
    case "development":
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    default:
        return nil, &errors.AppError{
            Code: errors.ErrCodeInvalid,
            Err: &ConfigError{
                Key:   "log.environment",
                Value: cfg.Environment,
            },
            Message: "invalid log environment",
        }
    }
    
    config.OutputPaths = []string{"stdout"}
    config.ErrorOutputPaths = []string{"stderr"}
    
    logger, err := config.Build()
    if err != nil {
        return nil, &errors.AppError{
            Code: errors.ErrCodeInternal,
            Err: &LoggerError{
                Op:  "create_logger",
                Err: err,
            },
            Message: "failed to create logger",
        }
    }
    
    return logger, nil
}

// Error envelope for logger operations
type LoggerError struct {
    Op  string // Operation that failed
    Err error  // Underlying error
}

func (e *LoggerError) Error() string {
    return fmt.Sprintf("logger %s: %v", e.Op, e.Err)
}

func (e *LoggerError) Unwrap() error {
    return e.Err
}
```

**Contextual logging with request tracking:**
```go
// GOOD - Structured logging with context envelope
func (s *Service) ProcessOrder(ctx context.Context, orderID int64) error {
    requestID := GetRequestID(ctx)
    userID := GetUserID(ctx)
    
    logger := s.logger.With(
        zap.String("request_id", requestID),
        zap.Int64("user_id", userID),
        zap.Int64("order_id", orderID),
    )
    
    logger.Info("Processing order started")
    
    order, err := s.repo.FindByID(ctx, orderID)
    if err != nil {
        logger.Error("Failed to find order", 
            zap.Error(err),
            zap.String("error_type", string(errors.GetCode(err))),
        )
        return err // Return envelope error from repository
    }
    
    logger.Info("Processing order completed", 
        zap.Int("items_count", len(order.Items)),
        zap.Float64("total_amount", order.TotalAmount),
    )
    
    return nil
}
```

**Zap rules:**
- ✅ Use structured logging (key-value pairs)
- ✅ Add request context (request_id, user_id)
- ✅ Use appropriate log levels (Debug, Info, Warn, Error)
- ✅ Log errors with context and error code
- ✅ Environment-based configuration
- ✅ Error envelope for logger operations
- ❌ NEVER log sensitive data (passwords, tokens, PII)
- ❌ NEVER use fmt.Println for logging
- ❌ NEVER log in hot loops (performance)

**Log levels with error envelope:**
```go
// Debug - Detailed debugging information (development only)
logger.Debug("Processing item", 
    zap.Int64("item_id", item.ID),
    zap.String("status", item.Status),
)

// Info - General informational messages
logger.Info("User created successfully", 
    zap.Int64("user_id", user.ID),
    zap.String("email", maskEmail(user.Email)),
)

// Warn - Unexpected but recoverable situations
logger.Warn("Cache miss", 
    zap.String("key", maskSensitive(cacheKey)),
    zap.Duration("stale_time", time.Since(item.CreatedAt)),
)

// Error - Error that requires attention
logger.Error("Failed to process payment", 
    zap.String("order_id", maskID(orderID)),
    zap.Error(err),
    zap.String("error_code", string(errors.GetCode(err))),
    zap.String("severity", getErrorSeverity(err)),
)

// Fatal - Critical error that stops the application
logger.Fatal("Failed to connect to database", 
    zap.Error(err),
    zap.String("host", maskHost(cfg.Host)),
)

// Helper function to get error severity from envelope
func getErrorSeverity(err error) string {
    if appErr, ok := err.(*errors.AppError); ok {
        if dbErr, ok := appErr.Err.(*DBError); ok {
            return dbErr.Severity
        }
    }
    return "unknown"
}
```

---

## Environment Configuration

**Configuration structure with error envelope:**
```go
// GOOD - Proper configuration with validation and error envelope
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Auth     AuthConfig     `mapstructure:"auth"`
    Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
    Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required"`
    WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required"`
}

// Error envelope for configuration errors
type ConfigError struct {
    Key   string // Configuration key that failed
    Value string // Invalid value
    Err   error  // Underlying error
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("config key %s=%s: %v", e.Key, e.Value, e.Err)
}

func (e *ConfigError) Unwrap() error {
    return e.Err
}

// Load and validate configuration
func LoadConfig(path string) (*Config, error) {
    viper.SetConfigFile(path)
    viper.SetConfigType("env")
    
    // Set defaults
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.read_timeout", "30s")
    viper.SetDefault("server.write_timeout", "30s")
    viper.SetDefault("database.sslmode", "disable")
    viper.SetDefault("log.level", "info")
    viper.SetDefault("log.environment", "development")
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, &errors.AppError{
            Code: errors.ErrCodeInvalid,
            Err: &ConfigError{
                Key: "config_file",
                Err: err,
            },
            Message: fmt.Sprintf("failed to read config from %s", path),
        }
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, &errors.AppError{
            Code: errors.ErrCodeInvalid,
            Err: &ConfigError{
                Key: "unmarshal",
                Err: err,
            },
            Message: "failed to unmarshal config",
        }
    }
    
    // Validate
    if err := config.Validate(); err != nil {
        return nil, err // Returns envelope error from Validate()
    }
    
    return &config, nil
}

func (c *Config) Validate() error {
    // Validate required fields
    if c.Database.Host == "" {
        return &errors.AppError{
            Code: errors.ErrCodeInvalid,
            Err: &ConfigError{
                Key:   "database.host",
                Value: "",
            },
            Message: "database host is required",
        }
    }
    if c.Database.User == "" {
        return &errors.AppError{
            Code: errors.ErrCodeInvalid,
            Err: &ConfigError{
                Key:   "database.user",
                Value: "",
            },
            Message: "database user is required",
        }
    }
    if c.Database.DBName == "" {
        return &errors.AppError{
            Code: errors.ErrCodeInvalid,
            Err: &ConfigError{
                Key:   "database.dbname",
                Value: "",
            },
            Message: "database name is required",
        }
    }
    
    // Validate ranges
    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return &errors.AppError{
            Code: errors.ErrCodeInvalid,
            Err: &ConfigError{
                Key:   "server.port",
                Value: fmt.Sprintf("%d", c.Server.Port),
            },
            Message: "port must be between 1 and 65535",
        }
    }
    
    return nil
}
```

**Configuration rules:**
- ✅ Use environment variables for secrets
- ✅ Provide sensible defaults
- ✅ Validate configuration on startup
- ✅ Use `.env.example` for documentation
- ✅ Return error envelope for config errors
- ✅ Validate ranges and required fields
- ❌ NEVER commit `.env` with real credentials
- ❌ NEVER hardcode URLs or credentials
- ❌ NEVER proceed with invalid configuration

**.env.example:**
```bash
# Server Configuration
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=golang_sample
DB_SSLMODE=disable

# Authentication
JWT_SECRET=your_jwt_secret_here
JWT_EXPIRATION=24h

# Logging
LOG_LEVEL=info
LOG_ENVIRONMENT=development
```

**Config loading with proper error handling:**
```go
// GOOD - Bootstrap with config error handling
func NewApp(cfgPath string) (*App, func(), error) {
    // Load configuration
    cfg, err := LoadConfig(cfgPath)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to load config: %w", err)
    }
    
    // Setup logger
    logger, err := NewLogger(cfg.Log)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to create logger: %w", err)
    }
    
    // Setup database
    db, cleanup, err := NewDatabase(cfg.Database)
    if err != nil {
        logger.Error("Failed to connect to database", zap.Error(err))
        return nil, nil, fmt.Errorf("failed to setup database: %w", err)
    }
    
    logger.Info("Database connected successfully")
    
    // Setup repositories
    authRepo := repository.NewUserRepository(db)
    
    // Setup services
    authService := usecase.NewAuthService(authRepo, logger)
    
    // Setup handlers
    authHandler := handler.NewAuthHandler(authService)
    
    // Setup HTTP server
    server := rest.NewServer(cfg.Server, authHandler)
    
    return &App{
        Server: server,
        Logger: logger,
        DB:     db,
    }, cleanup, nil
}
```

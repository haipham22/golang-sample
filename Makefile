# Govern Library Makefile

.PHONY: test build lint clean install-tools

test:
	@echo "Running govern library tests..."
	mise exec -- go test ./config/... ./cron/... ./database/... ./errors/... ./graceful/... ./healthcheck/... ./http/... ./log/... ./metrics/... ./mq/... ./retry/...

build:
	@echo "Building govern library packages..."
	mise exec -- go build ./config/... ./cron/... ./database/... ./errors/... ./graceful/... ./healthcheck/... ./http/... ./log/... ./metrics/... ./mq/... ./retry/...

lint:
	@echo "Running linters on govern library..."
	mise exec -- golangci-lint run ./config/... ./cron/... ./database/... ./errors/... ./graceful/... ./healthcheck/... ./http/... ./log/... ./metrics/... ./mq/... ./retry/...

clean:
	@echo "Cleaning govern library build artifacts..."
	find . -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true

install-tools:
	@echo "Installing development tools..."
	mise install

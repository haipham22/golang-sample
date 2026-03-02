.PHONY: all test mocks lint fmt tidy build serverd clean test-rest

# Default target
all: build test

# Run all tests
test:
	go test -v -race ./...

# Run tests for specific package
test-rest:
	go test -v -race ./internal/handler/rest/...

# Generate mocks with mockery
mocks:
	@echo "Generating mocks..."
	@mockery
	@echo "Mocks generated"

# Clean all mock files
clean-mocks:
	@echo "Removing all mock files..."
	@find internal -type f -path "*/mocks/*.go" -delete
	@rm -rf internal/storage/user/mocks
	@rm -rf internal/service/auth/mocks
	@echo "Mock files cleaned"

# Remove all mock files and regenerate them
regenerate-mocks: clean-mocks mocks
	@echo "Mock regeneration complete!"

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	goimports -local golang-sample -w .

# Tidy go modules
tidy:
	go mod tidy

# Build the application
build:
	go build -o bin/serverd .

# Run the application
serverd:
	go run main.go serverd

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

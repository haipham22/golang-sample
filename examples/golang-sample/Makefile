# Sample Application Makefile

.PHONY: test build run lint clean install-tools generate-mocks generate-wire

test:
	@echo "Running sample app tests..."
	mise exec -- go test ./...

build:
	@echo "Building sample app..."
	mise exec -- go build -o bin/serverd .

run:
	@echo "Running sample app..."
	mise exec -- go run .

lint:
	@echo "Running linters on sample app..."
	mise exec -- golangci-lint run ./...

clean:
	@echo "Cleaning sample app build artifacts..."
	rm -rf bin/ tmp/

install-tools:
	@echo "Installing development tools..."
	mise install

generate-mocks:
	@echo "Generating mocks..."
	mise exec -- mockery

generate-wire:
	@echo "Generating wire dependencies..."
	rm internal/handler/rest/wire_gen.go
	mise exec -- wire ./internal/handler/rest/

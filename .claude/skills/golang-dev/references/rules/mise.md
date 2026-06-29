# Mise Toolchain Rules

**Best practices for using mise for Go toolchain management and development workflow automation.**

---

## Mise Overview

**What is mise:**
- Toolchain manager for multiple programming languages
- Version management for Go, Node.js, Python, and more
- Cross-platform support (Linux, macOS, Windows)
- Alternative to tools like asdf, nvm, gvm

**Why use mise:**
- Single tool for all language versions
- Consistent toolchain across team
- Project-specific version pinning
- Easy setup for new developers

---

## Mise Configuration

**mise.toml structure:**
```toml
[tools]
go = "1.25.5"
"go:golang.org/x/tools/cmd/goimports" = "latest"
"go:honnef.co/go/tools/cmd/staticcheck" = "latest"
"go:github.com/kisielk/errcheck" = "latest"
golangci-lint = "latest"
mockery = "latest"
swag = "latest"

pre-commit = "latest"

"npm:@commitlint/cli" = "latest"
"npm:@commitlint/config-conventional" = "latest"
```

**Tool categories:**
- **Go runtime:** `go = "1.25.5"`
- **Go tools:** `go:github.com/...` (tool packages)
- **CLI tools:** `golangci-lint`, `mockery`, `swag`
- **Other:** `pre-commit`, npm packages

---

## Installation & Setup

**Initial setup:**
```bash
# Install mise (if not already installed)
curl https://mise.run | sh

# Or using Homebrew (macOS)
brew install mise

# Install mise from source
git clone https://github.com/jdx/mise ~/.mise
```

**Initialize mise in project:**
```bash
# Create mise.toml in project root
mise init

# Or create manually
touch mise.toml
```

**Install all tools:**
```bash
mise install
```

**Verify installation:**
```bash
mise exec -- go version
mise exec -- golangci-lint version
mise exec -- swag version
```

---

## Go Version Management

**Specify Go version:**
```toml
[tools]
go = "1.26.0"  # Latest stable
```

**Use mise exec for Go commands:**
```bash
# Run with mise-managed Go
mise exec -- go version
mise exec -- go build ./...
mise exec -- go test ./...

# Direct go command (system version)
go version  # Uses system Go, NOT mise
```

**Update Go version:**
```bash
# Update mise.toml
go = "1.26.0"

# Install new version
mise install

# Verify
mise exec -- go version
```

---

## Go Tools Management

**Install Go tool packages:**
```toml
# Tool packages
"go:golang.org/x/tools/cmd/goimports" = "latest"
"go:honnef.co/go/tools/cmd/staticcheck" = "latest"
"go:github.com/kisielk/errcheck" = "latest"
```

**Use Go tools via mise:**
```bash
# Format code
mise exec -- goimports -w .

# Static analysis
mise exec -- staticcheck ./...
mise exec -- errcheck -blank ./...

# Install Go tool from source
mise exec -- go install github.com/swaggo/swag/cmd/swag@latest
```

---

## CLI Tools Management

**Language-agnostic tools:**
```toml
# Go tools
golangci-lint = "latest"
mockery = "latest"
swag = "latest"

# Pre-commit (Python)
pre-commit = "latest"

# NPM packages (Node.js)
"npm:@commitlint/cli" = "latest"
"npm:@commitlint/config-conventional" = "latest"
```

**Use CLI tools:**
```bash
# Run golangci-lint
mise exec -- golangci-lint run

# Generate mocks
mise exec -- mockery --all

# Generate swagger docs
mise exec -- swag init -g cmd/api/main.go

# Run pre-commit hooks
mise exec -- pre-commit run --all-files
```

**Update all tools:**
```bash
mise install
```

---

## Development Workflow

**Before development:**
```bash
# 1. Ensure tools are installed
mise install

# 2. Verify Go version
mise exec -- go version

# 3. Check tool versions
mise list
```

**During development:**
```bash
# Format code
mise exec -- goimports -w .

# Run linters
mise exec -- golangci-lint run
mise exec -- staticcheck ./...
mise exec -- errcheck -blank ./...

# Run tests
mise exec -- go test ./...
mise exec -- go test -race ./...
mise exec -- go test -cover ./...

# Build to verify
mise exec -- go build ./...
```

**Before commit:**
```bash
# 1. Format code
mise exec -- goimports -w .

# 2. Run linters
mise exec -- golangci-lint run
mise exec -- staticcheck ./...
mise exec -- errcheck -blank ./...

# 3. Run tests
mise exec -- go test ./...
mise exec -- go test -race ./...
mise exec -- go test -cover ./...

# 4. Pre-commit hooks (if configured)
mise exec -- pre-commit run --all-files
```

---

## Version Pinning

**Pin specific versions:**
```toml
[tools]
# Exact version
go = "1.26.0"

# Tools with specific versions
golangci-lint = "v1.55.0"
mockery = "v2.14.0"

# Latest version (default)
swag = "latest"
```

**Update to latest:**
```bash
# Update single tool
mise install golangci-lint@latest

# Update all
mise install
```

---

## Project-Specific Tools

**Common Go tools for golang-sample:**
```toml
[tools]
# Go runtime
go = "1.26.0"

# Code formatting
"go:golang.org/x/tools/cmd/goimports" = "latest"

# Static analysis
"go:honnef.co/go/tools/cmd/staticcheck" = "latest"
"go:github.com/kisielk/errcheck" = "latest"

# Linting
golangci-lint = "latest"

# Testing
mockery = "latest"

# Documentation
swag = "latest"

# Commit hooks
pre-commit = "latest"

# NPM for commitlint
"npm:@commitlint/cli" = "latest"
"npm:@commitlint/config-conventional" = "lastest"
```

---

## Best Practices

**✅ DO:**
- Pin Go version in mise.toml
- Use `mise exec` for all tool commands
- Run `mise install` after updating mise.toml
- Use `latest` for frequently updated tools
- Pin versions for stability in CI/CD
- Keep mise.toml in project root

**❌ DON'T:**
- Mix system Go and mise Go
- Run `go install` for tools (use mise instead)
- Forget to run `mise install` after updating tools
- Use outdated tool versions
- Commit mise.local files

---

## Troubleshooting

**Tool not found:**
```bash
# Check tool is in mise.toml
grep "tool-name" mise.toml

# Install missing tool
mise install

# Verify installation
mise exec -- tool-name --version
```

**Wrong Go version:**
```bash
# Check which Go is being used
mise exec -- go version
which go

# Reinstall mise Go version
mise install go@1.26.0
```

**Tool installation failures:**
```bash
# Clear mise cache
rm -rf ~/.mise/cache

# Reinstall
mise install

# If still failing, check network
curl https://mise.run | sh
```

---

## CI/CD Integration

**GitHub Actions with mise:**
```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install mise
        run: curl https://mise.run | sh
      
      - name: Install tools
        run: mise install
      
      - name: Run tests
        run: mise exec -- go test ./...
      
      - name: Run linters
        run: mise exec -- golangci-lint run
```

**Docker with mise:**
```dockerfile
FROM golang:1.26

# Install mise
RUN curl https://mise.run | sh

# Install project tools
COPY mise.toml .
RUN mise install

# Use mise in commands
RUN mise exec -- go build ./...
```

---

## Quick Reference

**Common mise commands:**
```bash
mise install              # Install/update tools
mise exec -- go version    # Run command with mise
mise list                  # List installed tools
mise upgrade [tool]        # Upgrade specific tool
mise install [tool]@version # Install specific version
```

**Go workflow:**
```bash
mise exec -- goimports -w .     # Format
mise exec -- staticcheck ./...   # Static analysis
mise exec -- golangci-lint run  # Lint
mise exec -- go test ./...      # Test
mise exec -- go build ./...     # Build
```

**Tool management:**
```bash
mise install                 # Update all tools
mise install golangci-lint     # Update specific tool
mise list                     # Check versions
mise exec -- swag version     # Verify tool
```

---

## Comparison: mise vs Alternatives

| Feature | mise | asdf | nvm/gvm | Direct install |
|---------|-----|-----|----------|---------------|
| **Multi-language** | ✅ Yes | ✅ Yes | ❌ Single | ❌ Manual |
| **Go support** | ✅ Excellent | ✅ Good | ✅ Good | ❌ Manual |
| **Node support** | ✅ Excellent | ✅ Good | ❌ No | ❌ Manual |
| **Project configs** | ✅ Easy | ⚠️ Complex | ❌ No | ❌ No |
| **Speed** | 🚀 Fast | 🐢 Slower | 🐢 Slower | 🚀 Fast |
| **Team sync** | ✅ Simple file | ⚠️ Complex | ⚠️ Per-user | ❌ Manual |

**mise advantages:**
- Single `.toml` file for entire team
- Cross-platform consistency
- Simple version updates
- Great Go support
- Active development

---

## Version Strategy

**For Go projects:**
```toml
# Pin exact Go version (recommended)
go = "1.26.0"

# Use latest for Go tools
"go:golang.org/x/tools/cmd/goimports" = "latest"
```

**For production:**
```toml
# Pin all versions for reproducibility
go = "1.26.0"
golangci-lint = "v1.55.0"
mockery = "v2.14.0"
```

**For development:**
```toml
# Use latest for flexibility
go = "latest"
golangci-lint = "latest"
mockery = "latest"
```

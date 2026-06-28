# Contributing to Govern

Thanks for your interest in contributing to Govern!

## Development Setup

```bash
# Clone
git clone https://github.com/haipham22/govern.git
cd govern

# Install mise (manages Go version and tools)
curl https://mise.run | sh

# Install tools
mise install

# Verify
mise exec -- go version
```

## Repository

Monorepo with two modules (no `go.work`):

- **Govern library** (root) — `github.com/haipham22/govern`
- **Sample app** — `examples/golang-sample/` (`github.com/haipham22/golang-sample`, imports govern
  via `replace => ../../`)

## Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feat/your-feature-name
   ```

2. **Make changes** — follow existing style and patterns; add tests; update docs.

3. **Test (govern library, root)**
   ```bash
   mise exec -- goimports -w .
   mise exec -- golangci-lint run
   mise exec -- staticcheck ./...
   mise exec -- go test -race ./...
   mise exec -- go build ./...
   ```

4. **Test (sample app)**
   ```bash
   cd examples/golang-sample
   mise exec -- go test -race ./...
   mise exec -- go build .
   ```

5. **Commit** (see convention below).

## Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation changes
- `test:` — Test additions/changes
- `refactor:` — Code refactoring
- `chore:` — Maintenance tasks

**Note:** For changes in the `.claude/` directory, do not use `chore:` or `docs:` prefixes.

No AI references in commit messages.

## Pull Request Process

1. Push your branch and open a PR with a clear description.
2. Ensure CI checks pass (govern + sample app workflows).
3. Address review feedback.
4. Merge when approved.

## Code Style

- File names: `snake_case` (`user_service.go`)
- Exported types: `PascalCase` (`UserService`)
- Private variables: `camelCase` (`userRepo`)
- Packages: lowercase, singular (`auth`, `storage`)
- Keep files under ~200 lines; split when they grow.
- Add comments for complex logic; keep functions focused.

## Testing

- Table-driven tests for multiple cases.
- Mock external dependencies (Mockery).
- Test success and error paths; aim for 80%+ coverage.
- Always run with `-race`.

## Adding a Govern Package

1. Create the package directory at root: `your-package/`
2. Add package doc comment (`doc.go` or package comment).
3. Add usage examples (`example_test.go`).
4. Update root `README.md` package list.
5. Ensure the package is framework-agnostic.
6. Add comprehensive tests and include it in the root `Makefile`/`test.yml` paths.

## Questions?

- Open an issue for bugs or feature requests.
- Start a discussion for questions or ideas.
- Check existing issues first.

Thanks for contributing to Govern!

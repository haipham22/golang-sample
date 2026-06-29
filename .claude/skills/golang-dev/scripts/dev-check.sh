#!/usr/bin/env bash
# dev-check.sh — portable pre-commit quality gate for the sample app.
# Auto-detects sample root. Uses mise when available; falls back to bare tools.
#
# Runs: goimports, go vet, golangci-lint, staticcheck, errcheck, go test -race.
# A missing tool is reported but does not fail the gate (skip with warning).
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$("$HERE/detect-sample-root.sh" 2>/dev/null || true)"
[ -n "$ROOT" ] || { echo "ERROR: sample root not found." >&2; exit 1; }

# Wrap a command with `mise exec --` when mise is available.
run() {
  if command -v mise >/dev/null 2>&1; then mise exec -- "$@"; else "$@"; fi
}
have() { command -v "$1" >/dev/null 2>&1 || mise exec -- command -v "$1" >/dev/null 2>&1; }

cd "$ROOT"
FAIL=0
step() { echo "== $* =="; }

step "goimports"
if have goimports; then
  run goimports -l -w . || { echo "✗ goimports"; FAIL=1; }
else echo "  (skip: goimports not installed)"; fi

step "go vet"
run go vet ./... || { echo "✗ go vet"; FAIL=1; }

step "golangci-lint"
if have golangci-lint; then
  run golangci-lint run ./... || { echo "✗ golangci-lint"; FAIL=1; }
else echo "  (skip: golangci-lint not installed)"; fi

step "staticcheck"
if have staticcheck; then
  run staticcheck ./... || { echo "✗ staticcheck"; FAIL=1; }
else echo "  (skip: staticcheck not installed)"; fi

step "errcheck"
if have errcheck; then
  run errcheck -blank ./... || { echo "✗ errcheck"; FAIL=1; }
else echo "  (skip: errcheck not installed)"; fi

step "tests (race)"
run go test -race ./... || { echo "✗ tests"; FAIL=1; }

[ "$FAIL" -eq 0 ] && echo "✓ dev-check passed ($ROOT)" || echo "✗ dev-check FAILED"
exit "$FAIL"

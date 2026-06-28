#!/usr/bin/env bash
# validate.sh — Sample app validation: compile, vet, test, coverage, and (post-migration) no-Wire check.
# Run from examples/golang-sample/
set -euo pipefail

PASS=0
fail() { echo "✗ $1"; PASS=1; }
ok()   { echo "✓ $1"; }

echo "== Compilation =="
if mise exec -- go build ./...; then ok "build ./..."; else fail "build ./..."; fi

echo "== Vet =="
if mise exec -- go vet ./...; then ok "vet"; else fail "vet"; fi

echo "== Tests (race) =="
if mise exec -- go test -race ./...; then ok "tests (race)"; else fail "tests (race)"; fi

echo "== Coverage =="
mise exec -- go test -coverprofile=coverage.out ./... >/dev/null 2>&1 || true
if [ -f coverage.out ]; then
  COV=$(mise exec -- go tool cover -func=coverage.out 2>/dev/null | tail -1 | grep -oE '[0-9.]+%' | tail -1)
  echo "  total coverage: ${COV:-unknown}"
  ok "coverage report generated"
else
  fail "coverage report not generated"
fi

echo "== Wire removal check =="
# After Phase 11, no source file (excluding wire.go build-tag file once deleted) should import google/wire.
WIRE=$(grep -rl "google/wire" --include="*.go" . 2>/dev/null | grep -v 'wire_gen.go' || true)
if [ -z "$WIRE" ]; then ok "no google/wire imports"; else fail "google/wire still imported: $WIRE"; fi

echo "== govern/errors removal check =="
# Match only actual import paths (quoted), not prose mentions in comments.
ERR=$(grep -rn '"github.com/haipham22/govern/errors"' --include="*.go" . 2>/dev/null || true)
if [ -z "$ERR" ]; then ok "no govern/errors imports"; else fail "govern/errors imported: $ERR"; fi

exit $PASS

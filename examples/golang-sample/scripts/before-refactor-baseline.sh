#!/usr/bin/env bash
# before-refactor-baseline.sh — capture a test + coverage baseline before a
# refactor. Run from examples/golang-sample/. Output goes to tmp/baseline/.
#
# Usage:
#   ./scripts/before-refactor-baseline.sh [label]
# Example:
#   ./scripts/before-refactor-baseline.sh pre-wire-removal
set -euo pipefail

LABEL="${1:-baseline}"
OUT="tmp/baseline-${LABEL}"
mkdir -p "$(dirname "$OUT")"

echo "== Capturing baseline: ${LABEL} =="
echo "== Build =="
mise exec -- go build ./...

echo "== Test (race) =="
mise exec -- go test -race ./... 2>&1 | tee "${OUT}-tests.txt"

echo "== Coverage =="
mise exec -- go test -coverprofile="${OUT}.out" ./... >/dev/null 2>&1
mise exec -- go tool cover -func="${OUT}.out" > "${OUT}-coverage.txt" 2>/dev/null
echo "Total: $(tail -1 "${OUT}-coverage.txt")"

echo
echo "== Baseline saved to ${OUT}* =="
echo "   - ${OUT}-tests.txt      (test output)"
echo "   - ${OUT}.out            (raw coverage profile)"
echo "   - ${OUT}-coverage.txt   (per-function coverage)"
echo
echo "Compare after refactor with:"
echo "  diff ${OUT}-coverage.txt tmp/baseline-<new>-coverage.txt"

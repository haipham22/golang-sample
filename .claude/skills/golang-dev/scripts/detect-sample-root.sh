#!/usr/bin/env bash
# detect-sample-root.sh — print the Go sample app root.
# Portable: works whether the app is at repo root or under examples/<name>.
#
# Resolution order:
#   1. $SAMPLE_ROOT env var (if set)
#   2. current directory if it contains go.mod
#   3. examples/*/go.mod (first match)
#   4. search ancestors for go.mod
#
# Usage:
#   ROOT="$(./detect-sample-root.sh)" || exit 1
#   cd "$ROOT"
set -euo pipefail

if [ -n "${SAMPLE_ROOT:-}" ]; then
  printf '%s\n' "$SAMPLE_ROOT"
  exit 0
fi

# Prefer a sample app under examples/<app>/go.mod (monorepo layout).
shopt -s nullglob 2>/dev/null || true
for f in examples/*/go.mod; do
  printf '%s/%s\n' "$PWD" "$(dirname "$f")"
  exit 0
done

# Standalone project: go.mod at current directory.
if [ -f go.mod ]; then
  printf '%s\n' "$PWD"
  exit 0
fi

# Walk up ancestors.
dir="$PWD"
while [ "$dir" != "/" ]; do
  for f in "$dir"/examples/*/go.mod; do
    printf '%s\n' "$(dirname "$f")"
    exit 0
  done
  if [ -f "$dir/go.mod" ]; then
    printf '%s\n' "$dir"
    exit 0
  fi
  dir="$(dirname "$dir")"
done

echo "ERROR: no go.mod found (cwd=$PWD). Set SAMPLE_ROOT or run inside the app." >&2
exit 1

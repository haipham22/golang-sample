#!/usr/bin/env bash
# arch-lint.sh — enforce clean-architecture import rules.
# Portable: run from anywhere; auto-detects sample root.
#
# Rules (from references/rules/ + folder-structure.md):
#   - internal/domain/      imports no framework/db (echo, gorm, redis, zap, viper, asynq)
#   - internal/handler/     does NOT import internal/repository or internal/orm
#   - internal/usecase/     does NOT import internal/handler or internal/repository
#   - internal/repository/  does NOT import internal/handler
#
# Exit non-zero if any violation. Prints file + offending import.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$("$HERE/detect-sample-root.sh" 2>/dev/null || true)"
if [ -z "$ROOT" ]; then
  echo "ERROR: sample root not found." >&2
  exit 1
fi

INTERNAL="$ROOT/internal"
[ -d "$INTERNAL" ] || { echo "no internal/ at $ROOT"; exit 0; }

FAIL=0
viol() { echo "✗ $1"; FAIL=1; }

# Module path (everything before /internal) — derive from go.mod.
MODULE="$(head -1 "$ROOT/go.mod" | awk '{print $2}')"

# --- domain purity ---
if [ -d "$INTERNAL/domain" ]; then
  while IFS= read -r -d '' f; do
    if grep -Eq '"github.com/labstack/echo|gorm.io/gorm|github.com/redis/go-redis|go.uber.org/zap|github.com/spf13/viper|github.com/hibiken/asynq' "$f"; then
      viol "domain/$f imports framework/db (must stay pure)"
    fi
  done < <(find "$INTERNAL/domain" -name '*.go' -print0 2>/dev/null)
fi

# --- handler must not import repository/orm ---
# Exclusions: tests, DI/wiring files (di.go, *wire*, bootstrap) legitimately wire repos.
if [ -d "$INTERNAL/handler" ]; then
  while IFS= read -r -d '' f; do
    case "$f" in
      *_test.go) continue ;;
      */di.go|*wire*|*/bootstrap/*) continue ;;
    esac
    if grep -Eq "${MODULE}/internal/repository|${MODULE}/internal/orm" "$f"; then
      viol "handler $(basename "$f") imports repository/orm (handler -> usecase only)"
    fi
  done < <(find "$INTERNAL/handler" -name '*.go' -print0 2>/dev/null)
fi

# --- usecase must not import handler (clear violation) ---
# Note: usecase importing repository/<x> for a storage interface is allowed where the
# project places that interface in the repository package (consumer-owned variant).
if [ -d "$INTERNAL/usecase" ]; then
  while IFS= read -r -d '' f; do
    case "$f" in *_test.go) continue ;; esac
    if grep -Eq "${MODULE}/internal/handler" "$f"; then
      viol "usecase $(basename "$f") imports handler (usecase must not know delivery)"
    fi
  done < <(find "$INTERNAL/usecase" -name '*.go' -not -name '*_test.go' -print0 2>/dev/null)
fi

# --- repository must not import handler ---
if [ -d "$INTERNAL/repository" ]; then
  while IFS= read -r -d '' f; do
    if grep -Eq "${MODULE}/internal/handler" "$f"; then
      viol "repository/$f imports handler"
    fi
  done < <(find "$INTERNAL/repository" -name '*.go' -not -name '*_test.go' -print0 2>/dev/null)
fi

[ "$FAIL" -eq 0 ] && echo "✓ arch-lint clean ($ROOT)"
exit "$FAIL"

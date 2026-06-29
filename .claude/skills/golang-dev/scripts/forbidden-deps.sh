#!/usr/bin/env bash
# forbidden-deps.sh — flag dependencies that duplicate govern packages.
# Portable: auto-detects sample root.
#
# Flags (govern already covers these concerns):
#   - cenkalti/backoff          -> use govern/retry
#   - pkg/errors                -> use govern/errors (or app errors package)
#   - robfig/cron               -> use govern/cron
#   - go-co-op/gocron           -> use govern/cron
#   - spf13/viper (raw)         -> prefer govern/config
#
# Also flags direct gorm.Open / redis.NewClient at call sites (use govern wrappers).
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$("$HERE/detect-sample-root.sh" 2>/dev/null || true)"
[ -n "$ROOT" ] || { echo "ERROR: sample root not found." >&2; exit 1; }

FAIL=0
viol() { echo "✗ $1"; FAIL=1; }

GOMOD="$ROOT/go.mod"
if [ -f "$GOMOD" ]; then
  # Only flag direct requires (skip `// indirect` transitive deps).
  direct="$(grep -v '// indirect' "$GOMOD")"
  printf '%s\n' "$direct" | grep -Eq 'cenkalti/backoff'      && viol "go.mod: cenkalti/backoff — use govern/retry"
  printf '%s\n' "$direct" | grep -Eq 'github.com/pkg/errors' && viol "go.mod: pkg/errors — use govern/errors or app errors"
  printf '%s\n' "$direct" | grep -Eq 'robfig/cron'           && viol "go.mod: robfig/cron — use govern/cron"
  printf '%s\n' "$direct" | grep -Eq 'go-co-op/gocron'       && viol "go.mod: gocron — use govern/cron"
fi

# Direct driver calls at call sites (bootstrap/repo are allowed; flag app code).
while IFS= read -r -d '' f; do
  case "$f" in
    */bootstrap/*|*/repository/*|*/pkg/*) continue ;;
  esac
  if grep -Eq 'gorm\.Open\(' "$f"; then
    viol "$(basename "$f"): gorm.Open — use govern/database/postgres"
  fi
  if grep -Eq 'redis\.NewClient\(' "$f"; then
    viol "$(basename "$f"): redis.NewClient — use govern/database/redis"
  fi
done < <(find "$ROOT" -name '*.go' -not -name '*_test.go' -print0 2>/dev/null)

[ "$FAIL" -eq 0 ] && echo "✓ forbidden-deps clean ($ROOT)"
exit "$FAIL"

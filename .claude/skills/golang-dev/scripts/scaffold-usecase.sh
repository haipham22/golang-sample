#!/usr/bin/env bash
# scaffold-usecase.sh — generate a clean-arch usecase skeleton.
# Creates internal/usecase/<name>/{service.go,impl.go,dto.go}.
#
# Usage:
#   ./scaffold-usecase.sh product
#   ./scaffold-usecase.sh product /abs/app/root
#
# Idempotent: refuses to overwrite existing files.
set -euo pipefail

NAME="${1:?usage: scaffold-usecase.sh <name> [app-root]}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="${2:-$("$HERE/detect-sample-root.sh")}"

MODULE="$(head -1 "$ROOT/go.mod" | awk '{print $2}')"
DIR="$ROOT/internal/usecase/$NAME"
PKG="$NAME"

if [ -d "$DIR" ]; then
  echo "ERROR: $DIR already exists." >&2
  exit 1
fi
mkdir -p "$DIR"

# Capitalize first letter for exported type names.
Title="$(printf '%s' "$NAME" | sed 's/^./\U&/')"

cat > "$DIR/service.go" <<EOF
package $PKG

import "context"

// ${Title}Repository is the storage contract consumed by this usecase.
// Define interfaces here (consumer-owned), not in domain.
type ${Title}Repository interface {
	// TODO: add storage methods, e.g.
	// FindByID(ctx context.Context, id int64) (domain.${Title}, error)
}

// Service implements the $NAME usecase.
type Service interface {
	// TODO: add usecase methods.
}

type service struct {
	repo ${Title}Repository
}

// NewService constructs the $NAME usecase.
func NewService(repo ${Title}Repository) Service {
	return &service{repo: repo}
}

var _ Service = (*service)(nil)
EOF

cat > "$DIR/impl.go" <<EOF
package $PKG

// Use case implementations live here. Keep I/O methods context-first:
// func (s *service) DoThing(ctx context.Context, input Input) (Output, error) { ... }
EOF

cat > "$DIR/dto.go" <<EOF
package $PKG

// Request/Response DTOs for the $NAME usecase.
// Keep HTTP shape in internal/schemas; DTOs here are usecase-level.
EOF

echo "created: $DIR"
echo "  service.go  (interface + constructor)"
echo "  impl.go     (implementations)"
echo "  dto.go      (DTOs)"
echo "module: $MODULE"

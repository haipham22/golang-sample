// Package postgres contains GORM-based repository implementations for
// persistence. The Repository interface is defined here (consumer-defined,
// mirroring internal/repository/user) and implemented by the unexported repo
// struct. Upper layers (usecase) depend on the interface, not the concrete
// implementation.
package postgres

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/haipham22/golang-sample/internal/domain"
)

// ListParams carries pagination + filtering for product listing. Limit/Offset
// are applied when > 0; otherwise sensible defaults are used by the repo.
type ListParams struct {
	Limit  int
	Offset int
}

// Repository is the consumer-defined persistence interface for product
// aggregates. Defined here per the bxcodec/clean-arch pattern (interface lives
// with the consumer package; the implementation is in the same package for
// this sample but could move to a sub-package without breaking callers).
type Repository interface {
	Create(ctx context.Context, product *domain.Product) (*domain.Product, error)
	FindByID(ctx context.Context, id uint) (*domain.Product, error)
	List(ctx context.Context, params ListParams) ([]*domain.Product, int64, error)
	Delete(ctx context.Context, id uint) error
}

// repo is the GORM-backed implementation of Repository.
type repo struct {
	log *zap.SugaredLogger
	db  *gorm.DB
}

// Compile-time guard: repo implements Repository.
var _ Repository = (*repo)(nil)

// New wires the product repository with its dependencies and returns the
// Repository interface. Mirrors internal/repository/user.New(log, db).
func New(log *zap.SugaredLogger, db *gorm.DB) Repository {
	return &repo{
		log: log,
		db:  db,
	}
}

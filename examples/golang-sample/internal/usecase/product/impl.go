package product

import (
	"context"

	"go.uber.org/zap"

	"github.com/haipham22/golang-sample/internal/domain"
	apperrors "github.com/haipham22/golang-sample/internal/errors"
	"github.com/haipham22/golang-sample/internal/repository/postgres"
)

// impl is the default Service implementation. It owns no transport or ORM
// types — only the Repository interface and a logger.
type impl struct {
	log  *zap.SugaredLogger
	repo postgres.Repository
}

// Compile-time guard.
var _ Service = (*impl)(nil)

// NewService wires a product Service with its repository dependency.
func NewService(log *zap.SugaredLogger, repo postgres.Repository) Service {
	return &impl{log: log, repo: repo}
}

// Create validates the input, persists a new product, and returns the DTO.
func (s *impl) Create(ctx context.Context, req CreateRequest) (*domain.Product, error) {
	p := &domain.Product{Name: req.Name, Price: req.Price}
	if err := p.Validate(); err != nil {
		s.log.Warnf("product validation failed: %v", err)
		return nil, apperrors.WrapCode(apperrors.CodeInvalid, err)
	}

	created, err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	s.log.Infof("product created: ID=%d", created.ID)
	return created, nil
}

// GetByID returns a single product by ID or an apperrors.NotFound.
func (s *impl) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	if id == 0 {
		return nil, apperrors.NewCode(apperrors.CodeInvalid, "product id is required")
	}
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// List returns a page of products.
func (s *impl) List(ctx context.Context, req ListRequest) (*ListResponse, error) {
	items, total, err := s.repo.List(ctx, postgres.ListParams{Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		return nil, err
	}
	dtos := make([]*ProductDTO, 0, len(items))
	for _, p := range items {
		dtos = append(dtos, toDTO(p))
	}
	return &ListResponse{Items: dtos, Total: total}, nil
}

// Delete removes a product by ID.
func (s *impl) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return apperrors.NewCode(apperrors.CodeInvalid, "product id is required")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.log.Infof("product deleted: ID=%d", id)
	return nil
}

// toDTO converts a domain product to its DTO. Kept here (not in a separate
// converter) because the mapping is trivial and only used by this use case.
func toDTO(p *domain.Product) *ProductDTO {
	if p == nil {
		return nil
	}
	return &ProductDTO{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

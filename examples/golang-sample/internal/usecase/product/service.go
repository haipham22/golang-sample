// Package product implements the product business-logic use cases. The Service
// interface is defined here (consumer-facing contract); the implementation
// lives in impl.go. DTOs are kept inline to keep the surface small.
package product

import (
	"context"
	"time"

	"github.com/haipham22/golang-sample/internal/domain"
)

// Service is the product use case contract. Callers (handlers) depend on this
// interface, not the concrete implementation.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*domain.Product, error)
	GetByID(ctx context.Context, id uint) (*domain.Product, error)
	List(ctx context.Context, req ListRequest) (*ListResponse, error)
	Delete(ctx context.Context, id uint) error
}

// CreateRequest is the DTO for creating a product.
type CreateRequest struct {
	Name  string
	Price float64
}

// ListRequest carries pagination for listing products.
type ListRequest struct {
	Limit  int
	Offset int
}

// ProductDTO is the API-facing product representation returned by use cases.
type ProductDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListResponse wraps a page of products with the total count.
type ListResponse struct {
	Items []*ProductDTO
	Total int64
}

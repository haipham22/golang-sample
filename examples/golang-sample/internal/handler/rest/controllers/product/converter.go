package product

import (
	"github.com/haipham22/golang-sample/internal/domain"
	"github.com/haipham22/golang-sample/internal/schemas"
)

// modelToSchema converts a domain Product to its HTTP schema.
func modelToSchema(p *domain.Product) *schemas.Product {
	if p == nil {
		return nil
	}
	return &schemas.Product{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

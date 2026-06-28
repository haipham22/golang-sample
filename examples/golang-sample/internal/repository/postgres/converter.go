package postgres

import (
	"github.com/haipham22/golang-sample/internal/domain"
	"github.com/haipham22/golang-sample/internal/orm"
)

// ormToProduct converts an ORM Product to a domain Product.
func ormToProduct(p *orm.Product) *domain.Product {
	if p == nil {
		return nil
	}
	return &domain.Product{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// productToORM converts a domain Product to an ORM Product.
func productToORM(p *domain.Product) *orm.Product {
	if p == nil {
		return nil
	}
	return &orm.Product{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// ormSliceToProducts converts a slice of ORM Products to domain Products.
func ormSliceToProducts(items []*orm.Product) []*domain.Product {
	if items == nil {
		return nil
	}
	result := make([]*domain.Product, len(items))
	for i, p := range items {
		result[i] = ormToProduct(p)
	}
	return result
}

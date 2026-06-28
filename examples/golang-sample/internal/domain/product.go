package domain

import (
	"errors"
	"time"
)

// Product-specific domain errors.
var (
	ErrProductNameRequired = errors.New("product name is required")
	ErrProductNameTooLong  = errors.New("product name must be at most 255 characters")
	ErrProductPriceInvalid = errors.New("product price must be zero or positive")
)

// Product is a pure domain entity for a catalog product. It has no
// dependencies on persistence (ORM) or API (schemas) layers.
type Product struct {
	ID        uint
	Name      string
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks the product against business rules. Returns a typed domain
// error so callers (usecase/repository) can map it to an apperrors code at the
// boundary.
func (p *Product) Validate() error {
	if p == nil {
		return ErrProductNameRequired
	}
	if p.Name == "" {
		return ErrProductNameRequired
	}
	if len(p.Name) > 255 {
		return ErrProductNameTooLong
	}
	if p.Price < 0 {
		return ErrProductPriceInvalid
	}
	return nil
}

// IsNew reports whether the product has not yet been persisted (ID not set).
func (p *Product) IsNew() bool {
	if p == nil {
		return true
	}
	return p.ID == 0
}

// IsEqual reports whether two products share the same identity (ID + Name).
func (p *Product) IsEqual(other *Product) bool {
	if p == nil || other == nil {
		return false
	}
	return p.ID == other.ID && p.Name == other.Name
}

// Clone returns a deep copy of the product.
func (p *Product) Clone() *Product {
	if p == nil {
		return nil
	}
	return &Product{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

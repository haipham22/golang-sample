package schemas

import "time"

// Product is the API-facing product representation. Mirrors the shape of
// usecase/product.ProductDTO but stays in the schemas package to keep the HTTP
// boundary self-contained (no usecase import from schemas).
type Product struct {
	ID        uint      `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Price     float64   `json:"price,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductCreateRequest is the request body for creating a product.
type ProductCreateRequest struct {
	Name  string  `json:"name" validate:"required,min=1,max=255"`
	Price float64 `json:"price" validate:"gte=0"`
}

// ProductListRequest carries optional pagination query params.
type ProductListRequest struct {
	Limit  int `query:"limit" validate:"omitempty,gte=1,lte=100"`
	Offset int `query:"offset" validate:"omitempty,gte=0"`
}

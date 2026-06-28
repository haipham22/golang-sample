package orm

import "time"

// Product is the GORM entity for the products table.
type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Price     float64   `gorm:"type:decimal(10,2);not null;default:0" json:"price"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName overrides the default table name.
func (Product) TableName() string {
	return "products"
}

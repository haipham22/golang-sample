package domain

import (
	"testing"
	"time"
)

func TestProduct_Validate(t *testing.T) {
	tests := []struct {
		name    string
		product *Product
		wantErr error
	}{
		{
			name:    "valid product",
			product: &Product{ID: 1, Name: "Widget", Price: 9.99},
			wantErr: nil,
		},
		{
			name:    "valid product with zero price",
			product: &Product{ID: 1, Name: "Freebie", Price: 0},
			wantErr: nil,
		},
		{
			name:    "missing name",
			product: &Product{ID: 1, Name: "", Price: 1.0},
			wantErr: ErrProductNameRequired,
		},
		{
			name:    "name too long",
			product: &Product{ID: 1, Name: string(make([]byte, 256)), Price: 1.0},
			wantErr: ErrProductNameTooLong,
		},
		{
			name:    "negative price",
			product: &Product{ID: 1, Name: "Negative", Price: -1.0},
			wantErr: ErrProductPriceInvalid,
		},
		{
			name:    "nil product",
			product: nil,
			wantErr: ErrProductNameRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.Validate()
			if err != tt.wantErr {
				t.Errorf("Product.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProduct_Validate_NameLengthBoundaries(t *testing.T) {
	t.Run("exactly 255 character name is valid", func(t *testing.T) {
		p := &Product{Name: string(make([]byte, 255)), Price: 1.0}
		if err := p.Validate(); err != nil {
			t.Errorf("255 char name should be valid, got: %v", err)
		}
	})

	t.Run("exactly 256 character name is too long", func(t *testing.T) {
		p := &Product{Name: string(make([]byte, 256)), Price: 1.0}
		if err := p.Validate(); err != ErrProductNameTooLong {
			t.Errorf("256 char name should be too long, got: %v", err)
		}
	})
}

func TestProduct_IsNew(t *testing.T) {
	tests := []struct {
		name    string
		product *Product
		want    bool
	}{
		{name: "with ID not new", product: &Product{ID: 1}, want: false},
		{name: "without ID is new", product: &Product{Name: "x"}, want: true},
		{name: "zero ID is new", product: &Product{ID: 0}, want: true},
		{name: "nil is new", product: nil, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.product.IsNew(); got != tt.want {
				t.Errorf("Product.IsNew() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProduct_IsEqual(t *testing.T) {
	a := &Product{ID: 1, Name: "A"}
	tests := []struct {
		name  string
		a, b  *Product
		want  bool
	}{
		{name: "same identity", a: a, b: &Product{ID: 1, Name: "A"}, want: true},
		{name: "different ID", a: a, b: &Product{ID: 2, Name: "A"}, want: false},
		{name: "different name", a: a, b: &Product{ID: 1, Name: "B"}, want: false},
		{name: "both nil", a: nil, b: nil, want: false},
		{name: "one nil", a: a, b: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsEqual(tt.b); got != tt.want {
				t.Errorf("Product.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProduct_Clone(t *testing.T) {
	now := time.Now()
	original := &Product{ID: 1, Name: "Widget", Price: 9.99, CreatedAt: now, UpdatedAt: now}

	t.Run("deep copy", func(t *testing.T) {
		c := original.Clone()
		if c == nil {
			t.Fatal("Clone() returned nil")
		}
		if c.ID != original.ID || c.Name != original.Name || c.Price != original.Price {
			t.Errorf("Clone() = %+v, want %+v", c, original)
		}
		if !c.CreatedAt.Equal(original.CreatedAt) {
			t.Errorf("Clone() CreatedAt = %v, want %v", c.CreatedAt, original.CreatedAt)
		}
	})

	t.Run("independent from original", func(t *testing.T) {
		c := original.Clone()
		c.Name = "Modified"
		c.ID = 999
		if original.Name == "Modified" || original.ID == 999 {
			t.Error("modifying clone affected original")
		}
	})

	t.Run("nil returns nil", func(t *testing.T) {
		var p *Product
		if got := p.Clone(); got != nil {
			t.Errorf("Clone() of nil = %v, want nil", got)
		}
	})
}

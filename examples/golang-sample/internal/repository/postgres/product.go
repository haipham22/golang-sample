package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/haipham22/golang-sample/internal/domain"
	apperrors "github.com/haipham22/golang-sample/internal/errors"
	"github.com/haipham22/golang-sample/internal/orm"
)

// defaultListLimit caps List results when no limit is provided.
const defaultListLimit = 100

// Create persists a new product. The product must pass domain validation
// before this call (enforced by the usecase layer). Returns the created
// product with its generated ID and timestamps populated.
func (r *repo) Create(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	if product == nil {
		return nil, apperrors.NewCode(apperrors.CodeInvalid, "product is required")
	}

	ormProduct := productToORM(product)
	if err := r.db.WithContext(ctx).Create(ormProduct).Error; err != nil {
		r.log.Errorf("failed to create product: %v", err)
		return nil, apperrors.WrapCode(apperrors.CodeInternal, err)
	}
	return ormToProduct(ormProduct), nil
}

// FindByID loads a product by primary key. Returns an apperrors.NotFound when
// the row does not exist.
func (r *repo) FindByID(ctx context.Context, id uint) (*domain.Product, error) {
	var ormProduct orm.Product
	err := r.db.WithContext(ctx).First(&ormProduct, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound(fmt.Sprintf("product %d", id))
		}
		r.log.Errorf("failed to find product %d: %v", id, err)
		return nil, apperrors.WrapCode(apperrors.CodeInternal, err)
	}
	return ormToProduct(&ormProduct), nil
}

// List returns a paginated slice of products and the total matching count.
// When params.Limit <= 0 the default limit is applied.
func (r *repo) List(ctx context.Context, params ListParams) ([]*domain.Product, int64, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&orm.Product{}).Count(&total).Error; err != nil {
		r.log.Errorf("failed to count products: %v", err)
		return nil, 0, apperrors.WrapCode(apperrors.CodeInternal, err)
	}

	var ormProducts []*orm.Product
	query := r.db.WithContext(ctx).
		Model(&orm.Product{}).
		Order("id ASC").
		Limit(limit)
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}
	if err := query.Find(&ormProducts).Error; err != nil {
		r.log.Errorf("failed to list products: %v", err)
		return nil, 0, apperrors.WrapCode(apperrors.CodeInternal, err)
	}
	return ormSliceToProducts(ormProducts), total, nil
}

// Delete removes a product by ID. Returns apperrors.NotFound when the row does
// not exist (RowsAffected == 0).
func (r *repo) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&orm.Product{}, id)
	if result.Error != nil {
		r.log.Errorf("failed to delete product %d: %v", id, result.Error)
		return apperrors.WrapCode(apperrors.CodeInternal, result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFound(fmt.Sprintf("product %d", id))
	}
	return nil
}

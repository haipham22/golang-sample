// Package product contains the HTTP controllers for the product resource.
// Controllers are thin: they bind/validate input, delegate to the use case
// Service, and map the returned domain model to an HTTP schema. Error mapping
// is centralized in the Echo error handler (handler/rest/handler.go).
package product

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
	"github.com/haipham22/golang-sample/internal/schemas"
	productservice "github.com/haipham22/golang-sample/internal/usecase/product"
)

// Controller handles HTTP requests for /api/products.
type Controller struct {
	service productservice.Service
}

// New creates a product HTTP controller.
func New(service productservice.Service) *Controller {
	return &Controller{service: service}
}

// PostProduct godoc
//
//	@Summary	Create product
//	@Description	Create a new catalog product
//	@Tags		products
//	@Accept		json
//	@Produce	json
//	@Param		req	body		schemas.ProductCreateRequest	true	"Product to create"
//	@Success	201	{object}	schemas.Response[schemas.Product]
//	@Failure	400	{object}	apperrors.Response
//	@Router		/api/products [post]
func (h *Controller) PostProduct(c *echo.Context) error {
	var req schemas.ProductCreateRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.WrapCode(apperrors.CodeInvalid, err)
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	created, err := h.service.Create(c.Request().Context(), productservice.CreateRequest{
		Name: req.Name, Price: req.Price,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, schemas.NewResponse(*modelToSchema(created)))
}

// GetProduct godoc
//
//	@Summary	Get product
//	@Description	Get a product by ID
//	@Tags		products
//	@Produce	json
//	@Param		id	path		int	true	"Product ID"
//	@Success	200	{object}	schemas.Response[schemas.Product]
//	@Failure	404	{object}	apperrors.Response
//	@Router		/api/products/{id} [get]
func (h *Controller) GetProduct(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	p, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, schemas.NewResponse(*modelToSchema(p)))
}

// ListProducts godoc
//
//	@Summary	List products
//	@Description	List products with optional pagination
//	@Tags		products
//	@Produce	json
//	@Param		limit	query		int	false	"Page size (max 100)"	default(20)
//	@Param		offset	query		int	false	"Skip count"
//	@Success	200	{object}	schemas.Response[[]schemas.Product]
//	@Router		/api/products [get]
func (h *Controller) ListProducts(c *echo.Context) error {
	var req schemas.ProductListRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.WrapCode(apperrors.CodeInvalid, err)
	}
	resp, err := h.service.List(c.Request().Context(), productservice.ListRequest{
		Limit: req.Limit, Offset: req.Offset,
	})
	if err != nil {
		return err
	}
	items := make([]schemas.Product, 0, len(resp.Items))
	for _, dto := range resp.Items {
		items = append(items, schemas.Product{
			ID: dto.ID, Name: dto.Name, Price: dto.Price,
			CreatedAt: dto.CreatedAt, UpdatedAt: dto.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, schemas.NewResponse(items))
}

// DeleteProduct godoc
//
//	@Summary	Delete product
//	@Description	Delete a product by ID
//	@Tags		products
//	@Param		id	path	int	true	"Product ID"
//	@Success	204
//	@Failure	404	{object}	apperrors.Response
//	@Router		/api/products/{id} [delete]
func (h *Controller) DeleteProduct(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// parseID extracts and validates the {id} path parameter as a positive uint.
func parseID(c *echo.Context) (uint, error) {
	raw := c.Param("id")
	id64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id64 == 0 {
		return 0, apperrors.NewCode(apperrors.CodeInvalid, "product id must be a positive integer")
	}
	return uint(id64), nil
}

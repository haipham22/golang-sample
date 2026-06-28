package product

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
	productMocks "github.com/haipham22/golang-sample/internal/mocks/product"
	"github.com/haipham22/golang-sample/internal/schemas"
	productservice "github.com/haipham22/golang-sample/internal/usecase/product"
	apiValidator "github.com/haipham22/golang-sample/internal/validator"

	"github.com/haipham22/golang-sample/internal/domain"
)

func newController(svc productservice.Service) *Controller { return &Controller{service: svc} }

// newJSONCtx builds an echo.Context with a JSON body and optional path params
// supplied as (name, value) pairs.
func newJSONCtx(method, path string, body any, params ...string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = apiValidator.NewCustomValidator()

	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	for i := 0; i+1 < len(params); i += 2 {
		c.SetParamNames(params[i])
		c.SetParamValues(params[i+1])
	}
	return c, rec
}

func TestController_PostProduct_Success(t *testing.T) {
	m := productMocks.NewMockService(t)
	m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r productservice.CreateRequest) bool {
		return r.Name == "Widget" && r.Price == 9.99
	})).RunAndReturn(func(ctx context.Context, r productservice.CreateRequest) (*domain.Product, error) {
		return &domain.Product{ID: 1, Name: r.Name, Price: r.Price}, nil
	})

	c, rec := newJSONCtx(http.MethodPost, "/api/products", schemas.ProductCreateRequest{Name: "Widget", Price: 9.99})
	require.NoError(t, newController(m).PostProduct(c))
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"name":"Widget"`)
}

func TestController_PostProduct_ValidationErrors(t *testing.T) {
	t.Run("invalid JSON maps to Invalid", func(t *testing.T) {
		m := productMocks.NewMockService(t) // no expectations -> service unused
		e := echo.New()
		e.Validator = apiValidator.NewCustomValidator()
		req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewReader([]byte("{bad")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := newController(m).PostProduct(c)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("empty name rejected by validator", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		c, _ := newJSONCtx(http.MethodPost, "/api/products", schemas.ProductCreateRequest{Name: "", Price: 1})

		err := newController(m).PostProduct(c)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("negative price rejected by validator", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		c, _ := newJSONCtx(http.MethodPost, "/api/products", schemas.ProductCreateRequest{Name: "X", Price: -1})

		err := newController(m).PostProduct(c)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("service error propagated", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		m.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, apperrors.NewCode(apperrors.CodeConflict, "dup"))
		c, _ := newJSONCtx(http.MethodPost, "/api/products", schemas.ProductCreateRequest{Name: "X", Price: 1})

		err := newController(m).PostProduct(c)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeConflict))
	})
}

func TestController_GetProduct(t *testing.T) {
	t.Run("bad id -> Invalid", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		c, _ := newJSONCtx(http.MethodGet, "/api/products/abc", nil, "id", "abc")
		err := newController(m).GetProduct(c)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("zero id -> Invalid", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		c, _ := newJSONCtx(http.MethodGet, "/api/products/0", nil, "id", "0")
		err := newController(m).GetProduct(c)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("found", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		m.EXPECT().GetByID(mock.Anything, uint(5)).Return(&domain.Product{ID: 5, Name: "Gadget"}, nil)
		c, rec := newJSONCtx(http.MethodGet, "/api/products/5", nil, "id", "5")
		require.NoError(t, newController(m).GetProduct(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"name":"Gadget"`)
	})

	t.Run("not found propagated", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		m.EXPECT().GetByID(mock.Anything, uint(9)).Return(nil, apperrors.NotFound("product 9"))
		c, _ := newJSONCtx(http.MethodGet, "/api/products/9", nil, "id", "9")
		err := newController(m).GetProduct(c)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeNotFound))
	})
}

func TestController_ListProducts(t *testing.T) {
	m := productMocks.NewMockService(t)
	m.EXPECT().List(mock.Anything, mock.Anything).Return(&productservice.ListResponse{
		Items: []*productservice.ProductDTO{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}},
		Total: 2,
	}, nil)
	c, rec := newJSONCtx(http.MethodGet, "/api/products?limit=2", nil)
	require.NoError(t, newController(m).ListProducts(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"name":"A"`)
	assert.Contains(t, rec.Body.String(), `"name":"B"`)
}

func TestController_DeleteProduct(t *testing.T) {
	t.Run("bad id -> Invalid", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		c, _ := newJSONCtx(http.MethodDelete, "/api/products/x", nil, "id", "x")
		err := newController(m).DeleteProduct(c)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("success -> 204", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		m.EXPECT().Delete(mock.Anything, uint(3)).Return(nil)
		c, rec := newJSONCtx(http.MethodDelete, "/api/products/3", nil, "id", "3")
		require.NoError(t, newController(m).DeleteProduct(c))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("not found propagated", func(t *testing.T) {
		m := productMocks.NewMockService(t)
		m.EXPECT().Delete(mock.Anything, uint(7)).Return(apperrors.NotFound("product 7"))
		c, _ := newJSONCtx(http.MethodDelete, "/api/products/7", nil, "id", "7")
		err := newController(m).DeleteProduct(c)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeNotFound))
	})
}

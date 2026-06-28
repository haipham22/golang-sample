package product

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/haipham22/golang-sample/internal/domain"
	apperrors "github.com/haipham22/golang-sample/internal/errors"
	repoMocks "github.com/haipham22/golang-sample/internal/mocks/repository"
	"github.com/haipham22/golang-sample/internal/repository/postgres"
)

func newTestService(t *testing.T, repo postgres.Repository) Service {
	t.Helper()
	return NewService(zap.NewNop().Sugar(), repo)
}

func TestService_Create_Success(t *testing.T) {
	t.Parallel()
	m := repoMocks.NewMockRepository(t)
	m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
		return p.Name == "Widget" && p.Price == 9.99
	})).RunAndReturn(func(ctx context.Context, p *domain.Product) (*domain.Product, error) {
		p.ID = 1
		return p, nil
	})

	svc := newTestService(t, m)
	got, err := svc.Create(context.Background(), CreateRequest{Name: "Widget", Price: 9.99})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, uint(1), got.ID)
	assert.Equal(t, "Widget", got.Name)
}

func TestService_Create_ValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty name rejected before repo", func(t *testing.T) {
		t.Parallel()
		m := repoMocks.NewMockRepository(t) // no expectations => repo must not be called
		svc := newTestService(t, m)

		_, err := svc.Create(context.Background(), CreateRequest{Name: "", Price: 1})
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("negative price rejected before repo", func(t *testing.T) {
		t.Parallel()
		m := repoMocks.NewMockRepository(t)
		svc := newTestService(t, m)

		_, err := svc.Create(context.Background(), CreateRequest{Name: "X", Price: -1})
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("repo error is propagated", func(t *testing.T) {
		t.Parallel()
		m := repoMocks.NewMockRepository(t)
		m.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Product")).Return(nil, apperrors.NewCode(apperrors.CodeInternal, "boom"))
		svc := newTestService(t, m)

		_, err := svc.Create(context.Background(), CreateRequest{Name: "X", Price: 1})
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInternal))
	})
}

func TestService_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("zero id rejected", func(t *testing.T) {
		t.Parallel()
		m := repoMocks.NewMockRepository(t)
		svc := newTestService(t, m)

		_, err := svc.GetByID(context.Background(), 0)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		m := repoMocks.NewMockRepository(t)
		m.EXPECT().FindByID(mock.Anything, uint(5)).Return(&domain.Product{ID: 5, Name: "N"}, nil)
		svc := newTestService(t, m)

		got, err := svc.GetByID(context.Background(), 5)
		require.NoError(t, err)
		assert.Equal(t, uint(5), got.ID)
	})

	t.Run("not found propagated", func(t *testing.T) {
		t.Parallel()
		m := repoMocks.NewMockRepository(t)
		m.EXPECT().FindByID(mock.Anything, uint(7)).Return(nil, apperrors.NotFound("product 7"))
		svc := newTestService(t, m)

		_, err := svc.GetByID(context.Background(), 7)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeNotFound))
	})
}

func TestService_List(t *testing.T) {
	t.Parallel()
	m := repoMocks.NewMockRepository(t)
	m.EXPECT().List(mock.Anything, postgres.ListParams{Limit: 2, Offset: 0}).Return(
		[]*domain.Product{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}, int64(5), nil,
	)
	svc := newTestService(t, m)

	resp, err := svc.List(context.Background(), ListRequest{Limit: 2})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(5), resp.Total)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, "A", resp.Items[0].Name)
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	t.Run("zero id rejected", func(t *testing.T) {
		t.Parallel()
		m := repoMocks.NewMockRepository(t)
		svc := newTestService(t, m)

		err := svc.Delete(context.Background(), 0)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		m := repoMocks.NewMockRepository(t)
		m.EXPECT().Delete(mock.Anything, uint(3)).Return(nil)
		svc := newTestService(t, m)

		require.NoError(t, svc.Delete(context.Background(), 3))
	})

	t.Run("not found propagated", func(t *testing.T) {
		t.Parallel()
		m := repoMocks.NewMockRepository(t)
		m.EXPECT().Delete(mock.Anything, uint(9)).Return(apperrors.NotFound("product 9"))
		svc := newTestService(t, m)

		err := svc.Delete(context.Background(), 9)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeNotFound))
	})
}

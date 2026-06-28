package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/haipham22/golang-sample/internal/domain"
	apperrors "github.com/haipham22/golang-sample/internal/errors"
)

// fakeRepo is a hand-written test double for the user Repository. We use a
// fake (not a mockery mock) because the Repository interface is defined in
// this same package, and a generated mock would create an import cycle.
type fakeRepo struct {
	byID      *domain.User
	byIDErr   error
	listItems []*domain.User
	listTotal int64
	listErr   error
}

func (f *fakeRepo) FindUserByID(ctx context.Context, id uint) (*domain.User, error) {
	return f.byID, f.byIDErr
}

func (f *fakeRepo) ListUsers(ctx context.Context, params ListParams) ([]*domain.User, int64, error) {
	return f.listItems, f.listTotal, f.listErr
}

func newTestService(t *testing.T, repo Repository) Service {
	t.Helper()
	return NewService(zap.NewNop().Sugar(), repo)
}

func TestService_GetByID(t *testing.T) {
	t.Run("zero id rejected", func(t *testing.T) {
		svc := newTestService(t, &fakeRepo{})
		_, err := svc.GetByID(context.Background(), 0)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
	})

	t.Run("found", func(t *testing.T) {
		svc := newTestService(t, &fakeRepo{byID: &domain.User{ID: 5, Username: "bob"}})
		got, err := svc.GetByID(context.Background(), 5)
		require.NoError(t, err)
		assert.Equal(t, "bob", got.Username)
	})

	t.Run("absent -> NotFound", func(t *testing.T) {
		svc := newTestService(t, &fakeRepo{byID: nil})
		_, err := svc.GetByID(context.Background(), 7)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeNotFound))
	})

	t.Run("repo error propagated", func(t *testing.T) {
		svc := newTestService(t, &fakeRepo{byIDErr: apperrors.NewCode(apperrors.CodeInternal, "boom")})
		_, err := svc.GetByID(context.Background(), 9)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeInternal))
	})
}

func TestService_List(t *testing.T) {
	svc := newTestService(t, &fakeRepo{
		listItems: []*domain.User{{ID: 1, Username: "a"}, {ID: 2, Username: "b"}},
		listTotal: 2,
	})

	resp, err := svc.List(context.Background(), ListParams{Limit: 2})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(2), resp.Total)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, "a", resp.Items[0].Username)
}

func TestService_List_RepoError(t *testing.T) {
	svc := newTestService(t, &fakeRepo{listErr: apperrors.NewCode(apperrors.CodeInternal, "db down")})
	_, err := svc.List(context.Background(), ListParams{})
	require.Error(t, err)
	assert.True(t, apperrors.IsCode(err, apperrors.CodeInternal))
}

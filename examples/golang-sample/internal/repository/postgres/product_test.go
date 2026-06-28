package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/haipham22/golang-sample/internal/domain"
	apperrors "github.com/haipham22/golang-sample/internal/errors"
	"github.com/haipham22/golang-sample/internal/orm"
)

// openTestDB creates an isolated in-memory SQLite database for product tests.
// Mirrors internal/repository/user/user_test.go setup.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	dbSQL, err := db.DB()
	require.NoError(t, err)
	dbSQL.SetMaxOpenConns(1)
	dbSQL.SetMaxIdleConns(1)

	t.Cleanup(func() {
		if err := dbSQL.Close(); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
	})

	require.NoError(t, db.AutoMigrate(&orm.Product{}), "failed to migrate products")
	return db
}

func TestRepository_InterfaceCompliance(t *testing.T) {
	var _ Repository = (*repo)(nil)
}

func TestNew(t *testing.T) {
	db := openTestDB(t)
	r := New(zap.NewNop().Sugar(), db)
	assert.NotNil(t, r)
	assert.IsType(t, &repo{}, r)
}

func TestRepo_Create_Success(t *testing.T) {
	db := openTestDB(t)
	r := New(zap.NewNop().Sugar(), db)
	ctx := context.Background()

	created, err := r.Create(ctx, &domain.Product{Name: "Widget", Price: 9.99})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "Widget", created.Name)
	assert.Equal(t, 9.99, created.Price)
	assert.False(t, created.CreatedAt.IsZero())
}

func TestRepo_Create_NilInput(t *testing.T) {
	db := openTestDB(t)
	r := New(zap.NewNop().Sugar(), db)

	_, err := r.Create(context.Background(), nil)
	require.Error(t, err)
	assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalid))
}

func TestRepo_FindByID(t *testing.T) {
	db := openTestDB(t)
	r := New(zap.NewNop().Sugar(), db)
	ctx := context.Background()

	created, err := r.Create(ctx, &domain.Product{Name: "Gadget", Price: 4.5})
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		got, err := r.FindByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "Gadget", got.Name)
	})

	t.Run("not found returns typed error", func(t *testing.T) {
		_, err := r.FindByID(ctx, 9999)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeNotFound))
	})
}

func TestRepo_List(t *testing.T) {
	db := openTestDB(t)
	r := New(zap.NewNop().Sugar(), db)
	ctx := context.Background()

	for i := range 5 {
		_, err := r.Create(ctx, &domain.Product{Name: fmt.Sprintf("P%d", i), Price: float64(i)})
		require.NoError(t, err)
	}

	t.Run("returns all with default limit", func(t *testing.T) {
		items, total, err := r.List(ctx, ListParams{})
		require.NoError(t, err)
		assert.Len(t, items, 5)
		assert.Equal(t, int64(5), total)
	})

	t.Run("respects limit", func(t *testing.T) {
		items, total, err := r.List(ctx, ListParams{Limit: 2})
		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, int64(5), total)
	})

	t.Run("respects offset", func(t *testing.T) {
		items, _, err := r.List(ctx, ListParams{Limit: 2, Offset: 2})
		require.NoError(t, err)
		require.Len(t, items, 2)
		// IDs are 1-based ascending; offset 2 skips IDs 1 and 2.
		assert.Equal(t, uint(3), items[0].ID)
	})

	t.Run("empty table", func(t *testing.T) {
		dbEmpty := openTestDB(t)
		rEmpty := New(zap.NewNop().Sugar(), dbEmpty)
		items, total, err := rEmpty.List(ctx, ListParams{})
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, int64(0), total)
	})
}

func TestRepo_Delete(t *testing.T) {
	db := openTestDB(t)
	r := New(zap.NewNop().Sugar(), db)
	ctx := context.Background()

	created, err := r.Create(ctx, &domain.Product{Name: "Doomed", Price: 1.0})
	require.NoError(t, err)

	t.Run("delete existing", func(t *testing.T) {
		err := r.Delete(ctx, created.ID)
		require.NoError(t, err)

		_, err = r.FindByID(ctx, created.ID)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeNotFound))
	})

	t.Run("delete missing returns not found", func(t *testing.T) {
		err := r.Delete(ctx, 8888)
		require.Error(t, err)
		assert.True(t, apperrors.IsCode(err, apperrors.CodeNotFound))
	})
}

func TestConverter_NilSafety(t *testing.T) {
	assert.Nil(t, ormToProduct(nil))
	assert.Nil(t, productToORM(nil))
	assert.Nil(t, ormSliceToProducts(nil))
}

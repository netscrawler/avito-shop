package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMerchByName(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewMerchRepository(mock)
	ctx := context.Background()
	merchName := "test-item"

	t.Run("успешное получение товара", func(t *testing.T) {
		mock.ExpectQuery("SELECT name, price FROM merch WHERE name = \\$1").
			WithArgs(merchName).
			WillReturnRows(pgxmock.NewRows([]string{"name", "price"}).
				AddRow(merchName, uint64(100)))

		merch, err := repo.GetMerchByName(ctx, merchName)
		assert.NoError(t, err)
		assert.NotNil(t, merch)
		assert.Equal(t, merchName, merch.Name)
		assert.Equal(t, uint64(100), merch.Price)
	})

	t.Run("товар не найден", func(t *testing.T) {
		mock.ExpectQuery("SELECT name, price FROM merch WHERE name = \\$1").
			WithArgs(merchName).
			WillReturnError(pgx.ErrNoRows)

		merch, err := repo.GetMerchByName(ctx, merchName)
		assert.Error(t, err)
		assert.Nil(t, merch)
		assert.ErrorIs(t, err, domain.ErrMerchNotFound)
	})
}

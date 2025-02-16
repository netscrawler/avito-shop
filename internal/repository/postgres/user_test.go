package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestGetUserInfo(t *testing.T) {
	t.Run("успешное получение информации о пользователе", func(t *testing.T) {
		// Подготовка
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepository(mock)
		username := "testuser"

		// Ожидаемые данные
		expectedUser := &domain.User{
			Username: username,
			Password: []byte("hashedpassword"),
			Coins:    1000,
		}

		// Настройка ожиданий
		mock.ExpectQuery("SELECT username, password, coins FROM users WHERE username = \\$1").
			WithArgs(username).
			WillReturnRows(pgxmock.NewRows([]string{"username", "password", "coins"}).
				AddRow(username, []byte("hashedpassword"), uint64(1000)))

		// Добавляем ожидание для запроса инвентаря
		mock.ExpectQuery("SELECT item_name, quantity FROM user_inventory WHERE username = \\$1").
			WithArgs(username).
			WillReturnRows(pgxmock.NewRows([]string{"item_name", "quantity"}).
				AddRow("item1", 1).
				AddRow("item2", 2))

		// Действие
		user, err := repo.GetUserInfo(context.Background(), username)

		// Проверка
		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, expectedUser.Username, user.Username)
		require.Equal(t, expectedUser.Password, user.Password)
		require.Equal(t, expectedUser.Coins, user.Coins)
		require.Len(t, user.Inventory, 2)
		require.Equal(t, "item1", user.Inventory[0].Type)
		require.Equal(t, 1, user.Inventory[0].Quantity)
		require.Equal(t, "item2", user.Inventory[1].Type)
		require.Equal(t, 2, user.Inventory[1].Quantity)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		// Подготовка
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepository(mock)
		username := "nonexistent"

		// Настройка ожиданий
		mock.ExpectQuery("SELECT username, password, coins FROM users WHERE username = \\$1").
			WithArgs(username).
			WillReturnRows(pgxmock.NewRows([]string{"username", "password", "coins"}))

		// Действие
		user, err := repo.GetUserInfo(context.Background(), username)

		// Проверка
		require.ErrorIs(t, err, domain.ErrUserNotFound)
		require.Nil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateUserInventory(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepository(mock)
	username := "testuser"
	itemName := "test-item"
	quantity := 1

	t.Run("добавление нового предмета", func(t *testing.T) {
		// Начинаем транзакцию
		mock.ExpectBegin()

		// Проверяем текущее количество
		mock.ExpectQuery("SELECT quantity FROM user_inventory").
			WithArgs(username, itemName).
			WillReturnError(pgx.ErrNoRows)

		// Вставляем новую запись
		mock.ExpectExec("INSERT INTO user_inventory").
			WithArgs(username, itemName, quantity).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		err = repo.UpdateUserInventory(context.Background(), &domain.User{Username: username}, itemName, quantity)
		require.NoError(t, err)
	})

	t.Run("обновление существующего предмета", func(t *testing.T) {
		currentQuantity := 5

		// Начинаем транзакцию
		mock.ExpectBegin()

		// Проверяем текущее количество
		mock.ExpectQuery("SELECT quantity FROM user_inventory").
			WithArgs(username, itemName).
			WillReturnRows(pgxmock.NewRows([]string{"quantity"}).AddRow(currentQuantity))

		// Обновляем существующую запись
		mock.ExpectExec("UPDATE user_inventory").
			WithArgs(currentQuantity+quantity, username, itemName).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectCommit()

		err = repo.UpdateUserInventory(context.Background(), &domain.User{Username: username}, itemName, quantity)
		require.NoError(t, err)
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("успешное обновление пользователя", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepository(mock)
		user := &domain.User{
			Username: "testuser",
			Coins:    1000,
		}

		mock.ExpectExec("UPDATE users SET coins = \\$1 WHERE username = \\$2").
			WithArgs(user.Coins, user.Username).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.UpdateUser(context.Background(), user)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка обновления", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepository(mock)
		user := &domain.User{
			Username: "testuser",
			Coins:    1000,
		}

		expectedError := errors.New("ошибка обновления")
		mock.ExpectExec("UPDATE users SET coins = \\$1 WHERE username = \\$2").
			WithArgs(user.Coins, user.Username).
			WillReturnError(expectedError)

		err = repo.UpdateUser(context.Background(), user)
		require.Error(t, err)
		require.Contains(t, err.Error(), expectedError.Error())
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

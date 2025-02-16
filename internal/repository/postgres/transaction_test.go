package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTransaction(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTransactionRepository(mock)
	ctx := context.Background()
	now := time.Now()

	transaction := &domain.Transaction{
		SenderName:   "sender",
		ReceiverName: "receiver",
		Amount:       100,
		Type:         domain.TransactionTypeTransfer,
		Timestamp:    now,
	}

	t.Run("успешное создание транзакции", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO transactions").
			WithArgs(transaction.SenderName, transaction.ReceiverName, transaction.Amount,
				transaction.Type, transaction.Timestamp).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.CreateTransaction(ctx, transaction)
		assert.NoError(t, err)
	})

	t.Run("ошибка создания транзакции", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO transactions").
			WithArgs(transaction.SenderName, transaction.ReceiverName, transaction.Amount,
				transaction.Type, transaction.Timestamp).
			WillReturnError(pgx.ErrTxClosed)

		err := repo.CreateTransaction(ctx, transaction)
		assert.Error(t, err)
	})
}

func TestGetUserTransactions(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTransactionRepository(mock)
	ctx := context.Background()
	username := "testuser"
	now := time.Now()

	t.Run("успешное получение транзакций", func(t *testing.T) {
		mock.ExpectQuery("SELECT sender_name, receiver_name, amount, transfer_type, timestamp FROM transactions WHERE \\(sender_name = \\$1 OR receiver_name = \\$1\\) AND transfer_type = \\$2 ORDER BY timestamp DESC").
			WithArgs(username, domain.TransactionTypeTransfer).
			WillReturnRows(pgxmock.NewRows([]string{"sender_name", "receiver_name", "amount", "transfer_type", "timestamp"}).
				AddRow(username, "receiver1", uint64(100), domain.TransactionTypeTransfer, now).
				AddRow("sender2", username, uint64(200), domain.TransactionTypeTransfer, now))

		transactions, err := repo.GetUserTransactions(ctx, username)
		assert.NoError(t, err)
		assert.Len(t, transactions, 2)
		assert.Equal(t, username, transactions[0].SenderName)
		assert.Equal(t, username, transactions[1].ReceiverName)
	})

	t.Run("пустой список транзакций", func(t *testing.T) {
		mock.ExpectQuery("SELECT sender_name, receiver_name, amount, transfer_type, timestamp FROM transactions WHERE \\(sender_name = \\$1 OR receiver_name = \\$1\\) AND transfer_type = \\$2 ORDER BY timestamp DESC").
			WithArgs(username, domain.TransactionTypeTransfer).
			WillReturnRows(pgxmock.NewRows([]string{"sender_name", "receiver_name", "amount", "transfer_type", "timestamp"}))

		transactions, err := repo.GetUserTransactions(ctx, username)
		assert.NoError(t, err)
		assert.Empty(t, transactions)
	})
}

func TestExecuteTransfer(t *testing.T) {
	t.Run("успешный перевод", func(t *testing.T) {
		// Arrange
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewTransactionRepository(mock)
		ctx := context.Background()

		sender := "sender"
		receiver := "receiver"
		amount := uint64(100)

		// Начало транзакции
		mock.ExpectBegin()

		// Получение баланса отправителя
		mock.ExpectQuery("SELECT coins FROM users WHERE username = \\$1 FOR UPDATE").
			WithArgs(sender).
			WillReturnRows(pgxmock.NewRows([]string{"coins"}).AddRow(uint64(1000)))

		// Получение баланса получателя
		mock.ExpectQuery("SELECT coins FROM users WHERE username = \\$1 FOR UPDATE").
			WithArgs(receiver).
			WillReturnRows(pgxmock.NewRows([]string{"coins"}).AddRow(uint64(500)))

		// Обновление баланса отправителя
		mock.ExpectExec("UPDATE users SET coins = coins - \\$1 WHERE username = \\$2").
			WithArgs(amount, sender).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		// Обновление баланса получателя
		mock.ExpectExec("UPDATE users SET coins = coins \\+ \\$1 WHERE username = \\$2").
			WithArgs(amount, receiver).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		// Создание записи о транзакции
		mock.ExpectExec("INSERT INTO transactions \\(sender_name, receiver_name, amount, transfer_type, timestamp\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\)").
			WithArgs(sender, receiver, amount, domain.TransactionTypeTransfer, pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		// Подтверждение транзакции
		mock.ExpectCommit()

		// Act
		err = repo.ExecuteTransfer(ctx, sender, receiver, amount)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("недостаточно средств", func(t *testing.T) {
		// Arrange
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewTransactionRepository(mock)
		ctx := context.Background()

		sender := "sender"
		receiver := "receiver"
		amount := uint64(1500)

		// Начало транзакции
		mock.ExpectBegin()

		// Получение баланса отправителя
		mock.ExpectQuery("SELECT coins FROM users WHERE username = \\$1 FOR UPDATE").
			WithArgs(sender).
			WillReturnRows(pgxmock.NewRows([]string{"coins"}).AddRow(uint64(1000)))

		// Получение баланса получателя
		mock.ExpectQuery("SELECT coins FROM users WHERE username = \\$1 FOR UPDATE").
			WithArgs(receiver).
			WillReturnRows(pgxmock.NewRows([]string{"coins"}).AddRow(uint64(500)))

		// Ожидаем откат транзакции
		mock.ExpectRollback()

		// Act
		err = repo.ExecuteTransfer(ctx, sender, receiver, amount)

		// Assert
		require.ErrorIs(t, err, domain.ErrInsufficientFunds)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestExecutePurchase(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTransactionRepository(mock)
	ctx := context.Background()

	t.Run("успешная покупка", func(t *testing.T) {
		username := "buyer"
		merchName := "item1"
		price := uint64(100)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT coins FROM users WHERE username = \\$1 FOR UPDATE").
			WithArgs(username).
			WillReturnRows(pgxmock.NewRows([]string{"coins"}).AddRow(uint64(1000)))
		mock.ExpectExec("UPDATE users SET coins = coins - \\$1 WHERE username = \\$2").
			WithArgs(price, username).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectExec("INSERT INTO user_inventory \\(username, item_name, quantity\\) VALUES \\(\\$1, \\$2, 1\\) ON CONFLICT \\(username, item_name\\) DO UPDATE SET quantity = user_inventory.quantity \\+ 1").
			WithArgs(username, merchName).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mock.ExpectExec("INSERT INTO transactions").
			WithArgs(username, "SHOP", price, domain.TransactionTypePurchase, pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mock.ExpectCommit()

		err := repo.ExecutePurchase(ctx, username, merchName, price)
		assert.NoError(t, err)
	})

	t.Run("ошибка покупки", func(t *testing.T) {
		username := "buyer"
		merchName := "item1"
		price := uint64(100)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT coins FROM users WHERE username = \\$1 FOR UPDATE").
			WithArgs(username).
			WillReturnRows(pgxmock.NewRows([]string{"coins"}).AddRow(uint64(1000)))
		mock.ExpectExec("UPDATE users SET coins = coins - \\$1 WHERE username = \\$2").
			WithArgs(price, username).
			WillReturnError(pgx.ErrTxClosed)
		mock.ExpectRollback()

		err := repo.ExecutePurchase(ctx, username, merchName, price)
		assert.Error(t, err)
	})
}

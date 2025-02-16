package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/netscrawler/avito-shop/internal/repository"
)

// transaction реализует интерфейс TransactionRepository для работы с транзакциями в PostgreSQL
type transaction struct {
	db DBPool
}

// NewTransactionRepository создает новый экземпляр репозитория транзакций
func NewTransactionRepository(db DBPool) repository.TransactionRepository {
	return &transaction{db: db}
}

// CreateTransaction создает новую транзакцию в базе данных
func (t *transaction) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	const op = "TransactionRepository.CreateTransaction"

	_, err := t.db.Exec(ctx, `
		INSERT INTO transactions (sender_name, receiver_name, amount, transfer_type, timestamp)
		VALUES ($1, $2, $3, $4, $5)`,
		transaction.SenderName, transaction.ReceiverName, transaction.Amount,
		transaction.Type, transaction.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetTransactionsBySender возвращает список транзакций по отправителю
func (t *transaction) GetTransactionsBySender(ctx context.Context, senderName, transferType string) ([]*domain.Transaction, error) {
	const op = "TransactionRepository.GetTransactionsBySender"

	rows, err := t.db.Query(ctx, `
		SELECT sender_name, receiver_name, amount, timestamp
		FROM transactions
		WHERE sender_name = $1 AND transfer_type = $2
		ORDER BY timestamp DESC`,
		senderName, transferType,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		transaction := &domain.Transaction{}
		if err := rows.Scan(
			&transaction.SenderName,
			&transaction.ReceiverName,
			&transaction.Amount,
			&transaction.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		transaction.Type = domain.TransactionType(transferType)
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return transactions, nil
}

// GetTransactionsByReceiver возвращает список транзакций по получателю
func (t *transaction) GetTransactionsByReceiver(ctx context.Context, receiverName, transferType string) ([]*domain.Transaction, error) {
	const op = "TransactionRepository.GetTransactionsByReceiver"

	rows, err := t.db.Query(ctx, `
		SELECT sender_name, receiver_name, amount, timestamp
		FROM transactions
		WHERE receiver_name = $1 AND transfer_type = $2
		ORDER BY timestamp DESC`,
		receiverName, transferType,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		transaction := &domain.Transaction{}
		if err := rows.Scan(
			&transaction.SenderName,
			&transaction.ReceiverName,
			&transaction.Amount,
			&transaction.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		transaction.Type = domain.TransactionType(transferType)
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return transactions, nil
}

// ExecuteTransfer выполняет перевод монет между пользователями в рамках одной транзакции
func (t *transaction) ExecuteTransfer(ctx context.Context, fromUsername, toUsername string, amount uint64) error {
	const op = "TransactionRepository.ExecuteTransfer"

	tx, err := t.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: начало транзакции: %w", op, err)
	}

	var committed bool
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("%v, rollback error: %v", err, rollbackErr)
			}
		}
	}()

	// Получаем баланс отправителя
	var senderCoins uint64
	err = tx.QueryRow(ctx,
		"SELECT coins FROM users WHERE username = $1 FOR UPDATE",
		fromUsername,
	).Scan(&senderCoins)
	if err != nil {
		return fmt.Errorf("%s: получение данных первого пользователя: %w", op, err)
	}

	// Получаем баланс получателя
	var receiverCoins uint64
	err = tx.QueryRow(ctx,
		"SELECT coins FROM users WHERE username = $1 FOR UPDATE",
		toUsername,
	).Scan(&receiverCoins)
	if err != nil {
		return fmt.Errorf("%s: получение данных второго пользователя: %w", op, err)
	}

	// Проверяем достаточность средств
	if senderCoins < amount {
		return domain.ErrInsufficientFunds
	}

	// Обновляем баланс отправителя
	_, err = tx.Exec(ctx,
		"UPDATE users SET coins = coins - $1 WHERE username = $2",
		amount, fromUsername,
	)
	if err != nil {
		return fmt.Errorf("%s: обновление баланса отправителя: %w", op, err)
	}

	// Обновляем баланс получателя
	_, err = tx.Exec(ctx,
		"UPDATE users SET coins = coins + $1 WHERE username = $2",
		amount, toUsername,
	)
	if err != nil {
		return fmt.Errorf("%s: обновление баланса получателя: %w", op, err)
	}

	// Создаем запись о транзакции
	_, err = tx.Exec(ctx,
		"INSERT INTO transactions (sender_name, receiver_name, amount, transfer_type, timestamp) VALUES ($1, $2, $3, $4, $5)",
		fromUsername, toUsername, amount, domain.TransactionTypeTransfer, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("%s: создание записи о транзакции: %w", op, err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: фиксация транзакции: %w", op, err)
	}
	committed = true

	return nil
}

// GetUserTransactions возвращает все транзакции пользователя
func (t *transaction) GetUserTransactions(ctx context.Context, username string) ([]*domain.Transaction, error) {
	const op = "TransactionRepository.GetUserTransactions"

	rows, err := t.db.Query(ctx, `
		SELECT sender_name, receiver_name, amount, transfer_type, timestamp
		FROM transactions
		WHERE (sender_name = $1 OR receiver_name = $1)
		AND transfer_type = $2
		ORDER BY timestamp DESC`,
		username, domain.TransactionTypeTransfer,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		trx := &domain.Transaction{}
		if err := rows.Scan(
			&trx.SenderName,
			&trx.ReceiverName,
			&trx.Amount,
			&trx.Type,
			&trx.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("%s: сканирование строки: %w", op, err)
		}
		transactions = append(transactions, trx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: итерация по результатам: %w", op, err)
	}

	return transactions, nil
}

// ExecutePurchase выполняет покупку товара в рамках одной транзакции
func (t *transaction) ExecutePurchase(ctx context.Context, username string, merchName string, price uint64) error {
	const op = "TransactionRepository.ExecutePurchase"

	tx, err := t.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: начало транзакции: %w", op, err)
	}

	var committed bool
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("%v, rollback error: %v", err, rollbackErr)
			}
		}
	}()

	// Блокируем строку пользователя для обновления
	var coins uint64
	err = tx.QueryRow(ctx,
		"SELECT coins FROM users WHERE username = $1 FOR UPDATE",
		username,
	).Scan(&coins)
	if err != nil {
		return fmt.Errorf("%s: получение данных пользователя: %w", op, err)
	}

	// Проверяем достаточность средств
	if coins < price {
		return domain.ErrInsufficientFunds
	}

	// Обновляем баланс
	_, err = tx.Exec(ctx,
		"UPDATE users SET coins = coins - $1 WHERE username = $2",
		price, username,
	)
	if err != nil {
		return fmt.Errorf("%s: обновление баланса: %w", op, err)
	}

	// Обновляем или создаем запись в инвентаре
	_, err = tx.Exec(ctx, `
		INSERT INTO user_inventory (username, item_name, quantity)
		VALUES ($1, $2, 1)
		ON CONFLICT (username, item_name)
		DO UPDATE SET quantity = user_inventory.quantity + 1`,
		username, merchName,
	)
	if err != nil {
		return fmt.Errorf("%s: обновление инвентаря: %w", op, err)
	}

	// Создаем запись о транзакции
	_, err = tx.Exec(ctx,
		"INSERT INTO transactions (sender_name, receiver_name, amount, transfer_type, timestamp) VALUES ($1, $2, $3, $4, $5)",
		username, "SHOP", price, domain.TransactionTypePurchase, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("%s: создание записи о транзакции: %w", op, err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: фиксация транзакции: %w", op, err)
	}
	committed = true

	return nil
}

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/netscrawler/avito-shop/internal/repository"
)

// user реализует интерфейс UserRepository для работы с пользователями в PostgreSQL
type user struct {
	db DBPool
}

// NewUserRepository создает новый экземпляр репозитория пользователей
func NewUserRepository(db DBPool) repository.UserRepository {
	return &user{db: db}
}

// CreateUser создает нового пользователя в базе данных
func (u *user) CreateUser(ctx context.Context, user *domain.User) error {
	const op = "UserRepository.CreateUser"

	tx, err := u.db.Begin(ctx)
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

	// Проверяем существование пользователя с блокировкой
	var exists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1) FOR UPDATE",
		user.Username,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: проверка существования пользователя: %w", op, err)
	}

	if exists {
		return domain.ErrUserAlreadyExists
	}

	// Создаем пользователя
	_, err = tx.Exec(ctx,
		"INSERT INTO users (username, password, coins) VALUES ($1, $2, $3)",
		user.Username, user.Password, user.Coins,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // unique_violation
				return domain.ErrUserAlreadyExists
			}
			return fmt.Errorf("%s: ошибка PostgreSQL (код: %s): %w", op, pgErr.Code, err)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: фиксация транзакции: %w", op, err)
	}
	committed = true

	return nil
}

// UpdateUserInventory обновляет инвентарь пользователя
func (u *user) UpdateUserInventory(ctx context.Context, user *domain.User, item string, quantity int) error {
	const op = "UserRepository.UpdateUserInventory"

	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: начало транзакции: %w", op, err)
	}

	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	var currentQuantity int
	err = tx.QueryRow(ctx,
		"SELECT quantity FROM user_inventory WHERE username = $1 AND item_name = $2",
		user.Username, item,
	).Scan(&currentQuantity)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Если записи нет, создаем новую
			_, err = tx.Exec(ctx,
				"INSERT INTO user_inventory (username, item_name, quantity) VALUES ($1, $2, $3)",
				user.Username, item, quantity,
			)
		} else {
			return fmt.Errorf("%s: ошибка получения текущего количества: %w", op, err)
		}
	} else {
		// Если запись существует, обновляем её
		_, err = tx.Exec(ctx,
			"UPDATE user_inventory SET quantity = $1 WHERE username = $2 AND item_name = $3",
			currentQuantity+quantity, user.Username, item,
		)
	}

	if err != nil {
		return fmt.Errorf("%s: обновление инвентаря: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: подтверждение транзакции: %w", op, err)
	}
	committed = true

	return nil
}

// GetUserByUsername возвращает пользователя по его имени
func (u *user) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	const op = "UserRepository.GetUserByUsername"

	user := &domain.User{}
	err := u.db.QueryRow(ctx,
		"SELECT username, password, coins FROM users WHERE username = $1",
		username,
	).Scan(&user.Username, &user.Password, &user.Coins)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		if pgErr, ok := err.(*pgconn.PgError); ok {
			return nil, fmt.Errorf("%s: ошибка PostgreSQL (код: %s): %w", op, pgErr.Code, err)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// GetUserInfo возвращает полную информацию о пользователе, включая инвентарь
func (u *user) GetUserInfo(ctx context.Context, username string) (*domain.User, error) {
	const op = "UserRepository.GetUserInfo"

	user, err := u.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Инициализируем пустой инвентарь
	user.Inventory = make([]domain.UserInventory, 0)

	rows, err := u.db.Query(ctx,
		"SELECT item_name, quantity FROM user_inventory WHERE username = $1",
		username,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Если у пользователя нет инвентаря, возвращаем пустой список
			return user, nil
		}
		return nil, fmt.Errorf("%s: получение инвентаря: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.UserInventory
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			return nil, fmt.Errorf("%s: сканирование строки инвентаря: %w", op, err)
		}
		user.Inventory = append(user.Inventory, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: итерация по инвентарю: %w", op, err)
	}

	return user, nil
}

// UpdateUser обновляет информацию о пользователе
func (u *user) UpdateUser(ctx context.Context, user *domain.User) error {
	const op = "UserRepository.UpdateUser"

	_, err := u.db.Exec(ctx,
		"UPDATE users SET coins = $1 WHERE username = $2",
		user.Coins, user.Username,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

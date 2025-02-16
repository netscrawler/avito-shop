package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/netscrawler/avito-shop/internal/repository"
)

// merch реализует интерфейс MerchRepository для работы с товарами в PostgreSQL
type merch struct {
	db DBPool
}

// NewMerchRepository создает новый экземпляр репозитория товаров
func NewMerchRepository(db DBPool) repository.MerchRepository {
	return &merch{db: db}
}

// GetMerchByName возвращает товар по его названию
func (m *merch) GetMerchByName(ctx context.Context, name string) (*domain.Merch, error) {
	const op = "MerchRepository.GetMerchByName"

	row := m.db.QueryRow(ctx, "SELECT name, price FROM merch WHERE name = $1", name)

	merch := &domain.Merch{}
	if err := row.Scan(&merch.Name, &merch.Price); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrMerchNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return merch, nil
}

// GetMerchById возвращает товар по его идентификатору
func (m *merch) GetMerchById(ctx context.Context, id int) (*domain.Merch, error) {
	const op = "MerchRepository.GetMerchById"

	row := m.db.QueryRow(ctx, "SELECT name, price FROM merch WHERE id = $1", id)

	merch := &domain.Merch{}
	if err := row.Scan(&merch.Name, &merch.Price); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrMerchNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return merch, nil
}

// GetAllMerch возвращает список всех доступных товаров
func (m *merch) GetAllMerch(ctx context.Context) ([]*domain.Merch, error) {
	const op = "MerchRepository.GetAllMerch"

	rows, err := m.db.Query(ctx, "SELECT name, price FROM merch ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var items []*domain.Merch
	for rows.Next() {
		item := &domain.Merch{}
		if err := rows.Scan(&item.Name, &item.Price); err != nil {
			return nil, fmt.Errorf("%s: сканирование строки: %w", op, err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: итерация по результатам: %w", op, err)
	}

	return items, nil
}

// Предоставляет интерфейс для работы с пулом соединений базы данных
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBPool определяет интерфейс для работы с пулом соединений базы данных
type DBPool interface {
	Begin(context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// PoolAdapter адаптирует *pgxpool.Pool к интерфейсу DBPool
type PoolAdapter struct {
	*pgxpool.Pool
}

// NewPoolAdapter создает новый адаптер для пула соединений
func NewPoolAdapter(pool *pgxpool.Pool) DBPool {
	return &PoolAdapter{pool}
}

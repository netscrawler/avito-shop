package repository

import (
	"context"

	"github.com/netscrawler/avito-shop/internal/domain"
)

// UserRepository определяет методы для работы с пользователями
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserInfo(ctx context.Context, username string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	UpdateUserInventory(ctx context.Context, user *domain.User, item string, quantity int) error
}

// TransactionRepository определяет методы для работы с транзакциями
type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction *domain.Transaction) error
	GetUserTransactions(ctx context.Context, username string) ([]*domain.Transaction, error)
	ExecuteTransfer(ctx context.Context, fromUsername, toUsername string, amount uint64) error
	ExecutePurchase(ctx context.Context, username string, merchName string, price uint64) error
}

// MerchRepository определяет методы для работы с товарами
type MerchRepository interface {
	GetMerchByName(ctx context.Context, name string) (*domain.Merch, error)
	GetAllMerch(ctx context.Context) ([]*domain.Merch, error)
}

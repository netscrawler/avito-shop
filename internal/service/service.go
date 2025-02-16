package service

import (
	"context"

	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/netscrawler/avito-shop/internal/model"
)

const (
	Purchase = "PURCHASE"
	Transfer = "TRANSFER"
)

type UserService interface {
	RegisterUser(ctx context.Context, username, password string) error
	AuthenticateUser(ctx context.Context, username, password string) (string, error)
	GetUserInfo(ctx context.Context, username string) (*domain.User, error)
}

type TransferService interface {
	SendCoins(ctx context.Context, sender, receiver string, amount uint64) error
	GetTransactionHistory(ctx context.Context, username string) (model.CoinHistory, error)
}

type MerchService interface {
	BuyMerch(ctx context.Context, username, merchName string) error
	GetAllMerch(ctx context.Context) ([]*domain.Merch, error)
}

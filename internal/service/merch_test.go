package service

import (
	"context"
	"errors"
	"testing"

	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock репозиториев
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetUserInfo(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) UpdateUserInventory(ctx context.Context, user *domain.User, item string, quantity int) error {
	args := m.Called(ctx, user, item, quantity)
	return args.Error(0)
}

type mockMerchRepo struct {
	mock.Mock
}

func (m *mockMerchRepo) GetMerchByName(ctx context.Context, name string) (*domain.Merch, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Merch), args.Error(1)
}

func (m *mockMerchRepo) GetAllMerch(ctx context.Context) ([]*domain.Merch, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Merch), args.Error(1)
}

func (m *mockMerchRepo) GetMerchById(ctx context.Context, id int) (*domain.Merch, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Merch), args.Error(1)
}

type mockTransactionRepo struct {
	mock.Mock
}

func (m *mockTransactionRepo) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *mockTransactionRepo) GetTransactionsBySender(ctx context.Context, senderName, transferType string) ([]*domain.Transaction, error) {
	args := m.Called(ctx, senderName, transferType)
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func (m *mockTransactionRepo) GetTransactionsByReceiver(ctx context.Context, receiverName, transferType string) ([]*domain.Transaction, error) {
	args := m.Called(ctx, receiverName, transferType)
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func (m *mockTransactionRepo) ExecutePurchase(ctx context.Context, sender, receiver string, amount uint64) error {
	args := m.Called(ctx, sender, receiver, amount)
	return args.Error(0)
}

func (m *mockTransactionRepo) ExecuteTransfer(ctx context.Context, sender, receiver string, amount uint64) error {
	args := m.Called(ctx, sender, receiver, amount)
	return args.Error(0)
}

func (m *mockTransactionRepo) GetUserTransactions(ctx context.Context, username string) ([]*domain.Transaction, error) {
	args := m.Called(ctx, username)
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func TestBuyMerch_Success(t *testing.T) {
	// Подготовка
	userRepo := new(mockUserRepo)
	merchRepo := new(mockMerchRepo)
	transRepo := new(mockTransactionRepo)
	service := NewMerchService(userRepo, merchRepo, transRepo)

	username := "testuser"
	itemName := "test-item"
	price := uint64(100)

	merch := &domain.Merch{
		Name:  itemName,
		Price: price,
	}

	// Настройка ожиданий
	merchRepo.On("GetMerchByName", mock.Anything, itemName).Return(merch, nil)
	transRepo.On("ExecutePurchase", mock.Anything, username, itemName, price).Return(nil)

	// Действие
	err := service.BuyMerch(context.Background(), username, itemName)

	// Проверка
	require.NoError(t, err)
	merchRepo.AssertExpectations(t)
	transRepo.AssertExpectations(t)
}

func TestBuyMerch_InsufficientFunds(t *testing.T) {
	// Подготовка
	userRepo := new(mockUserRepo)
	merchRepo := new(mockMerchRepo)
	transRepo := new(mockTransactionRepo)
	service := NewMerchService(userRepo, merchRepo, transRepo)

	username := "testuser"
	itemName := "test-item"
	price := uint64(1000)

	merch := &domain.Merch{
		Name:  itemName,
		Price: price,
	}

	// Настройка ожиданий
	merchRepo.On("GetMerchByName", mock.Anything, itemName).Return(merch, nil)
	transRepo.On("ExecutePurchase", mock.Anything, username, itemName, price).Return(domain.ErrInsufficientFunds)

	// Действие
	err := service.BuyMerch(context.Background(), username, itemName)

	// Проверка
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInsufficientFunds))
	merchRepo.AssertExpectations(t)
	transRepo.AssertExpectations(t)
}

func TestBuyMerch_FailedTransaction(t *testing.T) {
	// Подготовка
	userRepo := new(mockUserRepo)
	merchRepo := new(mockMerchRepo)
	transRepo := new(mockTransactionRepo)
	service := NewMerchService(userRepo, merchRepo, transRepo)

	username := "testuser"
	itemName := "test-item"
	price := uint64(100)

	merch := &domain.Merch{
		Name:  itemName,
		Price: price,
	}

	expectedError := errors.New("ошибка транзакции")

	// Настройка ожиданий
	merchRepo.On("GetMerchByName", mock.Anything, itemName).Return(merch, nil)
	transRepo.On("ExecutePurchase", mock.Anything, username, itemName, price).Return(expectedError)

	// Действие
	err := service.BuyMerch(context.Background(), username, itemName)

	// Проверка
	require.Error(t, err)
	require.True(t, errors.Is(err, expectedError))
	merchRepo.AssertExpectations(t)
	transRepo.AssertExpectations(t)
}

func TestBuyMerch_CacheHit(t *testing.T) {
	// Подготовка
	userRepo := new(mockUserRepo)
	merchRepo := new(mockMerchRepo)
	transRepo := new(mockTransactionRepo)
	service := NewMerchService(userRepo, merchRepo, transRepo)

	username := "testuser"
	itemName := "test-item"
	price := uint64(100)

	merch := &domain.Merch{
		Name:  itemName,
		Price: price,
	}

	// Первый запрос - промах кэша
	merchRepo.On("GetMerchByName", mock.Anything, itemName).Return(merch, nil).Once()
	transRepo.On("ExecutePurchase", mock.Anything, username, itemName, price).Return(nil)

	// Первая покупка
	err := service.BuyMerch(context.Background(), username, itemName)
	require.NoError(t, err)

	// Вторая покупка - должна использовать кэш
	err = service.BuyMerch(context.Background(), username, itemName)
	require.NoError(t, err)

	// Проверяем, что GetMerchByName был вызван только один раз
	merchRepo.AssertNumberOfCalls(t, "GetMerchByName", 1)
	transRepo.AssertNumberOfCalls(t, "ExecutePurchase", 2)
}

func TestGetAllMerch_WithCaching(t *testing.T) {
	// Подготовка
	userRepo := new(mockUserRepo)
	merchRepo := new(mockMerchRepo)
	transRepo := new(mockTransactionRepo)
	service := NewMerchService(userRepo, merchRepo, transRepo)

	merch := []*domain.Merch{
		{Name: "item1", Price: 100},
		{Name: "item2", Price: 200},
	}

	// Настройка ожиданий
	merchRepo.On("GetAllMerch", mock.Anything).Return(merch, nil).Once()

	// Первый запрос
	items, err := service.GetAllMerch(context.Background())
	require.NoError(t, err)
	require.Len(t, items, 2)

	// После этого запроса товары должны быть в кэше
	// Проверяем каждый товар через BuyMerch - не должно быть обращений к GetMerchByName
	for _, item := range items {
		transRepo.On("ExecutePurchase", mock.Anything, "testuser", item.Name, item.Price).Return(nil).Once()
		err := service.BuyMerch(context.Background(), "testuser", item.Name)
		require.NoError(t, err)
	}

	// Проверяем, что GetMerchByName не вызывался
	merchRepo.AssertNotCalled(t, "GetMerchByName")
	merchRepo.AssertNumberOfCalls(t, "GetAllMerch", 1)
}

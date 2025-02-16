package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSendCoins_Success(t *testing.T) {
	// Подготовка
	userRepo := new(mockUserRepo)
	transRepo := new(mockTransactionRepo)
	service := NewTransferService(transRepo, userRepo)

	sender := "sender"
	receiver := "receiver"
	amount := uint64(300)

	senderUser := &domain.User{
		Username: sender,
		Coins:    1000,
	}

	receiverUser := &domain.User{
		Username: receiver,
		Coins:    500,
	}

	// Настройка ожиданий
	userRepo.On("GetUserByUsername", mock.Anything, sender).Return(senderUser, nil)
	userRepo.On("GetUserByUsername", mock.Anything, receiver).Return(receiverUser, nil)
	transRepo.On("ExecuteTransfer", mock.Anything, sender, receiver, amount).Return(nil)

	// Действие
	err := service.SendCoins(context.Background(), sender, receiver, amount)

	// Проверка
	require.NoError(t, err)
	userRepo.AssertExpectations(t)
	transRepo.AssertExpectations(t)
}

func TestSendCoins_InsufficientFunds(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	transRepo := new(mockTransactionRepo)

	service := NewTransferService(transRepo, userRepo)

	sender := &domain.User{
		Id:       1,
		Username: "sender",
		Coins:    100,
	}

	receiver := &domain.User{
		Id:       2,
		Username: "receiver",
		Coins:    500,
	}

	amount := uint64(300)

	// Настраиваем ожидания
	userRepo.On("GetUserByUsername", ctx, "sender").Return(sender, nil)
	userRepo.On("GetUserByUsername", ctx, "receiver").Return(receiver, nil)
	transRepo.On("ExecuteTransfer", ctx, "sender", "receiver", amount).Return(domain.ErrInsufficientFunds)

	// Act
	err := service.SendCoins(ctx, "sender", "receiver", amount)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
	assert.Equal(t, uint64(100), sender.Coins)
	assert.Equal(t, uint64(500), receiver.Coins)
	userRepo.AssertExpectations(t)
	transRepo.AssertExpectations(t)
}

func TestGetTransactionHistory_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	transRepo := new(mockTransactionRepo)

	service := NewTransferService(transRepo, userRepo)

	username := "testuser"
	now := time.Now()

	transactions := []*domain.Transaction{
		{
			SenderName:   username,
			ReceiverName: "user1",
			Amount:       100,
			Type:         domain.TransactionTypeTransfer,
			Timestamp:    now,
		},
		{
			SenderName:   "user2",
			ReceiverName: username,
			Amount:       200,
			Type:         domain.TransactionTypeTransfer,
			Timestamp:    now,
		},
	}

	// Настраиваем ожидания
	transRepo.On("GetUserTransactions", ctx, username).Return(transactions, nil)

	// Act
	history, err := service.GetTransactionHistory(ctx, username)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, history.Sent, 1)
	assert.Len(t, history.Received, 1)
	assert.Equal(t, uint64(100), history.Sent[0].Amount)
	assert.Equal(t, uint64(200), history.Received[0].Amount)
	transRepo.AssertExpectations(t)
}

func TestSendCoins_TransactionError(t *testing.T) {
	// Подготовка
	userRepo := new(mockUserRepo)
	transRepo := new(mockTransactionRepo)
	service := NewTransferService(transRepo, userRepo)

	sender := "sender"
	receiver := "receiver"
	amount := uint64(100)

	senderUser := &domain.User{
		Username: sender,
		Coins:    1000,
	}

	receiverUser := &domain.User{
		Username: receiver,
		Coins:    500,
	}

	expectedError := errors.New("ошибка транзакции")

	// Настройка ожиданий
	userRepo.On("GetUserByUsername", mock.Anything, sender).Return(senderUser, nil)
	userRepo.On("GetUserByUsername", mock.Anything, receiver).Return(receiverUser, nil)
	transRepo.On("ExecuteTransfer", mock.Anything, sender, receiver, amount).Return(expectedError)

	// Действие
	err := service.SendCoins(context.Background(), sender, receiver, amount)

	// Проверка
	require.Error(t, err)
	require.True(t, errors.Is(err, expectedError))
	userRepo.AssertExpectations(t)
	transRepo.AssertExpectations(t)
}

func TestSendCoins_SelfTransfer(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	transRepo := new(mockTransactionRepo)

	service := NewTransferService(transRepo, userRepo)

	username := "testuser"
	amount := uint64(100)

	// Act
	err := service.SendCoins(ctx, username, username, amount)

	// Assert
	assert.NoError(t, err)
	// Проверяем, что никакие методы репозиториев не вызывались
	userRepo.AssertNotCalled(t, "GetUserByUsername")
	userRepo.AssertNotCalled(t, "UpdateUser")
	transRepo.AssertNotCalled(t, "CreateTransaction")
}

func TestSendCoins_ReceiverUpdateError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	transRepo := new(mockTransactionRepo)
	service := NewTransferService(transRepo, userRepo)

	sender := &domain.User{
		Id:       1,
		Username: "sender",
		Coins:    1000,
	}
	receiver := &domain.User{
		Id:       2,
		Username: "receiver",
		Coins:    500,
	}

	amount := uint64(300)
	expectedError := errors.New("ошибка обновления баланса получателя")

	// Настраиваем ожидания
	userRepo.On("GetUserByUsername", ctx, "sender").Return(sender, nil)
	userRepo.On("GetUserByUsername", ctx, "receiver").Return(receiver, nil)
	transRepo.On("ExecuteTransfer", ctx, "sender", "receiver", amount).Return(expectedError)

	// Act
	err := service.SendCoins(ctx, "sender", "receiver", amount)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, expectedError)
	userRepo.AssertExpectations(t)
	transRepo.AssertExpectations(t)
}

func TestSendCoins_GetUserError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	transRepo := new(mockTransactionRepo)
	service := NewTransferService(transRepo, userRepo)

	// Настраиваем ожидания для ошибки получения отправителя
	userRepo.On("GetUserByUsername", ctx, "sender").Return(nil, errors.New("пользователь не найден"))

	// Act
	err := service.SendCoins(ctx, "sender", "receiver", uint64(100))

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "проверка отправителя")
	userRepo.AssertExpectations(t)

	// Сбрасываем мок и тестируем ошибку получения получателя
	userRepo = new(mockUserRepo)
	transRepo = new(mockTransactionRepo)
	service = NewTransferService(transRepo, userRepo)

	sender := &domain.User{
		Id:       1,
		Username: "sender",
		Coins:    1000,
	}

	userRepo.On("GetUserByUsername", ctx, "sender").Return(sender, nil)
	userRepo.On("GetUserByUsername", ctx, "receiver").Return(nil, errors.New("пользователь не найден"))

	// Act
	err = service.SendCoins(ctx, "sender", "receiver", uint64(100))

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "проверка получателя")
	userRepo.AssertExpectations(t)
}

func TestGetTransactionHistory_Error(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	transRepo := new(mockTransactionRepo)
	service := NewTransferService(transRepo, userRepo)

	username := "testuser"
	expectedError := errors.New("ошибка базы данных")

	// Тест ошибки получения транзакций
	t.Run("ошибка получения транзакций", func(t *testing.T) {
		transRepo.On("GetUserTransactions", ctx, username).Return([]*domain.Transaction(nil), expectedError)

		history, err := service.GetTransactionHistory(ctx, username)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "получение транзакций")
		assert.Empty(t, history)
		transRepo.AssertExpectations(t)
	})
}

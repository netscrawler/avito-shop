package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/netscrawler/avito-shop/internal/model"
	"github.com/netscrawler/avito-shop/internal/repository"
	"github.com/sirupsen/logrus"
)

// TransferService предоставляет методы для работы с переводами
type transferService struct {
	transRepo repository.TransactionRepository
	userRepo  repository.UserRepository
	locks     map[string]*sync.Mutex
	locksMu   sync.RWMutex
}

// NewTransferService создает новый экземпляр сервиса переводов
func NewTransferService(transRepo repository.TransactionRepository, userRepo repository.UserRepository) TransferService {
	return &transferService{
		transRepo: transRepo,
		userRepo:  userRepo,
		locks:     make(map[string]*sync.Mutex),
	}
}

func (s *transferService) getUserLock(username string) *sync.Mutex {
	s.locksMu.Lock()
	defer s.locksMu.Unlock()

	if lock, exists := s.locks[username]; exists {
		return lock
	}

	lock := &sync.Mutex{}
	s.locks[username] = lock
	return lock
}

// SendCoins выполняет перевод монет между пользователями
func (s *transferService) SendCoins(ctx context.Context, from, to string, amount uint64) error {
	const op = "TransferService.SendCoins"

	if from == to {
		return nil
	}

	// Получаем блокировки в определенном порядке для предотвращения взаимных блокировок
	firstLock, secondLock := from, to
	if from > to {
		firstLock, secondLock = to, from
	}

	firstMutex := s.getUserLock(firstLock)
	secondMutex := s.getUserLock(secondLock)

	firstMutex.Lock()
	defer firstMutex.Unlock()
	secondMutex.Lock()
	defer secondMutex.Unlock()

	// Проверяем существование отправителя
	if _, err := s.userRepo.GetUserByUsername(ctx, from); err != nil {
		if err == domain.ErrUserNotFound {
			logrus.Warnf("%s: отправитель %s не найден", op, from)
			return fmt.Errorf("%s: %w", op, domain.ErrSenderNotFound)
		}
		logrus.Errorf("%s: ошибка при проверке отправителя %s: %v", op, from, err)
		return fmt.Errorf("%s: проверка отправителя: %w", op, err)
	}

	// Проверяем существование получателя
	if _, err := s.userRepo.GetUserByUsername(ctx, to); err != nil {
		if err == domain.ErrUserNotFound {
			logrus.Warnf("%s: получатель %s не найден", op, to)
			return fmt.Errorf("%s: %w", op, domain.ErrRecipientNotFound)
		}
		logrus.Errorf("%s: ошибка при проверке получателя %s: %v", op, to, err)
		return fmt.Errorf("%s: проверка получателя: %w", op, err)
	}

	// Выполняем перевод в рамках транзакции
	err := s.transRepo.ExecuteTransfer(ctx, from, to, amount)
	if err != nil {
		logrus.Errorf("%s: ошибка при выполнении перевода: %v", op, err)
		return fmt.Errorf("%s: выполнение перевода: %w", op, err)
	}

	logrus.Infof("%s: успешно выполнен перевод %d монет от %s к %s", op, amount, from, to)
	return nil
}

// GetTransactionHistory возвращает историю транзакций пользователя
func (s *transferService) GetTransactionHistory(ctx context.Context, username string) (model.CoinHistory, error) {
	const op = "TransferService.GetTransactionHistory"

	// Получаем все транзакции пользователя
	transactions, err := s.transRepo.GetUserTransactions(ctx, username)
	if err != nil {
		return model.CoinHistory{}, fmt.Errorf("%s: получение транзакций: %w", op, err)
	}

	// Разделяем транзакции на отправленные и полученные
	var sent []model.SentTransaction
	var received []model.ReceivedTransaction

	for _, t := range transactions {
		if t.SenderName == username {
			sent = append(sent, model.SentTransaction{
				ToUser: t.ReceiverName,
				Amount: t.Amount,
			})
		} else if t.ReceiverName == username {
			received = append(received, model.ReceivedTransaction{
				FromUser: t.SenderName,
				Amount:   t.Amount,
			})
		}
	}

	return model.CoinHistory{
		Sent:     sent,
		Received: received,
	}, nil
}

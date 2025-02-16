package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/netscrawler/avito-shop/internal/repository"
	"github.com/sirupsen/logrus"
)

const (
	merchCacheTTL = 5 * time.Minute
)

type merchCache struct {
	merch     *domain.Merch
	timestamp time.Time
}

// MerchService предоставляет методы для работы с товарами
type merchService struct {
	userRepo  repository.UserRepository
	merchRepo repository.MerchRepository
	transRepo repository.TransactionRepository
	cache     map[string]merchCache
	cacheMu   sync.RWMutex
}

// NewMerchService создает новый экземпляр сервиса товаров
func NewMerchService(userRepo repository.UserRepository, merchRepo repository.MerchRepository, transRepo repository.TransactionRepository) MerchService {
	service := &merchService{
		userRepo:  userRepo,
		merchRepo: merchRepo,
		transRepo: transRepo,
		cache:     make(map[string]merchCache),
	}

	// Запускаем очистку кэша
	go service.cleanCache()

	return service
}

func (s *merchService) cleanCache() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		s.cacheMu.Lock()
		now := time.Now()
		for name, cached := range s.cache {
			if now.Sub(cached.timestamp) > merchCacheTTL {
				delete(s.cache, name)
			}
		}
		s.cacheMu.Unlock()
	}
}

func (s *merchService) getCachedMerch(name string) *domain.Merch {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if cached, ok := s.cache[name]; ok {
		if time.Since(cached.timestamp) < merchCacheTTL {
			return cached.merch
		}
	}
	return nil
}

func (s *merchService) cacheMerch(merch *domain.Merch) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache[merch.Name] = merchCache{
		merch:     merch,
		timestamp: time.Now(),
	}
}

// BuyMerch обрабатывает покупку товара пользователем
func (s *merchService) BuyMerch(ctx context.Context, username, merchName string) error {
	const op = "MerchService.BuyMerch"

	// Проверяем кэш
	merch := s.getCachedMerch(merchName)
	var err error
	if merch == nil {
		merch, err = s.merchRepo.GetMerchByName(ctx, merchName)
		if err != nil {
			if err == domain.ErrMerchNotFound {
				logrus.Warnf("%s: товар %s не найден", op, merchName)
				return fmt.Errorf("%s: %w", op, err)
			}
			logrus.Errorf("%s: ошибка при получении товара %s: %v", op, merchName, err)
			return fmt.Errorf("%s: получение товара: %w", op, err)
		}
		// Кэшируем товар
		s.cacheMerch(merch)
	}

	// Выполняем покупку в рамках одной транзакции
	if err := s.transRepo.ExecutePurchase(ctx, username, merchName, merch.Price); err != nil {
		logrus.Errorf("%s: ошибка при выполнении покупки: %v", op, err)
		return fmt.Errorf("%s: выполнение покупки: %w", op, err)
	}

	logrus.Infof("%s: пользователь %s успешно купил товар %s", op, username, merchName)
	return nil
}

// GetAllMerch возвращает список всех доступных товаров
func (s *merchService) GetAllMerch(ctx context.Context) ([]*domain.Merch, error) {
	const op = "MerchService.GetAllMerch"

	merch, err := s.merchRepo.GetAllMerch(ctx)
	if err != nil {
		logrus.Errorf("%s: ошибка при получении списка товаров: %v", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Кэшируем все товары
	for _, m := range merch {
		s.cacheMerch(m)
	}

	return merch, nil
}

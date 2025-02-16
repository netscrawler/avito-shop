package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netscrawler/avito-shop/internal/config"
	"github.com/netscrawler/avito-shop/internal/repository/postgres"
	"github.com/netscrawler/avito-shop/internal/service"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	db              *pgxpool.Pool
	userService     service.UserService
	merchService    service.MerchService
	transferService service.TransferService
	ctx             context.Context
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Настройка подключения к тестовой БД
	dbConfig := config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "avito_shop_test",
	}

	// Подключение к БД
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	var err error
	s.db, err = pgxpool.New(s.ctx, connString)
	require.NoError(s.T(), err)

	// Инициализация репозиториев
	merchRepo := postgres.NewMerchRepository(s.db)
	userRepo := postgres.NewUserRepository(s.db)
	transactionRepo := postgres.NewTransactionRepository(s.db)

	// Инициализация сервисов
	s.userService = service.NewUserService(userRepo, "your-secret-key")
	s.merchService = service.NewMerchService(userRepo, merchRepo, transactionRepo)
	s.transferService = service.NewTransferService(transactionRepo, userRepo)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *IntegrationTestSuite) TestUserOperations() {
	// Тест операций с пользователем
	username := "test-user"
	password := "test-password"

	// Регистрация пользователя
	err := s.userService.RegisterUser(s.ctx, username, password)
	s.Require().NoError(err)

	// Аутентификация пользователя
	token, err := s.userService.AuthenticateUser(s.ctx, username, password)
	s.Require().NoError(err)
	s.Require().NotEmpty(token)

	// Получение информации о пользователе
	user, err := s.userService.GetUserInfo(s.ctx, username)
	s.Require().NoError(err)
	s.Require().NotNil(user)
	s.Equal(username, user.Username)
}

func (s *IntegrationTestSuite) TestMerchOperations() {
	// Тест операций с товарами
	username := "merch-test-user"
	password := "test-password"

	// Регистрация пользователя
	err := s.userService.RegisterUser(s.ctx, username, password)
	s.Require().NoError(err)

	// Получение списка товаров
	merch, err := s.merchService.GetAllMerch(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(merch)

	if len(merch) > 0 {
		// Покупка первого доступного товара
		err = s.merchService.BuyMerch(s.ctx, username, merch[0].Name)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) TestTransferOperations() {
	// Тест операций с переводами
	sender := "sender-user"
	receiver := "receiver-user"
	password := "test-password"
	amount := uint64(1000)

	// Регистрация пользователей
	err := s.userService.RegisterUser(s.ctx, sender, password)
	s.Require().NoError(err)
	err = s.userService.RegisterUser(s.ctx, receiver, password)
	s.Require().NoError(err)

	// Перевод средств
	err = s.transferService.SendCoins(s.ctx, sender, receiver, amount)
	s.Require().NoError(err)

	// Проверка истории транзакций
	history, err := s.transferService.GetTransactionHistory(s.ctx, sender)
	s.Require().NoError(err)
	s.Require().NotEmpty(history)
}

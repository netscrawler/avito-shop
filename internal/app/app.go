package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netscrawler/avito-shop/internal/config"
	"github.com/netscrawler/avito-shop/internal/handler"
	"github.com/netscrawler/avito-shop/internal/middleware"
	"github.com/netscrawler/avito-shop/internal/repository/postgres"
	"github.com/netscrawler/avito-shop/internal/service"
	"github.com/sirupsen/logrus"
)

// App представляет основное приложение
type App struct {
	cfg    *config.Config
	logger *logrus.Logger
	router *gin.Engine
	db     *pgxpool.Pool
}

// New создает новый экземпляр приложения
func New(cfg *config.Config) (*App, error) {
	logger := setupLogger()

	// Подключаемся к БД
	db, err := setupDatabase(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Создаем роутер
	router := setupRouter(cfg, db, logger)

	return &App{
		cfg:    cfg,
		logger: logger,
		router: router,
		db:     db,
	}, nil
}

// Run запускает приложение
func (a *App) Run() error {
	// Создаем HTTP сервер
	server := &http.Server{
		Addr:           ":" + a.cfg.Server.Port,
		Handler:        a.router,
		ReadTimeout:    a.cfg.Server.ReadTimeout,
		WriteTimeout:   a.cfg.Server.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Канал для сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		a.logger.Infof("Сервер запущен на порту %s", a.cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	<-quit
	a.logger.Info("Получен сигнал завершения, выполняется graceful shutdown")

	// Даем 30 секунд на завершение текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("ошибка при graceful shutdown: %w", err)
	}

	// Закрываем соединение с БД
	a.db.Close()

	a.logger.Info("Сервер успешно остановлен")
	return nil
}

func setupLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	return logger
}

func setupDatabase(cfg *config.Config, logger *logrus.Logger) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgx.Connect(ctx, cfg.Database.DatabaseUrl())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}
	logger.Info("Успешное подключение к БД")

	// Создаем пул соединений
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.DatabaseUrl())
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфигурации пула: %w", err)
	}

	// Настраиваем пул соединений
	poolConfig.MaxConns = 500                              // Увеличиваем максимальное количество соединений
	poolConfig.MinConns = 50                               // Увеличиваем минимальное количество соединений
	poolConfig.MaxConnLifetime = 30 * time.Minute          // Уменьшаем время жизни соединения
	poolConfig.MaxConnIdleTime = 15 * time.Minute          // Уменьшаем время простоя
	poolConfig.HealthCheckPeriod = 30 * time.Second        // Чаще проверяем соединения
	poolConfig.ConnConfig.ConnectTimeout = 3 * time.Second // Уменьшаем таймаут подключения

	// Настраиваем параметры
	poolConfig.ConnConfig.RuntimeParams = map[string]string{
		"application_name":                    "avito_shop",
		"statement_timeout":                   "5000",  // 5 секунд
		"idle_in_transaction_session_timeout": "15000", // 15 секунд
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула соединений: %w", err)
	}

	return pool, nil
}
func setupRouter(cfg *config.Config, db *pgxpool.Pool, logger *logrus.Logger) *gin.Engine {
	// Создаем репозитории
	dbPool := postgres.NewPoolAdapter(db)
	userRepo := postgres.NewUserRepository(dbPool)
	merchRepo := postgres.NewMerchRepository(dbPool)
	transRepo := postgres.NewTransactionRepository(dbPool)

	// Создаем сервисы
	userService := service.NewUserService(userRepo, cfg.JWT.Secret)
	transferService := service.NewTransferService(transRepo, userRepo)
	merchService := service.NewMerchService(userRepo, merchRepo, transRepo)

	// Создаем обработчики
	h := handler.NewHandler(userService, transferService, merchService)

	// Настраиваем роутер
	router := gin.New()

	// Добавляем middleware
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware(logger))

	// Определяем маршруты
	router.GET("/health", h.HealthCheck)
	router.POST("/api/auth", h.Authenticate)

	// Группа защищенных маршрутов
	api := router.Group("/api")
	api.Use(middleware.JWTAuthMiddleware(cfg.JWT.Secret))
	api.GET("/info", h.GetInfo)
	api.POST("/sendCoin", h.SendCoin)
	api.GET("/buy/:item", h.BuyMerch)

	return router
}

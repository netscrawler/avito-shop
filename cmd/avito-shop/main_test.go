package main

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/netscrawler/avito-shop/internal/app"
	"github.com/netscrawler/avito-shop/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationInitialization(t *testing.T) {
	// Тест успешной инициализации
	t.Run("успешная инициализация приложения", func(t *testing.T) {
		cfg, err := config.New()
		require.NoError(t, err, "ошибка создания конфигурации")
		require.NotNil(t, cfg, "конфигурация не должна быть nil")

		// Устанавливаем тестовые параметры базы данных
		cfg.Database = config.DatabaseConfig{
			Host:     "db",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "avito_shop",
			SSLMode:  "disable",
		}

		// Находим свободный порт для сервера
		listener, err := net.Listen("tcp", ":0")
		require.NoError(t, err, "ошибка при поиске свободного порта")
		port := listener.Addr().(*net.TCPAddr).Port
		listener.Close()
		cfg.Server.Port = fmt.Sprintf("%d", port)

		instance, err := app.New(cfg)
		require.NoError(t, err, "ошибка создания приложения")
		require.NotNil(t, instance, "приложение не должно быть nil")
	})

	// Тест с некорректной конфигурацией
	t.Run("инициализация с некорректной конфигурацией", func(t *testing.T) {
		// Создаем некорректную конфигурацию
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Host: "invalid-host",
				Port: "invalid-port",
			},
		}
		instance, err := app.New(cfg)
		assert.Error(t, err, "должна быть ошибка при некорректной конфигурации")
		assert.Nil(t, instance, "приложение должно быть nil при ошибке")
	})
}

func TestConfigInitialization(t *testing.T) {
	t.Run("успешное создание конфигурации", func(t *testing.T) {
		cfg, err := config.New()
		assert.NoError(t, err, "не должно быть ошибки при создании конфигурации")
		assert.NotNil(t, cfg, "конфигурация не должна быть nil")

		// Проверяем наличие обязательных полей
		assert.NotEmpty(t, cfg.Server.Port, "порт сервера должен быть установлен")
		assert.NotEmpty(t, cfg.JWT.Secret, "секрет JWT должен быть установлен")
		assert.NotNil(t, cfg.Database, "конфигурация базы данных должна быть установлена")
	})
}

func TestIntegrationStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в коротком режиме")
	}

	t.Run("полный цикл запуска приложения", func(t *testing.T) {
		// Находим свободный порт
		listener, err := net.Listen("tcp", ":0")
		require.NoError(t, err, "ошибка при поиске свободного порта")
		port := listener.Addr().(*net.TCPAddr).Port
		listener.Close()

		// Создаем конфигурацию с найденным портом
		cfg, err := config.New()
		require.NoError(t, err, "ошибка создания конфигурации")
		cfg.Server.Port = fmt.Sprintf("%d", port)

		// Создаем контекст с отменой
		_, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Создаем приложение
		instance, err := app.New(cfg)
		require.NoError(t, err, "ошибка создания приложения")

		// Запускаем приложение в горутине
		errCh := make(chan error, 1)
		go func() {
			errCh <- instance.Run()
		}()

		// Проверяем, что приложение запустилось без ошибок
		select {
		case err := <-errCh:
			assert.NoError(t, err, "ошибка при запуске приложения")
		case <-time.After(5 * time.Second):
			// Если за 5 секунд нет ошибки, считаем что запуск успешен
			t.Log("приложение успешно запущено")
		}

		// Останавливаем приложение
		cancel()
		time.Sleep(time.Second) // Даем время на корректное завершение
	})
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netscrawler/avito-shop/internal/config"
	"github.com/netscrawler/avito-shop/internal/test/e2e"
)

func main() {
	// Загружаем конфигурацию для тестов из переменных окружения
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         getEnv("SERVER_PORT", "8081"),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		Database: config.DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", "test-db"),
			Port:     getEnv("DATABASE_PORT", "5432"),
			User:     getEnv("DATABASE_USER", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", "password"),
			DBName:   getEnv("DATABASE_NAME", "shop_test"),
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret: getEnv("JWT_SECRET", "test_secret"),
			TTL:    parseDuration(getEnv("JWT_TTL", "24h")),
		},
	}

	// Подключаемся к тестовой базе данных
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Инициализируем HTTP клиент
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Ждем, пока приложение станет доступным
	appHost := getEnv("APP_HOST", "app")
	baseURL := fmt.Sprintf("http://%s:%s", appHost, cfg.Server.Port)

	fmt.Println("Waiting for application to become ready...")
	for i := 0; i < 30; i++ {
		resp, err := httpClient.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				fmt.Printf("Failed to read health check response: %v\n", err)
				os.Exit(1)
			}

			var healthResp struct {
				Status string `json:"status"`
			}
			if err := json.Unmarshal(body, &healthResp); err != nil {
				fmt.Printf("Failed to parse health check response: %v\n", err)
				os.Exit(1)
			}

			if healthResp.Status == "ok" {
				fmt.Println("Application is ready!")
				break
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
		if i == 29 {
			fmt.Println("Timeout waiting for application")
			os.Exit(1)
		}
		time.Sleep(time.Second)
	}

	// Запускаем тесты
	suite := e2e.NewE2ETestSuite(cfg, dbPool, httpClient, baseURL)
	if err := suite.Run(); err != nil {
		fmt.Printf("Tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All tests passed successfully!")
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// parseDuration парсит строку длительности или возвращает значение по умолчанию
func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 24 * time.Hour // значение по умолчанию
	}
	return duration
}

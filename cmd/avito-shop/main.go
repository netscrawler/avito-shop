package main

import (
	"log"

	"github.com/netscrawler/avito-shop/internal/app"
	"github.com/netscrawler/avito-shop/internal/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("ошибка загрузки конфигурации: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("ошибка создания приложения: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("ошибка запуска приложения: %v", err)
	}
}

# Используем образ с Go для тестов
FROM golang:1.21-alpine

# Устанавливаем необходимые зависимости
RUN apk add --no-cache gcc musl-dev curl

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем только файлы с зависимостями
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Экспортируем порт
EXPOSE 8081

# Команда для запуска тестов
CMD ["go", "test", "-v", "./internal/test/e2e/..."]

# Команда запуска будет передана через docker-compose
CMD ["go", "run", "cmd/avito-shop/main.go"] 
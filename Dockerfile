# Этап сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые зависимости
RUN apk add --no-cache gcc musl-dev

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем только файлы с зависимостями
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/avito-shop

# Финальный этап
FROM alpine:3.18

# Устанавливаем curl для healthcheck
RUN apk add --no-cache curl

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/main .

# Экспортируем порт
EXPOSE 8081

# Команда запуска приложения
CMD ["./main"] 
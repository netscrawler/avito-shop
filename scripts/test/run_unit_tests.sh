#!/bin/bash

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color
BLUE='\033[0;34m'

# Функция для вывода сообщений
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

# Функция для проверки результата выполнения команды
check_result() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $1 успешно выполнено${NC}"
    else
        echo -e "${RED}✗ $1 завершилось с ошибкой${NC}"
        exit 1
    fi
}

# Переходим в корневую директорию проекта
cd "$(dirname "$0")/../.." || exit 1

# Запуск unit-тестов
log "Запуск unit-тестов..."
go test -v ./internal/service/... ./internal/repository/... ./internal/handler/...
check_result "Unit-тесты"

echo -e "\n${GREEN}Unit-тесты успешно пройдены!${NC}" 
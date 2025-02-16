#!/bin/bash

# Переходим в директорию с docker-compose
cd "$(dirname "$0")"

echo "Останавливаем предыдущие контейнеры..."
docker-compose down -v

echo "Запускаем тесты..."
docker-compose up \
    --build \
    --abort-on-container-exit \
    --exit-code-from test \
    --remove-orphans

# Получаем код возврата
exit_code=$?

echo "Очищаем контейнеры..."
docker-compose down -v

if [ $exit_code -eq 0 ]; then
    echo "Тесты успешно пройдены!"
else
    echo "Тесты завершились с ошибкой (код: $exit_code)"
fi

# Возвращаем код возврата тестов
exit $exit_code 
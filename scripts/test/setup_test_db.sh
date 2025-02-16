#!/bin/bash

# Переходим в корневую директорию проекта
cd "$(dirname "$0")/../.." || exit 1

# Остановка существующих контейнеров
docker-compose down

# Запуск PostgreSQL
docker-compose up -d db

# Ожидание готовности PostgreSQL
echo "Ожидание готовности PostgreSQL..."
for i in {1..30}; do
    if PGPASSWORD=postgres psql -h localhost -U postgres -c '\q' >/dev/null 2>&1; then
        break
    fi
    echo -n "."
    sleep 1
done
echo ""

# Создание тестовой базы данных
PGPASSWORD=postgres psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS avito_shop_test;"
PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE avito_shop_test;"

# Применение миграций к тестовой базе данных
PGPASSWORD=postgres psql -h localhost -U postgres -d avito_shop_test -f migrations/init.sql/001_create_tables.sql
PGPASSWORD=postgres psql -h localhost -U postgres -d avito_shop_test -f migrations/init.sql/002_add_foreign_keys.sql

# Добавление тестовых данных
PGPASSWORD=postgres psql -h localhost -U postgres -d avito_shop_test << EOF
-- Добавление тестовых товаров
INSERT INTO merch (name, price) VALUES 
    ('test-item-1', 1000),
    ('test-item-2', 2000),
    ('test-item-3', 3000);

-- Добавление тестовых пользователей с начальным балансом
INSERT INTO users (username, password_hash, balance) VALUES
    ('test-user-1', '\$2a\$10\$1234567890123456789012', 5000),
    ('test-user-2', '\$2a\$10\$1234567890123456789012', 5000);
EOF

echo "Тестовая база данных готова!" 
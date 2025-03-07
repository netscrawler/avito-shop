# Avito Shop

Сервис электронного магазина с поддержкой высоких нагрузок.

## Описание

Проект представляет собой RESTful API сервис для электронного магазина с следующим функционалом:
- Аутентификация пользователей
- Просмотр информации о товарах
- Система внутренних транзакций (отправка монет между пользователями)
- Покупка товаров

## Технологии

- Go (Golang)
- PostgreSQL
- Docker & Docker Compose
- Gin Web Framework
- JMeter (нагрузочное тестирование)

## Зависимости

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 15+
- Apache JMeter 5.6+

## Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/your-username/avito-shop.git
cd avito-shop
```

2. Запустите сервисы через Docker Compose:
```bash
docker-compose up -d
```

Сервис будет доступен по адресу: http://localhost:8080


## Результаты нагрузочного тестирования

### Конфигурация теста
- Длительность: 1 минута
- Количество пользователей: 200
- Время разгона (ramp-up): 20 секунд

### Общие результаты
- Всего запросов: 14,280
- Средняя пропускная способность: 236 запросов/сек
- Процент ошибок: 25.45%
- Среднее время отклика: 705 мс

### Результаты по эндпоинтам

#### Авторизация (Login Request)
- Успешность: 99.64%
- Среднее время: 1.4 сек
- 90-й процентиль: 2.27 сек
- Максимальное время: 3.8 сек

#### Информация (Get Info)
- Успешность: 99.64%
- Среднее время: 119 мс
- 90-й процентиль: 344 мс
- Максимальное время: 1.2 сек

#### Отправка монет (Send Coins)
- Успешность: 0% ()
- Среднее время: 977 мс
- 90-й процентиль: 1.73 сек
- Максимальное время: 3.58 сек
Низкая успешность из-за несуществующего получателя

#### Покупка товаров (Buy Item)
- Успешность: 99.51%
- Среднее время: 295 мс
- 90-й процентиль: 683 мс
- Максимальное время: 2.18 сек

##№P.S Нагрузочное тестрирование запускалось в docker контейнере, из-за ограничение железа показывают результат хуже, чем будет на сервере. 

## Запуск тестов

### Запуск всех тестов
```bash
./scripts/test/run_all_tests.sh
```

### Unit-тесты
```bash
./scripts/test/run_unit_tests.sh
```

### Интеграционные тесты
```bash
./scripts/test/run_integration_tests.sh
```

### E2E тесты
```bash
./scripts/test/run_e2e_tests.sh
```

### Нагрузочное тестирование
```bash
./scripts/test/run_load_test.sh
```

## Мониторинг

Результаты нагрузочного тестирования доступны в:
- `loadtest/report/index.html` - HTML отчет
- `loadtest/results.jtl` - сырые данные тестирования


## Автор

Netscrawler

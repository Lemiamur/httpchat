# HTTP Chat Service

Микросервис для обработки сообщений через HTTP API с сохранением в PostgreSQL и отправкой в Kafka.

## О проекте

HTTP Chat Service - это современный микросервис на языке Go, реализующий асинхронную обработку сообщений. 
Сервис следует принципам чистой архитектуры и использует лучшие практики Go разработки.

### Основные функции:
- Прием сообщений через HTTP API
- Хранение сообщений в PostgreSQL
- Асинхронная обработка через Apache Kafka
- Статистика по обработанным сообщениям

### Технологии:
- **Go 1.20** - язык программирования
- **PostgreSQL** - реляционная база данных
- **Apache Kafka** - распределенная потоковая платформа
- **Docker** - контейнеризация
- **Gin** - HTTP фреймворк
- **Zap** - структурированное логирование

## Быстрый старт

### Через Docker (рекомендуется):
```bash
docker-compose -f configs/docker-compose.yml up --build
```

### Локальный запуск:
```bash
# Запуск зависимостей
make docker-db-kafka

# Установка зависимостей
make deps

# Запуск приложения
make run
```

## API

Сервис предоставляет следующие эндпоинты:

### Создание сообщения
```http
POST /messages
```

```bash
curl -X POST http://localhost:8080/messages \
  -H \"Content-Type: application/json\" \
  -d '{\"content\": \"Привет, мир!\"}'
```

### Получение статистики
```http
GET /statistics
```

```bash
curl http://localhost:8080/statistics
```

### Обработка сообщения
```http
PUT /messages/{id}/process
```

```bash
curl -X PUT http://localhost:8080/messages/1/process
```

### Документация API
Swagger документация доступна по адресу: http://localhost:8080/swagger/

## Конфигурация

Сервис настраивается через переменные окружения:

- `SERVER_PORT` - Порт сервера (по умолчанию: 8080)
- `DATABASE_URL` - URL для подключения к PostgreSQL
- `KAFKA_BROKERS` - Список брокеров Kafka
- `KAFKA_TOPIC` - Топик Kafka для сообщений

## Разработка

### Команды Makefile:
```bash
make build        # Сборка приложения
make run          # Запуск приложения
make docker-run   # Запуск через Docker
make test         # Запуск тестов
make lint         # Запуск линтера
make swagger      # Генерация документации
```

### Архитектура:
```
Handler Layer (HTTP обработчики)
    ↓
Service Layer (бизнес-логика)
    ↓
Repository/Kafka Layer (работа с данными)
```

Сервис использует Kafka в режиме KRaft (Kafka Raft Metadata mode) без необходимости в ZooKeeper, что упрощает развертывание и обслуживание.

## Тестирование

```bash
# Запуск юнит-тестов
make test

# Запуск интеграционных тестов
make test-integration
```

## Лицензия

Apache 2.0
# Руководство по развертыванию HTTP Chat Service

## Требования к серверу

- ОС: Linux (рекомендуется Ubuntu 20.04 или новее)
- Docker и Docker Compose установлены
- Открытые порты: 8080 (приложение), 5432 (PostgreSQL), 9092 (Kafka)

## Шаги развертывания

### 1. Клонирование репозитория

```bash
git clone <URL_репозитория>
cd httpchat
```

### 2. Настройка конфигурации

Создайте файл `.env` в директории `configs` на основе примера:

```bash
cp configs/.env.local configs/.env
```

Отредактируйте файл `configs/.env` при необходимости, указав параметры подключения к базе данных и Kafka.

### 3. Запуск сервисов

Используйте Docker Compose для запуска всех сервисов:

```bash
docker-compose -f configs/docker-compose.yml up --build -d
```

### 4. Проверка состояния сервисов

Проверьте, что все контейнеры запущены:

```bash
docker-compose -f configs/docker-compose.yml ps
```

Вы должны увидеть три запущенных контейнера: `httpchat-app`, `httpchat-postgres` и `httpchat-kafka`.

### 5. Проверка работоспособности

После запуска вы можете проверить работоспособность сервиса:

```bash
# Создание сообщения
curl -X POST http://localhost:8080/messages -H "Content-Type: application/json" -d '{"content": "Hello, World!"}'

# Получение статистики
curl http://localhost:8080/statistics
```

## Мониторинг

Для мониторинга состояния сервисов можно использовать следующие команды:

```bash
# Просмотр логов приложения
docker-compose -f configs/docker-compose.yml logs app

# Просмотр логов базы данных
docker-compose -f configs/docker-compose.yml logs postgres

# Просмотр логов Kafka
docker-compose -f configs/docker-compose.yml logs kafka
```

## Обновление

Для обновления сервиса до новой версии:

1. Остановите текущие сервисы:
   ```bash
   docker-compose -f configs/docker-compose.yml down
   ```

2. Получите последние изменения из репозитория:
   ```bash
   git pull
   ```

3. Пересоберите и запустите сервисы:
   ```bash
   docker-compose -f configs/docker-compose.yml up --build -d
   ```
# HTTP Chat Service API

## Общая информация

Микросервис для обработки сообщений через HTTP API с сохранением в PostgreSQL и отправкой в Kafka.

## Базовый URL

```
http://localhost:8080
```

## Эндпоинты

### Создание сообщения

Создает новое сообщение и отправляет его в Kafka.

```
POST /messages
```

#### Тело запроса

```json
{
  "content": "string"
}
```

#### Ответы

```json
// 200 OK
{
  "id": 1
}
```

```json
// 400 Bad Request
{
  "error": "string"
}
```

```json
// 500 Internal Server Error
{
  "error": "string"
}
```

### Получение статистики

Возвращает статистику по обработанным и необработанным сообщениям.

```
GET /statistics
```

#### Ответы

```json
// 200 OK
{
  "total_messages": 100,
  "processed_messages": 75,
  "unprocessed_messages": 25
}
```

```json
// 500 Internal Server Error
{
  "error": "string"
}
```

### Обработка сообщения

Помечает сообщение как обработанное.

```
PUT /messages/{id}/process
```

#### Параметры пути

| Название | Тип   | Обязательный | Описание      |
|----------|-------|--------------|---------------|
| id       | int64 | Да           | ID сообщения  |

#### Ответы

```
// 200 OK
```

```json
// 400 Bad Request
{
  "error": "string"
}
```

```json
// 500 Internal Server Error
{
  "error": "string"
}
```
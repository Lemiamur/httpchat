# Примеры использования API

## Создание сообщения

```bash
curl -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "Привет, мир!"}'
```

Ответ:
```json
{
  "id": 1
}
```

## Получение статистики

```bash
curl http://localhost:8080/statistics
```

Ответ:
```json
{
  "total_messages": 1,
  "processed_messages": 0,
  "unprocessed_messages": 1
}
```

## Обработка сообщения

```bash
curl -X PUT http://localhost:8080/messages/1/process
```

Ответ:
```
200 OK
```

После обработки сообщения статистика изменится:

```bash
curl http://localhost:8080/statistics
```

Ответ:
```json
{
  "total_messages": 1,
  "processed_messages": 1,
  "unprocessed_messages": 0
}
```
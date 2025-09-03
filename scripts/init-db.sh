#!/bin/bash
# init-db.sh

# Подключение к PostgreSQL и создание таблицы messages

psql -U user -d messages_db -c "
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Создание индекса для улучшения производительности запросов по статусу обработки
CREATE INDEX IF NOT EXISTS idx_messages_processed ON messages(processed);

-- Создание индекса для улучшения производительности запросов по дате создания
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
"
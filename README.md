# File Forensics API

REST API на Go для анализа цифровых метаданных файлов.

## Что умеет

- Загрузка файла и вычисление хешей (MD5, SHA-1, SHA-256)
- Сохранение результатов в PostgreSQL
- Просмотр истории проверок

## Стек

- Go 1.22+
- Gin (HTTP)
- PostgreSQL 16
- pgx (драйвер + пул соединений)
- Docker / docker-compose

## Запуск

1. Поднять базу:
   ```bash
   docker compose up -d
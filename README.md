# File Forensics API

REST API на Go для анализа цифровых метаданных файлов.

## Что умеет

- Загрузка файла и вычисление хешей (MD5, SHA-1, SHA-256)
- Сохранение результатов в PostgreSQL
- Просмотр истории проверок
- Извлечение EXIF-метаданных из изображений (камера, дата, разрешение, GPS, ISO)

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

### `POST /upload`
Загружает файл, считает хеши, сохраняет в БД.
Если файл — изображение с EXIF, дополнительно извлекает метаданные.

**Пример 1: обычный файл**

```bash
curl -X POST http://localhost:8080/upload -F "file=@test.txt"
```

Ответ:
```json
{
  "id": 1,
  "filename": "test.txt",
  "size": 15,
  "md5": "...",
  "sha1": "...",
  "sha256": "..."
}
```

**Пример 2: изображение с EXIF**

```bash
curl -X POST http://localhost:8080/upload -F "file=@photo.jpg"
```

Ответ:
```json
{
  "id": 2,
  "filename": "photo.jpg",
  "size": 2456789,
  "md5": "...",
  "sha1": "...",
  "sha256": "...",
  "metadata": {
    "camera_make": "Apple",
    "camera_model": "iPhone 6",
    "datetime": "2019-02-03T09:37:52+05:00",
    "width": 3264,
    "height": 2448,
    "iso": 32
  }
}
```
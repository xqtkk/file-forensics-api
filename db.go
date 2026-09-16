package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB — глобальный пул соединений
var DB *pgxpool.Pool

// InitDB подключается к PostgreSQL и создаёт таблицу, если её нет
func InitDB() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://forensics:forensics@localhost:5432/forensics"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("не удалось создать пул: %w", err)
	}

	// Проверяем, что соединение реально работает
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	DB = pool

	// Создаём таблицу, если её нет
	_, err = DB.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS files (
			id          SERIAL PRIMARY KEY,
			filename    TEXT NOT NULL,
			size        BIGINT NOT NULL,
			md5         TEXT NOT NULL,
			sha1        TEXT NOT NULL,
			sha256      TEXT NOT NULL,
			created_at  TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("не удалось создать таблицу: %w", err)
	}

	return nil
}

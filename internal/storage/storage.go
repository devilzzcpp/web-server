package storage

import (
	"database/sql"
	"web-server/pkg/logger"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	logger.Log.Info("✅ Подключено к SQLite", "path", dbPath)
	return &Storage{db: db}, nil
}

// Migrate создаёт таблицы, если их нет
func (s *Storage) Migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Username TEXT NOT NULL,
			Role TEXT NOT NULL,
			Login TEXT NOT NULL UNIQUE,
			Password TEXT NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}

	logger.Log.Info("✅ Миграции применены")
	return nil
}

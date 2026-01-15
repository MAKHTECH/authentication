package postgres

import (
	"database/sql"
	"fmt"

	"sso/sso/internal/config"

	_ "github.com/lib/pq"
)

// Repository представляет PostgreSQL репозиторий
type Repository struct {
	db *sql.DB
}

// New создает новое подключение к PostgreSQL
func New(cfg config.PostgresConfig) (*Repository, error) {
	const op = "repository.postgres.New"

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Repository{db: db}, nil
}

// Close закрывает соединение с базой данных
func (r *Repository) Close() error {
	return r.db.Close()
}

// DB возвращает соединение с базой данных для использования в запросах
func (r *Repository) DB() *sql.DB {
	return r.db
}

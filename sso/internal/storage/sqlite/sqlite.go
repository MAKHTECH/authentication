package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"sso/sso/internal/domain/models"
	"sso/sso/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Включаем режим WAL для лучшей производительности и надежности
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		return nil, fmt.Errorf("%s: failed to enable WAL mode: %w", op, err)
	}

	// Включаем синхронизацию NORMAL для баланса производительности и надежности
	if _, err := db.Exec("PRAGMA synchronous = NORMAL;"); err != nil {
		return nil, fmt.Errorf("%s: failed to set synchronous mode: %w", op, err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	
	return &Storage{db: db}, nil
}

func (s *Storage) App(ctx context.Context, appID int32) (models.App, error) {
	const op = "storage.sqlite.App"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

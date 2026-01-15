package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/sso/internal/repository"

	"sso/sso/internal/domain/models"
)

// App возвращает приложение по ID
func (r *Repository) App(ctx context.Context, appID int32) (models.App, error) {
	const op = "repository.postgres.App"

	stmt, err := r.db.Prepare("SELECT id, name, secret FROM apps WHERE id = $1")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, repository.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

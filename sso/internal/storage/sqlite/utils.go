package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/storage"
)

func scanUser(row *sql.Row, op string) (models.User, error) {
	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

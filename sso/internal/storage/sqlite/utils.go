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
	var email, passHash, firstName, lastName, photoURL sql.NullString
	var tgID sql.NullInt64
	var authType string

	err := row.Scan(&user.ID, &email, &user.Username, &passHash, &tgID, &firstName, &lastName, &photoURL, &authType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	// Устанавливаем значения
	if email.Valid {
		user.Email = &email.String
	}
	if passHash.Valid {
		user.PassHash = &passHash.String
	}
	if tgID.Valid {
		user.TelegramID = &tgID.Int64
	}
	if firstName.Valid {
		user.FirstName = &firstName.String
	}
	if lastName.Valid {
		user.LastName = &lastName.String
	}
	if photoURL.Valid {
		user.PhotoURL = &photoURL.String
	}
	user.AuthType = models.AuthType(authType)

	return user, nil
}

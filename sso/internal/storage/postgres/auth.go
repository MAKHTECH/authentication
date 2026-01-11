package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/storage"

	"github.com/lib/pq"
)

// SaveUser сохраняет пользователя с email авторизацией
func (s *Storage) SaveUser(ctx context.Context, email, username, passHash string) (int64, error) {
	const op string = "storage.postgres.SaveUser"

	var id int64
	err := s.db.QueryRowContext(ctx,
		"INSERT INTO users(email, username, pass_hash, auth_type) VALUES($1, $2, $3, 'email') RETURNING id",
		email, username, passHash,
	).Scan(&id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return 0, storage.ErrUserExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// SaveTelegramUser сохраняет пользователя, авторизованного через Telegram
func (s *Storage) SaveTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) (int64, error) {
	const op string = "storage.postgres.SaveTelegramUser"

	var usernameVal, firstNameVal, lastNameVal, photoURLVal interface{}
	if username != "" {
		usernameVal = username
	}
	if firstName != "" {
		firstNameVal = firstName
	}
	if lastName != "" {
		lastNameVal = lastName
	}
	if photoURL != "" {
		photoURLVal = photoURL
	}

	var id int64
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users(telegram_id, username, first_name, last_name, photo_url, auth_type) 
		 VALUES($1, $2, $3, $4, $5, 'telegram') RETURNING id`,
		telegramID, usernameVal, firstNameVal, lastNameVal, photoURLVal,
	).Scan(&id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// UserByTelegramID находит пользователя по Telegram ID
func (s *Storage) UserByTelegramID(ctx context.Context, telegramID int64, appID int) (*models.User, error) {
	const op string = "storage.postgres.UserByTelegramID"

	query := `
		SELECT 
			u.id, 
			u.email, 
			u.username, 
			u.pass_hash,
			u.telegram_id,
			u.first_name,
			u.last_name,
			u.photo_url,
			u.balance,
			u.auth_type,
			COALESCE(uar.role, u.role) as role 
		FROM users u
		LEFT JOIN user_app_roles uar ON u.id = uar.user_id AND uar.app_id = $1
		WHERE u.telegram_id = $2`

	row := s.db.QueryRowContext(ctx, query, appID, telegramID)
	return s.scanUser(row, op)
}

// UpdateTelegramUser обновляет данные пользователя Telegram
func (s *Storage) UpdateTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) error {
	const op string = "storage.postgres.UpdateTelegramUser"

	var usernameVal, firstNameVal, lastNameVal, photoURLVal interface{}
	if username != "" {
		usernameVal = username
	}
	if firstName != "" {
		firstNameVal = firstName
	}
	if lastName != "" {
		lastNameVal = lastName
	}
	if photoURL != "" {
		photoURLVal = photoURL
	}

	_, err := s.db.ExecContext(ctx,
		`UPDATE users 
		 SET username = COALESCE($1, username), 
		     first_name = COALESCE($2, first_name), 
		     last_name = COALESCE($3, last_name), 
		     photo_url = COALESCE($4, photo_url)
		 WHERE telegram_id = $5`,
		usernameVal, firstNameVal, lastNameVal, photoURLVal, telegramID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// User находит пользователя по username
func (s *Storage) User(ctx context.Context, username string, appID int) (*models.User, error) {
	const op string = "storage.postgres.User"

	query := `
		SELECT 
			u.id, 
			u.email, 
			u.username, 
			u.pass_hash,
			u.telegram_id,
			u.first_name,
			u.last_name,
			u.photo_url,
			u.balance,
			u.auth_type,
			COALESCE(uar.role, u.role) as role 
		FROM users u
		LEFT JOIN user_app_roles uar ON u.id = uar.user_id AND uar.app_id = $1
		WHERE u.username = $2`

	row := s.db.QueryRowContext(ctx, query, appID, username)
	return s.scanUser(row, op)
}

// UserByID находит пользователя по ID
func (s *Storage) UserByID(ctx context.Context, id int) (*models.User, error) {
	const op string = "storage.postgres.UserByID"

	query := `
		SELECT id, email, username, pass_hash, telegram_id, first_name, last_name, photo_url, balance, auth_type, role 
		FROM users WHERE id = $1`

	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanUser(row, op)
}

// scanUser сканирует строку в модель User
func (s *Storage) scanUser(row *sql.Row, op string) (*models.User, error) {
	var user models.User
	var email, passHash, firstName, lastName, photoURL, roleStr sql.NullString
	var tgID sql.NullInt64
	var authType string

	err := row.Scan(&user.ID, &email, &user.Username, &passHash, &tgID, &firstName, &lastName, &photoURL, &user.Balance, &authType, &roleStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

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

	if roleStr.Valid {
		user.Role = models.RoleToProto(roleStr.String)
	} else {
		user.Role = ssov1.Role_USER
	}

	return &user, nil
}

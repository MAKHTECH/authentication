package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/storage"

	"github.com/mattn/go-sqlite3"
)

func (s *Storage) SaveUser(ctx context.Context, email, username, passHash string) (int64, error) {
	const op string = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users(email, username, pass_hash, auth_type) VALUES(?, ?, ?, 'email')")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, email, username, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, storage.ErrUserExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// SaveTelegramUser сохраняет пользователя, авторизованного через Telegram
func (s *Storage) SaveTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) (int64, error) {
	const op string = "storage.sqlite.SaveTelegramUser"

	stmt, err := s.db.Prepare(`
		INSERT INTO users(telegram_id, username, first_name, last_name, photo_url, auth_type) 
		VALUES(?, ?, ?, ?, ?, 'telegram')
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	// Используем sql.NullString для опциональных полей
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

	res, err := stmt.ExecContext(ctx, telegramID, usernameVal, firstNameVal, lastNameVal, photoURLVal)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// UserByTelegramID находит пользователя по Telegram ID
func (s *Storage) UserByTelegramID(ctx context.Context, telegramID int64, appID int) (*models.User, error) {
	const op string = "storage.sqlite.UserByTelegramID"

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
			uar.role 
		FROM users u
		LEFT JOIN user_app_roles uar ON u.id = uar.user_id AND uar.app_id = ?
		WHERE u.telegram_id = ?`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, appID, telegramID)

	var user models.User
	var email, passHash, firstName, lastName, photoURL, roleStr sql.NullString
	var tgID sql.NullInt64
	var authType string

	err = row.Scan(&user.ID, &email, &user.Username, &passHash, &tgID, &firstName, &lastName, &photoURL, &user.Balance, &authType, &roleStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
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

	if roleStr.Valid {
		user.Role = ssov1.Role(ssov1.Role_value[roleStr.String])
	} else {
		user.Role = ssov1.Role_USER
	}

	return &user, nil
}

// UpdateTelegramUser обновляет данные пользователя Telegram
func (s *Storage) UpdateTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) error {
	const op string = "storage.sqlite.UpdateTelegramUser"

	stmt, err := s.db.Prepare(`
		UPDATE users 
		SET username = COALESCE(?, username), 
		    first_name = COALESCE(?, first_name), 
		    last_name = COALESCE(?, last_name), 
		    photo_url = COALESCE(?, photo_url)
		WHERE telegram_id = ?
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

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

	_, err = stmt.ExecContext(ctx, usernameVal, firstNameVal, lastNameVal, photoURLVal, telegramID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) User(ctx context.Context, username string, appID int) (*models.User, error) {
	const op string = "storage.sqlite.User"

	// Подготовка запроса с JOIN для получения роли из user_app_roles
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
            uar.role 
        FROM users u
        LEFT JOIN user_app_roles uar ON u.id = uar.user_id AND uar.app_id = ?
        WHERE u.username = ?`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	// Выполняем запрос с app_id и username
	row := stmt.QueryRowContext(ctx, appID, username)

	var user models.User
	var email, passHash, firstName, lastName, photoURL, roleStr sql.NullString
	var tgID sql.NullInt64
	var authType string

	err = row.Scan(&user.ID, &email, &user.Username, &passHash, &tgID, &firstName, &lastName, &photoURL, &user.Balance, &authType, &roleStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
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

	// Если roleStr содержит валидное значение, используем его, иначе устанавливаем роль по умолчанию
	if roleStr.Valid {
		user.Role = ssov1.Role(ssov1.Role_value[roleStr.String])
	} else {
		user.Role = ssov1.Role_USER // Роль по умолчанию
	}

	return &user, nil
}

func (s *Storage) UserByID(ctx context.Context, id int) (*models.User, error) {
	const op string = "storage.sqlite.UserByID"

	stmt, err := s.db.Prepare(`
		SELECT id, email, username, pass_hash, telegram_id, first_name, last_name, photo_url, balance, auth_type 
		FROM users WHERE id = ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var user models.User
	var email, passHash, firstName, lastName, photoURL sql.NullString
	var tgID sql.NullInt64
	var authType string

	err = row.Scan(&user.ID, &email, &user.Username, &passHash, &tgID, &firstName, &lastName, &photoURL, &user.Balance, &authType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
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

	return &user, nil
}

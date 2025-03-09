package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/storage"
)

func (s *Storage) SaveUser(ctx context.Context, email, username, passHash string) (int64, error) {
	const op string = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users(email, username, pass_hash) VALUES(?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, email, username, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr, sqlite3.ErrConstraintUnique) {
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

func (s *Storage) User(ctx context.Context, username string, appID int) (*models.User, error) {
	const op string = "storage.sqlite.User"

	// Подготовка запроса с JOIN для получения роли из user_app_roles
	query := `
        SELECT 
            u.id, 
            u.email, 
            u.username, 
            u.pass_hash, 
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
	var roleStr sql.NullString // Используем sql.NullString для обработки NULL
	err = row.Scan(&user.ID, &user.Email, &user.Username, &user.PassHash, &roleStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

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

	stmt, err := s.db.Prepare("SELECT id, email, username, pass_hash FROM users WHERE id = ?")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRowContext(ctx, id)

	var user models.User

	err = row.Scan(&user.ID, &user.Email, &user.Username, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

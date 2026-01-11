package postgres

import (
	"context"
	"errors"
	"fmt"

	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/storage"

	"github.com/lib/pq"
)

// AssignRole assigns a role to a user
func (s *Storage) AssignRole(ctx context.Context, userID uint32, appID int, role ssov1.Role) error {
	const op string = "storage.postgres.user.AssignRole"

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO user_app_roles(user_id, app_id, role) VALUES($1, $2, $3)",
		userID, appID, role.String(),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return storage.ErrUserRoleExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// CheckPermission проверяет права пользователя
func (s *Storage) CheckPermission(ctx context.Context, userID int, appID int) error {
	panic("implement me")
}

// ChangePhoto обновляет URL фото профиля пользователя
func (s *Storage) ChangePhoto(ctx context.Context, userID int, photoURL string) error {
	const op string = "storage.postgres.user.ChangePhoto"

	result, err := s.db.ExecContext(ctx,
		"UPDATE users SET photo_url = $1 WHERE id = $2",
		photoURL, userID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

// ChangeUsername обновляет username пользователя
func (s *Storage) ChangeUsername(ctx context.Context, userID int, username string) error {
	const op string = "storage.postgres.user.ChangeUsername"

	result, err := s.db.ExecContext(ctx,
		"UPDATE users SET username = $1 WHERE id = $2",
		username, userID,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return storage.ErrUsernameUnique
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

// ChangeEmail обновляет email пользователя
func (s *Storage) ChangeEmail(ctx context.Context, userID int, newEmail string) error {
	const op string = "storage.postgres.user.ChangeEmail"

	result, err := s.db.ExecContext(ctx,
		"UPDATE users SET email = $1 WHERE id = $2",
		newEmail, userID,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return storage.ErrEmailUnique
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

// ChangePassword обновляет пароль пользователя
func (s *Storage) ChangePassword(ctx context.Context, userID int, newPassword string) error {
	const op string = "storage.postgres.user.ChangePassword"

	result, err := s.db.ExecContext(ctx,
		"UPDATE users SET pass_hash = $1 WHERE id = $2",
		newPassword, userID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

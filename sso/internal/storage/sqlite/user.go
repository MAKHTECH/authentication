package sqlite

import (
	"context"
	"errors"
	"fmt"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/storage"

	"github.com/mattn/go-sqlite3"
)

// AssignRole assigns a role to a user.
func (s *Storage) AssignRole(ctx context.Context, userID uint32, appID int, role ssov1.Role) error {
	const op string = "storage.sqlite.user.AssignRole"

	stmt, err := s.db.Prepare("INSERT INTO user_app_roles(user_id, app_id, role) VALUES(?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, userID, appID, role.String())
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return storage.ErrUserRoleExists
			} else if errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
				return storage.ErrUserRoleExists
			}
			fmt.Printf("SQLite error code: %d, message: %s\n", sqliteErr.Code, sqliteErr.Error())
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) CheckPermission(ctx context.Context, userID int, appID int) error {
	panic("implement me")
}

// ChangePhoto обновляет URL фото профиля пользователя
func (s *Storage) ChangePhoto(ctx context.Context, userID int, photoURL string) error {
	const op string = "storage.sqlite.user.ChangePhoto"

	stmt, err := s.db.Prepare("UPDATE users SET photo_url = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, photoURL, userID)
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

func (s *Storage) ChangeUsername(ctx context.Context, userID int, username string) error {
	const op string = "storage.sqlite.user.ChangeUsername"

	stmt, err := s.db.Prepare("UPDATE users SET username = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, username, userID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return storage.ErrUsernameUnique
		} else if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
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

func (s *Storage) ChangeEmail(ctx context.Context, userID int, newEmail string) error {
	const op string = "storage.sqlite.user.ChangeEmail"

	stmt, err := s.db.Prepare("UPDATE users SET email = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, newEmail, userID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return storage.ErrEmailUnique
		} else if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
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

func (s *Storage) ChangePassword(ctx context.Context, userID int, newPassword string) error {
	const op string = "storage.sqlite.user.ChangePassword"

	stmt, err := s.db.Prepare("UPDATE users SET pass_hash = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, newPassword, userID)
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

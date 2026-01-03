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

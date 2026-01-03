package user

import (
	"context"
	"errors"
	"log/slog"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/config"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/lib/logger/sl"
	"sso/sso/internal/storage"
	"time"
)

type User struct {
	log *slog.Logger
	db  DB
	rdb RDB
	cfg *config.Config
}

var (
	ErrUserRoleExists  = errors.New("user role already exists or (user, app) not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPhotoURL = errors.New("invalid photo URL")
)

type DB interface {
	CheckPermission(ctx context.Context, userID int, appID int) error
	AssignRole(ctx context.Context, userID uint32, appID int, role ssov1.Role) error
	ChangePhoto(ctx context.Context, userID int, photoURL string) error
}

type RDB interface {
	SaveRefreshSession(ctx context.Context, rs *models.RefreshSession, refreshTTL time.Duration) error
	GetRefreshSession(ctx context.Context, fingerprint string) (*models.RefreshSession, error)
	GetRefreshSessionsByUserId(ctx context.Context, id string) ([]*models.RefreshSession, error)
	DeleteRefreshSession(ctx context.Context, fingerprint, id string) error
}

func New(log *slog.Logger, db DB, rdb RDB, cfg *config.Config) *User {
	return &User{
		log: log,
		db:  db,
		rdb: rdb,
		cfg: cfg,
	}
}

func (u *User) AssignRole(ctx context.Context, role ssov1.Role, userID, appID int) (bool, error) {
	const op string = "services.user.AssignRole"

	log := u.log.With(
		"operation", op,
		"role", role,
		"userID", userID,
		"appID", appID,
	)

	log.Info("assigning role")

	err := u.db.AssignRole(ctx, uint32(userID), appID, role)
	if err != nil {
		if errors.Is(err, storage.ErrUserRoleExists) {
			return false, ErrUserRoleExists
		}

		log.Error("failed to assign role", sl.Err(err))
		return false, err
	}

	return true, nil
}

func (u *User) CheckPermission(ctx context.Context, userID int, appID int) (bool, error) {

	return true, nil
}

func (u *User) ChangePhoto(ctx context.Context, userID int, photoURL string) (bool, error) {
	const op string = "services.user.ChangePhoto"

	log := u.log.With(
		"operation", op,
		"userID", userID,
	)

	log.Info("changing user photo")

	// Валидация URL фото
	if photoURL == "" {
		log.Warn("empty photo URL provided")
		return false, ErrInvalidPhotoURL
	}

	err := u.db.ChangePhoto(ctx, userID, photoURL)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", "userID", userID)
			return false, ErrUserNotFound
		}

		log.Error("failed to change photo", sl.Err(err))
		return false, err
	}

	log.Info("photo changed successfully")
	return true, nil
}

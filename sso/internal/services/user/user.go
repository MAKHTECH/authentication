package user

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/config"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/lib/logger/sl"
	"sso/sso/internal/storage"
	"sso/sso/pkg/utils"
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

	ErrUsernameUnique       = errors.New("username must be unique")
	ErrEmailUnique          = errors.New("email must be unique")
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrInvalidUsername      = errors.New("invalid username")
	ErrWrongCurrentPassword = errors.New("current password is incorrect")
)

type DB interface {
	CheckPermission(ctx context.Context, userID int, appID int) error
	AssignRole(ctx context.Context, userID uint32, appID int, role ssov1.Role) error

	ChangePhoto(ctx context.Context, userID int, photoURL string) error

	ChangeUsername(ctx context.Context, userID int, username string) error
	ChangeEmail(ctx context.Context, userID int, newEmail string) error
	ChangePassword(ctx context.Context, userID int, newPassword string) error

	UserByID(ctx context.Context, id int) (*models.User, error)
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
	return false, nil
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

func (u *User) ChangeUsername(ctx context.Context, userID int, username string) (string, error) {
	const op = "services.user.ChangeUsername"
	log := u.log.With(
		"operation", op,
		"userID", userID,
	)

	log.Info("changing username")

	// Валидация username
	if username == "" || len(username) < 3 || len(username) > 50 {
		log.Warn("invalid username provided", "username", username)
		return "", ErrInvalidUsername
	}

	// меняем username
	err := u.db.ChangeUsername(ctx, userID, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", "userID", userID)
			return "", ErrUserNotFound
		} else if errors.Is(err, storage.ErrUsernameUnique) {
			log.Warn("username already taken", "username", username)
			return "", ErrUsernameUnique
		}

		log.Error("failed to change username", sl.Err(err))
		return "", err
	}

	log.Info("username changed successfully")
	return username, nil
}

func (u *User) ChangeEmail(ctx context.Context, userID int, newEmail string) (string, error) {
	const op = "services.user.ChangeEmail"
	log := u.log.With(
		"operation", op,
		"userID", userID,
	)

	log.Info("changing email")

	// Валидация email
	if newEmail == "" {
		log.Warn("empty email provided")
		return "", ErrInvalidEmail
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(newEmail) {
		log.Warn("invalid email format", "email", newEmail)
		return "", ErrInvalidEmail
	}

	err := u.db.ChangeEmail(ctx, userID, newEmail)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", "userID", userID)
			return "", ErrUserNotFound
		} else if errors.Is(err, storage.ErrEmailUnique) {
			log.Warn("email already taken", "email", newEmail)
			return "", ErrEmailUnique
		}

		log.Error("failed to change email", sl.Err(err))
		return "", err
	}

	log.Info("email changed successfully")
	return newEmail, nil
}

func (u *User) ChangePassword(ctx context.Context, userID int, newPassword, currentPassword string) (bool, error) {
	const op = "services.user.ChangePassword"
	log := u.log.With(
		"operation", op,
		"userID", userID,
	)

	log.Info("changing password")

	// Валидация нового пароля
	if newPassword == "" || len(newPassword) < 6 {
		log.Warn("invalid new password provided")
		return false, ErrInvalidPassword
	}

	// Валидация текущего пароля
	if currentPassword == "" {
		log.Warn("empty current password provided")
		return false, ErrWrongCurrentPassword
	}

	// Получаем пользователя для проверки текущего пароля
	user, err := u.db.UserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", "userID", userID)
			return false, ErrUserNotFound
		}
		log.Error("failed to get user", sl.Err(err))
		return false, err
	}

	// Проверяем, что у пользователя есть пароль (не Telegram авторизация)
	if user.PassHash == nil {
		log.Warn("user has no password (telegram auth)")
		return false, ErrWrongCurrentPassword
	}

	// Проверяем текущий пароль
	if !utils.ComparePasswordHash(currentPassword, u.cfg.PrivateKey, *user.PassHash) {
		log.Warn("current password is incorrect")
		return false, ErrWrongCurrentPassword
	}

	// Хешируем новый пароль
	passHash := utils.PasswordToHash(newPassword, u.cfg.PrivateKey)

	err = u.db.ChangePassword(ctx, userID, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", "userID", userID)
			return false, ErrUserNotFound
		}

		log.Error("failed to change password", sl.Err(err))
		return false, err
	}

	log.Info("password changed successfully")
	return true, nil
}

func (u *User) GetBalance(ctx context.Context, userID int) (float64, float64, float64, error) {
	const op = "services.user.GetBalance"
	log := u.log.With(
		"operation", op,
		"userID", userID,
	)

	log.Info("get user")

	// Получаем пользователя
	user, err := u.db.UserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", "userID", userID)
			return 0, 0, 0, ErrUserNotFound
		}
		log.Error("failed to get user", sl.Err(err))
		return 0, 0, 0, err
	}

	// Конвертируем из копеек в рубли для gRPC ответа
	return models.CopecksToRubles(user.Balance),
		models.CopecksToRubles(user.ReservedBalance),
		models.CopecksToRubles(user.AvailableBalance()), nil
}
